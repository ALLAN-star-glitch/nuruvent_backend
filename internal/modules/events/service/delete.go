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

// DeleteEvent soft deletes a single event with 2-tier authorization
func (s *eventService) DeleteEvent(ctx context.Context, id, deletedBy, accountID, teamType string) error {
	// 1. Validate input
	if id == "" {
		return errors.New("event ID is required")
	}
	if deletedBy == "" {
		return errors.New("deleted by is required")
	}

	// 2. Get event and check permissions using caller context fallbacks
	event, err := s.getEventAndCheckDeletePermission(ctx, id, deletedBy, accountID, teamType)
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

// PermanentlyDeleteEvent hard deletes a single event with 2-tier authorization
func (s *eventService) PermanentlyDeleteEvent(ctx context.Context, id, deletedBy, accountID, teamType string) error {
	// 1. Validate input
	if id == "" {
		return errors.New("event ID is required")
	}
	if deletedBy == "" {
		return errors.New("deleted by is required")
	}

	// 2. Get event including soft-deleted and check permissions
	if err := s.checkDeletePermissionForEvent(ctx, id, deletedBy, accountID, teamType); err != nil {
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

// RestoreEvent restores a soft-deleted event with 2-tier authorization
func (s *eventService) RestoreEvent(ctx context.Context, id, restoredBy, accountID, teamType string) (*domain.Event, error) {
	// 1. Validate input
	if id == "" {
		return nil, errors.New("event ID is required")
	}
	if restoredBy == "" {
		return nil, errors.New("restored by is required")
	}

	// 2. Get event (including soft-deleted) and check permissions
	event, err := s.getEventAndCheckUpdatePermissionIncludingDeleted(ctx, id, restoredBy, accountID, teamType)
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
func (s *eventService) DeleteEvents(ctx context.Context, ids []string, deletedBy, accountID, teamType string) (*BulkDeleteResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}
	if deletedBy == "" {
		return nil, errors.New("deleted by is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, id := range ids {
		if err := s.DeleteEvent(ctx, id, deletedBy, accountID, teamType); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// PermanentlyDeleteEvents hard deletes multiple events
func (s *eventService) PermanentlyDeleteEvents(ctx context.Context, ids []string, deletedBy, accountID, teamType string) (*BulkDeleteResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}
	if deletedBy == "" {
		return nil, errors.New("deleted by is required")
	}

	result := &BulkDeleteResult{
		DeletedCount: 0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, id := range ids {
		if err := s.PermanentlyDeleteEvent(ctx, id, deletedBy, accountID, teamType); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.DeletedCount++
	}

	return result, nil
}

// RestoreEvents restores multiple soft-deleted events
func (s *eventService) RestoreEvents(ctx context.Context, ids []string, restoredBy, accountID, teamType string) (*BulkRestoreResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}
	if restoredBy == "" {
		return nil, errors.New("restored by is required")
	}

	result := &BulkRestoreResult{
		RestoredCount: 0,
		FailedIDs:     []string{},
		Errors:        []string{},
	}

	for _, id := range ids {
		if _, err := s.RestoreEvent(ctx, id, restoredBy, accountID, teamType); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}
		result.RestoredCount++
	}

	return result, nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// validateDeletePermission performs 2-tier domain authorization check for event deletion
func (s *eventService) validateDeletePermission(ctx context.Context, event *domain.Event, userID, fallbackAccountID, fallbackTeamType string) error {
	// Fallback scoping parameters if event aggregate root lacks stored values
	if event.AccountID == "" {
		event.AccountID = fallbackAccountID
	}
	if event.TeamType == "" {
		event.TeamType = fallbackTeamType
	}

	teamDomain := s.resolveEventTeamDomain(event)

	// Tier 1: Check Team Domain
	allowed, err := s.permChecker.CanDeleteEvent(ctx, userID, teamDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Tier 2: Check Account Domain fallback (if present)
	if !allowed && event.AccountID != "" {
		accountDomain := domain.AccountDomain(event.AccountID)
		allowed, err = s.permChecker.CanDeleteEvent(ctx, userID, accountDomain)
		if err != nil {
			return fmt.Errorf("account permission check failed: %w", err)
		}
	}

	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot delete event %s in domain %s", userID, event.ID, teamDomain)
		return domain.ErrForbidden
	}

	return nil
}

// validateRestorePermission performs 2-tier domain authorization check for event restore (update action)
func (s *eventService) validateRestorePermission(ctx context.Context, event *domain.Event, userID, fallbackAccountID, fallbackTeamType string) error {
	// Fallback scoping parameters if event aggregate root lacks stored values
	if event.AccountID == "" {
		event.AccountID = fallbackAccountID
	}
	if event.TeamType == "" {
		event.TeamType = fallbackTeamType
	}

	teamDomain := s.resolveEventTeamDomain(event)

	// Tier 1: Check Team Domain
	allowed, err := s.permChecker.CanUpdateEvent(ctx, userID, teamDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Tier 2: Check Account Domain fallback (if present)
	if !allowed && event.AccountID != "" {
		accountDomain := domain.AccountDomain(event.AccountID)
		allowed, err = s.permChecker.CanUpdateEvent(ctx, userID, accountDomain)
		if err != nil {
			return fmt.Errorf("account permission check failed: %w", err)
		}
	}

	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot restore event %s in domain %s", userID, event.ID, teamDomain)
		return domain.ErrForbidden
	}

	return nil
}

// checkDeletePermissionForEvent checks if user has permission to delete an event
func (s *eventService) checkDeletePermissionForEvent(ctx context.Context, id, deletedBy, accountID, teamType string) error {
	event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	return s.validateDeletePermission(ctx, event, deletedBy, accountID, teamType)
}

// getEventAndCheckDeletePermission gets event and checks delete permission
func (s *eventService) getEventAndCheckDeletePermission(ctx context.Context, id, deletedBy, accountID, teamType string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if err := s.validateDeletePermission(ctx, event, deletedBy, accountID, teamType); err != nil {
		return nil, err
	}

	return event, nil
}

// getEventAndCheckUpdatePermissionIncludingDeleted gets event (including soft-deleted) and checks update permission
func (s *eventService) getEventAndCheckUpdatePermissionIncludingDeleted(ctx context.Context, id, restoredBy, accountID, teamType string) (*domain.Event, error) {
	event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if err := s.validateRestorePermission(ctx, event, restoredBy, accountID, teamType); err != nil {
		return nil, err
	}

	return event, nil
}