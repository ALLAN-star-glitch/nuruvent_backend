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
		return // sync not wired (e.g. during early boot or tests)
	}
	if event == nil || len(event.Schedules) == 0 {
		return
	}

	for _, schedule := range event.Schedules {
		cmd, err := buildAttendanceCommand(event, schedule)
		if err != nil {
			log.Printf("[events] skipping attendance sync for schedule %s: %v", schedule.ID, err)
			continue
		}
		if err := s.attendance.UpsertSession(ctx, cmd); err != nil {
			log.Printf("[events] attendance sync failed for event=%s schedule=%s: %v",
				event.ID, schedule.ID, err)
			// Continue — don't fail the caller.
		}
	}
}

// buildAttendanceCommand maps one EventSchedule + its parent Event
// into an AttendanceUpsertSessionCommand.
func buildAttendanceCommand(
	event *domain.Event,
	schedule domain.EventSchedule,
) (domain.AttendanceUpsertSessionCommand, error) {
	start, end, err := resolveScheduleTimes(schedule)
	if err != nil {
		return domain.AttendanceUpsertSessionCommand{}, err
	}

	provider, meetingID, providerURL := detectProvider(schedule)

	return domain.AttendanceUpsertSessionCommand{
		ExternalType:      "event",
		ExternalID:        event.ID,
		ProviderSessionID: schedule.ID,
		Title:             composeSessionTitle(event, schedule),
		ScheduledStart:    start,
		ScheduledEnd:      end,
		Provider:          provider,
		ProviderMeetingID: meetingID,
		ProviderURL:       providerURL,
	}, nil
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

// loadLocationOrDefault loads a timezone or falls back to the
// platform default (Africa/Nairobi).
func loadLocationOrDefault(name string) (*time.Location, error) {
	if strings.TrimSpace(name) == "" {
		return time.LoadLocation("Africa/Nairobi")
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.LoadLocation("Africa/Nairobi")
	}
	return loc, nil
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

// detectProvider classifies the schedule into a provider, extracts
// its meeting ID, and picks the join URL.
func detectProvider(schedule domain.EventSchedule) (provider, meetingID, providerURL string) {
	if !schedule.IsVirtual {
		return "in_person", "", schedule.Location
	}

	zoom := strings.TrimSpace(schedule.ZoomLink)
	if zoom != "" {
		return "zoom", parseZoomMeetingID(zoom), zoom
	}

	meet := strings.TrimSpace(schedule.MeetLink)
	if meet != "" {
		return "google_meet", parseMeetCode(meet), meet
	}

	return "none", "", ""
}

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

func parseMeetCode(link string) string {
	m := meetCodeRe.FindStringSubmatch(link)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}