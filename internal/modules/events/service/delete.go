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

// DeleteEventsByInstitution soft deletes all events for an institution
func (s *eventService) DeleteEventsByInstitution(ctx context.Context, institutionID string, deletedBy string) (*BulkDeleteResult, error) {
	if institutionID == "" {
		return nil, errors.New("institution ID is required")
	}

	events, err := s.getEventsByInstitution(ctx, institutionID)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &BulkDeleteResult{DeletedCount: 0}, nil
	}

	ids := s.extractEventIDs(events)
	return s.DeleteEvents(ctx, ids, deletedBy)
}

// PermanentlyDeleteEventsByInstitution hard deletes all events for an institution
func (s *eventService) PermanentlyDeleteEventsByInstitution(ctx context.Context, institutionID string, deletedBy string) (*BulkDeleteResult, error) {
	if institutionID == "" {
		return nil, errors.New("institution ID is required")
	}

	events, err := s.getEventsByInstitution(ctx, institutionID)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return &BulkDeleteResult{DeletedCount: 0}, nil
	}

	ids := s.extractEventIDs(events)
	return s.PermanentlyDeleteEvents(ctx, ids, deletedBy)
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

// getEventAndCheckDeletePermission gets event and checks delete permission
func (s *eventService) getEventAndCheckDeletePermission(ctx context.Context, id, deletedBy string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	institutionID := s.getInstitutionIDFromEvent(event)
	if !s.permChecker.CanDeleteEvent(ctx, deletedBy, institutionID) {
		return nil, errors.New("insufficient permissions to delete this event")
	}

	return event, nil
}

// checkDeletePermissionForEvent checks if user has permission to delete an event
// (doesn't return the event, only checks permission)
func (s *eventService) checkDeletePermissionForEvent(ctx context.Context, id, deletedBy string) error {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	institutionID := s.getInstitutionIDFromEvent(event)
	if !s.permChecker.CanDeleteEvent(ctx, deletedBy, institutionID) {
		return errors.New("insufficient permissions to delete this event")
	}

	return nil
}

// getEventAndCheckUpdatePermissionIncludingDeleted gets event (including soft-deleted) and checks update permission
func (s *eventService) getEventAndCheckUpdatePermissionIncludingDeleted(ctx context.Context, id, restoredBy string) (*domain.Event, error) {
	event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	institutionID := s.getInstitutionIDFromEvent(event)
	if !s.permChecker.CanUpdateEvent(ctx, restoredBy, institutionID) {
		return nil, errors.New("insufficient permissions to restore this event")
	}

	return event, nil
}

// getInstitutionIDFromEvent extracts institution ID from event
func (s *eventService) getInstitutionIDFromEvent(event *domain.Event) string {
	if event.InstitutionID != nil {
		return *event.InstitutionID
	}
	return ""
}

// getEventsByInstitution gets all events for an institution
func (s *eventService) getEventsByInstitution(ctx context.Context, institutionID string) ([]*domain.Event, error) {
	events, _, err := s.repo.GetEventsByInstitution(ctx, institutionID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for institution: %w", err)
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