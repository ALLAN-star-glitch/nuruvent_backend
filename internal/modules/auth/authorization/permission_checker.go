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

// HasPermission checks if a user has a specific permission in a scope
func (c *PermissionChecker) HasPermission(ctx context.Context, userID string, scope authdomain.Scope, resource, action string) (bool, error) {
	domain := scope.Domain()
	if domain == "" {
		return false, fmt.Errorf("invalid scope: %s", scope.String())
	}

	allowed, err := c.enforcer.Enforce(userID, domain, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return allowed, nil
}

// HasAnyPermission checks if a user has any of the given permissions in a scope
func (c *PermissionChecker) HasAnyPermission(ctx context.Context, userID string, scope authdomain.Scope, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := c.HasPermission(ctx, userID, scope, resource, action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the given permissions in a scope
func (c *PermissionChecker) HasAllPermissions(ctx context.Context, userID string, scope authdomain.Scope, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := c.HasPermission(ctx, userID, scope, resource, action)
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
// CONVENIENCE METHODS - READ (EVENTS)
// ============================================================

// CanReadAllEvents checks if user can read ALL events in a scope
func (c *PermissionChecker) CanReadAllEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionReadAll.String())
}

// CanReadOwnEvents checks if user can read OWN events in a scope
func (c *PermissionChecker) CanReadOwnEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionReadOwn.String())
}

// ============================================================
// CONVENIENCE METHODS - CREATE (EVENTS)
// ============================================================

// CanCreateEvent checks if user can create events in a scope
func (c *PermissionChecker) CanCreateEvent(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionCreate.String())
}

// ============================================================
// CONVENIENCE METHODS - UPDATE (EVENTS)
// ============================================================

// CanUpdateAllEvents checks if user can update ALL events in a scope
func (c *PermissionChecker) CanUpdateAllEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionUpdateAll.String())
}


// CanUpdateOwnEvents checks if user can update OWN events in a scope
func (c *PermissionChecker) CanUpdateOwnEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionUpdateOwn.String())
}

// ============================================================
// CONVENIENCE METHODS - DELETE (EVENTS)
// ============================================================

// CanDeleteAllEvents checks if user can delete ALL events in a scope
func (c *PermissionChecker) CanDeleteAllEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionDeleteAll.String())
}

// CanDeleteOwnEvents checks if user can delete OWN events in a scope
func (c *PermissionChecker) CanDeleteOwnEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionDeleteOwn.String())
}

// ============================================================
// CONVENIENCE METHODS - PUBLISH (EVENTS)
// ============================================================

// CanPublishAllEvents checks if user can publish ALL events in a scope
func (c *PermissionChecker) CanPublishAllEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionPublishAll.String())
}

// CanPublishOwnEvents checks if user can publish OWN events in a scope
func (c *PermissionChecker) CanPublishOwnEvents(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceEvent.String(), authdomain.ActionPublishOwn.String())
}

// ============================================================
// CONVENIENCE METHODS - COMPOSITE (EVENTS)
// ============================================================

// CanManageEvent checks if user can manage events in a scope
func (c *PermissionChecker) CanManageEvent(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasAnyPermission(ctx, userID, scope, authdomain.ResourceEvent.String(),
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionDeleteAll.String(),
		authdomain.ActionManage.String(),
	)
}

// CanViewEvent checks if user can view events in a scope
func (c *PermissionChecker) CanViewEvent(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasAnyPermission(ctx, userID, scope, authdomain.ResourceEvent.String(),
		authdomain.ActionRead.String(),     
		authdomain.ActionReadAll.String(),
		authdomain.ActionReadOwn.String(),
	)
}

// CanViewCreator checks if user can view creator details
func (p *PermissionChecker) CanViewCreator(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return p.HasPermission(ctx, userID, scope, "event", "view_creator")
}

// ============================================================
// ✅ CONVENIENCE METHODS - PROFILE
// ============================================================

// CanReadAllProfiles checks if user can read ALL profiles in a scope
func (c *PermissionChecker) CanReadAllProfiles(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceProfile.String(), authdomain.ActionReadAll.String())
}

// CanReadOwnProfile checks if user can read OWN profile in a scope
func (c *PermissionChecker) CanReadOwnProfile(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceProfile.String(), authdomain.ActionReadOwn.String())
}

// CanUpdateAllProfiles checks if user can update ALL profiles in a scope
func (c *PermissionChecker) CanUpdateAllProfiles(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceProfile.String(), authdomain.ActionUpdateAll.String())
}

// CanUpdateOwnProfile checks if user can update OWN profile in a scope
func (c *PermissionChecker) CanUpdateOwnProfile(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasPermission(ctx, userID, scope, authdomain.ResourceProfile.String(), authdomain.ActionUpdateOwn.String())
}

// CanViewProfile checks if user can view profiles in a scope (read_all OR read_own)
func (c *PermissionChecker) CanViewProfile(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasAnyPermission(ctx, userID, scope, authdomain.ResourceProfile.String(),
		authdomain.ActionRead.String(),     // - exact read permission
		authdomain.ActionReadAll.String(),
		authdomain.ActionReadOwn.String(),
	)
}

// CanManageProfile checks if user can manage profiles in a scope (update_all)
func (c *PermissionChecker) CanManageProfile(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	return c.HasAnyPermission(ctx, userID, scope, authdomain.ResourceProfile.String(),
		authdomain.ActionUpdate.String(),     //  exact update permission
		authdomain.ActionUpdateAll.String(),
		authdomain.ActionUpdateOwn.String(),
	)
}

// ============================================================
// TEAM ROLE CHECKS
// ============================================================

// IsTeamAdmin checks if user is an admin in the scope
func (c *PermissionChecker) IsTeamAdmin(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	domain := scope.Domain()
	if domain == "" {
		return false, fmt.Errorf("invalid scope: %s", scope.String())
	}

	return c.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleAccountAdmin.String(), domain), nil
}

// IsEventManager checks if user is an event manager in the scope
func (c *PermissionChecker) IsEventManager(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	domain := scope.Domain()
	if domain == "" {
		return false, fmt.Errorf("invalid scope: %s", scope.String())
	}

	return c.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleEventManager.String(), domain), nil
}

// IsTeamMember checks if user is a team member in the scope
func (c *PermissionChecker) IsTeamMember(ctx context.Context, userID string, scope authdomain.Scope) (bool, error) {
	domain := scope.Domain()
	if domain == "" {
		return false, fmt.Errorf("invalid scope: %s", scope.String())
	}

	return c.enforcer.HasRoleForUserInDomain(userID, authdomain.RoleTeamMember.String(), domain), nil
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

// GetUserRoles returns all roles for a user in a scope
func (c *PermissionChecker) GetUserRoles(ctx context.Context, userID string, scope authdomain.Scope) ([]string, error) {
	domain := scope.Domain()
	if domain == "" {
		return nil, fmt.Errorf("invalid scope: %s", scope.String())
	}

	roles := c.enforcer.GetRolesForUserInDomain(userID, domain)
	return roles, nil
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