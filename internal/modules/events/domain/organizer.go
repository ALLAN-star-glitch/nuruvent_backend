// internal/modules/events/domain/organizer.go

package domain

// OrganizerInfo represents public-facing organizer information
type OrganizerInfo struct {
	ID          string
	Name        string
	DisplayName string
	Type        string // "institution" or "personal"
	AvatarURL   string // ✅ Added
	Slug        string
}

// InstitutionInfo represents institution information
type InstitutionInfo struct {
	ID          string
	Name        string
	DisplayName string
	Slug        string
	Email       string
	Phone       string
	LogoURL     string // ✅ This is the institution's avatar/logo
	Website     string
	Description string
}