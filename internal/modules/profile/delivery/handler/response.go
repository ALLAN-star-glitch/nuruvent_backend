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
		Email:       user.Email,
		Phone:       user.Phone,
		AccountType: user.AccountType,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
		Location:    user.Location,
		Website:     user.Website,
		SocialLinks: user.SocialLinks,
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
// INSTITUTION PROFILE RESPONSES
// ============================================================

// InstitutionProfileResponse is the API response for institution profile
type InstitutionProfileResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// NewInstitutionProfileResponse creates a new InstitutionProfileResponse from domain InstitutionInfo
func NewInstitutionProfileResponse(institution *domain.InstitutionInfo) InstitutionProfileResponse {
	if institution == nil {
		return InstitutionProfileResponse{}
	}
	return InstitutionProfileResponse{
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
	}
}

// NewFullInstitutionProfileResponse creates a full InstitutionProfileResponse from domain Institution
func NewFullInstitutionProfileResponse(institution *domain.Institution) InstitutionProfileResponse {
	if institution == nil {
		return InstitutionProfileResponse{}
	}
	return InstitutionProfileResponse{
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
		IsActive:    institution.IsActive,
		CreatedAt:   institution.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   institution.UpdatedAt.Format(time.RFC3339),
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