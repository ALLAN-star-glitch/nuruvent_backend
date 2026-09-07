// internal/app/adapters/events/permission_adapter.go

package events

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// PermissionAdapter adapts the auth module's PermissionChecker to the events domain's PermissionChecker
type PermissionAdapter struct {
	authPermChecker authdomain.PermissionChecker
}

// NewPermissionAdapter creates a new permission adapter for events
func NewPermissionAdapter(authPermChecker authdomain.PermissionChecker) domain.PermissionChecker {
	return &PermissionAdapter{
		authPermChecker: authPermChecker,
	}
}

// ============================================================
// CORE PERMISSION METHODS
// ============================================================

func (a *PermissionAdapter) HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error) {
	return a.authPermChecker.HasPermission(ctx, userID, domain, resource, action)
}

func (a *PermissionAdapter) HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	return a.authPermChecker.HasAnyPermission(ctx, userID, domain, resource, actions...)
}

func (a *PermissionAdapter) HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	return a.authPermChecker.HasAllPermissions(ctx, userID, domain, resource, actions...)
}

// ============================================================
// CREATE PERMISSIONS
// ============================================================

func (a *PermissionAdapter) CanCreateEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanCreateEvent(ctx, userID, domain)
}

// ============================================================
// READ PERMISSIONS
// ============================================================

func (a *PermissionAdapter) CanReadAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanReadAllEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanReadOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanReadOwnEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanReadEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanViewEvent(ctx, userID, domain)
}

// ============================================================
// UPDATE PERMISSIONS
// ============================================================

func (a *PermissionAdapter) CanUpdateAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanUpdateAllEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanUpdateOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanUpdateOwnEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanUpdateEvent(ctx context.Context, userID string, domain string) (bool, error) {
	// CanUpdateEvent = CanUpdateAllEvents OR CanUpdateOwnEvents
	canUpdateAll, err := a.authPermChecker.CanUpdateAllEvents(ctx, userID, domain)
	if err != nil {
		return false, err
	}
	if canUpdateAll {
		return true, nil
	}
	return a.authPermChecker.CanUpdateOwnEvents(ctx, userID, domain)
}

// ============================================================
// DELETE PERMISSIONS
// ============================================================

func (a *PermissionAdapter) CanDeleteAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanDeleteAllEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanDeleteOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanDeleteOwnEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanDeleteEvent(ctx context.Context, userID string, domain string) (bool, error) {
	// CanDeleteEvent = CanDeleteAllEvents OR CanDeleteOwnEvents
	canDeleteAll, err := a.authPermChecker.CanDeleteAllEvents(ctx, userID, domain)
	if err != nil {
		return false, err
	}
	if canDeleteAll {
		return true, nil
	}
	return a.authPermChecker.CanDeleteOwnEvents(ctx, userID, domain)
}

// ============================================================
// PUBLISH PERMISSIONS
// ============================================================

func (a *PermissionAdapter) CanPublishAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanPublishAllEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanPublishOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanPublishOwnEvents(ctx, userID, domain)
}

func (a *PermissionAdapter) CanPublishEvent(ctx context.Context, userID string, domain string) (bool, error) {
	// CanPublishEvent = CanPublishAllEvents OR CanPublishOwnEvents
	canPublishAll, err := a.authPermChecker.CanPublishAllEvents(ctx, userID, domain)
	if err != nil {
		return false, err
	}
	if canPublishAll {
		return true, nil
	}
	return a.authPermChecker.CanPublishOwnEvents(ctx, userID, domain)
}

// ============================================================
// MANAGEMENT (Convenience)
// ============================================================

func (a *PermissionAdapter) CanManageEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanManageEvent(ctx, userID, domain)
}

func (a *PermissionAdapter) CanViewEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanViewEvent(ctx, userID, domain)
}

func (a *PermissionAdapter) CanViewCreator(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanViewCreator(ctx, userID, domain)
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

func (a *PermissionAdapter) GetUserTeamDomains(ctx context.Context, userID string) ([]string, error) {
	return a.authPermChecker.GetUserTeamDomains(ctx, userID)
}

func (a *PermissionAdapter) GetUserAccountIDs(ctx context.Context, userID string) ([]string, error) {
	return a.authPermChecker.GetUserAccountIDs(ctx, userID)
}