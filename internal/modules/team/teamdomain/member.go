// internal/modules/team/teamdomain/member.go

package teamdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Member represents a user's membership in a team.
//
// The identity fields at the bottom — Name, DisplayName, Email,
// AvatarURL, AccountRole — are populated only by the enriched read
// path (GetTeamMembersWithUsers) which batch-fetches user records and
// account roles. They are empty on mutation responses such as
// AddMember, LeaveTeam, and AcceptInvitation, and on the plain
// GetTeamMembers path.
//
// Storing them on the domain type rather than introducing a separate
// MemberWithUser type keeps the repository, service, and handler
// signatures uniform: every code path that returns a member returns
// the same shape. Consumers that don't need identity ignore the empty
// fields.
type Member struct {
	ID        string     `json:"id"`
	TeamID    string     `json:"team_id"`
	UserID    string     `json:"user_id"`
	IsActive  bool       `json:"is_active"`
	JoinedAt  time.Time  `json:"joined_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// ---- Identity fields (enriched read path only) ----
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`

	// ---- Account-level role of this member (enriched read path only) ----
	AccountRole string `json:"role,omitempty"`
}

// NewMember creates a new team member.
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

// Deactivate deactivates the member.
func (m *Member) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// Activate activates the member.
func (m *Member) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}