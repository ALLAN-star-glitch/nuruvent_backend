// internal/modules/profile/service/service_impl.go

package service

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

type profileService struct {
    repo        domain.Repository
    permChecker domain.PermissionChecker
    mediaSvc    domain.MediaService 
}

func NewProfileService(repo domain.Repository, permChecker domain.PermissionChecker, mediaSvc domain.MediaService,) domain.Service {
    return &profileService{
        repo:        repo,
        permChecker: permChecker,
        mediaSvc:    mediaSvc,
    }
}

// ============================================================
// USER PROFILE METHODS
// ============================================================

func (s *profileService) GetUserProfile(ctx context.Context, userID string) (*domain.UserInfo, error) {
    if userID == "" {
        return nil, domain.ErrInvalidUserID
    }

    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, domain.ErrUserNotFound
    }

    return &domain.UserInfo{
        ID:          user.ID,
        Name:        user.Name,
        DisplayName: user.DisplayName,
        AvatarURL:   user.AvatarURL,
    }, nil
}

func (s *profileService) GetUserProfileWithDetails(ctx context.Context, userID string) (*domain.UserInfo, error) {
    if userID == "" {
        return nil, domain.ErrInvalidUserID
    }
	

    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

	    log.Printf("🔍 GetUserProfileWithDetails: viewerID=%s, userID=%s, same=%v", 
        viewerID, userID, viewerID == userID)

    // If viewing own profile, always allowed
    if viewerID == userID {
        return s.getUserDetails(ctx, userID)
    }

    // ✅ First, check if viewer has read_all permission globally
    // Check in viewer's personal team first (for account admins)
    viewerPersonalScope := domain.NewPersonalTeamScope(viewerID)
    allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, viewerPersonalScope)
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    if allowed {
        return s.getUserDetails(ctx, userID)
    }

    // Check if viewer has read_all in any institution where target user is a member
    targetInstitutionIDs, err := s.permChecker.GetUserInstitutionTeamIDs(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get target user's institution memberships: %w", err)
    }

    for _, institutionID := range targetInstitutionIDs {
        institutionScope := domain.NewInstitutionTeamScope(institutionID)
        allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, institutionScope)
        if err != nil {
            continue
        }
        if allowed {
            return s.getUserDetails(ctx, userID)
        }
    }

    // Finally, check if viewer has read permission in any institution
    // (this checks read_own as a fallback)
    for _, institutionID := range targetInstitutionIDs {
        institutionScope := domain.NewInstitutionTeamScope(institutionID)
        allowed, err := s.permChecker.CanReadProfile(ctx, viewerID, institutionScope)
        if err != nil {
            continue
        }
        if allowed {
            return s.getUserDetails(ctx, userID)
        }
    }

    return nil, domain.ErrPermissionDenied
}

func (s *profileService) GetUserProfiles(ctx context.Context, userIDs []string) ([]*domain.UserInfo, error) {
    if len(userIDs) == 0 {
        return []*domain.UserInfo{}, nil
    }

    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

    // Check if viewer has read_all permission
    scope := domain.NewPersonalTeamScope(viewerID)
    allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, scope)
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    if !allowed {
        // If requesting only their own profile, allow it
        if len(userIDs) == 1 && userIDs[0] == viewerID {
            // Allow viewing own profile
        } else {
            return nil, domain.ErrPermissionDenied
        }
    }

    users, err := s.repo.GetUsersByIDs(ctx, userIDs)
    if err != nil {
        return nil, err
    }

    userInfos := make([]*domain.UserInfo, len(users))
    for i, user := range users {
        userInfos[i] = &domain.UserInfo{
            ID:          user.ID,
            Name:        user.Name,
            DisplayName: user.DisplayName,
            AvatarURL:   user.AvatarURL,
        }
    }
    return userInfos, nil
}

func (s *profileService) UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*domain.UserInfo, error) {
    if userID == "" {
        return nil, domain.ErrInvalidUserID
    }

    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

    // If updating own profile, always allowed
    if viewerID == userID {
        return s.updateUser(ctx, userID, updates)
    }

    // Get all institution teams where the target user is a member
    targetInstitutionIDs, err := s.permChecker.GetUserInstitutionTeamIDs(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get target user's institution memberships: %w", err)
    }

    // Check if viewer has update permission in any institution where target user is a member
    for _, institutionID := range targetInstitutionIDs {
        institutionScope := domain.NewInstitutionTeamScope(institutionID)
        allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, institutionScope)
        if err != nil {
            continue
        }
        if allowed {
            return s.updateUser(ctx, userID, updates)
        }
    }

    return nil, domain.ErrPermissionDenied
}

func (s *profileService) ListUsers(ctx context.Context, filters domain.ListUsersFilters) ([]*domain.UserInfo, int64, error) {
    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, 0, domain.ErrPermissionDenied
    }

    // If team filter is personal, only allow listing if viewer is the user
    if filters.Team.Type == "personal" && filters.Team.ID != viewerID {
        return nil, 0, domain.ErrPermissionDenied
    }

    // If team filter is institution, check if viewer has read_all permission
    if filters.Team.Type == "institution" {
        scope := domain.NewInstitutionTeamScope(filters.Team.ID)
        allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, scope)
        if err != nil {
            return nil, 0, fmt.Errorf("permission check failed: %w", err)
        }
        if !allowed {
            return nil, 0, domain.ErrPermissionDenied
        }
    }

    // If no team filter, check if viewer has read_all permission
    if filters.Team.Type == "" {
        scope := domain.NewPersonalTeamScope(viewerID)
        allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, scope)
        if err != nil {
            return nil, 0, fmt.Errorf("permission check failed: %w", err)
        }
        if !allowed {
            return nil, 0, domain.ErrPermissionDenied
        }
    }

    users, total, err := s.repo.ListUsers(ctx, filters)
    if err != nil {
        return nil, 0, err
    }

    userInfos := make([]*domain.UserInfo, len(users))
    for i, user := range users {
        userInfos[i] = &domain.UserInfo{
            ID:          user.ID,
            Name:        user.Name,
            DisplayName: user.DisplayName,
            AvatarURL:   user.AvatarURL,
        }
    }
    return userInfos, total, nil
}

// ============================================================
// INSTITUTION PROFILE METHODS
// ============================================================

func (s *profileService) GetInstitutionProfile(ctx context.Context, institutionID string) (*domain.InstitutionInfo, error) {
    if institutionID == "" {
        return nil, domain.ErrInvalidInstitutionID
    }

    institution, err := s.repo.GetInstitutionByID(ctx, institutionID)
    if err != nil {
        return nil, err
    }
    if institution == nil {
        return nil, domain.ErrInstitutionNotFound
    }

    return &domain.InstitutionInfo{
        ID:          institution.ID,
        Name:        institution.Name,
        DisplayName: institution.DisplayName,
        Slug:        institution.Slug,
        LogoURL:     institution.LogoURL,
    }, nil
}

func (s *profileService) GetInstitutionProfileWithDetails(ctx context.Context, institutionID string) (*domain.InstitutionInfo, error) {
    if institutionID == "" {
        return nil, domain.ErrInvalidInstitutionID
    }

    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

    // Check if viewer has read permission in the institution scope
    scope := domain.NewInstitutionTeamScope(institutionID)
    allowed, err := s.permChecker.CanReadProfile(ctx, viewerID, scope)
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    if !allowed {
        return nil, domain.ErrPermissionDenied
    }

    institution, err := s.repo.GetInstitutionByID(ctx, institutionID)
    if err != nil {
        return nil, err
    }
    if institution == nil {
        return nil, domain.ErrInstitutionNotFound
    }

    return &domain.InstitutionInfo{
        ID:          institution.ID,
        Name:        institution.Name,
        DisplayName: institution.DisplayName,
        Slug:        institution.Slug,
        Email:       institution.Email,
        Phone:       institution.Phone,
        Website:     institution.Website,
        Description: institution.Description,
        LogoURL:     institution.LogoURL,
        Address:     institution.Address,
        City:        institution.City,
        Country:     institution.Country,
    }, nil
}

func (s *profileService) GetInstitutionProfiles(ctx context.Context, institutionIDs []string) ([]*domain.InstitutionInfo, error) {
    if len(institutionIDs) == 0 {
        return []*domain.InstitutionInfo{}, nil
    }

    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

    institutions, err := s.repo.GetInstitutionsByIDs(ctx, institutionIDs)
    if err != nil {
        return nil, err
    }

    institutionInfos := make([]*domain.InstitutionInfo, len(institutions))
    for i, institution := range institutions {
        institutionInfos[i] = &domain.InstitutionInfo{
            ID:          institution.ID,
            Name:        institution.Name,
            DisplayName: institution.DisplayName,
            Slug:        institution.Slug,
            LogoURL:     institution.LogoURL,
        }
    }
    return institutionInfos, nil
}

func (s *profileService) UpdateInstitutionProfile(ctx context.Context, institutionID string, updates map[string]interface{}) (*domain.InstitutionInfo, error) {
    if institutionID == "" {
        return nil, domain.ErrInvalidInstitutionID
    }

    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

    // Check if viewer has update permission in the institution scope
    scope := domain.NewInstitutionTeamScope(institutionID)
    allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, scope)
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    if !allowed {
        return nil, domain.ErrPermissionDenied
    }

    institution, err := s.repo.GetInstitutionByID(ctx, institutionID)
    if err != nil {
        return nil, err
    }
    if institution == nil {
        return nil, domain.ErrInstitutionNotFound
    }

    // Apply updates
    if name, ok := updates["name"].(string); ok && name != "" {
        institution.Name = name
    }
    if displayName, ok := updates["display_name"].(string); ok {
        institution.DisplayName = displayName
    }
    if email, ok := updates["email"].(string); ok {
        institution.Email = email
    }
    if phone, ok := updates["phone"].(string); ok {
        institution.Phone = phone
    }
    if website, ok := updates["website"].(string); ok {
        institution.Website = website
    }
    if description, ok := updates["description"].(string); ok {
        institution.Description = description
    }
    if logoURL, ok := updates["logo_url"].(string); ok {
        institution.LogoURL = logoURL
    }
    if address, ok := updates["address"].(string); ok {
        institution.Address = address
    }
    if city, ok := updates["city"].(string); ok {
        institution.City = city
    }
    if country, ok := updates["country"].(string); ok {
        institution.Country = country
    }

    if err := s.repo.UpdateInstitution(ctx, institution); err != nil {
        return nil, fmt.Errorf("failed to update institution: %w", err)
    }

    log.Printf("✅ Institution profile updated: %s", institutionID)

    return &domain.InstitutionInfo{
        ID:          institution.ID,
        Name:        institution.Name,
        DisplayName: institution.DisplayName,
        Slug:        institution.Slug,
        Email:       institution.Email,
        Phone:       institution.Phone,
        Website:     institution.Website,
        Description: institution.Description,
        LogoURL:     institution.LogoURL,
        Address:     institution.Address,
        City:        institution.City,
        Country:     institution.Country,
    }, nil
}

func (s *profileService) ListInstitutions(ctx context.Context, filters domain.ListInstitutionsFilters) ([]*domain.InstitutionInfo, int64, error) {
    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, 0, domain.ErrPermissionDenied
    }

    institutions, total, err := s.repo.ListInstitutions(ctx, filters)
    if err != nil {
        return nil, 0, err
    }

    institutionInfos := make([]*domain.InstitutionInfo, len(institutions))
    for i, institution := range institutions {
        institutionInfos[i] = &domain.InstitutionInfo{
            ID:          institution.ID,
            Name:        institution.Name,
            DisplayName: institution.DisplayName,
            Slug:        institution.Slug,
            LogoURL:     institution.LogoURL,
        }
    }
    return institutionInfos, total, nil
}

// ============================================================
// ORGANIZER INFO (FOR EVENTS MODULE)
// ============================================================

func (s *profileService) GetOrganizerInfo(ctx context.Context, scope domain.Scope) (*domain.OrganizerInfo, error) {
    if scope.IsInstitution() {
        institution, err := s.repo.GetInstitutionByID(ctx, scope.ID)
        if err != nil {
            return nil, err
        }
        if institution == nil {
            return nil, domain.ErrInstitutionNotFound
        }
        return &domain.OrganizerInfo{
            ID:          institution.ID,
            Name:        institution.Name,
            DisplayName: institution.DisplayName,
            Type:        "institution",
            AvatarURL:   institution.LogoURL,
            Slug:        institution.Slug,
        }, nil
    }

    if scope.IsPersonal() {
        user, err := s.repo.GetUserByID(ctx, scope.ID)
        if err != nil {
            return nil, err
        }
        if user == nil {
            return nil, domain.ErrUserNotFound
        }
        return &domain.OrganizerInfo{
            ID:          user.ID,
            Name:        user.Name,
            DisplayName: user.DisplayName,
            Type:        "personal",
            AvatarURL:   user.AvatarURL,
            Slug:        user.Slug,
        }, nil
    }

    return nil, domain.ErrInvalidScope
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// getUserIDFromContext extracts user ID from context
func (s *profileService) getUserIDFromContext(ctx context.Context) string {
    if userID, ok := ctx.Value("user_id").(string); ok {
        return userID
    }
    return ""
}

// getUserDetails gets full user details without permission checks
func (s *profileService) getUserDetails(ctx context.Context, userID string) (*domain.UserInfo, error) {
    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, domain.ErrUserNotFound
    }

    return &domain.UserInfo{
        ID:          user.ID,
        Name:        user.Name,
        DisplayName: user.DisplayName,
        Email:       user.Email,
        Phone:       user.Phone,
        AccountType: user.AccountType,
        AvatarURL:   user.AvatarURL,
        Bio:         user.Bio,
        Location:    user.Location,
        Website:     user.Website,
        SocialLinks: user.SocialLinks,
        CreatedAt:   user.CreatedAt,  // ✅ Add this
        UpdatedAt:   user.UpdatedAt,
    }, nil
}

// updateUser updates a user without permission checks
func (s *profileService) updateUser(ctx context.Context, userID string, updates map[string]interface{}) (*domain.UserInfo, error) {
    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, domain.ErrUserNotFound
    }

    // ✅ Update all fields now
    if name, ok := updates["name"].(string); ok && name != "" {
        user.Name = name
    }
    if displayName, ok := updates["display_name"].(string); ok {
        user.DisplayName = displayName
    }
    if phone, ok := updates["phone"].(string); ok {
        user.Phone = phone
    }
    if avatarURL, ok := updates["avatar_url"].(string); ok {
        user.AvatarURL = avatarURL
    }
    if bio, ok := updates["bio"].(string); ok {
        user.Bio = bio
    }
    if location, ok := updates["location"].(string); ok {
        user.Location = location
    }
    if website, ok := updates["website"].(string); ok {
        user.Website = website
    }
    if socialLinks, ok := updates["social_links"].(map[string]string); ok {
        user.SocialLinks = socialLinks
    }

    if err := s.repo.UpdateUser(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }

    log.Printf("✅ User profile updated: %s", userID)

    return &domain.UserInfo{
        ID:          user.ID,
        Name:        user.Name,
        DisplayName: user.DisplayName,
        Email:       user.Email,
        Phone:       user.Phone,
        AccountType: user.AccountType,
        AvatarURL:   user.AvatarURL,
        Bio:         user.Bio,
        Location:    user.Location,
        Website:     user.Website,
        SocialLinks: user.SocialLinks,
    }, nil
}


// ============================================================
// MEDIA UPLOAD METHODS
// ============================================================

// uploadMedia uploads a file to the media service
func (s *profileService) uploadMedia(ctx context.Context, cmd domain.UploadMediaCommand) (*domain.MediaInfo, error) {
    media, err := s.mediaSvc.UploadFile(ctx, cmd)
    if err != nil {
        return nil, err
    }
    return media, nil
}

// internal/modules/profile/service/service_impl.go

// UploadUserAvatar uploads an avatar for a user
func (s *profileService) UploadUserAvatar(ctx context.Context, userID string, file []byte, filename, contentType string) (*domain.UserInfo, error) {
    if userID == "" {
        return nil, domain.ErrInvalidUserID
    }

    // 1. Validate file
    if err := s.validateImageFile(file, filename, contentType); err != nil {
        return nil, err
    }

    // 2. Check permission
    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }
    if viewerID != userID {
        scope := domain.NewPersonalTeamScope(viewerID)
        allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, scope)
        if err != nil {
            return nil, fmt.Errorf("permission check failed: %w", err)
        }
        if !allowed {
            return nil, domain.ErrPermissionDenied
        }
    }

    // 3. Get user
    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, domain.ErrUserNotFound
    }

    // 4. Delete old avatar if exists
    if user.AvatarURL != "" {
        log.Printf("🗑️ Deleting old avatar: %s", user.AvatarURL)
        mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_profile")
        if err != nil {
            log.Printf("⚠️ Failed to get media type: %v", err)
        } else if mediaType != nil {
            if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, userID, mediaType.ID); err != nil {
                log.Printf("⚠️ Failed to delete old avatar: %v", err)
            } else {
                log.Printf("✅ Old avatar deleted")
            }
        }
    }

    // ✅ 5. Generate clean filename - just the user ID and extension
    ext := filepath.Ext(filename)
    if ext == "" {
        // If no extension, use .jpg as default (but preserve original if possible)
        ext = ".jpg"
    }
    
    // Clean filename: just user_id + extension
    // Example: 058dfcd9-d152-4634-8e2a-02e8315b5d4c.jpg
    cleanFilename := userID + ext

    // 6. Upload to storage
    uploadCmd := domain.UploadMediaCommand{
        File:          file,
        FileName:      cleanFilename,  // ✅ Clean filename
        ContentType:   contentType,
        MediaTypeName: "media_type_profile",
        EntityID:      userID,
        UploadedBy:    viewerID,
    }

    mediaInfo, err := s.uploadMedia(ctx, uploadCmd)
    if err != nil {
        return nil, fmt.Errorf("failed to upload avatar: %w", err)
    }

    // 7. Update user with avatar URL
    user.AvatarURL = mediaInfo.URL
    if err := s.repo.UpdateUser(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }

    log.Printf("✅ User avatar uploaded for user: %s", userID)

    // 8. Return updated user info
    return &domain.UserInfo{
        ID:          user.ID,
        Name:        user.Name,
        DisplayName: user.DisplayName,
        Email:       user.Email,
        Phone:       user.Phone,
        AccountType: user.AccountType,
        AvatarURL:   user.AvatarURL,
        Bio:         user.Bio,
        Location:    user.Location,
        Website:     user.Website,
        SocialLinks: user.SocialLinks,
        IsActive:    user.IsActive,
        CreatedAt:   user.CreatedAt,
        UpdatedAt:   user.UpdatedAt,
    }, nil
}

// UploadInstitutionLogo uploads a logo for an institution
func (s *profileService) UploadInstitutionLogo(ctx context.Context, institutionID string, file []byte, filename, contentType string) (*domain.InstitutionInfo, error) {
    if institutionID == "" {
        return nil, domain.ErrInvalidInstitutionID
    }

    // 1. Validate file
    if err := s.validateImageFile(file, filename, contentType); err != nil {
        return nil, err
    }

    // 2. Get viewer ID
    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return nil, domain.ErrPermissionDenied
    }

    // 3. Check if viewer is an account admin in this institution
    scope := domain.NewInstitutionTeamScope(institutionID)
    
    // ✅ Check if user has account_admin role in this institution
    hasRole, err := s.permChecker.IsTeamAdmin(ctx, viewerID, scope)
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    if !hasRole {
        return nil, domain.ErrPermissionDenied
    }

    // 4. Get institution
    institution, err := s.repo.GetInstitutionByID(ctx, institutionID)
    if err != nil {
        return nil, err
    }
    if institution == nil {
        return nil, domain.ErrInstitutionNotFound
    }

    // 5. Delete old logo if exists
    if institution.LogoURL != "" {
        log.Printf("🗑️ Deleting old logo: %s", institution.LogoURL)
        mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_business")
        if err != nil {
            log.Printf("⚠️ Failed to get media type: %v", err)
        } else if mediaType != nil {
            if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, institutionID, mediaType.ID); err != nil {
                log.Printf("⚠️ Failed to delete old logo: %v", err)
            } else {
                log.Printf("✅ Old logo deleted")
            }
        }
    }

    // 6. Generate clean filename
    ext := filepath.Ext(filename)
    if ext == "" {
        ext = ".png"
    }
    cleanFilename := institutionID + ext

    // 7. Upload to storage
    uploadCmd := domain.UploadMediaCommand{
        File:          file,
        FileName:      cleanFilename,
        ContentType:   contentType,
        MediaTypeName: "media_type_business",
        EntityID:      institutionID,
        UploadedBy:    viewerID,
    }

    mediaInfo, err := s.uploadMedia(ctx, uploadCmd)
    if err != nil {
        return nil, fmt.Errorf("failed to upload logo: %w", err)
    }

    // 8. Update institution with logo URL
    institution.LogoURL = mediaInfo.URL
    if err := s.repo.UpdateInstitution(ctx, institution); err != nil {
        return nil, fmt.Errorf("failed to update institution: %w", err)
    }

    log.Printf("✅ Institution logo uploaded for institution: %s", institutionID)

    return &domain.InstitutionInfo{
        ID:          institution.ID,
        Name:        institution.Name,
        DisplayName: institution.DisplayName,
        Slug:        institution.Slug,
        Email:       institution.Email,
        Phone:       institution.Phone,
        Website:     institution.Website,
        Description: institution.Description,
        LogoURL:     institution.LogoURL,
        Address:     institution.Address,
        City:        institution.City,
        Country:     institution.Country,
    }, nil
}

// ============================================================
// MEDIA - DELETE METHODS
// ============================================================

// DeleteUserAvatar deletes a user's avatar
func (s *profileService) DeleteUserAvatar(ctx context.Context, userID string) error {
    if userID == "" {
        return domain.ErrInvalidUserID
    }

    // 1. Check permission
    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return domain.ErrPermissionDenied
    }
    if viewerID != userID {
        scope := domain.NewPersonalTeamScope(viewerID)
        allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, scope)
        if err != nil {
            return fmt.Errorf("permission check failed: %w", err)
        }
        if !allowed {
            return domain.ErrPermissionDenied
        }
    }

    // 2. Get user
    user, err := s.repo.GetUserByID(ctx, userID)
    if err != nil {
        return err
    }
    if user == nil {
        return domain.ErrUserNotFound
    }

    // 3. ✅ Delete from storage using media type (like events module)
    if user.AvatarURL != "" {
        log.Printf("🗑️ Deleting avatar for user: %s", userID)
        
        mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_profile")
        if err != nil {
            log.Printf("⚠️ Failed to get media type: %v", err)
        } else if mediaType != nil {
            if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, userID, mediaType.ID); err != nil {
                log.Printf("⚠️ Failed to delete avatar files: %v", err)
                // Continue to clear URL even if storage deletion fails
            } else {
                log.Printf("✅ Avatar files deleted from storage")
            }
        }
    } else {
        log.Printf("ℹ️ No avatar URL to delete")
    }

    // 4. Clear avatar URL in database
    user.AvatarURL = ""
    if err := s.repo.UpdateUser(ctx, user); err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }

    log.Printf("✅ User avatar deleted for user: %s", userID)
    return nil
}


func (s *profileService) DeleteInstitutionLogo(ctx context.Context, institutionID string) error {
    if institutionID == "" {
        return domain.ErrInvalidInstitutionID
    }

    // 1. Check permission - must be account admin
    viewerID := s.getUserIDFromContext(ctx)
    if viewerID == "" {
        return domain.ErrPermissionDenied
    }

    scope := domain.NewInstitutionTeamScope(institutionID)
    
    // ✅ Check if user has account_admin role in this institution
    hasRole, err := s.permChecker.IsTeamAdmin(ctx, viewerID, scope)
    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }
    if !hasRole {
        return domain.ErrPermissionDenied
    }

    // 2. Get institution
    institution, err := s.repo.GetInstitutionByID(ctx, institutionID)
    if err != nil {
        return err
    }
    if institution == nil {
        return domain.ErrInstitutionNotFound
    }

    // 3. Delete file from storage
    if institution.LogoURL != "" {
        log.Printf("🗑️ Deleting logo for institution: %s", institutionID)
        
        mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, "media_type_business")
        if err != nil {
            log.Printf("⚠️ Failed to get media type: %v", err)
        } else if mediaType != nil {
            if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, institutionID, mediaType.ID); err != nil {
                log.Printf("⚠️ Failed to delete logo files: %v", err)
            } else {
                log.Printf("✅ Logo files deleted from storage")
            }
        }
    }

    // 4. Clear logo URL in database
    institution.LogoURL = ""
    if err := s.repo.UpdateInstitution(ctx, institution); err != nil {
        return fmt.Errorf("failed to update institution: %w", err)
    }

    log.Printf("✅ Institution logo deleted for institution: %s", institutionID)
    return nil
}
// ============================================================
// PRIVATE HELPERS
// ============================================================

// validateImageFile validates an image file
func (s *profileService) validateImageFile(file []byte, filename, contentType string) error {
    // Check file size (max 5MB)
    maxSize := 5 * 1024 * 1024 // 5MB
    if len(file) > maxSize {
        return fmt.Errorf("file size exceeds maximum allowed size of 5MB")
    }

    // Check content type (allow common image types)
    allowedTypes := map[string]bool{
        "image/jpeg":    true,
        "image/png":     true,
        "image/gif":     true,
        "image/webp":    true,
        "image/svg+xml": true,
        "image/bmp":     true,
    }
    if !allowedTypes[contentType] {
        return fmt.Errorf("unsupported file type: %s. Allowed types: JPEG, PNG, GIF, WEBP, SVG, BMP", contentType)
    }

    // Check file extension
    ext := strings.ToLower(filepath.Ext(filename))
    allowedExtensions := map[string]bool{
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
        ".gif":  true,
        ".webp": true,
        ".svg":  true,
        ".bmp":  true,
    }
    if !allowedExtensions[ext] {
        return fmt.Errorf("unsupported file extension: %s. Allowed extensions: .jpg, .jpeg, .png, .gif, .webp, .svg, .bmp", ext)
    }

    return nil
}