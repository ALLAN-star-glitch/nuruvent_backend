// internal/modules/auth/authorization/role_manager.go

package authorization

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// RoleManager implements authdomain.RoleManager
type RoleManager struct {
	enforcer *Enforcer
}

// NewRoleManager creates a new role manager
func NewRoleManager(enforcer *Enforcer) authdomain.RoleManager {
	return &RoleManager{enforcer: enforcer}
}

// ============================================================
// GENERIC ROLE OPERATIONS
// ============================================================

// AssignRole assigns a role to a user in a scope
func (m *RoleManager) AssignRole(ctx context.Context, scope authdomain.Scope, userID string, role string) error {
	domain := scope.Domain()
	if domain == "" {
		return fmt.Errorf("invalid scope: %s", scope.String())
	}

	log.Printf("Assigning role %s to user %s in domain %s", role, userID, domain)

	_, err := m.enforcer.AddRoleForUserInDomain(userID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	log.Printf("✅ Assigned role %s to user %s in domain %s", role, userID, domain)
	return nil
}

// RemoveRole removes a role from a user in a scope
func (m *RoleManager) RemoveRole(ctx context.Context, scope authdomain.Scope, userID string, role string) error {
	domain := scope.Domain()
	if domain == "" {
		return fmt.Errorf("invalid scope: %s", scope.String())
	}

	log.Printf("Removing role %s from user %s in domain %s", role, userID, domain)

	_, err := m.enforcer.RemoveRoleForUserInDomain(userID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	log.Printf("✅ Removed role %s from user %s in domain %s", role, userID, domain)
	return nil
}

// RemoveAllRoles removes all roles from a user in a scope
func (m *RoleManager) RemoveAllRoles(ctx context.Context, scope authdomain.Scope, userID string) error {
	domain := scope.Domain()
	if domain == "" {
		return fmt.Errorf("invalid scope: %s", scope.String())
	}

	log.Printf("Removing all roles for user %s in domain %s", userID, domain)

	_, err := m.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
	if err != nil {
		return fmt.Errorf("failed to remove all roles: %w", err)
	}

	log.Printf("✅ Removed all roles for user %s in domain %s", userID, domain)
	return nil
}

// GetUserRoles returns all roles for a user in a scope
func (m *RoleManager) GetUserRoles(ctx context.Context, userID string, scope authdomain.Scope) ([]string, error) {
	domain := scope.Domain()
	if domain == "" {
		return nil, fmt.Errorf("invalid scope: %s", scope.String())
	}

	roles := m.enforcer.GetRolesForUserInDomain(userID, domain)
	return roles, nil
}

// ============================================================
// CONVENIENCE METHODS
// ============================================================

// AssignPersonalTeamAdmin assigns account_admin to a user's personal team
func (m *RoleManager) AssignPersonalTeamAdmin(ctx context.Context, userID string) error {
	scope := authdomain.NewPersonalTeamScope(userID)
	return m.AssignRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String())
}

// AssignInstitutionAdmin assigns account_admin in an institution team
func (m *RoleManager) AssignInstitutionAdmin(ctx context.Context, institutionID, userID string) error {
	scope := authdomain.NewInstitutionTeamScope(institutionID)
	return m.AssignRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String())
}

// AssignEventManager assigns event_manager in a scope
func (m *RoleManager) AssignEventManager(ctx context.Context, scope authdomain.Scope, userID string) error {
	return m.AssignRole(ctx, scope, userID, authdomain.RoleEventManager.String())
}

// AssignTeamMember assigns team_member in a scope
func (m *RoleManager) AssignTeamMember(ctx context.Context, scope authdomain.Scope, userID string) error {
	return m.AssignRole(ctx, scope, userID, authdomain.RoleTeamMember.String())
}

// RemovePersonalTeamAdmin removes account_admin from a user's personal team
func (m *RoleManager) RemovePersonalTeamAdmin(ctx context.Context, userID string) error {
	scope := authdomain.NewPersonalTeamScope(userID)
	return m.RemoveRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String())
}

// RemoveInstitutionAdmin removes account_admin from an institution team
func (m *RoleManager) RemoveInstitutionAdmin(ctx context.Context, institutionID, userID string) error {
	scope := authdomain.NewInstitutionTeamScope(institutionID)
	return m.RemoveRole(ctx, scope, userID, authdomain.RoleAccountAdmin.String())
}