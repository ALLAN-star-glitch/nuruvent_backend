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

// PublishEvent publishes a single event
func (s *eventService) PublishEvent(ctx context.Context, id, publishedBy string) (*domain.Event, error) {
	log.Printf("📤 Publishing event: %s", id)

	// 1. Get event and check permissions using Scope
	event, err := s.getEventAndCheckPublishPermission(ctx, id, publishedBy)
	if err != nil {
		return nil, err
	}

	// 2. Get published status
	status, err := s.getEventStatusBySlug(ctx, domain.EventStatusPublished.GetSlug())
	if err != nil {
		return nil, err
	}

	// 3. Validate event for publishing
	if err := event.ValidateForPublish(); err != nil {
		log.Printf("❌ Publish validation failed: %v", err)
		return nil, fmt.Errorf("cannot publish event: %w", err)
	}

	// 4. Set status and save
	event.EventStatusID = status.ID
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		log.Printf("❌ Failed to update event: %v", err)
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("✅ Event published successfully: %s", id)
	return event, nil
}

// CancelEvent cancels a single event
func (s *eventService) CancelEvent(ctx context.Context, id, cancelledBy string) (*domain.Event, error) {
	// 1. Get event and check permissions using Scope
	// ✅ Reuses getEventAndCheckUpdatePermission from media.go
	event, err := s.getEventAndCheckUpdatePermission(ctx, id, cancelledBy)
	if err != nil {
		return nil, err
	}

	// 2. Cancel event
	if err := event.Cancel(); err != nil {
		return nil, err
	}

	// 3. Save
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to cancel event: %w", err)
	}

	log.Printf("✅ Event cancelled: %s by %s", id, cancelledBy)
	return event, nil
}

// CompleteEvent marks a single event as completed
func (s *eventService) CompleteEvent(ctx context.Context, id string) (*domain.Event, error) {
	// 1. Get event (no permission check needed - completion is automatic)
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// 2. Complete event
	if err := event.Complete(); err != nil {
		return nil, err
	}

	// 3. Save
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to complete event: %w", err)
	}

	log.Printf("✅ Event completed: %s", id)
	return event, nil
}

// ============================================================
// STATUS - Bulk
// ============================================================

// BulkPublishEvents publishes multiple events
func (s *eventService) BulkPublishEvents(ctx context.Context, ids []string, publishedBy string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
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
			if strings.HasPrefix(errMsg, "validation failed: ") {
				errMsg = strings.TrimPrefix(errMsg, "validation failed: ")
			}
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		result.ProcessedCount++
	}

	s.logBulkStatusResult("publish", result, len(ids))
	return result, nil
}

// BulkCancelEvents cancels multiple events
func (s *eventService) BulkCancelEvents(ctx context.Context, ids []string, cancelledBy string) (*BulkStatusResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
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

// BulkCompleteEvents completes multiple events
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

// getEventAndCheckPublishPermission gets event and checks publish permission using Scope
func (s *eventService) getEventAndCheckPublishPermission(ctx context.Context, id, publishedBy string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// Create scope from event
	scope := s.getScopeFromEvent(event)

	// Check if user can publish events in this scope
	allowed, err := s.permChecker.CanPublishEvent(ctx, publishedBy, scope)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, errors.New("insufficient permissions to publish this event")
	}

	return event, nil
}

// getEventStatusBySlug retrieves an event status by slug
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
