// internal/modules/events/domain/account.go

package domain

import "time"

// AccountInfo represents basic account information for event creators
type AccountInfo struct {
    ID              string
    Name            string
    DisplayName     string
    Email           string
    Phone           string
    AccountType     string
    InstitutionID   string
    InstitutionName string
    Slug            string
    IsActive        bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// NewAccountInfo creates a new AccountInfo instance
func NewAccountInfo(id, name, email string) *AccountInfo {
    return &AccountInfo{
        ID:        id,
        Name:      name,
        Email:     email,
        IsActive:  true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// WithInstitution sets the institution details
func (a *AccountInfo) WithInstitution(id, name string) *AccountInfo {
    a.InstitutionID = id
    a.InstitutionName = name
    return a
}

// WithDisplayName sets the display name
func (a *AccountInfo) WithDisplayName(displayName string) *AccountInfo {
    a.DisplayName = displayName
    return a
}

// WithAccountType sets the account type
func (a *AccountInfo) WithAccountType(accountType string) *AccountInfo {
    a.AccountType = accountType
    return a
}

// IsInstitution checks if the account is an institution
func (a *AccountInfo) IsInstitution() bool {
    return a.AccountType == "institution"
}

// GetDisplayName returns the best available display name
func (a *AccountInfo) GetDisplayName() string {
    if a.DisplayName != "" {
        return a.DisplayName
    }
    if a.InstitutionName != "" {
        return a.InstitutionName
    }
    return a.Name
}