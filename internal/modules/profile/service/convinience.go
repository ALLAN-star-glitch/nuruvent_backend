// internal/modules/profile/service/convenience.go

package service

import (
    "context"
    "log"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// GetOrganizerInfoSafe returns organizer info without errors (logs instead)
func (s *profileService) GetOrganizerInfoSafe(ctx context.Context, scope domain.Scope) *domain.OrganizerInfo {
    organizer, err := s.GetOrganizerInfo(ctx, scope)
    if err != nil {
        log.Printf("⚠️ Failed to get organizer info for scope %s: %v", scope.String(), err)
        return nil
    }
    return organizer
}

// GetUserProfileSafe returns user profile without errors (logs instead)
func (s *profileService) GetUserProfileSafe(ctx context.Context, userID string) *domain.UserInfo {
    user, err := s.GetUserProfile(ctx, userID)
    if err != nil {
        log.Printf("⚠️ Failed to get user profile for %s: %v", userID, err)
        return nil
    }
    return user
}

// GetUserProfileWithDetailsSafe returns full user profile without errors (logs instead)
func (s *profileService) GetUserProfileWithDetailsSafe(ctx context.Context, userID string) *domain.UserInfo {
    user, err := s.GetUserProfileWithDetails(ctx, userID)
    if err != nil {
        log.Printf("⚠️ Failed to get full user profile for %s: %v", userID, err)
        return nil
    }
    return user
}

// GetInstitutionProfileSafe returns institution profile without errors (logs instead)
func (s *profileService) GetInstitutionProfileSafe(ctx context.Context, institutionID string) *domain.InstitutionInfo {
    institution, err := s.GetInstitutionProfile(ctx, institutionID)
    if err != nil {
        log.Printf("⚠️ Failed to get institution profile for %s: %v", institutionID, err)
        return nil
    }
    return institution
}

// GetInstitutionProfileWithDetailsSafe returns full institution profile without errors (logs instead)
func (s *profileService) GetInstitutionProfileWithDetailsSafe(ctx context.Context, institutionID string) *domain.InstitutionInfo {
    institution, err := s.GetInstitutionProfileWithDetails(ctx, institutionID)
    if err != nil {
        log.Printf("⚠️ Failed to get full institution profile for %s: %v", institutionID, err)
        return nil
    }
    return institution
}

// GetUserProfilesSafe returns multiple user profiles without errors (logs instead)
func (s *profileService) GetUserProfilesSafe(ctx context.Context, userIDs []string) []*domain.UserInfo {
    users, err := s.GetUserProfiles(ctx, userIDs)
    if err != nil {
        log.Printf("⚠️ Failed to get user profiles: %v", err)
        return []*domain.UserInfo{}
    }
    return users
}

// GetInstitutionProfilesSafe returns multiple institution profiles without errors (logs instead)
func (s *profileService) GetInstitutionProfilesSafe(ctx context.Context, institutionIDs []string) []*domain.InstitutionInfo {
    institutions, err := s.GetInstitutionProfiles(ctx, institutionIDs)
    if err != nil {
        log.Printf("⚠️ Failed to get institution profiles: %v", err)
        return []*domain.InstitutionInfo{}
    }
    return institutions
}