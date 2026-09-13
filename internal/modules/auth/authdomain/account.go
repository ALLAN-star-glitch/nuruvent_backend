// internal/modules/auth/authdomain/account.go

package authdomain

import (
	"errors"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
	"github.com/google/uuid"
)

// Account represents an account (personal or institution)
type Account struct {
	ID                string
	Name              string
	DisplayName       string
	Slug              string
	Email             string
	Phone             string
	AccountTypeID     string  // References account_types table (personal/institution)
	InstitutionTypeID *string // References institution_types table (company/institute/etc.) - NULL for personal
	Status            string
	LogoURL           string
	Website           string
	Description       string
	Address           string
	City              string
	Country           string
	BillingEmail      string
	SubscriptionPlan  string
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// Account status constants
const (
	AccountStatusActive    = "active"
	AccountStatusSuspended = "suspended"
	AccountStatusInactive  = "inactive"
)


// NewAccount creates a new account
func NewAccount(name, displayName, slug, email, phone, accountTypeID, createdBy string) (*Account, error) {
	if name == "" {
		return nil, errors.New("account name is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if accountTypeID == "" {
		return nil, errors.New("account type is required")
	}

	now := time.Now()
	return &Account{
		ID:                uuid.New().String(),
		Name:              name,
		DisplayName:       displayName,
		Slug:              slug,
		Email:             email,
		Phone:             phone,
		AccountTypeID:     accountTypeID,
		InstitutionTypeID: nil, // Will be set later for institution accounts
		Status:            AccountStatusActive,
		CreatedBy:         createdBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// NewPersonalAccount creates a new personal account
func NewPersonalAccount(name, displayName, slug, email, phone, accountTypeID, createdBy string) (*Account, error) {
	account, err := NewAccount(name, displayName, slug, email, phone, accountTypeID, createdBy)
	if err != nil {
		return nil, err
	}
	account.InstitutionTypeID = nil
	return account, nil
}

// NewInstitutionAccount creates a new institution account
func NewInstitutionAccount(name, displayName, slug, email, phone, accountTypeID, institutionTypeID, createdBy string) (*Account, error) {
	if institutionTypeID == "" {
		return nil, errors.New("institution type is required for institution accounts")
	}

	account, err := NewAccount(name, displayName, slug, email, phone, accountTypeID, createdBy)
	if err != nil {
		return nil, err
	}
	account.InstitutionTypeID = &institutionTypeID
	return account, nil
}

// ============================================================
// TYPE CHECKERS
// ============================================================

// IsPersonal checks if the account is a personal account
func (a *Account) IsPersonal() bool {
	return a.AccountTypeID == types.AccountTypePersonalName
}

// IsInstitution checks if the account is an institution account
func (a *Account) IsInstitution() bool {
	return a.AccountTypeID == types.AccountTypeInstitutionName
}

// HasInstitutionType checks if the account has an institution type
func (a *Account) HasInstitutionType() bool {
	return a.InstitutionTypeID != nil && *a.InstitutionTypeID != ""
}

// ============================================================
// UPDATE METHODS
// ============================================================

// UpdateName updates the account name
func (a *Account) UpdateName(name string) {
	if name != "" {
		a.Name = name
		a.UpdatedAt = time.Now()
	}
}

// UpdateDisplayName updates the account display name
func (a *Account) UpdateDisplayName(displayName string) {
	a.DisplayName = displayName
	a.UpdatedAt = time.Now()
}

// UpdateEmail updates the account email
func (a *Account) UpdateEmail(email string) {
	if email != "" {
		a.Email = email
		a.UpdatedAt = time.Now()
	}
}

// UpdatePhone updates the account phone
func (a *Account) UpdatePhone(phone string) {
	a.Phone = phone
	a.UpdatedAt = time.Now()
}

// UpdateLogoURL updates the account logo URL
func (a *Account) UpdateLogoURL(logoURL string) {
	a.LogoURL = logoURL
	a.UpdatedAt = time.Now()
}

// UpdateWebsite updates the account website
func (a *Account) UpdateWebsite(website string) {
	a.Website = website
	a.UpdatedAt = time.Now()
}

// UpdateDescription updates the account description
func (a *Account) UpdateDescription(description string) {
	a.Description = description
	a.UpdatedAt = time.Now()
}

// UpdateAddress updates the account address
func (a *Account) UpdateAddress(address, city, country string) {
	a.Address = address
	a.City = city
	a.Country = country
	a.UpdatedAt = time.Now()
}

// UpdateBillingEmail updates the billing email
func (a *Account) UpdateBillingEmail(billingEmail string) {
	a.BillingEmail = billingEmail
	a.UpdatedAt = time.Now()
}

// UpdateSubscriptionPlan updates the subscription plan
func (a *Account) UpdateSubscriptionPlan(plan string) {
	a.SubscriptionPlan = plan
	a.UpdatedAt = time.Now()
}

// ============================================================
// STATUS METHODS
// ============================================================

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

// IsActive checks if the account is active
func (a *Account) IsActive() bool {
	return a.Status == AccountStatusActive
}

// IsSuspended checks if the account is suspended
func (a *Account) IsSuspended() bool {
	return a.Status == AccountStatusSuspended
}

// IsInactive checks if the account is inactive
func (a *Account) IsInactive() bool {
	return a.Status == AccountStatusInactive
}

// ============================================================
// SOFT DELETE METHODS
// ============================================================

// SoftDelete soft deletes the account
func (a *Account) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.Status = AccountStatusInactive
	a.UpdatedAt = now
}

// Restore restores a soft-deleted account
func (a *Account) Restore() {
	a.DeletedAt = nil
	a.Status = AccountStatusActive
	a.UpdatedAt = time.Now()
}