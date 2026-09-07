// internal/modules/events/domain/organizer.go

package domain

// OrganizerInfo represents public-facing organizer information
type OrganizerInfo struct {
	ID          string
	Name        string
	DisplayName string
	Type        string // "institution" or "personal"
	AvatarURL   string
	Slug        string
}

// AccountInfo represents account information for the events module
// Replaces InstitutionInfo
type AccountInfo struct {
	ID          string
	Name        string
	DisplayName string
	Slug        string
	Type        string // "personal" or "institution"
	Email       string
	Phone       string
	LogoURL     string
	Website     string
	Description string
}