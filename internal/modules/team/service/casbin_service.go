// internal/modules/team/service/casbin_service.go

package service

import "context"

// ============================================================
// CASBIN SERVICE INTERFACE (Outbound Port)
// ============================================================

// CasbinService defines the permission operations needed by the team module
type CasbinService interface {
    // AssignRole assigns a role to a user in a domain
    // domain should be: "personal:team:{team_id}", "institution:team:{team_id}", or "account:{account_id}"
    AssignRole(ctx context.Context, domain string, userID, role string) error

    // RemoveRole removes a role from a user in a domain
    RemoveRole(ctx context.Context, domain string, userID, role string) error

    // GetUserRoles returns all roles for a user in a domain
    GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error)

    // HasPermission checks if a user has a specific permission in a domain
    HasPermission(ctx context.Context, domain string, userID, resource, action string) (bool, error)

    // AddTeamPolicies adds all policies for a new team
    AddTeamPolicies(ctx context.Context, domain string) error

    // RemoveTeamPolicies removes all policies for a team
    RemoveTeamPolicies(ctx context.Context, domain string) error

    // AddAccountPolicies adds all policies for a new account
    AddAccountPolicies(ctx context.Context, domain string) error

    // RemoveAccountPolicies removes all policies for an account
    RemoveAccountPolicies(ctx context.Context, domain string) error

    // ============================================================
    // TEAM ROLE CHECKS
    // ============================================================

    // IsAccountAdmin checks if a user is an account admin in the domain
    IsAccountAdmin(ctx context.Context, domain string, userID string) (bool, error)

    // IsTrainer checks if a user is a trainer in the domain
    IsTrainer(ctx context.Context, domain string, userID string) (bool, error)
}