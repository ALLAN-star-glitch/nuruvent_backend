// internal/modules/profile/service/organizer.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// ============================================================
// ORGANIZER INFO (FOR EVENTS MODULE)
// ============================================================

// GetOrganizerInfo returns organizer information for the events module
func (s *profileService) GetOrganizerInfo(ctx context.Context, organizerType string, organizerID string) (*domain.OrganizerInfo, error) {
	if organizerID == "" {
		return nil, domain.ErrInvalidOrganizerID
	}

	switch organizerType {
	case "institution":
		account, err := s.repo.GetAccountByID(ctx, organizerID)
		if err != nil {
			return nil, err
		}
		if account == nil {
			return nil, domain.ErrAccountNotFound
		}
		return &domain.OrganizerInfo{
			ID:          account.ID,
			Name:        account.Name,
			DisplayName: account.DisplayName,
			Type:        "institution",
			AvatarURL:   account.LogoURL,
			Slug:        account.Slug,
		}, nil

	case "personal":
		user, err := s.repo.GetUserByID(ctx, organizerID)
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

	default:
		return nil, domain.ErrInvalidOrganizerType
	}
}