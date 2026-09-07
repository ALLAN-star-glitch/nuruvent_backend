// internal/modules/auth/authorization/permission_checker.go

package authorization

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// PermissionChecker implements authdomain.PermissionChecker
type PermissionChecker struct {
	enforcer *Enforcer
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(enforcer *Enforcer) authdomain.PermissionChecker {
	return &PermissionChecker{enforcer: enforcer}
}

// ============================================================
// CORE PERMISSION CHECKS
// ============================================================

func (c *PermissionChecker) HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error) {
	if domain == "" {
		return false, fmt.Errorf("invalid domain: empty string")
	}
	return c.enforcer.Enforce(userID, domain, resource, action)
}

func (c *PermissionChecker) HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := c.HasPermission(ctx, userID, domain, resource, action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

func (c *PermissionChecker) HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := c.HasPermission(ctx, userID, domain, resource, action)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// ============================================================
// PROFILE PERMISSIONS
// ============================================================

func (c *PermissionChecker) CanReadAllProfiles(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionReadAll.String())
}

func (c *PermissionChecker) CanReadOwnProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionReadOwn.String())
}

func (c *PermissionChecker) CanReadProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceProfile.String(),
		authdomain.ActionRead.String(),
		authdomain.ActionReadAll.String(),
		authdomain.ActionReadOwn.String(),
	)
}

func (c *PermissionChecker) CanUpdateAllProfiles(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdateAll.String())
}

func (c *PermissionChecker) CanUpdateOwnProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdateOwn.String())
}

func (c *PermissionChecker) CanUpdateProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceProfile.String(),
		authdomain.ActionUpdate.String(),
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionUpdateOwn.String(),
	)
}

func (c *PermissionChecker) CanManageProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceProfile.String(),
		authdomain.ActionUpdate.String(),
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionManage.String(),
	)
}

func (c *PermissionChecker) CanViewProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.CanReadProfile(ctx, userID, domain)
}

// ============================================================
// ACCOUNT ROLE CHECKS
// ============================================================

func (c *PermissionChecker) IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error) {
	if domain == "" {
		return false, fmt.Errorf("invalid domain: empty string")
	}
	// Extract account ID from domain
	accountID := authdomain.ExtractAccountIDFromDomain(domain)
	if accountID == "" {
		return false, fmt.Errorf("invalid domain format: %s", domain)
	}
	return c.enforcer.IsAccountAdmin(userID, accountID), nil
}

func (c *PermissionChecker) IsTeamAdmin(ctx context.Context, userID string, domain string) (bool, error) {
	return c.IsAccountAdmin(ctx, userID, domain)
}

func (c *PermissionChecker) IsTrainer(ctx context.Context, userID string, domain string) (bool, error) {
	if domain == "" {
		return false, fmt.Errorf("invalid domain: empty string")
	}
	accountID := authdomain.ExtractAccountIDFromDomain(domain)
	if accountID == "" {
		return false, fmt.Errorf("invalid domain format: %s", domain)
	}
	return c.enforcer.IsAccountTrainer(userID, accountID), nil
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

func (c *PermissionChecker) GetUserTeamDomains(ctx context.Context, userID string) ([]string, error) {
	// Team domains are the account domains since roles are at account level
	return c.enforcer.GetUserAccountDomains(userID), nil
}

func (c *PermissionChecker) GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error) {
	// Personal teams don't exist as separate domains - roles are at account level
	// Return empty slice
	return []string{}, nil
}

func (c *PermissionChecker) GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error) {
	// Institution teams are accounts
	return c.enforcer.GetUserAccountIDs(userID), nil
}

func (c *PermissionChecker) GetUserAccountIDs(ctx context.Context, userID string) ([]string, error) {
	return c.enforcer.GetUserAccountIDs(userID), nil
}

func (c *PermissionChecker) GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error) {
	if domain == "" {
		return nil, fmt.Errorf("invalid domain: empty string")
	}
	return c.enforcer.GetRolesForUserInDomain(userID, domain), nil
}

func (c *PermissionChecker) HasTeamAccess(ctx context.Context, userID string) (bool, error) {
	return c.enforcer.HasAnyAccountRole(userID), nil
}

// ============================================================
// ACCOUNT CONVENIENCE METHODS
// ============================================================

func (c *PermissionChecker) CanManageAccount(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceAccount.String(),
		authdomain.ActionUpdate.String(),
		authdomain.ActionDelete.String(),
		authdomain.ActionManage.String(),
	)
}

func (c *PermissionChecker) CanManageAccountMembers(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceMember.String(),
		authdomain.ActionMemberAdd.String(),
		authdomain.ActionMemberRemove.String(),
	)
}

func (c *PermissionChecker) CanViewAccount(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String())
}

// ============================================================
// EVENT CONVENIENCE METHODS
// ============================================================

func (c *PermissionChecker) CanReadAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String())
}

func (c *PermissionChecker) CanReadOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadOwn.String())
}

func (c *PermissionChecker) CanCreateEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String())
}

func (c *PermissionChecker) CanUpdateAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String())
}

func (c *PermissionChecker) CanUpdateOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateOwn.String())
}

func (c *PermissionChecker) CanDeleteAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String())
}

func (c *PermissionChecker) CanDeleteOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteOwn.String())
}

func (c *PermissionChecker) CanPublishAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String())
}

func (c *PermissionChecker) CanPublishOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishOwn.String())
}

func (c *PermissionChecker) CanViewCreator(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionViewCreator.String())
}

func (c *PermissionChecker) CanManageEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceEvent.String(),
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionDeleteAll.String(),
		authdomain.ActionManage.String(),
	)
}

func (c *PermissionChecker) CanViewEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceEvent.String(),
		authdomain.ActionRead.String(),
		authdomain.ActionReadAll.String(),
		authdomain.ActionReadOwn.String(),
	)
}