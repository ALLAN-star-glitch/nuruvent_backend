// internal/modules/account/accountdomain/account_type.go

package accountdomain

import (
    "errors"
    "time"
    "github.com/google/uuid"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

// AccountType represents the type of account
type AccountType struct {
    ID          string
    Name        string
    DisplayName string
    Slug        string
    Description string
    Icon        string
    Color       string
    SortOrder   int
    IsActive    bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

// NewAccountType creates a new account type from shared types
func NewAccountType(accountType types.AccountType) (*AccountType, error) {
    sanitizer := validation.Sanitize{}
    
    if !accountType.IsValid() {
        return nil, errors.New("invalid account type")
    }

    name := accountType.GetName()
    displayName := accountType.GetDisplayName()
    //  Remove unused variable
    description := accountType.GetDescription()
    icon := accountType.GetIcon()
    color := accountType.GetColor()
    sortOrder := accountType.GetSortOrder()

    // Sanitize fields
    cleanName := sanitizer.Identifier(name)
    cleanDisplayName := sanitizer.DisplayName(displayName)
    cleanSlug := sanitizer.GenerateSlugFromName(displayName)
    cleanDescription := sanitizer.Description(description)

    now := time.Now()
    return &AccountType{
        ID:          uuid.New().String(),
        Name:        cleanName,
        DisplayName: cleanDisplayName,
        Slug:        cleanSlug,
        Description: cleanDescription,
        Icon:        icon,
        Color:       color,
        SortOrder:   sortOrder,
        IsActive:    true,
        CreatedAt:   now,
        UpdatedAt:   now,
    }, nil
}

// NewPersonalAccountType creates a personal account type
func NewPersonalAccountType() (*AccountType, error) {
    return NewAccountType(types.AccountTypePersonal)
}

// NewInstitutionAccountType creates an institution account type
func NewInstitutionAccountType() (*AccountType, error) {
    return NewAccountType(types.AccountTypeInstitution)
}

// IsPersonal checks if account type is personal
func (a *AccountType) IsPersonal() bool {
    return a.Name == types.AccountTypePersonalName
}

// IsInstitution checks if account type is institution
func (a *AccountType) IsInstitution() bool {
    return a.Name == types.AccountTypeInstitutionName
}

// GetSharedType returns the shared types.AccountType
func (a *AccountType) GetSharedType() types.AccountType {
    if a.IsPersonal() {
        return types.AccountTypePersonal
    }
    return types.AccountTypeInstitution
}

// ToSharedType converts to shared types.AccountType
func (a *AccountType) ToSharedType() types.AccountType {
    return a.GetSharedType()
}