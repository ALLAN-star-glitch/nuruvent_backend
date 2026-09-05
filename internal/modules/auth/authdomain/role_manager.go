// internal/modules/auth/authdomain/role_manager.go

package authdomain

import "context"

// RoleManager handles role assignments and revocations (state mutations)
type RoleManager interface {
	// AssignRole assigns a role to a user in a domain
	AssignRole(ctx context.Context, domain string, userID string, role string) error

	// RemoveRole removes a role from a user in a domain
	RemoveRole(ctx context.Context, domain string, userID string, role string) error

	// RemoveAllRoles removes all roles for a user from a domain
	RemoveAllRoles(ctx context.Context, domain string, userID string) error

	// GetUserRoles returns all roles for a user in a domain
	GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error)
}