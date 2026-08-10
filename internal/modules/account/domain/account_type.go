package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// ACCOUNT TYPE - Domain Entity (Database Representation)
// ============================================================

// AccountType is a domain entity stored in the database
type AccountType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}



// ============================================================
// FACTORY
// ============================================================

// NewAccountType creates a new AccountType entity
func NewAccountType(slug, name, displayName, description, icon, color string, sortOrder int) (*AccountType, error) {
	if slug == "" {
		return nil, ErrInvalidAccountType
	}
	if name == "" {
		return nil, ErrInvalidAccountType
	}
	if displayName == "" {
		return nil, ErrInvalidAccountType
	}

	now := time.Now()
	return &AccountType{
		ID:          generateID(),
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewAccountTypeFromRepository recreates AccountType from repository data
func NewAccountTypeFromRepository(
	id, slug, name, displayName, description, icon, color string,
	sortOrder int, isActive bool,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *AccountType {
	return &AccountType{
		ID:          id,
		Slug:        slug,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		Color:       color,
		SortOrder:   sortOrder,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	}
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

// IsValid checks if the account type has a valid ID
func (a *AccountType) IsValid() bool {
	return a.ID != ""
}

// IsActiveType checks if the account type is active
func (a *AccountType) IsActiveType() bool {
	return a.IsActive
}

// Activate activates the account type
func (a *AccountType) Activate() {
	a.IsActive = true
	a.UpdatedAt = time.Now()
}

// Deactivate deactivates the account type
func (a *AccountType) Deactivate() {
	a.IsActive = false
	a.UpdatedAt = time.Now()
}

// UpdateName updates the name of the account type
func (a *AccountType) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidAccountType
	}
	a.Name = name
	a.UpdatedAt = time.Now()
	return nil
}

// UpdateDisplayName updates the display name
func (a *AccountType) UpdateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrInvalidAccountType
	}
	a.DisplayName = displayName
	a.UpdatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (a *AccountType) UpdateDescription(description string) {
	a.Description = description
	a.UpdatedAt = time.Now()
}

// UpdateIcon updates the icon
func (a *AccountType) UpdateIcon(icon string) {
	a.Icon = icon
	a.UpdatedAt = time.Now()
}

// UpdateColor updates the color
func (a *AccountType) UpdateColor(color string) {
	a.Color = color
	a.UpdatedAt = time.Now()
}

// UpdateSortOrder updates the sort order
func (a *AccountType) UpdateSortOrder(sortOrder int) {
	a.SortOrder = sortOrder
	a.UpdatedAt = time.Now()
}

// ============================================================
// HELPERS
// ============================================================

func generateID() string {
	return uuid.New().String()
}