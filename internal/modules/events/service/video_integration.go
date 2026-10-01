// internal/modules/events/service/video_integration.go

package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// attachVideoMeetings creates, updates, or skips video meetings to
// match the current state of the event's schedules.
//
// Called during published event creation, during update whenever
// schedules are present in the command, and from the create-meeting
// endpoint.
//
// Platform selection:
//   - If the schedule has an explicit Platform, that is used.
//   - Otherwise, platformOverride is used if it is non-empty. This
//     covers schedules created before the picker existed — the caller
//     supplies a platform at click time.
//   - Otherwise, the URL shape of an existing link is inspected
//     (legacy rows with a link but no platform).
//   - If none of the above yields a platform, the schedule is skipped
//     and logged; publish validation surfaces the missing provider.
//
// After a meeting is created, the chosen platform and both meeting
// identifiers are written back to the schedule:
//
//   - VideoMeetingID          — the Nuruvent UUID (video_meetings.id)
//   - VideoMeetingExternalID  — the platform code (spaces/..., Zoom ID)
//
// The two are distinct and are stored in distinct fields. Attendance
// sync reads VideoMeetingExternalID so it never has to parse a URL.
//
// A schedule is considered to "have a meeting" only when BOTH IDs are
// present. A schedule with a UUID but no external ID is a half-broken
// row from a previous write; it is recreated rather than updated, so
// the platform code gets populated.
//
// Rules:
//   - In-person schedules are skipped.
//   - Schedules with a manual link and no valid meeting are skipped
//     (true manual mode).
//   - Schedules with a valid meeting are updated on the platform
//     (topic, start time, duration, timezone).
//   - Schedules with no meeting (or an incomplete one) are created.
//   - If the host is not connected, the schedule is skipped and a
//     warning is logged. Publish validation surfaces the missing
//     link to the caller.
//
// Best-effort: a failure on one schedule doesn't abort the others.
// Errors are logged and returned aggregated at the end so the caller
// can decide whether to fail the request.
func (s *eventService) attachVideoMeetings(
	ctx context.Context,
	event *domain.Event,
	hostUserID string,
	platformOverride string,
) error {
	if s.video == nil {
		return nil
	}

	if len(event.Schedules) == 0 {
		return nil
	}

	var errs []string
	for i := range event.Schedules {
		sched := &event.Schedules[i]

		if !sched.IsVirtual {
			continue
		}

		// The platform the schedule *should* have. Resolved before we
		// decide whether the existing meeting is valid, so we can
		// check the two for agreement.
		platform := virtualPlatformForSchedule(*sched, platformOverride)

		// A manual link with no valid meeting is true manual mode.
		// Checked before hasMeeting so a legacy row with a pasted
		// link and no meeting row doesn't get auto-created over.
		hasAnyID := sched.VideoMeetingID != nil &&
			*sched.VideoMeetingID != "" &&
			sched.VideoMeetingExternalID != ""
		hasManualLink := sched.ZoomLink != "" || sched.MeetLink != ""
		if hasManualLink && !hasAnyID {
			continue
		}

		// Determine whether the schedule has a *usable* meeting:
		// both IDs present AND the external ID shape matches the
		// platform we'd create on.
		hasMeeting := hasAnyID && externalIDMatchesPlatform(
			sched.VideoMeetingExternalID, platform,
		)

		// Stale meeting: a fully-linked row whose external ID doesn't
		// match the schedule's current platform. e.g. the platform
		// was switched in the editor but the meeting row still points
		// at the previous provider.
		hasStaleMeeting := hasAnyID && !hasMeeting

		if platform == "" {
			log.Printf("video: schedule %d is virtual but has no provider",
				sched.SessionNumber)
			continue
		}

		connected, err := s.video.IsConnected(ctx, hostUserID, platform)
		if err != nil {
			errs = append(errs, fmt.Sprintf("session %d: check connection: %v",
				sched.SessionNumber, err))
			continue
		}
		if !connected {
			log.Printf("video: host %s not connected to %s; skipping session %d",
				hostUserID, platform, sched.SessionNumber)
			continue
		}

		start := combineScheduleStart(*sched)
		duration := scheduleDuration(*sched)

		// ─────────────────────────────────────────────────────────
		// Stale meeting: delete the old one before creating a new one
		// so we don't orphan a row on the previous platform.
		//
		// Delete is best-effort. If it fails we still clear the local
		// fields so the create branch below runs.
		// ─────────────────────────────────────────────────────────
		if hasStaleMeeting {
			// We don't know which platform the OLD external ID
			// belongs to. Try both; the module will reject the one
			// that doesn't match.
			for _, p := range []string{
				domain.VideoPlatformZoom,
				domain.VideoPlatformGoogleMeet,
			} {
				if err := s.video.DeleteMeeting(ctx, domain.VideoMeetingDeleteRequest{
					UserID:     hostUserID,
					Platform:   p,
					ExternalID: sched.VideoMeetingExternalID,
				}); err != nil {
					// eslint-style: don't fail the whole pass over a
					// stale cleanup miss.
					log.Printf("video: cleanup stale meeting (platform=%s) session %d: %v",
						p, sched.SessionNumber, err)
					continue
				}
				// First successful delete wins.
				break
			}

			// Clear the local pointers so the create branch below
			// runs cleanly. The links are cleared too — they
			// pointed at the wrong provider.
			sched.VideoMeetingID = nil
			sched.VideoMeetingExternalID = ""
			sched.ZoomLink = ""
			sched.MeetLink = ""

			log.Printf("video: cleared stale meeting for session %d (was platform-mismatched)",
				sched.SessionNumber)
		}

		// ─────────────────────────────────────────────────────────
		// Case 1: schedule has a valid, platform-consistent meeting
		// — update it.
		// ─────────────────────────────────────────────────────────
		if hasMeeting {
			sched.Platform = platform

			updated, err := s.video.UpdateMeeting(ctx, domain.UpdateVideoMeetingRequest{
				UserID:     hostUserID,
				Platform:   platform,
				ExternalID: sched.VideoMeetingExternalID,
				Topic:      sched.SessionName,
				StartTime:  start,
				Duration:   duration,
				Timezone:   sched.Timezone,
				Agenda:     event.Description,
			})
			if err != nil {
				errs = append(errs, fmt.Sprintf("session %d: update meeting: %v",
					sched.SessionNumber, err))
				continue
			}

			if updated.JoinURL != "" {
				assignJoinLink(sched, platform, updated.JoinURL)
			}
			if updated.ExternalID != "" {
				sched.VideoMeetingExternalID = updated.ExternalID
			}

			log.Printf("video: updated meeting %s for session %d on %s",
				sched.VideoMeetingExternalID, sched.SessionNumber, platform)
			continue
		}

		// ─────────────────────────────────────────────────────────
		// Case 2: create (no meeting, half-broken row, or stale
		// meeting we just cleared).
		// ─────────────────────────────────────────────────────────
		result, err := s.video.CreateMeeting(ctx, domain.VideoMeetingRequest{
			UserID:    hostUserID,
			Platform:  platform,
			Topic:     sched.SessionName,
			StartTime: start,
			Duration:  duration,
			Timezone:  sched.Timezone,
			Agenda:    event.Description,
		})
		if err != nil {
			errs = append(errs, fmt.Sprintf("session %d: create meeting: %v",
				sched.SessionNumber, err))
			continue
		}

		sched.Platform = platform
		assignJoinLink(sched, platform, result.JoinURL)

		mid := result.MeetingID
		sched.VideoMeetingID = &mid
		sched.VideoMeetingExternalID = result.ExternalID

		log.Printf("video: created meeting %s (external=%s) for session %d on %s",
			result.MeetingID, result.ExternalID, sched.SessionNumber, platform)
	}

	if len(errs) > 0 {
		return fmt.Errorf("video: %d session(s) failed: %s",
			len(errs), strings.Join(errs, "; "))
	}
	return nil
}

// externalIDForSchedule returns the platform-side meeting ID for a
// schedule that already has an auto-created meeting.
//
// Only VideoMeetingExternalID is consulted. VideoMeetingID holds the
// Nuruvent UUID, not a platform code, so falling back to it would
// produce a value the video module cannot resolve.
//
// Deprecated: prefer reading sched.VideoMeetingExternalID directly.
// Kept only for callers outside this file that still reference it.
func externalIDForSchedule(sched domain.EventSchedule) string {
	return sched.VideoMeetingExternalID
}

// virtualPlatformForSchedule determines which video platform a
// schedule should use.
//
// Priority:
//  1. An explicit Platform field on the schedule.
//  2. platformOverride, when non-empty. This lets the caller supply a
//     platform for schedules that were created before the picker
//     existed.
//  3. The URL shape of an existing link (legacy rows with a link but
//     no platform column set).
//  4. Empty string — the caller skips and logs.
//
// Returning "" rather than defaulting to Zoom is deliberate.
// Auto-creation requires an explicit choice; either the schedule
// carries a platform, the caller supplies one, or a link exists whose
// shape identifies the provider.
func virtualPlatformForSchedule(s domain.EventSchedule, override string) string {
	if s.Platform != "" {
		return s.Platform
	}

	if override != "" {
		return override
	}

	link := s.ZoomLink
	if link == "" {
		link = s.MeetLink
	}

	switch {
	case strings.Contains(link, "zoom.us"):
		return domain.VideoPlatformZoom
	case strings.Contains(link, "meet.google.com"):
		return domain.VideoPlatformGoogleMeet
	}

	return ""
}

// assignJoinLink writes a platform's join URL to the schedule field
// that matches the platform.
//
// Keeping Zoom and Meet links in their respective fields means
// URL-shape inference remains reliable for legacy rows and display
// code can show the correct link without inspecting prefixes.
func assignJoinLink(sched *domain.EventSchedule, platform, url string) {
	switch platform {
	case domain.VideoPlatformZoom:
		sched.ZoomLink = url
	case domain.VideoPlatformGoogleMeet:
		sched.MeetLink = url
	default:
		// Unknown platform: fall back to ZoomLink for backward
		// compatibility with anything that reads it.
		sched.ZoomLink = url
	}
}

// combineScheduleStart merges StartDate + StartTime in the schedule's
// timezone. Returns zero if either is missing.
func combineScheduleStart(s domain.EventSchedule) time.Time {
	if s.StartDate.IsZero() || s.StartTime == "" {
		return time.Time{}
	}

	loc := time.UTC
	if s.Timezone != "" {
		if loaded, err := time.LoadLocation(s.Timezone); err == nil {
			loc = loaded
		}
	}

	parsed, err := time.Parse("15:04:05", s.StartTime)
	if err != nil {
		parsed, err = time.Parse("15:04", s.StartTime)
	}
	if err != nil {
		return time.Time{}
	}

	return time.Date(
		s.StartDate.Year(), s.StartDate.Month(), s.StartDate.Day(),
		parsed.Hour(), parsed.Minute(), parsed.Second(), 0, loc,
	)
}

// scheduleDuration returns the wall-clock duration of a single schedule.
func scheduleDuration(s domain.EventSchedule) time.Duration {
	start := combineScheduleStart(s)
	if start.IsZero() || s.EndTime == "" {
		return 0
	}

	endDate := s.StartDate
	if s.EndDate != nil && !s.EndDate.IsZero() {
		endDate = *s.EndDate
	}

	end := combineScheduleStart(domain.EventSchedule{
		StartDate: endDate,
		StartTime: s.EndTime,
		Timezone:  s.Timezone,
	})
	if end.IsZero() || end.Before(start) {
		return 0
	}
	return end.Sub(start)
}

// externalIDMatchesPlatform reports whether an external meeting ID
// looks like it belongs to the given platform.
//
// Google Meet external IDs always start with "spaces/". Zoom meeting
// IDs are numeric strings. Anything else (empty, unknown shape) is
// treated as a mismatch so the caller falls into the recreate branch
// rather than trusting the row.
func externalIDMatchesPlatform(externalID, platform string) bool {
	if externalID == "" {
		return false
	}
	switch platform {
	case domain.VideoPlatformZoom:
		return isAllDigits(externalID)
	case domain.VideoPlatformGoogleMeet:
		return strings.HasPrefix(externalID, "spaces/")
	default:
		return false
	}
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}