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

// attachVideoMeetings creates a meeting via the video module for every
// virtual schedule that doesn't already have a link.
//
// Called during published event creation and, for new schedules only,
// during update. Drafts never call this.
//
// Rules:
//   - In-person schedules are skipped.
//   - Schedules with ZoomLink or MeetLink already set are skipped (manual mode).
//   - Schedules with VideoMeetingID already set are skipped (already created).
//   - If the host is not connected, the schedule is skipped and a
//     warning is logged. Publish validation will surface the missing
//     link to the caller.
//
// Best-effort: a failure on one schedule doesn't abort the others.
// Errors are logged and returned aggregated at the end so the caller
// can decide whether to fail the request.
func (s *eventService) attachVideoMeetings(
	ctx context.Context,
	event *domain.Event,
	hostUserID string,
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
		if sched.ZoomLink != "" || sched.MeetLink != "" {
			// Manual mode — host pasted a link. Nothing to create.
			continue
		}
		if sched.VideoMeetingID != nil && *sched.VideoMeetingID != "" {
			// Already created for this schedule.
			continue
		}

		platform := virtualPlatformForSchedule(*sched)
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

		sched.ZoomLink = result.JoinURL
		mid := result.MeetingID
		sched.VideoMeetingID = &mid

		log.Printf("video: created meeting %s for session %d on %s",
			result.MeetingID, sched.SessionNumber, platform)
	}

	if len(errs) > 0 {
		return fmt.Errorf("video: %d session(s) failed: %s",
			len(errs), strings.Join(errs, "; "))
	}
	return nil
}

// virtualPlatformForSchedule infers the platform from a schedule.
// Returns "" when the schedule doesn't indicate one.
func virtualPlatformForSchedule(s domain.EventSchedule) string {
	if s.ZoomLink != "" {
		return "zoom"
	}
	if s.MeetLink != "" {
		return "google_meet"
	}
	// No link yet — this is the auto case. Default to zoom until the
	// frontend sends a platform field explicitly.
	return "zoom"
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