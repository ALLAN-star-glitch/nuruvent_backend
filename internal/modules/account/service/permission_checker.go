// internal/modules/account/service/permission_checker.go

package service

import "context"

// PermissionChecker defines the permission checking operations needed by the account module
type PermissionChecker interface {
    // HasPermission checks if a user has a specific permission in a scope
    HasPermission(ctx context.Context, scope Scope, userID, resource, action string) (bool, error)

    // HasAnyPermission checks if a user has any of the given permissions in a scope
    HasAnyPermission(ctx context.Context, scope Scope, userID, resource string, actions ...string) (bool, error)

    // HasAllPermissions checks if a user has all of the given permissions in a scope
    HasAllPermissions(ctx context.Context, scope Scope, userID, resource string, actions ...string) (bool, error)

    // CanManageAccount checks if user can manage an account
    CanManageAccount(ctx context.Context, scope Scope, userID string) (bool, error)

    // CanManageAccountMembers checks if user can manage account members
    CanManageAccountMembers(ctx context.Context, scope Scope, userID string) (bool, error)

    // CanViewAccount checks if user can view an account
    CanViewAccount(ctx context.Context, scope Scope, userID string) (bool, error)
}