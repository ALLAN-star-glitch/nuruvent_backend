// internal/modules/profile/domain/role_manager.go

package domain

import "context"

// ============================================================
// ROLE MANAGER - Outbound Port for Profile Module
// ============================================================

// RoleManager defines the role management operations that the profile module requires
type RoleManager interface {
    // AssignRole assigns a role to a user in a scope
    AssignRole(ctx context.Context, scope Scope, userID string, role string) error
    
    // RemoveRole removes a role from a user in a scope
    RemoveRole(ctx context.Context, scope Scope, userID string, role string) error
    
    // GetUserRoles returns all roles for a user in a scope
    GetUserRoles(ctx context.Context, userID string, scope Scope) ([]string, error)
}