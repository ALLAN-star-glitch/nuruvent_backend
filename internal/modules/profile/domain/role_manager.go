// internal/modules/profile/domain/role_manager.go

package domain

import "context"

// ============================================================
// ROLE MANAGER - Outbound Port for Profile Module
// ============================================================

// RoleManager defines the role management operations that the profile module requires
type RoleManager interface {
    // AssignRole assigns a role to a user in a domain
    // domain: "personal:team:{user_id}", "institution:team:{institution_id}", "account:{account_id}", or "platform"
    AssignRole(ctx context.Context, domain string, userID string, role string) error
    
    // RemoveRole removes a role from a user in a domain
    RemoveRole(ctx context.Context, domain string, userID string, role string) error
    
    // GetUserRoles returns all roles for a user in a domain
    GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error)
}