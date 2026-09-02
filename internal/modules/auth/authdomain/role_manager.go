// internal/modules/auth/authdomain/role_manager.go

package authdomain

import "context"

// RoleManager handles role assignments and revocations (state mutations)
// This interface is focused ONLY on writing/modifying roles
type RoleManager interface {
	// AssignRole assigns a role to a user in a scope
	// Returns error if assignment fails
	AssignRole(ctx context.Context, scope Scope, userID string, role string) error

	// RemoveRole removes a role from a user in a scope
	RemoveRole(ctx context.Context, scope Scope, userID string, role string) error

	// RemoveAllRoles removes all roles for a user from a scope
	RemoveAllRoles(ctx context.Context, scope Scope, userID string) error

	// GetUserRoles returns all roles for a user in a scope
	GetUserRoles(ctx context.Context, userID string, scope Scope) ([]string, error)

	// ============================================================
	// CONVENIENCE METHODS (Built on AssignRole/RemoveRole)
	// ============================================================

	// AssignPersonalTeamAdmin assigns account_admin to a user's personal team
	AssignPersonalTeamAdmin(ctx context.Context, userID string) error

	// AssignInstitutionAdmin assigns account_admin in an institution team
	AssignInstitutionAdmin(ctx context.Context, institutionID, userID string) error

	// AssignEventManager assigns event_manager in a scope
	AssignEventManager(ctx context.Context, scope Scope, userID string) error

	// AssignTeamMember assigns team_member in a scope
	AssignTeamMember(ctx context.Context, scope Scope, userID string) error

	// RemovePersonalTeamAdmin removes account_admin from a user's personal team
	RemovePersonalTeamAdmin(ctx context.Context, userID string) error

	// RemoveInstitutionAdmin removes account_admin from an institution team
	RemoveInstitutionAdmin(ctx context.Context, institutionID, userID string) error
}