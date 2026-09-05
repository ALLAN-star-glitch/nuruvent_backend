// internal/modules/profile/domain/service.go

package domain

import "context"

// Service defines the profile module's business logic interface
type Service interface {
    // ============================================================
    // USER PROFILE METHODS
    // ============================================================
    
    GetUserProfile(ctx context.Context, userID string) (*UserInfo, error)
    GetUserProfileWithDetails(ctx context.Context, userID string) (*UserInfo, error)
    GetUserProfiles(ctx context.Context, userIDs []string) ([]*UserInfo, error)
    UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*UserInfo, error)
    ListUsers(ctx context.Context, filters ListUsersFilters) ([]*UserInfo, int64, error)
    
    // ============================================================
    // INSTITUTION PROFILE METHODS
    // ============================================================
    
    GetInstitutionProfile(ctx context.Context, institutionID string) (*InstitutionInfo, error)
    GetInstitutionProfileWithDetails(ctx context.Context, institutionID string) (*InstitutionInfo, error)
    GetInstitutionProfiles(ctx context.Context, institutionIDs []string) ([]*InstitutionInfo, error)
    UpdateInstitutionProfile(ctx context.Context, institutionID string, updates map[string]interface{}) (*InstitutionInfo, error)
    ListInstitutions(ctx context.Context, filters ListInstitutionsFilters) ([]*InstitutionInfo, int64, error)
    
    // ============================================================
    // ORGANIZER INFO (FOR EVENTS MODULE)
    // ============================================================
    
    GetOrganizerInfo(ctx context.Context, scope Scope) (*OrganizerInfo, error)

    // ============================================================
    // MEDIA UPLOAD METHODS
    // ============================================================
    
    UploadUserAvatar(ctx context.Context, userID string, file []byte, filename, contentType string) (*UserInfo, error)
    UploadInstitutionLogo(ctx context.Context, institutionID string, file []byte, filename, contentType string) (*InstitutionInfo, error)
    DeleteUserAvatar(ctx context.Context, userID string) error
    DeleteInstitutionLogo(ctx context.Context, institutionID string) error

}

