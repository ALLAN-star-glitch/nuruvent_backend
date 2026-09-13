// internal/modules/auth/authdomain/permission_checker.go

package authdomain

import "context"

// PermissionChecker handles access evaluation queries.
// This interface is focused ONLY on reading/checking permissions.
//
// ============================================================
// DOMAIN MODEL (post-revamp)
// ============================================================
//
// Casbin uses exactly two domain families:
//
//   - "platform"        — Nuruvent staff only
//   - "account:<uuid>"  — every tenant account
//
// Policies in the account family use the wildcard pattern "account:*"
// (matched at check time via keyMatch2). Grouping rules (user role
// assignments) always use concrete domains ("account:<uuid>").
//
// Teams are NOT authorization domains. Team membership is enforced as
// data (team_members) in the service layer, not through this interface.
//
// USAGE GUIDANCE:
//
//   - The middleware uses HasPermission / HasAnyPermission / HasAllPermissions.
//   - Services should prefer the convenience methods (CanManageMembers,
//     CanCreateEvent, etc.) — they express intent and go through the same
//     Casbin check without hard-coding role names.
//   - Role checks (IsAccountAdmin, IsTrainer) are informational. Prefer a
//     permission check when authorizing an action.
type PermissionChecker interface {
	// ============================================================
	// CORE PERMISSION METHODS
	// ============================================================

	// HasPermission checks if a user has a specific permission in a domain.
	//
	// domain must be one of:
	//   - "platform"
	//   - "account:<uuid>"
	//
	// Returns (true, nil) if allowed, (false, nil) if denied, (false, error)
	// if the check itself failed.
	HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error)

	// HasAnyPermission checks if a user has any of the given permissions in a domain.
	HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// HasAllPermissions checks if a user has all of the given permissions in a domain.
	HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error)

	// ============================================================
	// ACCOUNT CONVENIENCE METHODS
	// ============================================================

	// CanViewAccount checks if user can view an account (account:read).
	CanViewAccount(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageAccount checks if user can manage an account
	// (account:update or account:delete).
	CanManageAccount(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// MEMBER CONVENIENCE METHODS
	// ============================================================

	// CanViewMembers checks if user can view the member roster (member:read).
	CanViewMembers(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageMembers checks if user can manage members
	// (member:create, member:update, member:delete, or member:invite).
	CanManageMembers(ctx context.Context, userID string, domain string) (bool, error)

	// CanAddMember checks if user can add a member (member:create).
	CanAddMember(ctx context.Context, userID string, domain string) (bool, error)

	// CanRemoveMember checks if user can remove a member (member:delete).
	CanRemoveMember(ctx context.Context, userID string, domain string) (bool, error)

	// CanInviteMember checks if user can invite a member (member:invite).
	CanInviteMember(ctx context.Context, userID string, domain string) (bool, error)

	// CanLeaveAccount checks if user can leave the account (member:leave).
	CanLeaveAccount(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// TEAM CONVENIENCE METHODS
	// ============================================================

	// CanViewTeams checks if user can view teams (team:read).
	CanViewTeams(ctx context.Context, userID string, domain string) (bool, error)

	// CanCreateTeam checks if user can create a team (team:create).
	CanCreateTeam(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageTeam checks if user can manage a team
	// (team:update or team:manage).
	CanManageTeam(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteTeam checks if user can delete a team (team:delete).
	CanDeleteTeam(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// EVENT CONVENIENCE METHODS
	// ============================================================

	// CanViewEvent checks if user can view events (event:read, read_all,
	// or read_own).
	CanViewEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadAllEvents checks if user can read ALL events (event:read_all).
	CanReadAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadOwnEvents checks if user can read OWN events (event:read_own).
	CanReadOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanCreateEvent checks if user can create events (event:create).
	CanCreateEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateAllEvents checks if user can update ALL events (event:update_all).
	CanUpdateAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateOwnEvents checks if user can update OWN events (event:update_own).
	CanUpdateOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteAllEvents checks if user can delete ALL events (event:delete_all).
	CanDeleteAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanDeleteOwnEvents checks if user can delete OWN events (event:delete_own).
	CanDeleteOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanPublishAllEvents checks if user can publish ALL events (event:publish_all).
	CanPublishAllEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanPublishOwnEvents checks if user can publish OWN events (event:publish_own).
	CanPublishOwnEvents(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageEvent checks if user can fully manage events (event:manage
	// or the combination of update_all + delete_all).
	CanManageEvent(ctx context.Context, userID string, domain string) (bool, error)

	// CanViewCreator checks if user can view creator details
	// (event:view_creator).
	CanViewCreator(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// PROFILE CONVENIENCE METHODS
	// ============================================================

	// CanViewProfile checks if user can view profiles (profile:read,
	// read_all, or read_own).
	CanViewProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadAllProfiles checks if user can read ALL profiles (profile:read_all).
	CanReadAllProfiles(ctx context.Context, userID string, domain string) (bool, error)

	// CanReadOwnProfile checks if user can read OWN profile (profile:read_own).
	CanReadOwnProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateAllProfiles checks if user can update ALL profiles (profile:update_all).
	CanUpdateAllProfiles(ctx context.Context, userID string, domain string) (bool, error)

	// CanUpdateOwnProfile checks if user can update OWN profile (profile:update_own).
	CanUpdateOwnProfile(ctx context.Context, userID string, domain string) (bool, error)

	// CanManageProfile checks if user can manage profiles (profile:manage
	// or the combination of update variants).
	CanManageProfile(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// ROLE CHECKS (INFORMATIONAL)
	// ============================================================
	//
	// Prefer the permission methods above when authorizing an action. These
	// role checks exist for UI hints and diagnostics (e.g. "show admin
	// badge"), not for gating.

	// IsAccountAdmin reports whether user holds the account_admin role.
	IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error)

	// IsTrainer reports whether user holds the trainer role.
	IsTrainer(ctx context.Context, userID string, domain string) (bool, error)

	// ============================================================
	// USER INFORMATION METHODS
	// ============================================================

	// GetUserRoles returns all roles for a user in a domain.
	GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error)

	// GetUserAccountIDs returns all account IDs where a user has membership.
	// The returned values are bare UUIDs, not domain strings.
	GetUserAccountIDs(ctx context.Context, userID string) ([]string, error)
}