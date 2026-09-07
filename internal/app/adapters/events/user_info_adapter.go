// internal/app/adapters/events/user_info_adapter.go

package events

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	profileService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/service"
)

// UserInfoAdapter adapts the profile module's Service to the events domain's UserInfoProvider
type UserInfoAdapter struct {
	profileSvc profileService.Service
}

// NewUserInfoAdapter creates a new user info adapter for events
func NewUserInfoAdapter(profileSvc profileService.Service) domain.UserInfoProvider {
	return &UserInfoAdapter{
		profileSvc: profileSvc,
	}
}

// GetUserByID retrieves basic user information by ID
func (a *UserInfoAdapter) GetUserByID(ctx context.Context, userID string) (*domain.UserInfo, error) {
	if userID == "" {
		return nil, nil
	}

	userInfo, err := a.profileSvc.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userInfo == nil {
		return nil, nil
	}

	return &domain.UserInfo{
		ID:          userInfo.ID,
		Name:        userInfo.Name,
		DisplayName: userInfo.DisplayName,
		AvatarURL:   userInfo.AvatarURL,
	}, nil
}

// GetUserByIDWithDetails retrieves full user information including email, phone, etc.
func (a *UserInfoAdapter) GetUserByIDWithDetails(ctx context.Context, userID string) (*domain.UserInfo, error) {
	if userID == "" {
		return nil, nil
	}

	userInfo, err := a.profileSvc.GetUserProfileWithDetails(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userInfo == nil {
		return nil, nil
	}

	return &domain.UserInfo{
		ID:          userInfo.ID,
		Name:        userInfo.Name,
		DisplayName: userInfo.DisplayName,
		Email:       userInfo.Email,
		Phone:       userInfo.Phone,
		AccountType: userInfo.AccountType,
		AvatarURL:   userInfo.AvatarURL,
		IsActive:    userInfo.IsActive,
		CreatedAt:   userInfo.CreatedAt,
		UpdatedAt:   userInfo.UpdatedAt,
	}, nil
}

// GetAccountByID retrieves account information by ID
func (a *UserInfoAdapter) GetAccountByID(ctx context.Context, accountID string) (*domain.AccountInfo, error) {
	if accountID == "" {
		return nil, nil
	}

	accountInfo, err := a.profileSvc.GetAccountProfile(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if accountInfo == nil {
		return nil, nil
	}

	return &domain.AccountInfo{
		ID:          accountInfo.ID,
		Name:        accountInfo.Name,
		DisplayName: accountInfo.DisplayName,
		Slug:        accountInfo.Slug,
		Type:        accountInfo.Type,
		LogoURL:     accountInfo.LogoURL,
		Email:       accountInfo.Email,
		Phone:       accountInfo.Phone,
		Website:     accountInfo.Website,
		Description: accountInfo.Description,
	}, nil
}