// internal/modules/events/domain/permission.go

package domain

import "context"

// PermissionChecker defines what the events domain needs for authorization
type PermissionChecker interface {
	// ============================================================
	// CORE PERMISSION METHODS
	// ============================================================

	// HasPermission checks if a user has a specific permission in a domain
	// domain: "personal:team:{user_id}", "institution:team:{account_id}", "account:{account_id}", or "platform"
	HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error)

	// HasAnyPermission checks if a user has any of the given permissions in a domain
	HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// HasAllPermissions checks if a user has all of the given permissions in a domain
	HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// ============================================================
	// CREATE PERMISSIONS
	// ============================================================

	// CanCreateEvent checks if user can create events in a domain
	CanCreateEvent(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// READ PERMISSIONS
	// ============================================================

	// CanReadAllEvents checks if user can read ALL events in a domain
	CanReadAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadOwnEvents checks if user can read OWN events in a domain
	CanReadOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadEvent checks if user can read events in a domain (ALL or OWN)
	CanReadEvent(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// UPDATE PERMISSIONS
	// ============================================================

	// CanUpdateAllEvents checks if user can update ALL events in a domain
	CanUpdateAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateOwnEvents checks if user can update OWN events in a domain
	CanUpdateOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateEvent checks if user can update events in a domain (ALL or OWN)
	CanUpdateEvent(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// DELETE PERMISSIONS
	// ============================================================

	// CanDeleteAllEvents checks if user can delete ALL events in a domain
	CanDeleteAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteOwnEvents checks if user can delete OWN events in a domain
	CanDeleteOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteEvent checks if user can delete events in a domain (ALL or OWN)
	CanDeleteEvent(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// PUBLISH PERMISSIONS
	// ============================================================

	// CanPublishAllEvents checks if user can publish ALL events in a domain
	CanPublishAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanPublishOwnEvents checks if user can publish OWN events in a domain
	CanPublishOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanPublishEvent checks if user can publish events in a domain (ALL or OWN)
	CanPublishEvent(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// MANAGEMENT (Convenience)
	// ============================================================

	// CanManageEvent checks if user can manage events in a domain
	CanManageEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewEvent checks if user can view events in a domain
	CanViewEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewCreator checks if user can view creator details
	CanViewCreator(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// USER INFORMATION METHODS
	// ============================================================

	// GetUserTeamDomains returns all team domains where a user has membership
	// Returns domains in format: "personal:team:{user_id}" and "institution:team:{account_id}"
	GetUserTeamDomains(ctx context.Context, userID string) ([]string, error)

	// GetUserAccountIDs returns all account IDs where a user has membership
	GetUserAccountIDs(ctx context.Context, userID string) ([]string, error)
}