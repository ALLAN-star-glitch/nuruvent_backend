// internal/modules/events/domain/user.go

package domain

import "time"

// UserInfo represents basic user information for event creators
type UserInfo struct {
	ID              string
	Name            string
	DisplayName     string
	Email           string
	Phone           string
	Username        string
	AccountType     string
	InstitutionID   string
	InstitutionName string
	Slug            string
	AvatarURL       string // ✅ Added
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewUserInfo creates a new UserInfo instance
func NewUserInfo(id, name, email string) *UserInfo {
	return &UserInfo{
		ID:        id,
		Name:      name,
		Email:     email,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// WithInstitution sets the institution details
func (u *UserInfo) WithInstitution(id, name string) *UserInfo {
	u.InstitutionID = id
	u.InstitutionName = name
	return u
}

// WithDisplayName sets the display name
func (u *UserInfo) WithDisplayName(displayName string) *UserInfo {
	u.DisplayName = displayName
	return u
}

// WithAccountType sets the account type
func (u *UserInfo) WithAccountType(accountType string) *UserInfo {
	u.AccountType = accountType
	return u
}

// WithUsername sets the username
func (u *UserInfo) WithUsername(username string) *UserInfo {
	u.Username = username
	return u
}

// WithPhone sets the phone number
func (u *UserInfo) WithPhone(phone string) *UserInfo {
	u.Phone = phone
	return u
}

// WithAvatarURL sets the avatar URL
func (u *UserInfo) WithAvatarURL(avatarURL string) *UserInfo {
	u.AvatarURL = avatarURL
	return u
}

// IsInstitution checks if the user is an institution account
func (u *UserInfo) IsInstitution() bool {
	return u.AccountType == "account_type_institution"
}

// IsPersonal checks if the user is a personal account
func (u *UserInfo) IsPersonal() bool {
	return u.AccountType == "account_type_personal"
}

// GetDisplayName returns the best available display name
func (u *UserInfo) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.InstitutionName != "" {
		return u.InstitutionName
	}
	return u.Name
}

// HasInstitution checks if the user belongs to an institution
func (u *UserInfo) HasInstitution() bool {
	return u.InstitutionID != ""
}

// GetFullName returns the full name (or display name if available)
func (u *UserInfo) GetFullName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Name
}

// GetAvatarURL returns the avatar URL or empty string if not set
func (u *UserInfo) GetAvatarURL() string {
	return u.AvatarURL
}