package authdomain

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

// Account entity
type Account struct {
    ID             string
    Slug           string
    Name           string
    DisplayName    string
    Email          string
    PasswordHash       string
    Phone          string
    AccountTypeID  string
    ProfessionalTypeID *string
    InstitutionID  *string
    EmailVerified  bool
    EmailVerifiedAt *time.Time
    IdentityVerified bool
    IsActive       bool
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
}

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
    if accountTypeID == "" {
        return nil, errors.New("account type is required")
    }

    now := time.Now()
    return &Account{
        ID:            uuid.New().String(),
        Slug:          slugify(name),
        Name:          name,
        DisplayName:   name,
        Email:         email,
        PasswordHash:  passwordHash,
        Phone:         phone,
        AccountTypeID: accountTypeID,
        EmailVerified: false,
        IsActive:      true,
        CreatedAt:     now,
        UpdatedAt:     now,
    }, nil
}

// Behaviors
func (a *Account) VerifyEmail() {
    now := time.Now()
    a.EmailVerified = true
    a.EmailVerifiedAt = &now
    a.UpdatedAt = now
}

func (a *Account) Deactivate() error {
    if !a.IsActive {
        return errors.New("account already inactive")
    }
    a.IsActive = false
    a.UpdatedAt = time.Now()
    return nil
}

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

func (a *Account) IsActiveAccount() bool {
    return a.IsActive
}

func (a *Account) IsEmailVerified() bool {
    return a.EmailVerified
}

func (a *Account) UpdateName(name string) error {
    if name == "" {
        return errors.New("name cannot be empty")
    }
    a.Name = name
    a.UpdatedAt = time.Now()
    return nil
}

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

func (a *Account) UpdatePhone(phone string) error {
    if phone == "" {
        return errors.New("phone cannot be empty")
    }
    a.Phone = phone
    a.UpdatedAt = time.Now()
    return nil
}

func (a *Account) UpdatePassword(hashedPassword string) {
    a.PasswordHash = hashedPassword
    a.UpdatedAt = time.Now()
}

func slugify(name string) string {
    return "user-" + uuid.New().String()[:8]
}

