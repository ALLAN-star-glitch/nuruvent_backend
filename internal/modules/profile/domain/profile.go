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
	Slug        string
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

// Account represents an account (personal or institution)
type Account struct {
	ID          string
	Name        string
	DisplayName string
	Slug        string
	Type        string // "personal" or "institution"
	Email       string
	Phone       string
	Website     string
	Description string
	LogoURL     string
	Address     string
	City        string
	Country     string
	Status      string // "active", "suspended", "inactive"
	KYCStatus   string // "pending", "submitted", "verified", "rejected", "not_required"
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
	Slug        string
	Email       string
	Phone       string
	AccountType string
	AvatarURL   string
	Bio         string
	Location    string
	Website     string
	IsActive    bool
	SocialLinks map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AccountInfo (minimal) for cross-module usage (events, etc.)
type AccountInfo struct {
	ID          string
	Name        string
	DisplayName string
	Slug        string
	Type        string // "personal" or "institution"
	Email       string
	Phone       string
	Website     string
	Description string
	LogoURL     string
	Address     string
	City        string
	Country     string
	Status      string
	KYCStatus   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ============================================================
// PUBLIC-FACING ORGANIZER INFO (FOR EVENTS MODULE)
// ============================================================

// OrganizerInfo represents public-facing organizer information
// Used by events module to display who is organizing an event
type OrganizerInfo struct {
	ID          string
	Name        string
	DisplayName string
	Type        string // "personal" or "institution"
	AvatarURL   string
	Slug        string
}

// ============================================================
// TEAM TYPE (For team member management)
// ============================================================

// TeamType represents a team type/role
type TeamType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	IsActive    bool
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}