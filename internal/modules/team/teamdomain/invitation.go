// internal/modules/team/teamdomain/invitation.go

package teamdomain

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
    InvitationStatusPending  InvitationStatus = "pending"
    InvitationStatusAccepted InvitationStatus = "accepted"
    InvitationStatusDeclined InvitationStatus = "declined"
    InvitationStatusExpired  InvitationStatus = "expired"
)

// Invitation represents an invitation to join a team
type Invitation struct {
    ID            string
    TeamID        string
    Email         string
    Role          MemberRole
    Token         string
    Status        InvitationStatus
    InvitedBy     string
    ExpiresAt     time.Time
    AcceptedAt    *time.Time
    DeclinedAt    *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

// NewInvitation creates a new team invitation
func NewInvitation(teamID, email, invitedBy string, role MemberRole, token string, expiresAt time.Time) (*Invitation, error) {
    if teamID == "" {
        return nil, errors.New("team ID is required")
    }
    if email == "" {
        return nil, errors.New("email is required")
    }
    if invitedBy == "" {
        return nil, errors.New("invited by is required")
    }
    if role == "" {
        return nil, errors.New("role is required")
    }
    if !IsValidRole(string(role)) {
        return nil, errors.New("invalid role")
    }
    if token == "" {
        return nil, errors.New("token is required")
    }
    if expiresAt.IsZero() {
        return nil, errors.New("expires at is required")
    }

    now := time.Now()
    return &Invitation{
        ID:         uuid.New().String(),
        TeamID:     teamID,
        Email:      email,
        Role:       role,
        Token:      token,
        Status:     InvitationStatusPending,
        InvitedBy:  invitedBy,
        ExpiresAt:  expiresAt,
        CreatedAt:  now,
        UpdatedAt:  now,
    }, nil
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
    return time.Now().After(i.ExpiresAt)
}

// IsPending checks if the invitation is pending
func (i *Invitation) IsPending() bool {
    return i.Status == InvitationStatusPending
}

// Accept marks the invitation as accepted
func (i *Invitation) Accept() error {
    if i.Status != InvitationStatusPending {
        return errors.New("invitation is not pending")
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

// Decline marks the invitation as declined
func (i *Invitation) Decline() error {
    if i.Status != InvitationStatusPending {
        return errors.New("invitation is not pending")
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

// Expire marks the invitation as expired
func (i *Invitation) Expire() {
    i.Status = InvitationStatusExpired
    i.UpdatedAt = time.Now()
}

// GenerateToken generates a unique token for an invitation
func GenerateToken() string {
    return uuid.New().String()
}