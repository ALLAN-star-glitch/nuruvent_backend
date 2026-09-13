// internal/modules/auth/authdomain/policy_manager.go

package authdomain

import "context"

// PolicyManager handles policy lifecycle.
//
// ============================================================
// SCOPE (post-revamp)
// ============================================================
//
// Policies are static. The account-domain policy set uses the wildcard
// "account:*" and is seeded once at bootstrap. Individual accounts and
// teams do NOT get their own policies materialized at runtime.
//
// This interface therefore covers:
//
//   - Platform policy seeding (bootstrap)
//   - Platform policy removal (reset only)
//   - Account cleanup (removal of a deleted account's grouping rules)
//
// It does NOT cover:
//
//   - Adding policies for new accounts (the wildcard covers them)
//   - Any team policy lifecycle (teams are not authz domains)
//
// User role assignments (g rules) are managed by RoleManager.
type PolicyManager interface {
	// AddPlatformPolicies seeds the static platform policies. Runs once at
	// bootstrap. Safe to call multiple times.
	AddPlatformPolicies(ctx context.Context) error

	// RemovePlatformPolicies removes all platform-domain policies and
	// grouping rules. Used only for full resets.
	RemovePlatformPolicies(ctx context.Context) error

	// RemoveAccountPolicies removes all Casbin grouping rules for a
	// specific account domain. Call this when an account is deleted.
	//
	// Only the account's user-to-role bindings are removed. The static
	// "account:*" policies are left untouched.
	//
	// domain must be a concrete account domain: "account:<uuid>".
	RemoveAccountPolicies(ctx context.Context, domain string) error
}