package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// ACCOUNT - Domain Entity
// ============================================================

type Account struct {
	ID                 string
	Slug               string
	Name               string
	DisplayName        string
	Email              string
	PasswordHash       string     // Changed from Password to match database
	Phone              string
	AccountTypeID      string
	ProfessionalTypeID *string
	InstitutionID      *string
	EmailVerified      bool
	EmailVerifiedAt    *time.Time
	PhoneVerified      bool       // Added for individual registration
	PhoneVerifiedAt    *time.Time // Added for individual registration
	IdentityVerified   bool
	KYCStatus          string     // Added for individual registration
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewAccount creates a validated account
func NewAccount(email, passwordHash, name, phone, accountTypeID string) (*Account, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if passwordHash == "" {
		return nil, errors.New("password is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if phone == "" {
		return nil, errors.New("phone is required")
	}
	if accountTypeID == "" {
		return nil, errors.New("account type is required")
	}

	now := time.Now()
	return &Account{
		ID:                 uuid.New().String(),
		Slug:               generateSlug(name),
		Name:               name,
		DisplayName:        name,
		Email:              email,
		PasswordHash:       passwordHash,
		Phone:              phone,
		AccountTypeID:      accountTypeID,
		EmailVerified:      false,
		PhoneVerified:      false,
		IdentityVerified:   false,
		KYCStatus:          "pending",
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

// VerifyEmail marks the account's email as verified
func (a *Account) VerifyEmail() {
	now := time.Now()
	a.EmailVerified = true
	a.EmailVerifiedAt = &now
	a.UpdatedAt = now
}

// VerifyPhone marks the account's phone as verified
func (a *Account) VerifyPhone() {
	now := time.Now()
	a.PhoneVerified = true
	a.PhoneVerifiedAt = &now
	a.UpdatedAt = now
}

// UpdateKYCStatus updates the KYC status of the account
func (a *Account) UpdateKYCStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":      true,
		"submitted":    true,
		"verified":     true,
		"rejected":     true,
		"not_required": true,
	}
	if !validStatuses[status] {
		return errors.New("invalid KYC status")
	}
	a.KYCStatus = status
	a.UpdatedAt = time.Now()
	return nil
}

// Deactivate deactivates the account
func (a *Account) Deactivate() error {
	if !a.IsActive {
		return errors.New("account already inactive")
	}
	a.IsActive = false
	a.UpdatedAt = time.Now()
	return nil
}

// Activate activates the account
func (a *Account) Activate() error {
	if a.IsActive {
		return errors.New("account already active")
	}
	if !a.EmailVerified {
		return errors.New("cannot activate unverified account")
	}
	a.IsActive = true
	a.UpdatedAt = time.Now()
	return nil
}

// IsActiveAccount checks if the account is active
func (a *Account) IsActiveAccount() bool {
	return a.IsActive
}

// IsEmailVerified checks if the email is verified
func (a *Account) IsEmailVerified() bool {
	return a.EmailVerified
}

// IsPhoneVerified checks if the phone is verified
func (a *Account) IsPhoneVerified() bool {
	return a.PhoneVerified
}

// UpdateName updates the account name
func (a *Account) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	a.Name = name
	a.UpdatedAt = time.Now()
	return nil
}

// UpdateDisplayName updates the display name
func (a *Account) UpdateDisplayName(displayName string) {
	a.DisplayName = displayName
	a.UpdatedAt = time.Now()
}

// UpdateEmail updates the account email
func (a *Account) UpdateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	a.Email = email
	a.EmailVerified = false
	a.EmailVerifiedAt = nil
	a.UpdatedAt = time.Now()
	return nil
}

// UpdatePhone updates the account phone
func (a *Account) UpdatePhone(phone string) error {
	if phone == "" {
		return errors.New("phone cannot be empty")
	}
	a.Phone = phone
	a.PhoneVerified = false
	a.PhoneVerifiedAt = nil
	a.UpdatedAt = time.Now()
	return nil
}

// UpdatePassword updates the account password hash
func (a *Account) UpdatePassword(passwordHash string) {
	a.PasswordHash = passwordHash
	a.UpdatedAt = time.Now()
}

// ============================================================
// HELPERS
// ============================================================

func generateSlug(name string) string {
	return "user-" + uuid.New().String()[:8]
}