// internal/modules/profile/delivery/handler/request.go

package handler

// ============================================================
// REQUEST DTOS
// ============================================================

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name        string            `json:"name,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	Location    string            `json:"location,omitempty"`
	Website     string            `json:"website,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
}

// ToMap converts UpdateProfileRequest to a map of updates
func (r *UpdateProfileRequest) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if r.Name != "" {
		updates["name"] = r.Name
	}
	if r.DisplayName != "" {
		updates["display_name"] = r.DisplayName
	}
	if r.Phone != "" {
		updates["phone"] = r.Phone
	}
	if r.AvatarURL != "" {
		updates["avatar_url"] = r.AvatarURL
	}
	if r.Bio != "" {
		updates["bio"] = r.Bio
	}
	if r.Location != "" {
		updates["location"] = r.Location
	}
	if r.Website != "" {
		updates["website"] = r.Website
	}
	if r.SocialLinks != nil {
		updates["social_links"] = r.SocialLinks
	}
	return updates
}

// UpdateInstitutionRequest represents an institution update request
type UpdateInstitutionRequest struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
}

// ToMap converts UpdateInstitutionRequest to a map of updates
func (r *UpdateInstitutionRequest) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if r.Name != "" {
		updates["name"] = r.Name
	}
	if r.DisplayName != "" {
		updates["display_name"] = r.DisplayName
	}
	if r.Email != "" {
		updates["email"] = r.Email
	}
	if r.Phone != "" {
		updates["phone"] = r.Phone
	}
	if r.Website != "" {
		updates["website"] = r.Website
	}
	if r.Description != "" {
		updates["description"] = r.Description
	}
	if r.LogoURL != "" {
		updates["logo_url"] = r.LogoURL
	}
	if r.Address != "" {
		updates["address"] = r.Address
	}
	if r.City != "" {
		updates["city"] = r.City
	}
	if r.Country != "" {
		updates["country"] = r.Country
	}
	return updates
}

// ============================================================
// LIST REQUEST DTOS
// ============================================================

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	// Team Filter (from query params)
	TeamID   string `json:"team_id" query:"team_id"`
	TeamType string `json:"team_type" query:"team_type"` // "personal" or "institution"

	// User Filters
	UserID string `json:"user_id" query:"user_id"`
	Search string `json:"search" query:"search"`

	// Deletion Filters
	IncludeDeleted bool `json:"include_deleted" query:"include_deleted"`
	OnlyDeleted    bool `json:"only_deleted" query:"only_deleted"`

	// Pagination
	Limit  int `json:"limit" query:"limit"`
	Offset int `json:"offset" query:"offset"`

	// Sorting
	SortBy    string `json:"sort_by" query:"sort_by"`
	SortOrder string `json:"sort_order" query:"sort_order"` // asc, desc
}

// ListInstitutionsRequest represents a request to list institutions
type ListInstitutionsRequest struct {
	// Team Filter (from query params)
	TeamID   string `json:"team_id" query:"team_id"`
	TeamType string `json:"team_type" query:"team_type"` // "institution"

	// Institution Filters
	InstitutionID string `json:"institution_id" query:"institution_id"`
	Search        string `json:"search" query:"search"`

	// Deletion Filters
	IncludeDeleted bool `json:"include_deleted" query:"include_deleted"`
	OnlyDeleted    bool `json:"only_deleted" query:"only_deleted"`

	// Pagination
	Limit  int `json:"limit" query:"limit"`
	Offset int `json:"offset" query:"offset"`

	// Sorting
	SortBy    string `json:"sort_by" query:"sort_by"`
	SortOrder string `json:"sort_order" query:"sort_order"` // asc, desc
}

// GetOrganizerInfoRequest represents a request to get organizer info
type GetOrganizerInfoRequest struct {
	Scope string `json:"scope" query:"scope"` // "personal:user_id" or "institution:institution_id"
}