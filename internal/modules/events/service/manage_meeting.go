// internal/modules/events/service/manage_meeting_service.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// CREATE
// ============================================================

// CreateEventMeeting creates a meeting on the host's connected video
// platform for every virtual schedule that doesn't already have one.
//
// Idempotent: schedules with an existing video_meeting_id are skipped.
// If every virtual schedule already has a meeting, the call is a no-op
// and returns the event unchanged.
//
// platformOverride is an optional fallback for schedules whose
// platform column is empty. It is only used when the schedule itself
// does not carry a platform. An empty string means "use whatever the
// schedule already has."
func (s *eventService) CreateEventMeeting(
	ctx context.Context,
	eventID, userID, platformOverride string,
) (*domain.Event, error) {
	log.Printf("🆕 Creating meetings for event: %s by %s", eventID, userID)

	event, err := s.getEventAndCheckUpdatePermission(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if !hasVirtualSessions(event) {
		return nil, domain.ErrEventNotVirtual
	}
	if len(event.Schedules) == 0 {
		return nil, domain.ErrEventScheduleRequired
	}

	if err := s.attachVideoMeetings(ctx, event, userID, platformOverride); err != nil {
		return nil, fmt.Errorf("create meeting: %w", err)
	}

	// Re-derive event-level fields (zoom_link, virtual_platform_url,
	// etc.) from the freshly populated schedules.
	deriveEventFromSchedules(event)

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("create meeting: persist: %w", err)
	}

	if reloaded, err := s.repo.GetEventByID(ctx, eventID); err == nil && reloaded != nil {
		event = reloaded
	}

	log.Printf("✅ Meetings created for event: %s", eventID)
	return event, nil
}

// ============================================================
// DELETE
// ============================================================

// DeleteEventMeeting removes every virtual schedule's meeting from the
// platform and clears the local meeting fields.
func (s *eventService) DeleteEventMeeting(
	ctx context.Context,
	eventID, userID string,
) (*domain.Event, error) {
	log.Printf("🗑️  Deleting meetings for event: %s by %s", eventID, userID)

	event, err := s.getEventAndCheckUpdatePermission(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if !hasVirtualSessions(event) {
		return nil, domain.ErrEventNotVirtual
	}

	targets := s.schedulesWithMeetings(event)
	if len(targets) == 0 {
		return nil, domain.ErrNoMeetingToManage
	}

	for _, idx := range targets {
		sched := &event.Schedules[idx]

		// Guard: schedulesWithMeetings already filters these, but
		// be defensive so a bug elsewhere never causes a panic.
		if sched.VideoMeetingID == nil || *sched.VideoMeetingID == "" {
			continue
		}

		externalID := *sched.VideoMeetingID // capture BEFORE clearing
		platform := virtualPlatformForSchedule(*sched, "")

		err := s.video.DeleteMeeting(ctx, domain.VideoMeetingDeleteRequest{
			UserID:     userID,
			Platform:   platform,
			ExternalID: externalID,
		})
		if err != nil {
			log.Printf("⚠️ delete meeting %s: %v", externalID, err)
			// Continue — clear the local state regardless.
		}

		sched.VideoMeetingID = nil
		sched.ZoomLink = ""
		log.Printf("🗑️  Cleared meeting %s on session %d",
			externalID, sched.SessionNumber)
	}

	// Re-derive event-level fields from the now-cleared schedules so
	// the event's zoom_link / virtual_platform_url reflect reality.
	deriveEventFromSchedules(event)

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("delete meeting: persist: %w", err)
	}

	if reloaded, err := s.repo.GetEventByID(ctx, eventID); err == nil && reloaded != nil {
		event = reloaded
	}

	log.Printf("✅ Meetings deleted for event: %s", eventID)
	return event, nil
}

// ============================================================
// REGENERATE
// ============================================================

// RegenerateEventMeeting deletes every virtual schedule's meeting on
// the platform and creates a fresh one.
func (s *eventService) RegenerateEventMeeting(
	ctx context.Context,
	eventID, userID string,
) (*domain.Event, error) {
	log.Printf("🔄 Regenerating meetings for event: %s by %s", eventID, userID)

	event, err := s.getEventAndCheckUpdatePermission(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if !hasVirtualSessions(event) {
		return nil, domain.ErrEventNotVirtual
	}
	if len(event.Schedules) == 0 {
		return nil, domain.ErrEventScheduleRequired
	}

	// 1. Delete existing meetings. Best-effort.
	for _, idx := range s.schedulesWithMeetings(event) {
		sched := &event.Schedules[idx]

		if sched.VideoMeetingID == nil || *sched.VideoMeetingID == "" {
			continue
		}

		externalID := *sched.VideoMeetingID // capture BEFORE clearing
		platform := virtualPlatformForSchedule(*sched, "")

		err := s.video.DeleteMeeting(ctx, domain.VideoMeetingDeleteRequest{
			UserID:     userID,
			Platform:   platform,
			ExternalID: externalID,
		})
		if err != nil {
			log.Printf("⚠️ regenerate: delete %s: %v", externalID, err)
		}

		sched.VideoMeetingID = nil
		sched.ZoomLink = ""
	}

	// 2. Create fresh meetings. No override: the schedules already
	//    carry a platform from their previous lifecycle.
	if err := s.attachVideoMeetings(ctx, event, userID, ""); err != nil {
		return nil, fmt.Errorf("regenerate meeting: %w", err)
	}

	// 3. Re-derive event-level fields from the freshly populated
	//    schedules.
	deriveEventFromSchedules(event)

	// 4. Persist.
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("regenerate meeting: persist: %w", err)
	}

	if reloaded, err := s.repo.GetEventByID(ctx, eventID); err == nil && reloaded != nil {
		event = reloaded
	}

	log.Printf("✅ Meetings regenerated for event: %s", eventID)
	return event, nil
}

// ============================================================
// HELPERS
// ============================================================

// schedulesWithMeetings returns the indexes of virtual schedules
// that currently have a video_meeting_id set.
func (s *eventService) schedulesWithMeetings(event *domain.Event) []int {
	var out []int
	for i := range event.Schedules {
		sched := &event.Schedules[i]
		if !sched.IsVirtual {
			continue
		}
		if sched.VideoMeetingID == nil || *sched.VideoMeetingID == "" {
			continue
		}
		out = append(out, i)
	}
	return out
}

// hasVirtualSessions reports whether the event has at least one virtual
// schedule. This is the source of truth for "can we manage meetings on
// this event" — the event-level IsVirtual/IsHybrid flags are derived
// and can lag behind schedule edits.
func hasVirtualSessions(event *domain.Event) bool {
	if event == nil {
		return false
	}
	for _, s := range event.Schedules {
		if s.IsVirtual {
			return true
		}
	}
	return false
}