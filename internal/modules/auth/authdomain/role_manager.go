// internal/modules/auth/authdomain/role_manager.go

package authdomain

import "context"

// RoleManager manages user-to-role bindings in Casbin.
type RoleManager interface {
	// AssignRole writes a grouping rule binding a user to a role in a domain.
	AssignRole(ctx context.Context, domain, userID, role string) error

	// RemoveRole removes a grouping rule.
	RemoveRole(ctx context.Context, domain, userID, role string) error

	// GetUserRoles returns all roles a user holds in a domain.
	GetUserRoles(ctx context.Context, userID, domain string) ([]string, error)

	// RemoveAllRolesInDomain removes every grouping rule for a domain.
	// Used when an account is deleted.
	RemoveAllRolesInDomain(ctx context.Context, domain string) error

	// ReloadPolicies forces the enforcer to re-read its in-memory model from
	// the adapter (DB).
	//
	// Call this after writes that must be visible to subsequent permission
	// checks in the same request. The enforcer's auto-load runs on an
	// interval (10s by default); without this, a just-written rule may not
	// be visible until the next cycle.
	ReloadPolicies(ctx context.Context) error
}