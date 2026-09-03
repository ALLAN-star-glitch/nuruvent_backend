// internal/modules/profile/domain/service.go

package domain

import "context"

// Service defines the profile module's business logic interface
type Service interface {
    // ============================================================
    // USER PROFILE METHODS
    // ============================================================
    
    // GetUserProfile retrieves basic user information by ID
    GetUserProfile(ctx context.Context, userID string) (*UserInfo, error)
    
    // GetUserProfileWithDetails retrieves full user information with sensitive data
    // Requires proper permissions (viewer must have read_all or be the user themselves)
    GetUserProfileWithDetails(ctx context.Context, userID string) (*UserInfo, error)
    
    // GetUserProfiles retrieves basic information for multiple users
    GetUserProfiles(ctx context.Context, userIDs []string) ([]*UserInfo, error)
    
    // UpdateUserProfile updates a user's profile
    // Requires proper permissions (user must be updating themselves or have update_all)
    UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*UserInfo, error)
    
    // ListUsers returns a paginated list of users with flexible filtering
    ListUsers(ctx context.Context, filters ListUsersFilters) ([]*UserInfo, int64, error)
    
    // ============================================================
    // INSTITUTION PROFILE METHODS
    // ============================================================
    
    // GetInstitutionProfile retrieves basic institution information by ID
    GetInstitutionProfile(ctx context.Context, institutionID string) (*InstitutionInfo, error)
    
    // GetInstitutionProfileWithDetails retrieves full institution information with sensitive data
    // Requires proper permissions (viewer must have read_all or be admin)
    GetInstitutionProfileWithDetails(ctx context.Context, institutionID string) (*InstitutionInfo, error)
    
    // GetInstitutionProfiles retrieves basic information for multiple institutions
    GetInstitutionProfiles(ctx context.Context, institutionIDs []string) ([]*InstitutionInfo, error)
    
    // UpdateInstitutionProfile updates an institution's profile
    // Requires proper permissions (user must have update_all in the institution)
    UpdateInstitutionProfile(ctx context.Context, institutionID string, updates map[string]interface{}) (*InstitutionInfo, error)
    
    // ListInstitutions returns a paginated list of institutions with flexible filtering
    ListInstitutions(ctx context.Context, filters ListInstitutionsFilters) ([]*InstitutionInfo, int64, error)
    
    // ============================================================
    // ORGANIZER INFO (FOR EVENTS MODULE)
    // ============================================================
    
    // GetOrganizerInfo returns public-facing organizer information based on scope
    // Used by events module to display who is organizing an event
    // Scope determines whether to return institution or user info
    GetOrganizerInfo(ctx context.Context, scope Scope) (*OrganizerInfo, error)
}