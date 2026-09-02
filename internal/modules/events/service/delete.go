// internal/modules/events/service/delete.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// DELETE - Single Event
// ============================================================

// DeleteEvent soft deletes a single event
func (s *eventService) DeleteEvent(ctx context.Context, id, deletedBy string) error {
	// 1. Validate input
	if id == "" {
		return errors.New("event ID is required")
	}

	// 2. Get event and check permissions
	event, err := s.getEventAndCheckDeletePermission(ctx, id, deletedBy)
	if err != nil {
		return err
	}

	// 3. Soft delete and save
	if err := event.SoftDelete(deletedBy); err != nil {
		return err
	}
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	log.Printf("✅ Event soft deleted: %s by %s", id, deletedBy)
	return nil
}

// PermanentlyDeleteEvent hard deletes a single event
func (s *eventService) PermanentlyDeleteEvent(ctx context.Context, id, deletedBy string) error {
	// 1. Validate input
	if id == "" {
		return errors.New("event ID is required")
	}

	// 2. Check permissions (we don't need the event object)
	if err := s.checkDeletePermissionForEvent(ctx, id, deletedBy); err != nil {
		return err
	}

	// 3. Delete associated media
	if s.mediaSvc != nil {
		if err := s.mediaSvc.DeleteFilesByEntity(ctx, id); err != nil {
			log.Printf("⚠️ Failed to delete media for event %s: %v", id, err)
		}
	}

	// 4. Permanently delete from database
	if err := s.repo.PermanentlyDeleteEvent(ctx, id); err != nil {
		return fmt.Errorf("failed to permanently delete event: %w", err)
	}

	log.Printf("✅ Event permanently deleted: %s by %s", id, deletedBy)
	return nil
}

// RestoreEvent restores a soft-deleted event
func (s *eventService) RestoreEvent(ctx context.Context, id, restoredBy string) (*domain.Event, error) {
	// 1. Validate input
	if id == "" {
		return nil, errors.New("event ID is required")
	}

	// 2. Get event (including soft-deleted) and check permissions
	event, err := s.getEventAndCheckUpdatePermissionIncludingDeleted(ctx, id, restoredBy)
	if err != nil {
		return nil, err
	}

	// 3. Restore and save
	if err := event.Restore(restoredBy); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to restore event: %w", err)
	}

	log.Printf("✅ Event restored: %s by %s", id, restoredBy)
	return event, nil
}

// ============================================================
// DELETE - Bulk Events
// ============================================================

// DeleteEvents soft deletes multiple events
func (s *eventService) DeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, id := range ids {
		if err := s.DeleteEvent(ctx, id, deletedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// PermanentlyDeleteEvents hard deletes multiple events
func (s *eventService) PermanentlyDeleteEvents(ctx context.Context, ids []string, deletedBy string) (*BulkDeleteResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, id := range ids {
		if err := s.PermanentlyDeleteEvent(ctx, id, deletedBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// RestoreEvents restores multiple soft-deleted events
func (s *eventService) RestoreEvents(ctx context.Context, ids []string, restoredBy string) (*BulkRestoreResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}

	result := &BulkRestoreResult{
		RestoredCount: 0,
		FailedIDs:     []string{},
		Errors:        []string{},
	}

	for _, id := range ids {
		if _, err := s.RestoreEvent(ctx, id, restoredBy); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.RestoredCount++
	}

	return result, nil
}



// checkDeletePermissionForEvent checks if user has permission to delete an event
func (s *eventService) checkDeletePermissionForEvent(ctx context.Context, id, deletedBy string) error {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	scope := s.getScopeFromEvent(event)

	allowed, err := s.permChecker.CanDeleteEvent(ctx, deletedBy, scope)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return errors.New("insufficient permissions to delete this event")
	}

	return nil
}

// getEventAndCheckUpdatePermissionIncludingDeleted gets event (including soft-deleted) and checks update permission
// ✅ FIXED: Uses ListEvents with IncludeDeleted filter instead of GetEventByIDIncludingDeleted
func (s *eventService) getEventAndCheckUpdatePermissionIncludingDeleted(ctx context.Context, id, restoredBy string) (*domain.Event, error) {
	// Use ListEvents with IncludeDeleted and OnlyDeleted filters to get the event
	filters := domain.ListEventsFilters{
		IncludeDeleted: true,
		Limit:          1,
		Offset:         0,
	}

	// We need to find the event by ID - we'll filter after getting results
	events, _, err := s.repo.ListEvents(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Find the event with the matching ID
	var event *domain.Event
	for _, e := range events {
		if e.ID == id {
			event = e
			break
		}
	}

	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	scope := s.getScopeFromEvent(event)

	allowed, err := s.permChecker.CanUpdateEvent(ctx, restoredBy, scope)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, errors.New("insufficient permissions to restore this event")
	}

	return event, nil
}


// getEventsByTeam gets all events for a team (personal or institution)
func (s *eventService) getEventsByTeam(ctx context.Context, team domain.TeamFilter) ([]*domain.Event, error) {
	events, _, err := s.repo.ListEvents(ctx, domain.ListEventsFilters{
		Team:  team,
		Limit: 0,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for team: %w", err)
	}
	return events, nil
}

// extractEventIDs extracts IDs from a slice of events
func (s *eventService) extractEventIDs(events []*domain.Event) []string {
	ids := make([]string, len(events))
	for i, event := range events {
		ids[i] = event.ID
	}
	return ids
}