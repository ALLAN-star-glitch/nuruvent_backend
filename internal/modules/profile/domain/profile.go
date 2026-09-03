// internal/modules/profile/domain/profile.go

package domain

import "time"

// ============================================================
// FULL DOMAIN MODELS
// ============================================================

// User represents a user profile
type User struct {
    ID          string
    Name        string
    DisplayName string
	Slug		string
    Email       string
    Phone       string
    AccountType string
    AvatarURL   string
    Bio         string
    Location    string
    Website     string
    SocialLinks map[string]string
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Institution represents an institution profile
type Institution struct {
    ID          string
    Name        string
    DisplayName string
    Slug        string
    Email       string
    Phone       string
    Website     string
    Description string
    LogoURL     string
    Address     string
    City        string
    Country     string
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// ============================================================
// MINIMAL DTOS FOR CROSS-MODULE USAGE
// ============================================================

// UserInfo (minimal) for cross-module usage (events, etc.)
type UserInfo struct {
    ID          string
    Name        string
    DisplayName string
    Email       string
    Phone       string
    AccountType string
    AvatarURL   string
    Bio         string
    Location    string
    Website     string
    SocialLinks map[string]string
}

// InstitutionInfo (minimal) for cross-module usage (events, etc.)
type InstitutionInfo struct {
    ID          string
    Name        string
    DisplayName string
    Slug        string
    Email       string
    Phone       string
    Website     string
    Description string
    LogoURL     string
    Address     string
    City        string
    Country     string
}

// ============================================================
// PUBLIC-FACING ORGANIZER INFO (FOR EVENTS MODULE)
// ============================================================

// OrganizerInfo represents public-facing organizer information
// Used by events module to display who is organizing an event
type OrganizerInfo struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    Type        string `json:"type"` // "institution" or "personal"
    AvatarURL   string `json:"avatar_url,omitempty"`
    Slug        string `json:"slug,omitempty"`
}