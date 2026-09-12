// internal/app/adapters/team/casbin_adapter.go

package team

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// CasbinAdapter adapts the auth module's permission and role managers to
// the team module's CasbinService port.
//
// The team service talks in high-level permission terms
// (CanRemoveMember, CanCreateTeam) rather than raw Casbin terms
// (HasPermission with resource+action). Each method here maps one
// high-level concept to a call on the auth module's PermissionChecker.
//
// Domain: always pass the team's PARENT ACCOUNT domain
// ("account:<team.account_id>"). Teams are not authz domains.
type CasbinAdapter struct {
	permChecker authDomain.PermissionChecker
	roleManager authDomain.RoleManager
}

// NewCasbinAdapter creates a new team casbin adapter.
func NewCasbinAdapter(
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
) teamService.CasbinService {
	return &CasbinAdapter{
		permChecker: permChecker,
		roleManager: roleManager,
	}
}

// ============================================================
// ROLE (GROUPING) MANAGEMENT
// ============================================================

func (a *CasbinAdapter) AssignRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.AssignRole(ctx, domain, userID, role)
}

func (a *CasbinAdapter) RemoveRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.RemoveRole(ctx, domain, userID, role)
}

func (a *CasbinAdapter) GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error) {
	return a.roleManager.GetUserRoles(ctx, userID, domain)
}

// ============================================================
// TEAM PERMISSION CHECKS
// ============================================================

func (a *CasbinAdapter) CanViewTeams(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanViewTeams(ctx, userID, domain)
}

func (a *CasbinAdapter) CanCreateTeam(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanCreateTeam(ctx, userID, domain)
}


func (a *CasbinAdapter) CanManageTeam(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanManageTeam(ctx, userID, domain)
}


func (a *CasbinAdapter) CanDeleteTeam(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanDeleteTeam(ctx, userID, domain)
}

// ============================================================
// MEMBER PERMISSION CHECKS
// ============================================================

func (a *CasbinAdapter) CanViewMembers(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanViewMembers(ctx, userID, domain)
}

func (a *CasbinAdapter) CanManageMembers(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanManageMembers(ctx, userID, domain)
}

func (a *CasbinAdapter) CanAddMember(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanAddMember(ctx, userID, domain)
}

func (a *CasbinAdapter) CanRemoveMember(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanRemoveMember(ctx, userID, domain)
}

func (a *CasbinAdapter) CanInviteMember(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanInviteMember(ctx, userID, domain)
}

func (a *CasbinAdapter) CanLeaveAccount(ctx context.Context, userID, domain string) (bool, error) {
	return a.permChecker.CanLeaveAccount(ctx, userID, domain)
}