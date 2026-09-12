// internal/modules/events/service/delete_event.go

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

// DeleteEvent soft deletes a single event.
//
// Authorization is checked against the event's parent account domain.
// The accountID and teamType parameters are accepted for backward
// compatibility but are no longer used for authz — the event's own
// AccountID is authoritative.
func (s *eventService) DeleteEvent(ctx context.Context, id, deletedBy, accountID, teamType string) error {
	if id == "" {
		return errors.New("event ID is required")
	}
	if deletedBy == "" {
		return errors.New("deleted by is required")
	}
	_ = teamType // ignored

	event, err := s.getEventAndCheckDeletePermission(ctx, id, deletedBy, accountID)
	if err != nil {
		return err
	}

	if err := event.SoftDelete(deletedBy); err != nil {
		return err
	}
	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	log.Printf("✅ Event soft deleted: %s by %s", id, deletedBy)
	return nil
}

// PermanentlyDeleteEvent hard deletes a single event.
func (s *eventService) PermanentlyDeleteEvent(ctx context.Context, id, deletedBy, accountID, teamType string) error {
	if id == "" {
		return errors.New("event ID is required")
	}
	if deletedBy == "" {
		return errors.New("deleted by is required")
	}
	_ = teamType

	if err := s.checkDeletePermissionForEvent(ctx, id, deletedBy, accountID); err != nil {
		return err
	}

	if s.mediaSvc != nil {
		if err := s.mediaSvc.DeleteFilesByEntity(ctx, id); err != nil {
			log.Printf("⚠️ Failed to delete media for event %s: %v", id, err)
		}
	}

	if err := s.repo.PermanentlyDeleteEvent(ctx, id); err != nil {
		return fmt.Errorf("failed to permanently delete event: %w", err)
	}

	log.Printf("✅ Event permanently deleted: %s by %s", id, deletedBy)
	return nil
}

// RestoreEvent restores a soft-deleted event.
func (s *eventService) RestoreEvent(ctx context.Context, id, restoredBy, accountID, teamType string) (*domain.Event, error) {
	if id == "" {
		return nil, errors.New("event ID is required")
	}
	if restoredBy == "" {
		return nil, errors.New("restored by is required")
	}
	_ = teamType

	event, err := s.getEventAndCheckUpdatePermissionIncludingDeleted(ctx, id, restoredBy, accountID)
	if err != nil {
		return nil, err
	}

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

// DeleteEvents soft deletes multiple events.
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

// PermanentlyDeleteEvents hard deletes multiple events.
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

// RestoreEvents restores multiple soft-deleted events.
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

// validateDeletePermission checks whether the user may delete this event.
//
// POST-REVAMP: single Casbin check against the account domain.
// `fallbackAccountID` is used only when the loaded event has no AccountID
// (legacy rows). Passing an empty value and hitting a legacy event is an
// error.
func (s *eventService) validateDeletePermission(
	ctx context.Context,
	event *domain.Event,
	userID, fallbackAccountID string,
) error {
	accountID := event.AccountID
	if accountID == "" {
		accountID = fallbackAccountID
	}
	if accountID == "" {
		return errors.New("event has no account ID; cannot authorize delete")
	}

	accountDomain := domain.AccountDomain(accountID)
	log.Printf("🔍 AUTHZ: event:delete check user=%s domain=%s event=%s",
		userID, accountDomain, event.ID)

	allowed, err := s.permChecker.CanDeleteEvent(ctx, userID, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot delete event %s in domain %s",
			userID, event.ID, accountDomain)
		return domain.ErrForbidden
	}

	return nil
}

// validateRestorePermission checks whether the user may restore this event.
//
// POST-REVAMP: single Casbin check against the account domain.
func (s *eventService) validateRestorePermission(
	ctx context.Context,
	event *domain.Event,
	userID, fallbackAccountID string,
) error {
	accountID := event.AccountID
	if accountID == "" {
		accountID = fallbackAccountID
	}
	if accountID == "" {
		return errors.New("event has no account ID; cannot authorize restore")
	}

	accountDomain := domain.AccountDomain(accountID)
	log.Printf("🔍 AUTHZ: event:update check user=%s domain=%s event=%s (restore)",
		userID, accountDomain, event.ID)

	allowed, err := s.permChecker.CanUpdateEvent(ctx, userID, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot restore event %s in domain %s",
			userID, event.ID, accountDomain)
		return domain.ErrForbidden
	}

	return nil
}

// checkDeletePermissionForEvent loads the event (including soft-deleted)
// and verifies the user may delete it.
func (s *eventService) checkDeletePermissionForEvent(
	ctx context.Context,
	id, deletedBy, accountID string,
) error {
	event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return domain.ErrEventNotFound
	}

	return s.validateDeletePermission(ctx, event, deletedBy, accountID)
}

// getEventAndCheckDeletePermission loads the event and verifies the user
// may delete it.
func (s *eventService) getEventAndCheckDeletePermission(
	ctx context.Context,
	id, deletedBy, accountID string,
) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if err := s.validateDeletePermission(ctx, event, deletedBy, accountID); err != nil {
		return nil, err
	}

	return event, nil
}

// getEventAndCheckUpdatePermissionIncludingDeleted loads the event
// (including soft-deleted) and verifies the user may update/restore it.
func (s *eventService) getEventAndCheckUpdatePermissionIncludingDeleted(
	ctx context.Context,
	id, restoredBy, accountID string,
) (*domain.Event, error) {
	event, err := s.repo.GetEventByIDIncludingDeleted(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	if err := s.validateRestorePermission(ctx, event, restoredBy, accountID); err != nil {
		return nil, err
	}

	return event, nil
}