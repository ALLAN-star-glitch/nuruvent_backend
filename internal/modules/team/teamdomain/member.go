// internal/modules/team/teamdomain/member.go

package teamdomain

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

// MemberRole represents a member's role in a team
type MemberRole string

const (
    RoleAccountAdmin MemberRole = "account_admin"
    RoleTrainer      MemberRole = "trainer"
)

// AllRoles returns all available roles
func AllRoles() []MemberRole {
    return []MemberRole{RoleAccountAdmin, RoleTrainer}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
    for _, r := range AllRoles() {
        if string(r) == role {
            return true
        }
    }
    return false
}

// Member represents a user's membership in a team
type Member struct {
    ID            string
    TeamID        string
    UserID        string
    Role          MemberRole
    IsActive      bool
    JoinedAt      time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

// NewMember creates a new team member
func NewMember(teamID, userID string, role MemberRole) (*Member, error) {
    if teamID == "" {
        return nil, errors.New("team ID is required")
    }
    if userID == "" {
        return nil, errors.New("user ID is required")
    }
    if role == "" {
        return nil, errors.New("role is required")
    }
    if !IsValidRole(string(role)) {
        return nil, errors.New("invalid role")
    }

    now := time.Now()
    return &Member{
        ID:        uuid.New().String(),
        TeamID:    teamID,
        UserID:    userID,
        Role:      role,
        IsActive:  true,
        JoinedAt:  now,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

// UpdateRole updates the member's role
func (m *Member) UpdateRole(role MemberRole) error {
    if !IsValidRole(string(role)) {
        return errors.New("invalid role")
    }
    m.Role = role
    m.UpdatedAt = time.Now()
    return nil
}

// Deactivate deactivates the member
func (m *Member) Deactivate() {
    m.IsActive = false
    m.UpdatedAt = time.Now()
}

// Activate activates the member
func (m *Member) Activate() {
    m.IsActive = true
    m.UpdatedAt = time.Now()
}