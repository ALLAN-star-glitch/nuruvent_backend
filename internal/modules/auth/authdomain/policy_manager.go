// internal/modules/auth/authdomain/policy_manager.go

package authdomain

import "context"

// PolicyManager handles lifecycle policies for team domains
// This interface is focused ONLY on policy lifecycle management
type PolicyManager interface {
	// AddTeamPolicies adds default policies for a team scope
	// Domain: personal:team:{user_id} or institution:team:{institution_id}
	AddTeamPolicies(ctx context.Context, scope Scope) error

	// RemoveTeamPolicies removes all policies for a team scope
	RemoveTeamPolicies(ctx context.Context, scope Scope) error

	// AddPlatformPolicies adds default platform policies
	AddPlatformPolicies(ctx context.Context) error

	// RemovePlatformPolicies removes all platform policies
	RemovePlatformPolicies(ctx context.Context) error
}