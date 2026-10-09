// internal/modules/account/accountdomain/account.go

package accountdomain

import (
    "errors"
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
    "github.com/google/uuid"
)

// Account represents a personal or institution account
type Account struct {
    ID            string
    Name          string
    DisplayName   string
    Slug          string
    Type          string // "personal" or "institution"
    Email         string
    Phone         string
    AccountTypeID string
    Status        string
    LogoURL       string
    Website       string
    Description   string
    Address       string
    City          string
    Country       string
    KYCStatus     string
    CreatedBy     string
    IsActive      bool
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

// Account type constants
const (
    AccountTypePersonal    = "personal"
    AccountTypeInstitution = "institution"
)

// KYC status constants
const (
    KYCStatusPending     = "pending"
    KYCStatusSubmitted   = "submitted"
    KYCStatusVerified    = "verified"
    KYCStatusRejected    = "rejected"
    KYCStatusNotRequired = "not_required"
)

// ============================================================
// CONSTRUCTORS
// ============================================================

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

    // Sanitize fields.
    //
    // Email uses Sanitize.Email, not Sanitize.Identifier — Identifier
    // strips `@` and `.` and would corrupt the address.
    cleanName := sanitizer.DisplayName(name)
    displayName := cleanName
    slug := sanitizer.GenerateSlugFromName(cleanName)
    cleanEmail := sanitizer.Email(email)
    cleanPhone := sanitizer.Phone(phone)

    now := time.Now()
    return &Account{
        ID:            uuid.New().String(),
        Name:          cleanName,
        DisplayName:   displayName,
        Slug:          slug,
        Type:          AccountTypePersonal,
        Email:         cleanEmail,
        Phone:         cleanPhone,
        AccountTypeID: accountTypeID,
        Status:        AccountStatusActive,
        KYCStatus:     KYCStatusNotRequired,
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

    // Sanitize fields. Email and URL have dedicated sanitizers that
    // preserve their syntax characters.
    cleanName := sanitizer.DisplayName(name)
    displayName := cleanName
    slug := sanitizer.GenerateSlugFromName(cleanName)
    cleanEmail := sanitizer.Email(email)
    cleanPhone := sanitizer.Phone(phone)
    cleanWebsite := sanitizer.URL(website)
    cleanDescription := sanitizer.Description(description)

    now := time.Now()
    return &Account{
        ID:            uuid.New().String(),
        Name:          cleanName,
        DisplayName:   displayName,
        Slug:          slug,
        Type:          AccountTypeInstitution,
        Email:         cleanEmail,
        Phone:         cleanPhone,
        AccountTypeID: accountTypeID,
        Status:        AccountStatusActive,
        KYCStatus:     KYCStatusPending,
        LogoURL:       "",
        Website:       cleanWebsite,
        Description:   cleanDescription,
        CreatedBy:     createdBy,
        CreatedAt:     now,
        UpdatedAt:     now,
    }, nil
}

// ============================================================
// UPDATE
// ============================================================

// Update updates the account with new values.
//
// Only non-empty fields are applied. Email, phone, and website use
// their dedicated sanitizers (Email / Phone / URL) so syntax characters
// are preserved.
func (a *Account) Update(name, email, phone, website, description, logoURL, address, city, country string) {
    sanitizer := validation.Sanitize{}

    if name != "" {
        a.Name = sanitizer.DisplayName(name)
        a.DisplayName = a.Name
        a.Slug = sanitizer.GenerateSlugFromName(a.Name)
    }
    if email != "" {
        a.Email = sanitizer.Email(email)
    }
    if phone != "" {
        a.Phone = sanitizer.Phone(phone)
    }
    if website != "" {
        a.Website = sanitizer.URL(website)
    }
    if description != "" {
        a.Description = sanitizer.Description(description)
    }
    if logoURL != "" {
        a.LogoURL = logoURL
    }
    if address != "" {
        a.Address = address
    }
    if city != "" {
        a.City = city
    }
    if country != "" {
        a.Country = country
    }
    a.UpdatedAt = time.Now()
}

// ============================================================
// STATUS
// ============================================================

// UpdateStatus updates the account status
func (a *Account) UpdateStatus(status string) error {
    validStatuses := map[string]bool{
        AccountStatusActive:    true,
        AccountStatusSuspended: true,
        AccountStatusInactive:  true,
    }
    if !validStatuses[status] {
        return errors.New("invalid status")
    }
    a.Status = status
    a.UpdatedAt = time.Now()
    return nil
}

// UpdateKYCStatus updates the KYC status
func (a *Account) UpdateKYCStatus(status string) error {
    validStatuses := map[string]bool{
        KYCStatusPending:     true,
        KYCStatusSubmitted:   true,
        KYCStatusVerified:    true,
        KYCStatusRejected:    true,
        KYCStatusNotRequired: true,
    }
    if !validStatuses[status] {
        return errors.New("invalid KYC status")
    }
    a.KYCStatus = status
    a.UpdatedAt = time.Now()
    return nil
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

// ============================================================
// TYPE CHECKS
// ============================================================

// IsPersonal checks if the account is a personal account
func (a *Account) IsPersonal() bool {
    return a.Type == AccountTypePersonal
}

// IsInstitution checks if the account is an institution account
func (a *Account) IsInstitution() bool {
    return a.Type == AccountTypeInstitution
}

// IsAccountActive checks if the account is active
func (a *Account) IsAccountActive() bool {
    return a.Status == AccountStatusActive
}

// IsSuspended checks if the account is suspended
func (a *Account) IsSuspended() bool {
    return a.Status == AccountStatusSuspended
}

// IsDeleted checks if the account is deleted
func (a *Account) IsDeleted() bool {
    return a.DeletedAt != nil && !a.DeletedAt.IsZero()
}