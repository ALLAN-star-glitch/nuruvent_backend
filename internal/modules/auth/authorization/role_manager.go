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

// AssignRole assigns a role to a user in a domain
func (m *RoleManager) AssignRole(ctx context.Context, domain string, userID string, role string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	log.Printf("Assigning role %s to user %s in domain %s", role, userID, domain)

	_, err := m.enforcer.AddRoleForUserInDomain(userID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	log.Printf("✅ Assigned role %s to user %s in domain %s", role, userID, domain)
	return nil
}

// RemoveRole removes a role from a user in a domain
func (m *RoleManager) RemoveRole(ctx context.Context, domain string, userID string, role string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	log.Printf("Removing role %s from user %s in domain %s", role, userID, domain)

	_, err := m.enforcer.RemoveRoleForUserInDomain(userID, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	log.Printf("✅ Removed role %s from user %s in domain %s", role, userID, domain)
	return nil
}

// RemoveAllRoles removes all roles from a user in a domain
func (m *RoleManager) RemoveAllRoles(ctx context.Context, domain string, userID string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	log.Printf("Removing all roles for user %s in domain %s", userID, domain)

	_, err := m.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", domain)
	if err != nil {
		return fmt.Errorf("failed to remove all roles: %w", err)
	}

	log.Printf("✅ Removed all roles for user %s in domain %s", userID, domain)
	return nil
}

// GetUserRoles returns all roles for a user in a domain
func (m *RoleManager) GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error) {
	if domain == "" {
		return nil, fmt.Errorf("invalid domain: empty string")
	}

	roles := m.enforcer.GetRolesForUserInDomain(userID, domain)
	return roles, nil
}

// ============================================================
// ACCOUNT-LEVEL CONVENIENCE METHODS (Only)
// ============================================================

// AssignAccountAdmin assigns account_admin to a user in an account domain
// domain: account:{account_id}
func (m *RoleManager) AssignAccountAdmin(ctx context.Context, domain string, userID string) error {
	return m.AssignRole(ctx, domain, userID, authdomain.RoleAccountAdmin.String())
}

// AssignTrainer assigns trainer to a user in an account domain
// domain: account:{account_id}
func (m *RoleManager) AssignTrainer(ctx context.Context, domain string, userID string) error {
	return m.AssignRole(ctx, domain, userID, authdomain.RoleTrainer.String())
}

// RemoveAccountAdmin removes account_admin from a user in an account domain
// domain: account:{account_id}
func (m *RoleManager) RemoveAccountAdmin(ctx context.Context, domain string, userID string) error {
	return m.RemoveRole(ctx, domain, userID, authdomain.RoleAccountAdmin.String())
}

// RemoveTrainer removes trainer from a user in an account domain
// domain: account:{account_id}
func (m *RoleManager) RemoveTrainer(ctx context.Context, domain string, userID string) error {
	return m.RemoveRole(ctx, domain, userID, authdomain.RoleTrainer.String())
}