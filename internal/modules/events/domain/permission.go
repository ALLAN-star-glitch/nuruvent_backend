// internal/modules/events/domain/permission.go

package domain

import "context"

// PermissionChecker defines what the events domain needs for authorization
type PermissionChecker interface {
    // HasPermission checks if a user has a specific permission in a scope
    HasPermission(ctx context.Context, userID string, scope Scope, resource, action string) (bool, error)

    // HasAnyPermission checks if a user has any of the given permissions in a scope
    HasAnyPermission(ctx context.Context, userID string, scope Scope, resource string, actions ...string) (bool, error)

    // HasAllPermissions checks if a user has all of the given permissions in a scope
    HasAllPermissions(ctx context.Context, userID string, scope Scope, resource string, actions ...string) (bool, error)

    // --- CREATE ---
    CanCreateEvent(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- READ ---
    CanReadAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanReadOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanReadEvent(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- UPDATE ---
    CanUpdateAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanUpdateOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanUpdateEvent(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- DELETE ---
    CanDeleteAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanDeleteOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanDeleteEvent(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- PUBLISH ---
    CanPublishAllEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanPublishOwnEvents(ctx context.Context, userID string, scope Scope) (bool, error)
    CanPublishEvent(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- MANAGEMENT (Convenience) ---
    CanManageEvent(ctx context.Context, userID string, scope Scope) (bool, error)
    CanViewEvent(ctx context.Context, userID string, scope Scope) (bool, error)


    CanViewCreator(ctx context.Context, userID string, scope Scope) (bool, error)
}