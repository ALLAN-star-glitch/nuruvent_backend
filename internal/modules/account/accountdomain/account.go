// internal/modules/account/accountdomain/account.go

package accountdomain

import (
    "errors"
    "time"

    "github.com/google/uuid"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

// Account represents a personal or institution account
type Account struct {
    ID            string
    Name          string
    DisplayName   string
    Slug          string
    Email         string
    Phone         string
    AccountTypeID string
    Status        string
    LogoURL       string
    Website       string
    Description   string
    CreatedBy     string
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

// Account status constants
const (
    AccountStatusActive    = "active"
    AccountStatusSuspended = "suspended"
    AccountStatusInactive  = "inactive"
)

// NewPersonalAccount creates a new personal account
func NewPersonalAccount(name, email, phone, createdBy string, accountTypeID string) (*Account, error) {
    sanitizer := validation.Sanitize{}
    
    if name == "" {
        return nil, errors.New("name is required")
    }
    if email == "" {
        return nil, errors.New("email is required")
    }
    if accountTypeID == "" {
        return nil, errors.New("account type is required")
    }

    // Sanitize fields
    cleanName := sanitizer.DisplayName(name)           // Preserves case, removes extra spaces
    displayName := cleanName
    slug := sanitizer.GenerateSlugFromName(cleanName)  // "john-doe"
    cleanEmail := sanitizer.Identifier(email)          // Lowercase, no spaces

    now := time.Now()
    return &Account{
        ID:            uuid.New().String(),
        Name:          cleanName,
        DisplayName:   displayName,
        Slug:          slug,
        Email:         cleanEmail,
        Phone:         phone,
        AccountTypeID: accountTypeID,
        Status:        AccountStatusActive,
        CreatedBy:     createdBy,
        CreatedAt:     now,
        UpdatedAt:     now,
    }, nil
}

// NewInstitutionAccount creates a new institution account
func NewInstitutionAccount(name, email, phone, website, description, createdBy string, accountTypeID string) (*Account, error) {
    sanitizer := validation.Sanitize{}
    
    if name == "" {
        return nil, errors.New("institution name is required")
    }
    if email == "" {
        return nil, errors.New("institution email is required")
    }
    if accountTypeID == "" {
        return nil, errors.New("account type is required")
    }

    // Sanitize fields
    cleanName := sanitizer.DisplayName(name)
    displayName := cleanName
    slug := sanitizer.GenerateSlugFromName(cleanName)  // "nuruvent-ltd"
    cleanEmail := sanitizer.Identifier(email)
    cleanWebsite := sanitizer.Identifier(website)
    cleanDescription := sanitizer.Description(description)

    now := time.Now()
    return &Account{
        ID:            uuid.New().String(),
        Name:          cleanName,
        DisplayName:   displayName,
        Slug:          slug,
        Email:         cleanEmail,
        Phone:         phone,
        AccountTypeID: accountTypeID,
        Status:        AccountStatusActive,
        LogoURL:       "",
        Website:       cleanWebsite,
        Description:   cleanDescription,
        CreatedBy:     createdBy,
        CreatedAt:     now,
        UpdatedAt:     now,
    }, nil
}

// Update updates the account with new values
func (a *Account) Update(name, email, phone, website, description, logoURL string) {
    sanitizer := validation.Sanitize{}
    
    if name != "" {
        a.Name = sanitizer.DisplayName(name)
        a.DisplayName = a.Name
        a.Slug = sanitizer.GenerateSlugFromName(a.Name)
    }
    if email != "" {
        a.Email = sanitizer.Identifier(email)
    }
    if phone != "" {
        a.Phone = phone
    }
    if website != "" {
        a.Website = sanitizer.Identifier(website)
    }
    if description != "" {
        a.Description = sanitizer.Description(description)
    }
    if logoURL != "" {
        a.LogoURL = logoURL
    }
    a.UpdatedAt = time.Now()
}

// Suspend suspends the account
func (a *Account) Suspend() {
    a.Status = AccountStatusSuspended
    a.UpdatedAt = time.Now()
}

// Activate activates the account
func (a *Account) Activate() {
    a.Status = AccountStatusActive
    a.UpdatedAt = time.Now()
}

// Deactivate deactivates the account
func (a *Account) Deactivate() {
    a.Status = AccountStatusInactive
    a.UpdatedAt = time.Now()
}