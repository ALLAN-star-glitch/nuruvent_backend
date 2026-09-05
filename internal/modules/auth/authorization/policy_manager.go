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

// AddTeamPolicies adds default policies for a team domain
func (m *PolicyManager) AddTeamPolicies(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	var policies [][]string
	if authdomain.IsPersonalTeamDomain(domain) {
		log.Printf("Adding personal team policies for domain: %s", domain)
		policies = GetPersonalTeamPolicies(domain)
	} else if authdomain.IsInstitutionTeamDomain(domain) {
		log.Printf("Adding institution team policies for domain: %s", domain)
		policies = GetInstitutionTeamPolicies(domain)
	} else {
		return fmt.Errorf("invalid team domain: %s", domain)
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

// RemoveTeamPolicies removes all policies for a team domain
func (m *PolicyManager) RemoveTeamPolicies(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
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

// ============================================================
// ACCOUNT POLICY MANAGEMENT (NEW)
// ============================================================

// AddAccountPolicies adds default policies for an account domain
func (m *PolicyManager) AddAccountPolicies(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	if !authdomain.IsAccountDomain(domain) {
		return fmt.Errorf("invalid account domain: %s", domain)
	}

	log.Printf("Adding account policies for domain: %s", domain)

	policies := GetAccountPolicies(domain)
	_, err := m.enforcer.AddPolicies(policies)
	if err != nil {
		return fmt.Errorf("failed to add account policies: %w", err)
	}

	// Add role hierarchy for account
	hierarchy := GetAccountRoleHierarchy(domain)
	_, err = m.enforcer.AddGroupingPolicies(hierarchy)
	if err != nil {
		return fmt.Errorf("failed to add account role hierarchy: %w", err)
	}

	log.Printf("✅ Added %d account policies and %d hierarchy rules for domain: %s",
		len(policies), len(hierarchy), domain)
	return nil
}

// RemoveAccountPolicies removes all policies for an account domain
func (m *PolicyManager) RemoveAccountPolicies(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	if !authdomain.IsAccountDomain(domain) {
		return fmt.Errorf("invalid account domain: %s", domain)
	}

	log.Printf("Removing account policies for domain: %s", domain)

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

	log.Printf("✅ Removed %d account policies and %d grouping policies for domain: %s",
		len(policies), len(groupingPolicies), domain)
	return nil
}

// ============================================================
// PLATFORM POLICY MANAGEMENT
// ============================================================

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

// ensureTeamPoliciesExist ensures policies exist for a team domain
func (m *PolicyManager) ensureTeamPoliciesExist(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	policies, err := m.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	if len(policies) == 0 {
		log.Printf("No policies found for domain %s, adding default policies", domain)
		return m.AddTeamPolicies(ctx, domain)
	}

	return nil
}

// ensureAccountPoliciesExist ensures policies exist for an account domain
func (m *PolicyManager) ensureAccountPoliciesExist(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	if !authdomain.IsAccountDomain(domain) {
		return fmt.Errorf("invalid account domain: %s", domain)
	}

	policies, err := m.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	if len(policies) == 0 {
		log.Printf("No policies found for account domain %s, adding default policies", domain)
		return m.AddAccountPolicies(ctx, domain)
	}

	return nil
}