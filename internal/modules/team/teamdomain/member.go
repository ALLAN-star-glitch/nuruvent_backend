// internal/modules/team/teamdomain/member.go

package teamdomain

import (
    "errors"
    "time"

    "github.com/google/uuid"
)




// Member represents a user's membership in a team
type Member struct {
    ID            string
    TeamID        string
    UserID        string
    IsActive      bool
    JoinedAt      time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

// NewMember creates a new team member
func NewMember(teamID, userID string) (*Member, error) {
    if teamID == "" {
        return nil, errors.New("team ID is required")
    }
    if userID == "" {
        return nil, errors.New("user ID is required")
    }



    now := time.Now()
    return &Member{
        ID:        uuid.New().String(),
        TeamID:    teamID,
        UserID:    userID,
        IsActive:  true,
        JoinedAt:  now,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
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