// internal/modules/profile/service/user_profile.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// ============================================================
// USER PROFILE METHODS
// ============================================================

// GetUserProfile returns basic user profile information
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
		Slug:        user.Slug,
		AvatarURL:   user.AvatarURL,
	}, nil
}

// GetUserProfileWithDetails returns detailed user profile information with permission checks
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

	// Get target user's account domains
	targetAccountDomains, err := s.permChecker.GetUserTeamDomains(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user's account memberships: %w", err)
	}

	// Check if viewer has read_all permission in any shared account
	for _, domainStr := range targetAccountDomains {
		if domain.IsAccountDomain(domainStr) {
			allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, domainStr)
			if err != nil {
				continue
			}
			if allowed {
				return s.getUserDetails(ctx, userID)
			}
		}
	}

	// Check if viewer has read_all in their personal account
	viewerPersonalDomain := domain.PersonalTeamDomain(viewerID)
	if viewerPersonalDomain != "" {
		allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, viewerPersonalDomain)
		if err == nil && allowed {
			return s.getUserDetails(ctx, userID)
		}
	}

	// Finally, check if viewer has read permission in any shared account
	for _, domainStr := range targetAccountDomains {
		if domain.IsAccountDomain(domainStr) {
			allowed, err := s.permChecker.CanReadProfile(ctx, viewerID, domainStr)
			if err != nil {
				continue
			}
			if allowed {
				return s.getUserDetails(ctx, userID)
			}
		}
	}

	return nil, domain.ErrPermissionDenied
}

// GetUserProfiles returns multiple user profiles
func (s *profileService) GetUserProfiles(ctx context.Context, userIDs []string) ([]*domain.UserInfo, error) {
	if len(userIDs) == 0 {
		return []*domain.UserInfo{}, nil
	}

	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, domain.ErrPermissionDenied
	}

	// Check if viewer has read_all permission in their personal account
	viewerPersonalDomain := domain.PersonalTeamDomain(viewerID)
	if viewerPersonalDomain != "" {
		allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, viewerPersonalDomain)
		if err != nil {
			return nil, fmt.Errorf("permission check failed: %w", err)
		}
		if allowed {
			return s.getUsersByIDs(ctx, userIDs)
		}
	}

	// If requesting only their own profile, allow it
	if len(userIDs) == 1 && userIDs[0] == viewerID {
		return s.getUsersByIDs(ctx, userIDs)
	}

	return nil, domain.ErrPermissionDenied
}

// UpdateUserProfile updates a user's profile
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

	// Get target user's account domains
	targetAccountDomains, err := s.permChecker.GetUserTeamDomains(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user's account memberships: %w", err)
	}

	// Check if viewer has update permission in any account where target user is a member
	for _, domainStr := range targetAccountDomains {
		if domain.IsAccountDomain(domainStr) {
			allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, domainStr)
			if err != nil {
				continue
			}
			if allowed {
				return s.updateUser(ctx, userID, updates)
			}
		}
	}

	return nil, domain.ErrPermissionDenied
}

// ListUsers returns a paginated list of users
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
		accountDomain := domain.AccountDomain(filters.Team.ID)
		if accountDomain == "" {
			return nil, 0, domain.ErrInvalidAccountID
		}
		allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, accountDomain)
		if err != nil {
			return nil, 0, fmt.Errorf("permission check failed: %w", err)
		}
		if !allowed {
			return nil, 0, domain.ErrPermissionDenied
		}
	}

	// If no team filter, check if viewer has read_all permission in personal account
	if filters.Team.Type == "" {
		viewerPersonalDomain := domain.PersonalTeamDomain(viewerID)
		if viewerPersonalDomain == "" {
			return nil, 0, domain.ErrInvalidUserID
		}
		allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, viewerPersonalDomain)
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
			Slug:        user.Slug,
			AvatarURL:   user.AvatarURL,
		}
	}
	return userInfos, total, nil
}

// ============================================================
// PRIVATE HELPERS FOR USER
// ============================================================

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
		Slug:        user.Slug,
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

// getUsersByIDs gets multiple users without permission checks
func (s *profileService) getUsersByIDs(ctx context.Context, userIDs []string) ([]*domain.UserInfo, error) {
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
			Slug:        user.Slug,
			AvatarURL:   user.AvatarURL,
		}
	}
	return userInfos, nil
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

	// Apply updates
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
	if accountType, ok := updates["account_type"].(string); ok {
		user.AccountType = accountType
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("✅ User profile updated: %s", userID)

	return &domain.UserInfo{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Slug:        user.Slug,
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