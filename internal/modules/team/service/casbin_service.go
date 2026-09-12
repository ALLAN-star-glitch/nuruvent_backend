// internal/modules/team/service/casbin_service.go

package service

import "context"

// ============================================================
// CASBIN SERVICE INTERFACE (Outbound Port)
// ============================================================
//
// DOMAIN MODEL:
//
//   - "platform"        — Nuruvent staff only
//   - "account:<uuid>"  — every tenant account
//
// Teams are NOT authz domains. Team-scoped checks always pass the account
// domain of the team:
//
//     accountdomain.AccountDomain(team.AccountID)
//
// The port speaks in permission terms (CanRemoveMember) rather than raw
// Casbin terms (HasPermission with resource+action). This keeps the team
// service free of stringly-typed Casbin calls and lets the port's
// implementation decide what "can remove member" means.
//
// MEMBERSHIP IS DATA:
//
//   - team_members stores who's on a team.
//   - Casbin g rules store user→role bindings in ACCOUNT domains.
//   - Adding or removing a team member does NOT touch Casbin.
type CasbinService interface {
    // ============================================================
    // ROLE (GROUPING) MANAGEMENT
    // ============================================================

    // AssignRole writes a grouping rule binding a user to a role in a
    // domain. Used only at ACCOUNT-level boundaries (account_members
    // synced to Casbin). Never called for team membership.
    AssignRole(ctx context.Context, domain string, userID, role string) error

    // RemoveRole removes a grouping rule.
    RemoveRole(ctx context.Context, domain string, userID, role string) error

    // GetUserRoles returns all roles a user holds in a domain.
    GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error)

    // ============================================================
    // TEAM PERMISSION CHECKS
    // ============================================================

    // CanViewTeams reports whether userID may list/view teams in the account.
    CanViewTeams(ctx context.Context, userID, domain string) (bool, error)

    // CanCreateTeam reports whether userID may create a team in the account.
    CanCreateTeam(ctx context.Context, userID, domain string) (bool, error)

    // CanManageTeam reports whether userID may modify team settings.
    CanManageTeam(ctx context.Context, userID, domain string) (bool, error)

    // CanDeleteTeam reports whether userID may delete a team.
    CanDeleteTeam(ctx context.Context, userID, domain string) (bool, error)

    // ============================================================
    // MEMBER PERMISSION CHECKS
    // ============================================================

    // CanViewMembers reports whether userID may view the member roster.
    CanViewMembers(ctx context.Context, userID, domain string) (bool, error)

    // CanManageMembers reports whether userID may perform any management
    // action on members (add, update, remove, invite).
    CanManageMembers(ctx context.Context, userID, domain string) (bool, error)

    // CanAddMember reports whether userID may add a member.
    CanAddMember(ctx context.Context, userID, domain string) (bool, error)

    // CanRemoveMember reports whether userID may remove a member.
    CanRemoveMember(ctx context.Context, userID, domain string) (bool, error)

    // CanInviteMember reports whether userID may invite a member.
    CanInviteMember(ctx context.Context, userID, domain string) (bool, error)

    // CanLeaveAccount reports whether userID may leave the account
    // (removing themselves).
    CanLeaveAccount(ctx context.Context, userID, domain string) (bool, error)

    // ============================================================
    // POLICY MANAGEMENT
    // ============================================================
    //
    // NOTE: platform/account policy seeding is the auth module's job.
    // This port does not include those methods. If you find yourself
    // wanting to seed policies from the team module, that's a layering
    // violation.
}

// ============================================================
// IMPLEMENTATION NOTES
// ============================================================
//
// The concrete implementation of CanRemoveMember, CanAddMember, etc. may
// itself call a lower-level HasPermission helper — that's fine. The
// important thing is that the TEAM service only sees the high-level
// methods, not raw Casbin calls.
//
// Example implementation shape:
//
//     func (s *casbinService) CanRemoveMember(ctx, userID, domain string) (bool, error) {
//         return s.hasPermission(ctx, domain, userID, "member", "delete")
//     }
//
// where hasPermission is an unexported helper on the concrete type.