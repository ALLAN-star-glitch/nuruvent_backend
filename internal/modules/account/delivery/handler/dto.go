package handler

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
)

// internal/modules/account/delivery/handler/dto.go

type UserInfoResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	AvatarURL   string `json:"avatar_url"`
}

func NewUserInfoResponse(u *accountdomain.UserInfo) UserInfoResponse {
	if u == nil {
		return UserInfoResponse{}
	}
	return UserInfoResponse{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Phone:       u.Phone,
		AvatarURL:   u.AvatarURL,
	}
}


// UpdateProfileRequest carries partial profile updates from the client.
// nil fields are "not provided" — the service leaves them unchanged.
type UpdateProfileRequest struct {
	DisplayName *string            `json:"display_name"`
	Phone       *string            `json:"phone"`
	Bio         *string            `json:"bio"`
	Location    *string            `json:"location"`
	Website     *string            `json:"website"`
	SocialLinks *map[string]string `json:"social_links"`
}

// ToService converts the request into the service-layer command.
func (r UpdateProfileRequest) ToService() service.ProfileUpdates {
	return service.ProfileUpdates{
		DisplayName: r.DisplayName,
		Phone:       r.Phone,
		Bio:         r.Bio,
		Location:    r.Location,
		Website:     r.Website,
		SocialLinks: r.SocialLinks,
	}
}

// ProfileResponse is the full profile returned to the user themselves.
type ProfileResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Email       string            `json:"email"`
	Phone       string            `json:"phone"`
	AvatarURL   string            `json:"avatar_url"`
	Bio         string            `json:"bio"`
	Location    string            `json:"location"`
	Website     string            `json:"website"`
	SocialLinks map[string]string `json:"social_links"`
}

func NewProfileResponse(u *accountdomain.UserInfo) ProfileResponse {
	if u == nil {
		return ProfileResponse{}
	}
	return ProfileResponse{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Phone:       u.Phone,
		AvatarURL:   u.AvatarURL,
		Bio:         u.Bio,
		Location:    u.Location,
		Website:     u.Website,
		SocialLinks: u.SocialLinks,
	}
}

// PublicProfileResponse is the reduced profile shown to other users.
type PublicProfileResponse struct {
	ID          string            `json:"id"`
	DisplayName string            `json:"display_name"`
	AvatarURL   string            `json:"avatar_url"`
	Bio         string            `json:"bio"`
	Location    string            `json:"location"`
	Website     string            `json:"website"`
	SocialLinks map[string]string `json:"social_links"`
}

func NewPublicProfileResponse(p *accountdomain.PublicProfile) PublicProfileResponse {
	if p == nil {
		return PublicProfileResponse{}
	}
	return PublicProfileResponse{
		ID:          p.ID,
		DisplayName: p.DisplayName,
		AvatarURL:   p.AvatarURL,
		Bio:         p.Bio,
		Location:    p.Location,
		Website:     p.Website,
		SocialLinks: p.SocialLinks,
	}
}