// internal/modules/profile/delivery/handler/response.go

package handler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// ============================================================
// USER PROFILE RESPONSES
// ============================================================

// UserProfileResponse is the API response for user profile
type UserProfileResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Slug        string            `json:"slug,omitempty"`
	Email       string            `json:"email,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	AccountType string            `json:"account_type,omitempty"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	Location    string            `json:"location,omitempty"`
	Website     string            `json:"website,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// NewUserProfileResponse creates a new UserProfileResponse from domain UserInfo
func NewUserProfileResponse(user *domain.UserInfo) UserProfileResponse {
	if user == nil {
		return UserProfileResponse{}
	}
	return UserProfileResponse{
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
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}
}

// NewFullUserProfileResponse creates a full UserProfileResponse from domain User
func NewFullUserProfileResponse(user *domain.User) UserProfileResponse {
	if user == nil {
		return UserProfileResponse{}
	}
	return UserProfileResponse{
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
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
	}
}

// ============================================================
// ACCOUNT PROFILE RESPONSES (replaces Institution)
// ============================================================

// AccountProfileResponse is the API response for account profile
type AccountProfileResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"` // "personal" or "institution"
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	Status      string `json:"status,omitempty"`
	KYCStatus   string `json:"kyc_status,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// NewAccountProfileResponse creates a new AccountProfileResponse from domain AccountInfo
func NewAccountProfileResponse(account *domain.AccountInfo) AccountProfileResponse {
	if account == nil {
		return AccountProfileResponse{}
	}
	return AccountProfileResponse{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		Email:       account.Email,
		Phone:       account.Phone,
		Website:     account.Website,
		Description: account.Description,
		LogoURL:     account.LogoURL,
		Address:     account.Address,
		City:        account.City,
		Country:     account.Country,
		Status:      account.Status,
		KYCStatus:   account.KYCStatus,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	}
}

// NewFullAccountProfileResponse creates a full AccountProfileResponse from domain Account
func NewFullAccountProfileResponse(account *domain.Account) AccountProfileResponse {
	if account == nil {
		return AccountProfileResponse{}
	}
	return AccountProfileResponse{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		Email:       account.Email,
		Phone:       account.Phone,
		Website:     account.Website,
		Description: account.Description,
		LogoURL:     account.LogoURL,
		Address:     account.Address,
		City:        account.City,
		Country:     account.Country,
		Status:      account.Status,
		KYCStatus:   account.KYCStatus,
		IsActive:    account.IsActive,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	}
}

// ============================================================
// ORGANIZER INFO RESPONSE
// ============================================================

// OrganizerInfoResponse is the API response for organizer info
type OrganizerInfoResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"` // "institution" or "personal"
	AvatarURL   string `json:"avatar_url,omitempty"`
	Slug        string `json:"slug,omitempty"`
}

// NewOrganizerInfoResponse creates a new OrganizerInfoResponse from domain OrganizerInfo
func NewOrganizerInfoResponse(organizer *domain.OrganizerInfo) OrganizerInfoResponse {
	if organizer == nil {
		return OrganizerInfoResponse{}
	}
	return OrganizerInfoResponse{
		ID:          organizer.ID,
		Name:        organizer.Name,
		DisplayName: organizer.DisplayName,
		Type:        organizer.Type,
		AvatarURL:   organizer.AvatarURL,
		Slug:        organizer.Slug,
	}
}