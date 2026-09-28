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
// After a meeting is created, the chosen platform is written back to
// the schedule so subsequent updates do not need to infer it.
//
// Rules:
//   - In-person schedules are skipped.
//   - Schedules with a manual link and no VideoMeetingID are skipped
//     (true manual mode).
//   - Schedules with a VideoMeetingID are updated on the platform
//     (topic, start time, duration, timezone).
//   - Schedules with no meeting and no manual link are created.
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
		// Video integration disabled (adapter not wired).
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

		// A schedule with a manual link but no Nuruvent-managed meeting
		// is a true manual link — the host pasted a URL and we should
		// leave it alone. A schedule with a VideoMeetingID is an
		// auto-created meeting that should be updated.
		hasManualLink := sched.ZoomLink != "" || sched.MeetLink != ""
		hasMeeting := sched.VideoMeetingID != nil && *sched.VideoMeetingID != ""
		if hasManualLink && !hasMeeting {
			continue
		}

		platform := virtualPlatformForSchedule(*sched, platformOverride)
		if platform == "" {
			// Virtual but no provider indicated. Skip and let
			// publish validation reject it.
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
		// Case 1: schedule already has a meeting — update it.
		// ─────────────────────────────────────────────────────────
		if hasMeeting {
			updated, err := s.video.UpdateMeeting(ctx, domain.UpdateVideoMeetingRequest{
				UserID:     hostUserID,
				Platform:   platform,
				ExternalID: *sched.VideoMeetingID,
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

			// The join URL is unchanged on update, but reassigning is
			// harmless and covers the case where the provider returns
			// it anyway. Route it to the field that matches the
			// platform so URL-shape inference remains reliable.
			if updated.JoinURL != "" {
				assignJoinLink(sched, platform, updated.JoinURL)
			}

			log.Printf("video: updated meeting %s for session %d on %s",
				updated.ExternalID, sched.SessionNumber, platform)
			continue
		}

		// ─────────────────────────────────────────────────────────
		// Case 2: schedule has no meeting — create one.
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

		// Persist the platform so subsequent updates do not need to
		// infer it from the link.
		sched.Platform = platform

		// Route the URL to the correct field. A Google Meet join URL
		// must not land in ZoomLink; that would break subsequent
		// inference and confuse display code.
		assignJoinLink(sched, platform, result.JoinURL)

		mid := result.ExternalID // platform-side ID, not the Nuruvent UUID
		sched.VideoMeetingID = &mid

		log.Printf("video: created meeting %s for session %d on %s",
			result.ExternalID, sched.SessionNumber, platform)
	}

	if len(errs) > 0 {
		return fmt.Errorf("video: %d session(s) failed: %s",
			len(errs), strings.Join(errs, "; "))
	}
	return nil
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