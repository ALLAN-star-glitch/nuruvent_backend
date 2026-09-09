// internal/modules/team/service/casbin_service.go

package service

import "context"

// ============================================================
// CASBIN SERVICE INTERFACE (Outbound Port)
// ============================================================

// CasbinService defines the permission operations needed by the team module
// ✅ Updated: Roles are inherited from account, but we still need to assign
// users to team domains in Casbin for permission checks
type CasbinService interface {
    // ============================================================
    // ROLE MANAGEMENT
    // ============================================================

    // AssignRole assigns a role to a user in a domain
    // Used to link users to team domains in Casbin
    // Example: g, {user_id}, {role}, {domain}
    AssignRole(ctx context.Context, domain string, userID, role string) error

    // RemoveRole removes a role from a user in a domain
    RemoveRole(ctx context.Context, domain string, userID, role string) error

    // GetUserRoles returns all roles for a user in a domain
    GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error)

    // ============================================================
    // PERMISSION CHECKS
    // ============================================================

    // HasPermission checks if a user has a specific permission in a domain
    // This queries Casbin policies which are based on account-level roles
    HasPermission(ctx context.Context, domain string, userID, resource, action string) (bool, error)

    // ============================================================
    // POLICY MANAGEMENT
    // ============================================================

    // AddTeamPolicies adds all policies for a new team
    // Policies define what actions are allowed for each role
    AddTeamPolicies(ctx context.Context, domain string) error

    // RemoveTeamPolicies removes all policies for a team
    RemoveTeamPolicies(ctx context.Context, domain string) error

    // AddAccountPolicies adds all policies for a new account
    AddAccountPolicies(ctx context.Context, domain string) error

    // RemoveAccountPolicies removes all policies for an account
    RemoveAccountPolicies(ctx context.Context, domain string) error

    // ============================================================
    // TEAM ROLE CHECKS - QUERY ACCOUNT MODULE
    // ============================================================

    // IsAccountAdmin checks if a user is an account admin in the domain
    // ✅ UPDATED: Queries account_members table via Auth service
    // Domain should be: "personal:team:{team_id}" or "institution:team:{team_id}"
    // The team's AccountID is used to look up the user's role in account_members
    IsAccountAdmin(ctx context.Context, domain string, userID string) (bool, error)

    // IsTrainer checks if a user is a trainer in the domain
    // ✅ UPDATED: Queries account_members table via Auth service
    // Domain should be: "personal:team:{team_id}" or "institution:team:{team_id}"
    // The team's AccountID is used to look up the user's role in account_members
    IsTrainer(ctx context.Context, domain string, userID string) (bool, error)
}