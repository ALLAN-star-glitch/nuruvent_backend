// internal/modules/auth/authorization/policy_manager.go

package authorization

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// PolicyManager implements authdomain.PolicyManager.
//
// POST-REVAMP SCOPE:
//
// Policies (p rules) are static. They are seeded once via GetAccountPolicies()
// and GetPlatformPolicies() with wildcard domains ("account:*"). No per-account
// policy materialization happens at runtime.
//
// This manager handles:
//
//   - Platform policy seeding (one-time, at bootstrap)
//   - Account cleanup when an account is deleted (removes the account's g rules)
//
// It does NOT:
//
//   - Seed policies for new accounts (the wildcard covers them)
//   - Seed policies for teams (teams are not authz domains)
//   - Manage team membership (that's data, not policy)
//
// User role assignments (g rules) are managed by the RoleManager, not here.
type PolicyManager struct {
	enforcer *Enforcer
}

// NewPolicyManager creates a new policy manager.
func NewPolicyManager(enforcer *Enforcer) authdomain.PolicyManager {
	return &PolicyManager{enforcer: enforcer}
}

// ============================================================
// PLATFORM POLICY MANAGEMENT
// ============================================================

// AddPlatformPolicies seeds the static platform policies. Runs once at
// system bootstrap. Safe to call more than once — Casbin's AddPolicies
// is idempotent for identical rules.
func (m *PolicyManager) AddPlatformPolicies(ctx context.Context) error {
	log.Println("Adding platform policies")

	platformPolicies := GetPlatformPolicies()
	if _, err := m.enforcer.AddPolicies(platformPolicies); err != nil {
		return fmt.Errorf("failed to add platform policies: %w", err)
	}

	hierarchy := GetPlatformRoleHierarchy()
	if _, err := m.enforcer.AddGroupingPolicies(hierarchy); err != nil {
		return fmt.Errorf("failed to add platform role hierarchy: %w", err)
	}

	log.Printf("✅ Added %d platform policies and %d hierarchy rules",
		len(platformPolicies), len(hierarchy))
	return nil
}

// RemovePlatformPolicies removes all platform-domain policies and grouping
// rules. Used only for full resets, not in normal operation.
func (m *PolicyManager) RemovePlatformPolicies(ctx context.Context) error {
	log.Println("Removing platform policies")

	policies, err := m.enforcer.GetFilteredPolicy(1, authdomain.DomainPlatform)
	if err != nil {
		return fmt.Errorf("failed to get platform policies: %w", err)
	}

	if len(policies) > 0 {
		if _, err := m.enforcer.RemovePolicies(policies); err != nil {
			return fmt.Errorf("failed to remove platform policies: %w", err)
		}
	}

	groupingPolicies, err := m.enforcer.GetFilteredGroupingPolicy(2, authdomain.DomainPlatform)
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	if len(groupingPolicies) > 0 {
		if _, err := m.enforcer.RemoveGroupingPolicies(groupingPolicies); err != nil {
			return fmt.Errorf("failed to remove grouping policies: %w", err)
		}
	}

	log.Printf("✅ Removed %d platform policies and %d grouping policies",
		len(policies), len(groupingPolicies))
	return nil
}

// ============================================================
// ACCOUNT POLICY MANAGEMENT
// ============================================================

// RemoveAccountPolicies removes all Casbin rules tied to a specific account
// domain. Call this when an account is deleted.
//
// NOTE: Static policies use the wildcard "account:*", so they are NOT
// per-account. Only the account's grouping rules (user → role bindings)
// need to be removed.
func (m *PolicyManager) RemoveAccountPolicies(ctx context.Context, domain string) error {
	if domain == "" {
		return fmt.Errorf("invalid domain: empty string")
	}

	if !authdomain.IsAccountDomain(domain) {
		return fmt.Errorf("invalid account domain: %s", domain)
	}

	log.Printf("Removing account grouping policies for domain: %s", domain)

	// Remove user-to-role g rules for this concrete account domain.
	// Do NOT touch the wildcard "account:*" policies.
	groupingPolicies, err := m.enforcer.GetFilteredGroupingPolicy(2, domain)
	if err != nil {
		return fmt.Errorf("failed to get grouping policies: %w", err)
	}

	if len(groupingPolicies) > 0 {
		if _, err := m.enforcer.RemoveGroupingPolicies(groupingPolicies); err != nil {
			return fmt.Errorf("failed to remove grouping policies: %w", err)
		}
	}

	log.Printf("✅ Removed %d grouping policies for account domain: %s",
		len(groupingPolicies), domain)
	return nil
}