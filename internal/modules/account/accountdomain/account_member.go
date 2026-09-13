// internal/modules/account/accountdomain/account_member.go

package accountdomain

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

// AccountMember represents a user's membership in an account
type AccountMember struct {
    ID          string
    AccountID   string
    UserID      string
    Role        string
    IsActive    bool
    InvitedBy   string
    JoinedAt    time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

// Account role constants
const (
    RoleAccountAdmin = "account_admin"
    RoleTrainer      = "trainer"
)

// AllAccountRoles returns all valid account roles
func AllAccountRoles() []string {
    return []string{RoleAccountAdmin, RoleTrainer}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
    for _, r := range AllAccountRoles() {
        if r == role {
            return true
        }
    }
    return false
}

// NewAccountMember creates a new account member
func NewAccountMember(accountID, userID, role, invitedBy string) (*AccountMember, error) {
    if accountID == "" {
        return nil, errors.New("account ID is required")
    }
    if userID == "" {
        return nil, errors.New("user ID is required")
    }
    if role == "" {
        return nil, errors.New("role is required")
    }
    if !IsValidRole(role) {
        return nil, errors.New("invalid role")
    }

    now := time.Now()
    return &AccountMember{
        ID:          uuid.New().String(),
        AccountID:   accountID,
        UserID:      userID,
        Role:        role,
        IsActive:    true,
        InvitedBy:   invitedBy,
        JoinedAt:    now,
        CreatedAt:   now,
        UpdatedAt:   now,
    }, nil
}

// IsAdmin checks if the member is an account admin
func (m *AccountMember) IsAdmin() bool {
    return m.Role == RoleAccountAdmin
}

// IsTrainer checks if the member is a trainer
func (m *AccountMember) IsTrainer() bool {
    return m.Role == RoleTrainer
}

// UpdateRole updates the member's role
func (m *AccountMember) UpdateRole(role string) error {
    if !IsValidRole(role) {
        return errors.New("invalid role")
    }
    m.Role = role
    m.UpdatedAt = time.Now()
    return nil
}

// Deactivate deactivates the member
func (m *AccountMember) Deactivate() {
    m.IsActive = false
    m.UpdatedAt = time.Now()
}

// Activate activates the member
func (m *AccountMember) Activate() {
    m.IsActive = true
    m.UpdatedAt = time.Now()
}