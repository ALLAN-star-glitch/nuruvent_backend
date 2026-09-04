// internal/modules/profile/domain/permission.go

package domain

import "context"

// PermissionChecker defines what the profile domain needs for authorization
type PermissionChecker interface {
    // ============================================================
    // CORE PERMISSION METHODS
    // ============================================================

    // HasPermission checks if a user has a specific permission in a scope
    HasPermission(ctx context.Context, userID string, scope Scope, resource, action string) (bool, error)

    // HasAnyPermission checks if a user has any of the given permissions in a scope
    HasAnyPermission(ctx context.Context, userID string, scope Scope, resource string, actions ...string) (bool, error)

    // HasAllPermissions checks if a user has all of the given permissions in a scope
    HasAllPermissions(ctx context.Context, userID string, scope Scope, resource string, actions ...string) (bool, error)

    // ============================================================
    // PROFILE PERMISSIONS
    // ============================================================

    // --- READ ---
    CanReadAllProfiles(ctx context.Context, userID string, scope Scope) (bool, error)
    CanReadOwnProfile(ctx context.Context, userID string, scope Scope) (bool, error)
    CanReadProfile(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- UPDATE ---
    CanUpdateAllProfiles(ctx context.Context, userID string, scope Scope) (bool, error)
    CanUpdateOwnProfile(ctx context.Context, userID string, scope Scope) (bool, error)
    CanUpdateProfile(ctx context.Context, userID string, scope Scope) (bool, error)

    // --- MANAGEMENT (Convenience) ---
    CanManageProfile(ctx context.Context, userID string, scope Scope) (bool, error)
    CanViewProfile(ctx context.Context, userID string, scope Scope) (bool, error)

    // ============================================================
    // ✅ TEAM ROLE CHECKS (NEW)
    // ============================================================
    
    // IsTeamAdmin checks if user is an admin in the scope
    IsTeamAdmin(ctx context.Context, userID string, scope Scope) (bool, error)
    
    // IsEventManager checks if user is an event manager in the scope
    IsEventManager(ctx context.Context, userID string, scope Scope) (bool, error)
    
    // IsTeamMember checks if user is a team member in the scope
    IsTeamMember(ctx context.Context, userID string, scope Scope) (bool, error)

    // ============================================================
    // ✅ USER INFORMATION METHODS (NEW)
    // ============================================================
    
    // GetUserInstitutionTeamIDs returns all institution team IDs where a user has roles
    GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)
}