// internal/modules/events/service/attendance_sync.go

package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// syncEventSchedulesToAttendance mirrors every schedule on the event
// into the attendance module as an attendance session.
//
// Best-effort: if one session fails to sync, we log and continue.
// Publishing the event is more important than getting attendance
// perfect; a reconciliation job can catch missed sessions later.
//
// Idempotent: safe to call multiple times. Attendance's UpsertSession
// keys on (external ref, provider session id) and updates in place.
func (s *eventService) syncEventSchedulesToAttendance(
	ctx context.Context,
	event *domain.Event,
) {
	if s.attendance == nil {
		if event != nil {
			log.Printf("[events] sync SKIP event=%s reason=attendance-nil", event.ID)
		} else {
			log.Printf("[events] sync SKIP reason=attendance-nil event=nil")
		}
		return
	}
	if event == nil {
		log.Printf("[events] sync SKIP reason=event-nil")
		return
	}
	if len(event.Schedules) == 0 {
		log.Printf("[events] sync SKIP event=%s reason=no-schedules", event.ID)
		return
	}

	log.Printf("[events] sync START event=%s schedules=%d",
		event.ID, len(event.Schedules))

	// Resolve the organizer display name once per event — it doesn't
	// vary by schedule.
	//
	// The event loaded by most code paths has Organizer == nil,
	// because the events repository doesn't eagerly load it. Fall
	// back to the injected organizer provider, which knows how to
	// resolve it from the team.
	organizerName := s.resolveOrganizerForSync(ctx, event)

	var succeeded, failed int
	for _, schedule := range event.Schedules {
		cmd, err := buildAttendanceCommand(event, schedule, organizerName)
		if err != nil {
			log.Printf("[events] sync SKIP event=%s schedule=%s reason=build-command err=%v",
				event.ID, schedule.ID, err)
			failed++
			continue
		}
		if err := s.attendance.UpsertSession(ctx, cmd); err != nil {
			log.Printf("[events] sync FAIL event=%s schedule=%s provider=%s provider_session_id=%s err=%v",
				event.ID, schedule.ID, cmd.Provider, cmd.ProviderSessionID, err)
			failed++
			continue
		}
		log.Printf("[events] sync OK event=%s schedule=%s provider=%s provider_session_id=%s",
			event.ID, schedule.ID, cmd.Provider, cmd.ProviderSessionID)
		succeeded++
	}

	log.Printf("[events] sync DONE event=%s schedules=%d ok=%d failed=%d",
	
		event.ID, len(event.Schedules), succeeded, failed)


		// Register the host as an attendee for the event. Runs after the
	// schedules are mirrored so the host's session-status rows find
	// sessions to attach to.
	s.registerHostAttendeeForSync(ctx, event)
}





// registerHostAttendeeForSync creates or updates the host's attendee
// row in the attendance module.
//
// Best-effort, same as the schedule sync: a failure here does not
// fail the caller. The host's attendance will be missing until a
// later sync succeeds.
func (s *eventService) registerHostAttendeeForSync(
	ctx context.Context,
	event *domain.Event,
) {
	if s.attendance == nil {
		return
	}
	if event == nil || event.CreatedBy == "" {
		return
	}

	cmd := domain.AttendanceRegisterHostCommand{
		EventID:    event.ID,
		HostUserID: event.CreatedBy,
	}

	// Resolve the host's display name and email. Same helper the
	// organizer lookup uses.
	if event.Creator != nil {
		cmd.HostDisplayName = strings.TrimSpace(event.Creator.DisplayName)
		if cmd.HostDisplayName == "" {
			cmd.HostDisplayName = strings.TrimSpace(event.Creator.Name)
		}
		cmd.HostEmail = strings.TrimSpace(event.Creator.Email)
		cmd.HostPhone = strings.TrimSpace(event.Creator.Phone) 
	    cmd.HostUsername = strings.TrimSpace(event.Creator.Username)
	}
	if cmd.HostDisplayName == "" {
		cmd.HostDisplayName = s.resolveOrganizerForSync(ctx, event)
	}
	if cmd.HostDisplayName == "" {
		cmd.HostDisplayName = "Host"
	}

	// Resolve the host's Google Meet user id from their active
	// connection so the first Meet fetch matches without a manual
	// roster link. Zoom is handled by username (HostUsername), so no
	// platform id is needed there.
	if s.videoIdentity != nil {
		if id, err := s.videoIdentity.ExternalUserIDForPlatform(
			ctx, event.CreatedBy, "google_meet",
		); err != nil {
			log.Printf(
				"[events] register host: resolve google_meet id user=%s err=%v",
				event.CreatedBy, err,
			)
		} else if id != "" {
			cmd.HostGoogleMeetUserID = id
		}
	}

	if err := s.attendance.RegisterHostAttendee(ctx, cmd); err != nil {
		log.Printf(
			"[events] register host attendee FAIL event=%s host=%s err=%v",
			event.ID, event.CreatedBy, err,
		)
		return
	}

	log.Printf(
		"[events] register host attendee OK event=%s host=%s google_meet_id=%q",
		event.ID, event.CreatedBy, cmd.HostGoogleMeetUserID,
	)
}

// resolveOrganizerForSync returns the organizer's display name for an
// event, preferring what's already on the loaded struct and falling
// back to the organizer provider.
//
// Never returns an error — a missing organizer is not fatal to the
// sync; the join redirect just carries an empty host param, which
// the frontend handles gracefully.
func (s *eventService) resolveOrganizerForSync(
	ctx context.Context,
	event *domain.Event,
) string {
	// Fast path: already loaded.
	if name := resolveOrganizerDisplayName(event); name != "" {
		return name
	}

	// Slow path: resolve through the provider. This is the same
	// helper the API handler uses to build the organizer block on
	// GET responses.
	info, err := s.getOrganizerInfo(ctx, event)
	if err != nil {
		log.Printf("[events] sync: organizer lookup failed event=%s: %v",
			event.ID, err)
		return ""
	}
	if info == nil {
		return ""
	}
	if strings.TrimSpace(info.DisplayName) != "" {
		return strings.TrimSpace(info.DisplayName)
	}
	return strings.TrimSpace(info.Name)
}

// buildAttendanceCommand maps one EventSchedule + its parent Event
// into an AttendanceUpsertSessionCommand.
//
// organizerName is resolved once by the caller and passed in — it's
// the same for every schedule under the event.
func buildAttendanceCommand(
	event *domain.Event,
	schedule domain.EventSchedule,
	organizerName string,
) (domain.AttendanceUpsertSessionCommand, error) {
	start, end, err := resolveScheduleTimes(schedule)
	if err != nil {
		return domain.AttendanceUpsertSessionCommand{}, err
	}

	provider, meetingID, providerURL := detectProvider(schedule)

	return domain.AttendanceUpsertSessionCommand{
		ExternalType:         "event",
		ExternalID:           event.ID,
		ProviderSessionID:    schedule.ID,
		Title:                composeSessionTitle(event, schedule),
		ScheduledStart:       start,
		ScheduledEnd:         end,
		Provider:             provider,
		ProviderMeetingID:    meetingID,
		ProviderURL:          providerURL,
		EventDisplayName:     event.DisplayName,
		OrganizerDisplayName: organizerName,
	}, nil
}

// resolveOrganizerDisplayName extracts the organizer's display name
// from the event. Falls back through DisplayName → Name → "" so a
// partial load never blocks the sync.
//
// The event passed to the sync may or may not have the organizer
// loaded depending on which code path invoked us. When it's nil, we
// return "" and the redirect URL just carries an empty host param —
// the frontend already handles that gracefully.
func resolveOrganizerDisplayName(event *domain.Event) string {
	if event == nil || event.Organizer == nil {
		return ""
	}
	if strings.TrimSpace(event.Organizer.DisplayName) != "" {
		return event.Organizer.DisplayName
	}
	return strings.TrimSpace(event.Organizer.Name)
}

// resolveScheduleTimes combines the schedule's date(s) and time(s)
// with its timezone into absolute times.
func resolveScheduleTimes(schedule domain.EventSchedule) (time.Time, time.Time, error) {
	loc, err := loadLocationOrDefault(schedule.Timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid timezone %q: %w", schedule.Timezone, err)
	}

	startTime, err := parseTimeOfDay(schedule.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_time %q: %w", schedule.StartTime, err)
	}
	endTime, err := parseTimeOfDay(schedule.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_time %q: %w", schedule.EndTime, err)
	}

	startDate := schedule.StartDate
	endDate := startDate
	if schedule.EndDate != nil {
		endDate = *schedule.EndDate
	}

	start := time.Date(
		startDate.Year(), startDate.Month(), startDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, loc,
	)
	end := time.Date(
		endDate.Year(), endDate.Month(), endDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, loc,
	)
	return start, end, nil
}

// loadLocationOrDefault loads a timezone or falls back to the platform
// default (Africa/Nairobi), and then to UTC if neither is available.
//
// The UTC last resort ensures a missing or invalid tz database can
// never block attendance sync — a schedule will still be created, just
// with UTC times, which the host can correct.
func loadLocationOrDefault(name string) (*time.Location, error) {
	if strings.TrimSpace(name) != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc, nil
		}
	}
	if loc, err := time.LoadLocation("Africa/Nairobi"); err == nil {
		return loc, nil
	}
	return time.UTC, nil
}

// parseTimeOfDay parses "HH:MM" or "HH:MM:SS" into a time.Time whose
// date portion is zero — only the clock part is meaningful.
func parseTimeOfDay(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format")
}

// The events module already composes "Session N: Name" on the way in.
// If SessionName is set, use it as-is — don't prepend the session
// number again.
func composeSessionTitle(event *domain.Event, schedule domain.EventSchedule) string {
	name := strings.TrimSpace(schedule.SessionName)
	if name != "" {
		return name
	}
	if schedule.SessionNumber > 0 {
		return fmt.Sprintf("Session %d", schedule.SessionNumber)
	}
	if event.DisplayName != "" {
		return event.DisplayName
	}
	return event.Name
}

// ============================================================
// PROVIDER DETECTION
// ============================================================

// detectProvider classifies the schedule into a provider, extracts
// its meeting ID, and picks the join URL.
//
// Priority:
//
//  1. The platform external ID we stored at creation time
//     (VideoMeetingExternalID). This value came from the same
//     video.CreateMeeting call that persisted video_meetings.external_id,
//     so it matches by construction. No string parsing needed.
//
//  2. Legacy fallback for rows created before the external-ID column
//     existed: parse the join URL. Rows in this state are also the
//     ones most likely to produce mismatches, so prefer #1 wherever
//     possible.
//
// Returns ("none", "", "") when no provider can be determined.
func detectProvider(schedule domain.EventSchedule) (provider, meetingID, providerURL string) {
	if !schedule.IsVirtual {
		return "in_person", "", schedule.Location
	}

	// ── Preferred: the external ID stored at meeting-creation time ──
	if schedule.VideoMeetingExternalID != "" {
		platform := schedule.Platform
		if platform == "" {
			// Platform is missing on some legacy rows. Infer it from
			// the external ID shape so downstream code still gets a
			// sane provider string.
			if strings.HasPrefix(schedule.VideoMeetingExternalID, "spaces/") {
				platform = domain.VideoPlatformGoogleMeet
			} else {
				platform = domain.VideoPlatformZoom
			}
		}

		url := schedule.MeetLink
		if url == "" {
			url = schedule.ZoomLink
		}

		return platform, schedule.VideoMeetingExternalID, url
	}

	// ── Legacy fallback: parse the URL ──
	//
	// Rows here predate video_meeting_external_id. If they were
	// created through the video module, video_meeting_id holds a
	// platform code (from the bug this change fixes); otherwise the
	// link field is the only signal we have.
	if schedule.VideoMeetingID != nil && *schedule.VideoMeetingID != "" {
		// The legacy field may hold either:
		//   - a platform code (e.g. "spaces/xxx", or a Zoom numeric
		//     ID), for rows created before the column split, or
		//   - a Nuruvent UUID, for rows created during the brief
		//     window after the split when the semantics were wrong.
		//
		// If it looks like a platform code, use it as-is.
		legacy := *schedule.VideoMeetingID
		if isPlatformMeetingCode(legacy) {
			platform := schedule.Platform
			if platform == "" {
				if strings.HasPrefix(legacy, "spaces/") {
					platform = domain.VideoPlatformGoogleMeet
				} else {
					platform = domain.VideoPlatformZoom
				}
			}
			url := schedule.MeetLink
			if url == "" {
				url = schedule.ZoomLink
			}
			return platform, legacy, url
		}
		// Otherwise it's a UUID we can't use — fall through to URL
		// parsing below.
	}

	zoom := strings.TrimSpace(schedule.ZoomLink)
	if zoom != "" {
		return domain.VideoPlatformZoom, parseZoomMeetingID(zoom), zoom
	}

	meet := strings.TrimSpace(schedule.MeetLink)
	if meet != "" {
		return domain.VideoPlatformGoogleMeet, canonicalMeetResourceName(meet), meet
	}

	return "none", "", ""
}

// isPlatformMeetingCode reports whether a legacy video_meeting_id
// value looks like a platform meeting code rather than a Nuruvent
// UUID.
//
// Google Meet codes are "spaces/<id>"; Zoom codes are 9–11 digit
// numeric strings. Nuruvent UUIDs are 36-character dashed strings.
func isPlatformMeetingCode(s string) bool {
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "spaces/") {
		return true
	}
	// Zoom: all digits, 9–11 characters.
	if len(s) >= 9 && len(s) <= 11 {
		allDigits := true
		for _, r := range s {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return true
		}
	}
	return false
}

// ============================================================
// MEETING CODE PARSERS
// ============================================================

// zoomMeetingIDRe extracts the numeric meeting ID from a Zoom URL.
var zoomMeetingIDRe = regexp.MustCompile(`/j/(\d{9,11})`)

func parseZoomMeetingID(link string) string {
	m := zoomMeetingIDRe.FindStringSubmatch(link)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// meetCodeRe extracts the meeting code from a Google Meet URL.
var meetCodeRe = regexp.MustCompile(`meet\.google\.com/([a-z]{3}-[a-z]{4}-[a-z]{3})`)

// meetBareCodeRe matches a bare Google Meet code like "tbz-qpwt-vni".
var meetBareCodeRe = regexp.MustCompile(`^[a-z]{3}-[a-z]{4}-[a-z]{3}$`)

// canonicalMeetResourceName converts a Google Meet URL or bare code
// into the "spaces/<code>" resource name that video_meetings.external_id
// stores. Google's API always returns the resource-name form, and the
// attendance module matches sessions to meetings by exact string
// equality on this value.
//
// Returns "" if no code can be extracted.
//
// Kept for the legacy fallback path in detectProvider. Once every
// schedule has VideoMeetingExternalID set, this function is only
// reachable from migration-era rows.
func canonicalMeetResourceName(linkOrCode string) string {
	linkOrCode = strings.TrimSpace(linkOrCode)
	if linkOrCode == "" {
		return ""
	}

	// Already in canonical form.
	if strings.HasPrefix(linkOrCode, "spaces/") {
		return linkOrCode
	}

	// From a full URL.
	if m := meetCodeRe.FindStringSubmatch(linkOrCode); len(m) >= 2 {
		return "spaces/" + m[1]
	}

	// Bare code that matches the Meet format.
	if meetBareCodeRe.MatchString(linkOrCode) {
		return "spaces/" + linkOrCode
	}

	return ""
}

// parseMeetCode is retained for backwards compatibility with anything
// that still calls it. New code should use canonicalMeetResourceName
// (via detectProvider) instead, so the value matches
// video_meetings.external_id.
func parseMeetCode(link string) string {
	m := meetCodeRe.FindStringSubmatch(link)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

