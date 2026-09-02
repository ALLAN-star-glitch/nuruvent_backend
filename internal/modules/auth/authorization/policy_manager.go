// internal/modules/auth/authorization/policy_manager.go

package authorization

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// PolicyManager implements authdomain.PolicyManager
type PolicyManager struct {
	enforcer *Enforcer
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(enforcer *Enforcer) authdomain.PolicyManager {
	return &PolicyManager{enforcer: enforcer}
}

// ============================================================
// TEAM POLICY MANAGEMENT
// ============================================================

// AddTeamPolicies adds default policies for a team scope
func (m *PolicyManager) AddTeamPolicies(ctx context.Context, scope authdomain.Scope) error {
	domain := scope.Domain()
	if domain == "" {
		return fmt.Errorf("invalid scope: %s", scope.String())
	}

	var policies [][]string
	if scope.IsPersonalTeam() {
		log.Printf("Adding personal team policies for domain: %s", domain)
		policies = GetPersonalTeamPolicies(domain)
	} else if scope.IsInstitutionTeam() {
		log.Printf("Adding institution team policies for domain: %s", domain)
		policies = GetInstitutionTeamPolicies(domain)
	} else {
		return fmt.Errorf("invalid team scope: %s", scope.String())
	}

	// Add policies
	_, err := m.enforcer.AddPolicies(policies)
	if err != nil {
		return fmt.Errorf("failed to add team policies: %w", err)
	}

	// Add role hierarchy
	hierarchy := GetTeamRoleHierarchy(domain)
	_, err = m.enforcer.AddGroupingPolicies(hierarchy)
	if err != nil {
		return fmt.Errorf("failed to add role hierarchy: %w", err)
	}

	log.Printf("✅ Added %d policies and %d hierarchy rules for domain: %s",
		len(policies), len(hierarchy), domain)
	return nil
}

// RemoveTeamPolicies removes all policies for a team scope
func (m *PolicyManager) RemoveTeamPolicies(ctx context.Context, scope authdomain.Scope) error {
	domain := scope.Domain()
	if domain == "" {
		return fmt.Errorf("invalid scope: %s", scope.String())
	}

	log.Printf("Removing team policies for domain: %s", domain)

	// Remove policy rules
	policies, err := m.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to get policies: %w", err)
	}

	if len(policies) > 0 {
		_, err := m.enforcer.RemovePolicies(policies)
		if err != nil {
			return fmt.Errorf("failed to remove policies: %w", err)
		}
	}

	// Remove grouping policies
	groupingPolicies, err := m.enforcer.GetFilteredGroupingPolicy(2, domain)
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	if len(groupingPolicies) > 0 {
		_, err := m.enforcer.RemoveGroupingPolicies(groupingPolicies)
		if err != nil {
			return fmt.Errorf("failed to remove grouping policies: %w", err)
		}
	}

	log.Printf("✅ Removed %d policies and %d grouping policies for domain: %s",
		len(policies), len(groupingPolicies), domain)
	return nil
}

// AddPlatformPolicies adds default platform policies
func (m *PolicyManager) AddPlatformPolicies(ctx context.Context) error {
	log.Println("Adding platform policies")

	platformPolicies := GetPlatformPolicies()
	_, err := m.enforcer.AddPolicies(platformPolicies)
	if err != nil {
		return fmt.Errorf("failed to add platform policies: %w", err)
	}

	hierarchy := GetPlatformRoleHierarchy()
	_, err = m.enforcer.AddGroupingPolicies(hierarchy)
	if err != nil {
		return fmt.Errorf("failed to add platform role hierarchy: %w", err)
	}

	log.Printf("✅ Added %d platform policies and %d hierarchy rules",
		len(platformPolicies), len(hierarchy))
	return nil
}

// RemovePlatformPolicies removes all platform policies
func (m *PolicyManager) RemovePlatformPolicies(ctx context.Context) error {
	log.Println("Removing platform policies")

	// Remove policy rules
	policies, err := m.enforcer.GetFilteredPolicy(1, authdomain.DomainPlatform)
	if err != nil {
		return fmt.Errorf("failed to get platform policies: %w", err)
	}

	if len(policies) > 0 {
		_, err := m.enforcer.RemovePolicies(policies)
		if err != nil {
			return fmt.Errorf("failed to remove platform policies: %w", err)
		}
	}

	// Remove grouping policies
	groupingPolicies, err := m.enforcer.GetFilteredGroupingPolicy(2, authdomain.DomainPlatform)
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	if len(groupingPolicies) > 0 {
		_, err := m.enforcer.RemoveGroupingPolicies(groupingPolicies)
		if err != nil {
			return fmt.Errorf("failed to remove grouping policies: %w", err)
		}
	}

	log.Printf("✅ Removed %d platform policies and %d grouping policies",
		len(policies), len(groupingPolicies))
	return nil
}

// ============================================================
// HELPER METHODS (Internal)
// ============================================================

// ensureTeamPoliciesExist ensures policies exist for a team scope
func (m *PolicyManager) ensureTeamPoliciesExist(ctx context.Context, scope authdomain.Scope) error {
	domain := scope.Domain()
	if domain == "" {
		return fmt.Errorf("invalid scope: %s", scope.String())
	}

	policies, err := m.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	if len(policies) == 0 {
		log.Printf("No policies found for domain %s, adding default policies", domain)
		return m.AddTeamPolicies(ctx, scope)
	}

	return nil
}