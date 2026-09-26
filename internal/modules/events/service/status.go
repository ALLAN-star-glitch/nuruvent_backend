// internal/modules/events/service/status.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

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

	// NEW: Retry video meeting creation for any virtual schedule that
	// doesn't yet have a link. Covers drafts saved before the host
	// connected their platform. attachVideoMeetings is idempotent —
	// schedules that already have a VideoMeetingID or manual link are
	// skipped.
	if err := s.attachVideoMeetings(ctx, event, publishedBy); err != nil {
		log.Printf("⚠️ video integration on publish: %v", err)
	}

	if err := event.ValidateForPublish(); err != nil {
		log.Printf("❌ Publish validation failed: %v", err)
		return nil, fmt.Errorf("cannot publish event: %w", err)
	}

	event.EventStatusID = status.ID
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to update event: %v", err)
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	// Reload so DB-assigned schedule IDs are populated on the domain
	// struct before we mirror schedules into attendance.
	if reloaded, reloadErr := s.repo.GetEventByID(ctx, id); reloadErr == nil && reloaded != nil {
		event.Schedules = reloaded.Schedules
	} else if reloadErr != nil {
		log.Printf("⚠️ Could not reload event for attendance sync: %v", reloadErr)
	}

	s.syncEventSchedulesToAttendance(ctx, event)

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

	log.Printf("✅ Event completed: %s", id)
	return event, nil
}

// ============================================================
// STATUS - Bulk
// ============================================================

// BulkPublishEvents publishes multiple events.
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

// BulkCancelEvents cancels multiple events.
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

// BulkCompleteEvents completes multiple events.
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

// getEventAndCheckPublishPermission loads the event and verifies the user
// may publish it.
//
// POST-REVAMP: single Casbin check against the event's parent account
// domain. Teams are not authorization domains.
//
// The account is resolved through `resolveEventAccountID`, which walks
// events.team_id → teams.account_id when the event doesn't carry its own
// account reference. See helpers.go for details.
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
	log.Printf("🔍 PUBLISH CHECK: user=%s accountID=%s domain=%s",
		publishedBy, accountID, accountDomain)

	allowed, err := s.permChecker.CanPublishAllEvents(ctx, publishedBy, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		log.Printf("❌ Publish permission denied: user=%s accountDomain=%s", publishedBy, accountDomain)
		return nil, errors.New("insufficient permissions to publish this event")
	}

	return event, nil
}

// getEventStatusBySlug retrieves an event status by slug.
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






