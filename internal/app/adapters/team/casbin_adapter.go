// internal/app/adapters/team/casbin_adapter.go

package team

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// CasbinAdapter adapts auth domain permission checker, role manager, and policy manager to team module's CasbinService
type CasbinAdapter struct {
	permChecker authDomain.PermissionChecker
	roleManager authDomain.RoleManager
	policyMgr   authDomain.PolicyManager
}


// NewCasbinAdapter creates a new team casbin adapter
func NewCasbinAdapter(
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
	policyMgr authDomain.PolicyManager,
) teamService.CasbinService {
	return &CasbinAdapter{
		permChecker: permChecker,
		roleManager: roleManager,
		policyMgr:   policyMgr,
	}
}

// AssignRole assigns a role to a user in a domain
func (a *CasbinAdapter) AssignRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.AssignRole(ctx, domain, userID, role)
}

// RemoveRole removes a role from a user in a domain
func (a *CasbinAdapter) RemoveRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.RemoveRole(ctx, domain, userID, role)
}

// GetUserRoles returns all roles for a user in a domain
func (a *CasbinAdapter) GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error) {
	return a.roleManager.GetUserRoles(ctx, userID, domain)
}

// HasPermission checks if a user has a specific permission in a domain
func (a *CasbinAdapter) HasPermission(ctx context.Context, domain string, userID, resource, action string) (bool, error) {
	return a.permChecker.HasPermission(ctx, userID, domain, resource, action)
}

// AddTeamPolicies adds all policies for a new team
func (a *CasbinAdapter) AddTeamPolicies(ctx context.Context, domain string) error {
	// Just pass the domain string directly to policy manager
	return a.policyMgr.AddTeamPolicies(ctx, domain)
}

// RemoveTeamPolicies removes all policies for a team
func (a *CasbinAdapter) RemoveTeamPolicies(ctx context.Context, domain string) error {
	return a.policyMgr.RemoveTeamPolicies(ctx, domain)
}

// AddAccountPolicies adds all policies for a new account
func (a *CasbinAdapter) AddAccountPolicies(ctx context.Context, domain string) error {
	return a.policyMgr.AddAccountPolicies(ctx, domain)
}

// RemoveAccountPolicies removes all policies for an account
func (a *CasbinAdapter) RemoveAccountPolicies(ctx context.Context, domain string) error {
	return a.policyMgr.RemoveAccountPolicies(ctx, domain)
}

// IsAccountAdmin checks if a user is an account admin in the domain
func (a *CasbinAdapter) IsAccountAdmin(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.IsAccountAdmin(ctx, userID, domain)
}

// IsTrainer checks if a user is a trainer in the domain
func (a *CasbinAdapter) IsTrainer(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.IsTrainer(ctx, userID, domain)
}