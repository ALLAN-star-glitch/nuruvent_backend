// internal/modules/profile/domain/permission.go

package domain

import "context"

// PermissionChecker defines what the profile domain needs for authorization
// This interface is implemented by the auth module's authorization.PermissionChecker
type PermissionChecker interface {
	// ============================================================
	// CORE PERMISSION METHODS
	// ============================================================

	// HasPermission checks if a user has a specific permission in a domain
	// domain formats:
	//   - "personal:team:{user_id}" for personal teams
	//   - "institution:team:{account_id}" for institution teams
	//   - "account:{account_id}" for account-level permissions
	//   - "platform" for platform-wide permissions
	HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error)

	// HasAnyPermission checks if a user has any of the given permissions in a domain
	HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// HasAllPermissions checks if a user has all of the given permissions in a domain
	HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// ============================================================
	// PROFILE PERMISSIONS - READ
	// ============================================================

	// CanReadAllProfiles checks if user can read ALL profiles in a domain
	// This requires account_admin role in the domain
	CanReadAllProfiles(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadOwnProfile checks if user can read OWN profile in a domain
	// This requires being a member of the domain (trainer or account_admin)
	CanReadOwnProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadProfile checks if user can read profiles in a domain (ALL or OWN)
	// Returns true if user has either CanReadAllProfiles or CanReadOwnProfile
	CanReadProfile(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// PROFILE PERMISSIONS - UPDATE
	// ============================================================

	// CanUpdateAllProfiles checks if user can update ALL profiles in a domain
	// This requires account_admin role in the domain
	CanUpdateAllProfiles(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateOwnProfile checks if user can update OWN profile in a domain
	// This requires being a member of the domain (trainer or account_admin)
	CanUpdateOwnProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateProfile checks if user can update profiles in a domain (ALL or OWN)
	// Returns true if user has either CanUpdateAllProfiles or CanUpdateOwnProfile
	CanUpdateProfile(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// PROFILE PERMISSIONS - MANAGEMENT (Convenience)
	// ============================================================

	// CanManageProfile checks if user can manage profiles in a domain
	// This is a convenience method that checks CanUpdateAllProfiles
	CanManageProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewProfile checks if user can view profiles in a domain (ALL or OWN)
	// This is a convenience method that checks CanReadProfile
	CanViewProfile(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// ACCOUNT ROLE CHECKS
	// ============================================================

	// IsAccountAdmin checks if user is an account admin in the domain
	// domain should be "institution:team:{account_id}" or "account:{account_id}"
	IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error)

	// IsTeamAdmin checks if user is an account admin (alias for IsAccountAdmin)
	// This is used by the profile service for consistency
	IsTeamAdmin(ctx context.Context, userID string, domain string) (bool, error)

	// IsTrainer checks if user is a trainer in the domain
	// domain should be "institution:team:{account_id}" or "account:{account_id}"
	IsTrainer(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// USER INFORMATION METHODS
	// ============================================================

	// GetUserTeamDomains returns all team domains where a user has membership
	// Returns domains in format: "personal:team:{user_id}" and "institution:team:{account_id}"
	GetUserTeamDomains(ctx context.Context, userID string) ([]string, error)

	// GetUserPersonalTeamIDs returns personal team IDs where a user has roles
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserInstitutionTeamIDs returns institution team IDs where a user has roles
	// Returns account IDs for institution teams
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserAccountIDs returns all account IDs where a user has membership
	GetUserAccountIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserRoles returns all roles for a user in a domain
	GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error)

	// HasTeamAccess checks if a user has any team role
	HasTeamAccess(ctx context.Context, userID string) (bool, error)
}