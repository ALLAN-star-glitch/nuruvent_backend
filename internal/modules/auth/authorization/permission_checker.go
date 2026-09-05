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

// HasPermission checks if a user has a specific permission in a domain
func (c *PermissionChecker) HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error) {
	if domain == "" {
		return false, fmt.Errorf("invalid domain: empty string")
	}

	allowed, err := c.enforcer.Enforce(userID, domain, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return allowed, nil
}

// HasAnyPermission checks if a user has any of the given permissions in a domain
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

// HasAllPermissions checks if a user has all of the given permissions in a domain
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
// ACCOUNT CONVENIENCE METHODS
// ============================================================

// CanManageAccount checks if user can manage an account
func (c *PermissionChecker) CanManageAccount(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceAccount.String(),
		authdomain.ActionUpdate.String(),
		authdomain.ActionDelete.String(),
		authdomain.ActionManage.String(),
	)
}

// CanManageAccountMembers checks if user can manage account members
func (c *PermissionChecker) CanManageAccountMembers(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceMember.String(),
		authdomain.ActionMemberAdd.String(),
		authdomain.ActionMemberRemove.String(),
	)
}

// CanViewAccount checks if user can view an account
func (c *PermissionChecker) CanViewAccount(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceAccount.String(), authdomain.ActionRead.String())
}

// ============================================================
// EVENT CONVENIENCE METHODS
// ============================================================

// CanReadAllEvents checks if user can read ALL events in a domain
func (c *PermissionChecker) CanReadAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String())
}

// CanReadOwnEvents checks if user can read OWN events in a domain
func (c *PermissionChecker) CanReadOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionReadOwn.String())
}

// CanCreateEvent checks if user can create events in a domain
func (c *PermissionChecker) CanCreateEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String())
}

// CanUpdateAllEvents checks if user can update ALL events in a domain
func (c *PermissionChecker) CanUpdateAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String())
}

// CanUpdateOwnEvents checks if user can update OWN events in a domain
func (c *PermissionChecker) CanUpdateOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionUpdateOwn.String())
}

// CanDeleteAllEvents checks if user can delete ALL events in a domain
func (c *PermissionChecker) CanDeleteAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String())
}

// CanDeleteOwnEvents checks if user can delete OWN events in a domain
func (c *PermissionChecker) CanDeleteOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionDeleteOwn.String())
}

// CanPublishAllEvents checks if user can publish ALL events in a domain
func (c *PermissionChecker) CanPublishAllEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String())
}

// CanPublishOwnEvents checks if user can publish OWN events in a domain
func (c *PermissionChecker) CanPublishOwnEvents(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionPublishOwn.String())
}

// CanViewCreator checks if user can view creator details
func (c *PermissionChecker) CanViewCreator(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceEvent.String(), authdomain.ActionViewCreator.String())
}

// CanManageEvent checks if user can manage events in a domain
func (c *PermissionChecker) CanManageEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceEvent.String(),
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionDeleteAll.String(),
		authdomain.ActionManage.String(),
	)
}

// CanViewEvent checks if user can view events in a domain
func (c *PermissionChecker) CanViewEvent(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceEvent.String(),
		authdomain.ActionRead.String(),
		authdomain.ActionReadAll.String(),
		authdomain.ActionReadOwn.String(),
	)
}

// ============================================================
// PROFILE CONVENIENCE METHODS
// ============================================================

// CanReadAllProfiles checks if user can read ALL profiles in a domain
func (c *PermissionChecker) CanReadAllProfiles(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionReadAll.String())
}

// CanReadOwnProfile checks if user can read OWN profile in a domain
func (c *PermissionChecker) CanReadOwnProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionReadOwn.String())
}

// CanUpdateAllProfiles checks if user can update ALL profiles in a domain
func (c *PermissionChecker) CanUpdateAllProfiles(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdateAll.String())
}

// CanUpdateOwnProfile checks if user can update OWN profile in a domain
func (c *PermissionChecker) CanUpdateOwnProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasPermission(ctx, userID, domain, authdomain.ResourceProfile.String(), authdomain.ActionUpdateOwn.String())
}

// CanViewProfile checks if user can view profiles in a domain (read_all OR read_own)
func (c *PermissionChecker) CanViewProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceProfile.String(),
		authdomain.ActionRead.String(),
		authdomain.ActionReadAll.String(),
		authdomain.ActionReadOwn.String(),
	)
}

// CanManageProfile checks if user can manage profiles in a domain (update_all)
func (c *PermissionChecker) CanManageProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return c.HasAnyPermission(ctx, userID, domain, authdomain.ResourceProfile.String(),
		authdomain.ActionUpdate.String(),
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionUpdateOwn.String(),
	)
}

// ============================================================
// ROLE CHECKS
// ============================================================

// IsAccountAdmin checks if user is an account admin in the domain
func (c *PermissionChecker) IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error) {
	if domain == "" {
		return false, fmt.Errorf("invalid domain: empty string")
	}
	return c.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain), nil
}

// IsTrainer checks if user is a trainer in the domain
func (c *PermissionChecker) IsTrainer(ctx context.Context, userID string, domain string) (bool, error) {
	if domain == "" {
		return false, fmt.Errorf("invalid domain: empty string")
	}
	return c.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleTrainer.String(), domain), nil
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

// GetUserRoles returns all roles for a user in a domain
func (c *PermissionChecker) GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error) {
	if domain == "" {
		return nil, fmt.Errorf("invalid domain: empty string")
	}
	return c.enforcer.GetRolesForUserInDomain(userID, domain), nil
}

// GetUserTeamIDs returns all team IDs where a user has roles
func (c *PermissionChecker) GetUserTeamIDs(ctx context.Context, userID string) ([]string, error) {
	return c.enforcer.GetUserTeamIDs(userID), nil
}

// GetUserPersonalTeamIDs returns personal team IDs where a user has roles
func (c *PermissionChecker) GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error) {
	return c.enforcer.GetUserPersonalTeamIDs(userID), nil
}

// GetUserInstitutionTeamIDs returns institution team IDs where a user has roles
func (c *PermissionChecker) GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error) {
	return c.enforcer.GetUserInstitutionTeamIDs(userID), nil
}

// HasTeamAccess checks if a user has any team role
func (c *PermissionChecker) HasTeamAccess(ctx context.Context, userID string) (bool, error) {
	return c.enforcer.HasAnyTeamRole(userID), nil
}