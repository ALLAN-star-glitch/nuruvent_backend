// internal/modules/auth/authdomain/account.go

package authdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Account represents an account (personal or institution)
type Account struct {
	ID            string
	Name          string
	DisplayName   string
	Slug          string
	Email         string
	Phone         string
	AccountTypeID string // References account_types table
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
		ID:            uuid.New().String(),
		Name:          name,
		DisplayName:   displayName,
		Slug:          slug,
		Email:         email,
		Phone:         phone,
		AccountTypeID: accountTypeID,
		Status:        AccountStatusActive,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
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