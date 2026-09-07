// internal/app/adapters/profile/role_manager_adapter.go

package profile

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	profileDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// RoleManagerAdapter adapts auth domain's RoleManager to profile module's RoleManager port
type RoleManagerAdapter struct {
	roleMgr authdomain.RoleManager
}

// NewRoleManagerAdapter creates a new role manager adapter for the profile module
func NewRoleManagerAdapter(roleMgr authdomain.RoleManager) profileDomain.RoleManager {
	return &RoleManagerAdapter{
		roleMgr: roleMgr,
	}
}

// AssignRole assigns a role to a user in a domain
func (a *RoleManagerAdapter) AssignRole(ctx context.Context, domain string, userID string, role string) error {
	return a.roleMgr.AssignRole(ctx, domain, userID, role)
}

// RemoveRole removes a role from a user in a domain
func (a *RoleManagerAdapter) RemoveRole(ctx context.Context, domain string, userID string, role string) error {
	return a.roleMgr.RemoveRole(ctx, domain, userID, role)
}

// GetUserRoles returns all roles for a user in a domain
func (a *RoleManagerAdapter) GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error) {
	return a.roleMgr.GetUserRoles(ctx, userID, domain)
}