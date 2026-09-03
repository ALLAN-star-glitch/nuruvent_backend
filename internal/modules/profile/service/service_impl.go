// internal/modules/profile/service/service_impl.go

package service

import (
    "context"
    "fmt"
    "log"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

type profileService struct {
    repo        domain.Repository
    permChecker domain.PermissionChecker
}

func NewProfileService(repo domain.Repository, permChecker domain.PermissionChecker) domain.Service {
    return &profileService{
        repo:        repo,
        permChecker: permChecker,
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