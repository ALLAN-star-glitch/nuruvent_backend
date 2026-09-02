// internal/modules/auth/authdomain/permission_checker.go

package authdomain

import "context"

// PermissionChecker handles access evaluation queries
// This interface is focused ONLY on reading/checking permissions
type PermissionChecker interface {
	// HasPermission checks if a user has a specific permission in a scope
	// Returns (true, nil) if allowed, (false, nil) if denied, (false, error) if check failed
	HasPermission(ctx context.Context, userID string, scope Scope, resource, action string) (bool, error)

	// HasAnyPermission checks if a user has any of the given permissions in a scope
	HasAnyPermission(ctx context.Context, userID string, scope Scope, resource string, actions ...string) (bool, error)

	// HasAllPermissions checks if a user has all of the given permissions in a scope
	HasAllPermissions(ctx context.Context, userID string, scope Scope, resource string, actions ...string) (bool, error)

	// ============================================================
	// CONVENIENCE METHODS (Built on HasPermission)
	// ============================================================

	// CanReadAllEvents checks if user can read ALL events in a scope
	CanReadAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanReadOwnEvents checks if user can read OWN events in a scope
	CanReadOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanCreateEvent checks if user can create events in a scope
	CanCreateEvent(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanUpdateAllEvents checks if user can update ALL events in a scope
	CanUpdateAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanUpdateOwnEvents checks if user can update OWN events in a scope
	CanUpdateOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanDeleteAllEvents checks if user can delete ALL events in a scope
	CanDeleteAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanDeleteOwnEvents checks if user can delete OWN events in a scope
	CanDeleteOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanPublishAllEvents checks if user can publish ALL events in a scope
	CanPublishAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanPublishOwnEvents checks if user can publish OWN events in a scope
	CanPublishOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanManageEvent checks if user can manage events in a scope
	// (Combines CanUpdateAllEvents + CanDeleteAllEvents)
	CanManageEvent(ctx context.Context, userID string, scope Scope) (bool, error)

	// CanViewEvent checks if user can view events in a scope
	// (Combines CanReadAllEvents + CanReadOwnEvents)
	CanViewEvent(ctx context.Context, userID string, scope Scope) (bool, error)

	// ============================================================
	// TEAM ROLE CHECKS (Built on HasPermission)
	// ============================================================

	// IsTeamAdmin checks if user is an admin in the scope
	IsTeamAdmin(ctx context.Context, userID string, scope Scope) (bool, error)

	// IsEventManager checks if user is an event manager in the scope
	IsEventManager(ctx context.Context, userID string, scope Scope) (bool, error)

	// IsTeamMember checks if user is a team member in the scope
	IsTeamMember(ctx context.Context, userID string, scope Scope) (bool, error)

	// ============================================================
	// USER INFORMATION METHODS
	// ============================================================

	// GetUserRoles returns all roles for a user in a scope
	GetUserRoles(ctx context.Context, userID string, scope Scope) ([]string, error)

	// GetUserTeamIDs returns all team IDs where a user has roles
	GetUserTeamIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserPersonalTeamIDs returns personal team IDs where a user has roles
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserInstitutionTeamIDs returns institution team IDs where a user has roles
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)

	// HasTeamAccess checks if a user has any team role
	HasTeamAccess(ctx context.Context, userID string) (bool, error)
}