// internal/modules/auth/authdomain/policy_manager.go

package authdomain

import "context"

// PolicyManager handles lifecycle policies for team and account domains
// This interface is focused ONLY on policy lifecycle management
type PolicyManager interface {
	// AddTeamPolicies adds default policies for a team domain
	// domain: "personal:team:{user_id}" or "institution:team:{institution_id}"
	AddTeamPolicies(ctx context.Context, domain string) error

	// RemoveTeamPolicies removes all policies for a team domain
	RemoveTeamPolicies(ctx context.Context, domain string) error

	// AddAccountPolicies adds default policies for an account domain
	// domain: "account:{account_id}"
	AddAccountPolicies(ctx context.Context, domain string) error

	// RemoveAccountPolicies removes all policies for an account domain
	RemoveAccountPolicies(ctx context.Context, domain string) error

	// AddPlatformPolicies adds default platform policies
	AddPlatformPolicies(ctx context.Context) error

	// RemovePlatformPolicies removes all platform policies
	RemovePlatformPolicies(ctx context.Context) error
}