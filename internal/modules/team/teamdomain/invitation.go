// internal/modules/team/teamdomain/invitation.go

package teamdomain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// INVITATION STATUS
// ============================================================

// InvitationStatus represents the status of an invitation.
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// ============================================================
// ACCOUNT ROLE CONSTANTS
// ============================================================
//
// These mirror authdomain.RoleAccountAdmin and authdomain.RoleTrainer.
// Duplicated here to avoid the team module importing auth.
//
// The `account_members.role` column has a CHECK constraint limiting
// values to exactly these two strings. Keep them in sync.

const (
	RoleAccountAdmin = "account_admin"
	RoleTrainer      = "trainer"
)

// IsValidAccountRole reports whether role is one of the two account roles
// that may be granted via a team invitation.
func IsValidAccountRole(role string) bool {
	return role == RoleAccountAdmin || role == RoleTrainer
}

// ============================================================
// INVITATION ENTITY
// ============================================================

// Invitation represents an invitation to join a team.
//
// An invitation carries:
//   - The team the invitee is being added to.
//   - The account role the invitee should receive, if they aren't
//     already an account member. Existing members keep their role.
//   - The user who sent the invitation.
//   - A single-use token.
//
// Roles are not team-scoped. The `Role` field describes the invitee's
// role in the parent account, not a team-local role.
type Invitation struct {
	ID         string
	TeamID     string
	Email      string
	Role       string // "account_admin" or "trainer"
	Token      string
	Status     InvitationStatus
	InvitedBy  string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
	DeclinedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewInvitation creates a new pending team invitation.
//
// The role must be one of RoleAccountAdmin or RoleTrainer. It describes
// the role the invitee will receive in the account if they are not
// already an account member.
func NewInvitation(
	teamID, email, role, invitedBy, token string,
	expiresAt time.Time,
) (*Invitation, error) {
	if teamID == "" {
		return nil, errors.New("team ID is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if role == "" {
		return nil, errors.New("role is required")
	}
	if !IsValidAccountRole(role) {
		return nil, fmt.Errorf("invalid role: %q (must be %q or %q)",
			role, RoleAccountAdmin, RoleTrainer)
	}
	if invitedBy == "" {
		return nil, errors.New("invited by is required")
	}
	if token == "" {
		return nil, errors.New("token is required")
	}
	if expiresAt.IsZero() {
		return nil, errors.New("expires at is required")
	}
	if !expiresAt.After(time.Now()) {
		return nil, errors.New("expires at must be in the future")
	}

	now := time.Now()
	return &Invitation{
		ID:        uuid.New().String(),
		TeamID:    teamID,
		Email:     email,
		Role:      role,
		Token:     token,
		Status:    InvitationStatusPending,
		InvitedBy: invitedBy,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ============================================================
// STATE CHECKS
// ============================================================

// IsExpired reports whether the invitation has passed its expiry.
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending reports whether the invitation is still pending.
func (i *Invitation) IsPending() bool {
	return i.Status == InvitationStatusPending
}

// IsAccepted reports whether the invitation has been accepted.
func (i *Invitation) IsAccepted() bool {
	return i.Status == InvitationStatusAccepted
}

// IsDeclined reports whether the invitation has been declined.
func (i *Invitation) IsDeclined() bool {
	return i.Status == InvitationStatusDeclined
}

// IsUsable reports whether the invitation can be acted on.
// An invitation is usable when it is pending and not expired.
func (i *Invitation) IsUsable() bool {
	return i.IsPending() && !i.IsExpired()
}

// ============================================================
// STATE TRANSITIONS
// ============================================================

// Accept marks the invitation as accepted.
//
// Returns an error if the invitation is not pending or has expired.
func (i *Invitation) Accept() error {
	if i.Status != InvitationStatusPending {
		return fmt.Errorf("cannot accept invitation in status %q", i.Status)
	}
	if i.IsExpired() {
		return errors.New("invitation has expired")
	}
	now := time.Now()
	i.Status = InvitationStatusAccepted
	i.AcceptedAt = &now
	i.UpdatedAt = now
	return nil
}

// Decline marks the invitation as declined.
//
// Returns an error if the invitation is not pending or has expired.
func (i *Invitation) Decline() error {
	if i.Status != InvitationStatusPending {
		return fmt.Errorf("cannot decline invitation in status %q", i.Status)
	}
	if i.IsExpired() {
		return errors.New("invitation has expired")
	}
	now := time.Now()
	i.Status = InvitationStatusDeclined
	i.DeclinedAt = &now
	i.UpdatedAt = now
	return nil
}

// Expire marks the invitation as expired.
// Idempotent: safe to call on an already-expired invitation.
func (i *Invitation) Expire() {
	if i.Status == InvitationStatusExpired {
		return
	}
	i.Status = InvitationStatusExpired
	i.UpdatedAt = time.Now()
}

// Reissue refreshes the token and expiry on a pending invitation.
//
// Used by the "resend invitation" flow. Resets the state to pending
// so a previously expired invitation can be reactivated.
func (i *Invitation) Reissue(newToken string, newExpiry time.Time) error {
	if newToken == "" {
		return errors.New("new token is required")
	}
	if newExpiry.IsZero() || !newExpiry.After(time.Now()) {
		return errors.New("new expiry must be in the future")
	}
	if i.Status == InvitationStatusAccepted {
		return errors.New("cannot reissue an accepted invitation")
	}

	i.Token = newToken
	i.ExpiresAt = newExpiry
	i.Status = InvitationStatusPending
	i.AcceptedAt = nil
	i.DeclinedAt = nil
	i.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// TOKEN GENERATION
// ============================================================

// GenerateToken returns a new single-use invitation token.
func GenerateToken() string {
	return uuid.New().String()
}