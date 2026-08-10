package domain

import "context"

// ============================================================
// OUTBOUND PORT: PermissionService
// Defines what the auth module NEEDS for authorization
// ============================================================

type PermissionService interface {
    // Role Assignment
    AssignAccountAdminRole(ctx context.Context, accountID, userID string) error
    AssignEventManagerRole(ctx context.Context, accountID, userID string) error
    AssignTeamMemberRole(ctx context.Context, accountID, userID string) error
    AssignAdminRole(ctx context.Context, userID string) error
    AssignSuperAdminRole(ctx context.Context, userID string) error

    // Role Removal
    RemoveRole(ctx context.Context, accountID, userID string, role string) error
    RemoveAllAccountRoles(ctx context.Context, accountID, userID string) error
    RemovePlatformRole(ctx context.Context, userID string, role string) error

    // Account Policy Management
    AddAccountPolicies(ctx context.Context, accountID string) error
    RemoveAccountPolicies(ctx context.Context, accountID string) error
    AddPlatformPolicies(ctx context.Context) error

    // Permission Checks
    HasPermission(ctx context.Context, userID, domain, resource, action string) bool
    CanManageEvent(ctx context.Context, accountID, userID string) bool
    CanIssueCertificate(ctx context.Context, accountID, userID string) bool
    CanManageAccount(ctx context.Context, accountID, userID string) bool

    // Account Role Checks
    IsAccountAdmin(ctx context.Context, accountID, userID string) bool
    IsEventManager(ctx context.Context, accountID, userID string) bool
    IsTeamMember(ctx context.Context, accountID, userID string) bool

    // User Information
    GetUserRoles(ctx context.Context, userID, domain string) []string
    GetUserAccounts(ctx context.Context, userID string) []string
    HasAccountAccess(ctx context.Context, userID string) bool
}