// internal/modules/team/service/casbin_service.go

package service

import "context"

// ============================================================
// CASBIN SERVICE INTERFACE (Outbound Port)
// ============================================================

// CasbinService defines the permission operations needed by the team module
type CasbinService interface {
    // AssignRole assigns a role to a user in a team scope
    // scope.String() should return "personal:team:{id}" or "institution:team:{id}"
    AssignRole(ctx context.Context, scope Scope, userID, role string) error

    // RemoveRole removes a role from a user in a team scope
    RemoveRole(ctx context.Context, scope Scope, userID, role string) error

    // GetUserRoles returns all roles for a user in a team scope
    GetUserRoles(ctx context.Context, scope Scope, userID string) ([]string, error)

    // HasPermission checks if a user has a specific permission in a scope
    HasPermission(ctx context.Context, scope Scope, userID, resource, action string) (bool, error)

    // AddTeamPolicies adds all policies for a new team
    AddTeamPolicies(ctx context.Context, scope Scope) error

    // RemoveTeamPolicies removes all policies for a team
    RemoveTeamPolicies(ctx context.Context, scope Scope) error

    // ============================================================
    // TEAM ROLE CHECKS
    // ============================================================

    // IsTeamAdmin checks if a user is an admin in the scope
    IsTeamAdmin(ctx context.Context, scope Scope, userID string) (bool, error)

    // IsEventManager checks if a user is an event manager in the scope
    IsEventManager(ctx context.Context, scope Scope, userID string) (bool, error)

    // IsTeamMember checks if a user is a team member in the scope
    IsTeamMember(ctx context.Context, scope Scope, userID string) (bool, error)
}

