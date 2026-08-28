// internal/modules/events/domain/account.go

package domain

import "time"

// UserInfo represents basic user information for event creators
// (formerly AccountInfo)
type UserInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Username        string `json:"username"`
	AccountType     string `json:"account_type"`
	InstitutionID   string `json:"institution_id"`
	InstitutionName string `json:"institution_name"`
	Slug            string `json:"slug"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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

// ============================================================
// DEPRECATED - Keep for backward compatibility during migration
// ============================================================

// AccountInfo is deprecated. Use UserInfo instead.
// Deprecated: Renamed to UserInfo to align with users table
type AccountInfo = UserInfo

// NewAccountInfo is deprecated. Use NewUserInfo instead.
// Deprecated: Renamed to NewUserInfo to align with users table
func NewAccountInfo(id, name, email string) *AccountInfo {
	return NewUserInfo(id, name, email)
}