// internal/modules/events/service/status.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// STATUS - Single
// ============================================================

// PublishEvent publishes a single event.
//
// Also mirrors the event's schedules into the attendance module, so
// attendance has sessions ready the moment the event goes live. The
// reload after the write ensures schedule IDs are populated on the
// domain struct before the sync runs — otherwise the sync would emit
// sessions with empty provider_session_id.
//
// Video meetings are provisioned ASYNCHRONOUSLY after this method
// returns. The worker is enqueued below and does the actual Zoom/Meet
// API calls in the background.
func (s *eventService) PublishEvent(ctx context.Context, id, publishedBy string) (*domain.Event, error) {
	log.Printf("📤 Publishing event: %s", id)

	event, err := s.getEventAndCheckPublishPermission(ctx, id, publishedBy)
	if err != nil {
		return nil, err
	}

	status, err := s.getEventStatusBySlug(ctx, domain.EventStatusPublished.GetSlug())
	if err != nil {
		return nil, err
	}

	if err := event.ValidateForPublish(); err != nil {
		log.Printf("❌ Publish validation failed: %v", err)
		return nil, s.translatePublishError(err, nil)
	}

	event.EventStatusID = status.ID

	// Set PublishedAt / LastPublishedAt so IsPublished() returns true.
	// Without this, later UpdateEvent calls on this event can't detect
	// the published state and skip the async provisioning enqueue.
	now := time.Now()
	if event.PublishedAt == nil {
		event.PublishedAt = &now
	}
	event.LastPublishedAt = &now

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to update event: %v", err)
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	// Reload the whole event so the returned struct reflects what's
	// actually in the DB.
	if reloaded, reloadErr := s.repo.GetEventByID(ctx, id); reloadErr == nil && reloaded != nil {
		*event = *reloaded
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event after publish: %v", reloadErr)
	}

	// Mirror schedules into the attendance module. Best-effort.
	s.syncEventSchedulesToAttendance(ctx, event)

	// Fire-and-forget: provision video meetings for in-window
	// schedules. The worker skips schedules that already have
	// video_meeting_id, so this is idempotent.
	s.enqueueVideoProvisioning(ctx, event.ID)

	log.Printf("✅ Event published successfully: %s", id)
	return event, nil
}

// CancelEvent cancels a single event.
func (s *eventService) CancelEvent(ctx context.Context, id, cancelledBy string) (*domain.Event, error) {
	if cancelledBy == "" {
		return nil, errors.New("cancelled by is required")
	}

	event, err := s.getEventAndCheckUpdatePermission(ctx, id, cancelledBy)
	if err != nil {
		return nil, err
	}

	if err := event.Cancel(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to cancel event: %w", err)
	}

	if reloaded, reloadErr := s.repo.GetEventByID(ctx, id); reloadErr == nil && reloaded != nil {
		*event = *reloaded
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event after cancel: %v", reloadErr)
	}

	log.Printf("✅ Event cancelled: %s by %s", id, cancelledBy)
	return event, nil
}

// CompleteEvent marks a single event as completed.
//
// No permission check — completion is a scheduled/system action, not a
// user-initiated one.
func (s *eventService) CompleteEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if err := event.Complete(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to complete event: %w", err)
	}

	if reloaded, reloadErr := s.repo.GetEventByID(ctx, id); reloadErr == nil && reloaded != nil {
		*event = *reloaded
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event after complete: %v", reloadErr)
	}

	log.Printf("✅ Event completed: %s", id)
	return event, nil
}

// ============================================================
// STATUS - Bulk
// ============================================================

func (s *eventService) BulkPublishEvents(ctx context.Context, ids []string, publishedBy string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}
	if publishedBy == "" {
		return nil, errors.New("published by is required")
	}

	result := &BulkStatusResult{
		ProcessedCount: 0,
		FailedIDs:      []string{},
		Errors:         []string{},
	}

	for _, id := range ids {
		if _, err := s.PublishEvent(ctx, id, publishedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			errMsg := err.Error()
			if after, ok := strings.CutPrefix(errMsg, "validation failed: "); ok {
				errMsg = after
			}
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		result.ProcessedCount++
	}

	s.logBulkStatusResult("publish", result, len(ids))
	return result, nil
}

func (s *eventService) BulkCancelEvents(ctx context.Context, ids []string, cancelledBy string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}
	if cancelledBy == "" {
		return nil, errors.New("cancelled by is required")
	}

	result := &BulkStatusResult{
		ProcessedCount: 0,
		FailedIDs:      []string{},
		Errors:         []string{},
	}

	for _, id := range ids {
		if _, err := s.CancelEvent(ctx, id, cancelledBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.ProcessedCount++
	}

	s.logBulkStatusResult("cancel", result, len(ids))
	return result, nil
}

func (s *eventService) BulkCompleteEvents(ctx context.Context, ids []string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkStatusResult{
		ProcessedCount: 0,
		FailedIDs:      []string{},
		Errors:         []string{},
	}

	for _, id := range ids {
		if _, err := s.CompleteEvent(ctx, id); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.ProcessedCount++
	}

	s.logBulkStatusResult("complete", result, len(ids))
	return result, nil
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

func (s *eventService) getEventAndCheckPublishPermission(ctx context.Context, id, publishedBy string) (*domain.Event, error) {
	if publishedBy == "" {
		return nil, errors.New("published by is required")
	}

	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	accountID, err := s.resolveEventAccountID(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("cannot authorize publish: %w", err)
	}

	accountDomain := domain.AccountDomain(accountID)
	log.Printf("🔍 PUBLISH CHECK: user=%s accountID=%s domain=%s event=%s",
		publishedBy, accountID, accountDomain, event.ID)

	canAll, err := s.permChecker.CanPublishAllEvents(ctx, publishedBy, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if canAll {
		return event, nil
	}

	canOwn, err := s.permChecker.CanPublishOwnEvents(ctx, publishedBy, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if canOwn && event.CreatedBy == publishedBy {
		return event, nil
	}

	log.Printf("❌ Publish permission denied: user=%s accountDomain=%s event=%s",
		publishedBy, accountDomain, event.ID)
	return nil, errors.New("insufficient permissions to publish this event")
}

func (s *eventService) getEventStatusBySlug(ctx context.Context, slug string) (*domain.EventStatus, error) {
	status, err := s.repo.GetEventStatusBySlug(ctx, slug)
	if err != nil {
		log.Printf("❌ Failed to get status: %v", err)
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	if status == nil {
		log.Printf("❌ Status not found for slug: %s", slug)
		return nil, domain.ErrEventStatusNotFound
	}
	return status, nil
}