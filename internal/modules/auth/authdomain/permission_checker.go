// internal/modules/auth/authdomain/permission_checker.go

package authdomain

import "context"

// PermissionChecker handles access evaluation queries
// This interface is focused ONLY on reading/checking permissions
type PermissionChecker interface {
	// HasPermission checks if a user has a specific permission in a domain
	// domain: "personal:team:{user_id}", "institution:team:{institution_id}", "account:{account_id}", or "platform"
	// Returns (true, nil) if allowed, (false, nil) if denied, (false, error) if check failed
	HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error)

	// HasAnyPermission checks if a user has any of the given permissions in a domain
	HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// HasAllPermissions checks if a user has all of the given permissions in a domain
	HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// ============================================================
	// ACCOUNT CONVENIENCE METHODS
	// ============================================================

	// CanManageAccount checks if user can manage an account (account:update, account:delete, etc.)
	CanManageAccount(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageAccountMembers checks if user can manage account members (account:member_add, account:member_remove)
	CanManageAccountMembers(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewAccount checks if user can view an account (account:read)
	CanViewAccount(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// EVENT CONVENIENCE METHODS
	// ============================================================

	// CanReadAllEvents checks if user can read ALL events in a domain
	CanReadAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadOwnEvents checks if user can read OWN events in a domain
	CanReadOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanCreateEvent checks if user can create events in a domain
	CanCreateEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateAllEvents checks if user can update ALL events in a domain
	CanUpdateAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateOwnEvents checks if user can update OWN events in a domain
	CanUpdateOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteAllEvents checks if user can delete ALL events in a domain
	CanDeleteAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteOwnEvents checks if user can delete OWN events in a domain
	CanDeleteOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanPublishAllEvents checks if user can publish ALL events in a domain
	CanPublishAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanPublishOwnEvents checks if user can publish OWN events in a domain
	CanPublishOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewCreator checks if user can view creator details
	CanViewCreator(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageEvent checks if user can manage events in a domain
	CanManageEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewEvent checks if user can view events in a domain
	CanViewEvent(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// PROFILE CONVENIENCE METHODS
	// ============================================================

	// CanReadAllProfiles checks if user can read ALL profiles in a domain
	CanReadAllProfiles(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadOwnProfile checks if user can read OWN profile in a domain
	CanReadOwnProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateAllProfiles checks if user can update ALL profiles in a domain
	CanUpdateAllProfiles(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateOwnProfile checks if user can update OWN profile in a domain
	CanUpdateOwnProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewProfile checks if user can view profiles in a domain
	CanViewProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageProfile checks if user can manage profiles in a domain
	CanManageProfile(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// TEAM ROLE CHECKS
	// ============================================================

	// IsAccountAdmin checks if user is an account admin in the domain
	IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error)

	// IsTrainer checks if user is a trainer in the domain
	IsTrainer(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// USER INFORMATION METHODS
	// ============================================================

	// GetUserRoles returns all roles for a user in a domain
	GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error)

	// GetUserTeamIDs returns all team IDs where a user has roles
	GetUserTeamIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserPersonalTeamIDs returns personal team IDs where a user has roles
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)

	// GetUserInstitutionTeamIDs returns institution team IDs where a user has roles
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)

	// HasTeamAccess checks if a user has any team role
	HasTeamAccess(ctx context.Context, userID string) (bool, error)
}