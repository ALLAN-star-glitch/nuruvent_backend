package attendancedomain

import (
	"fmt"
	"strings"
	"time"
)

// JoinToken is a per-(attendee, session) credential that authorizes a
// join-click and maps it back to the (attendee, session) pair.
//
// The raw token is never stored — only its hash. The raw token is
// returned to the caller once, at issue time, and never again.
type JoinToken struct {
	ID         string
	AttendeeID string
	SessionID  string
	TokenHash  string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

// NewJoinToken constructs a new join token.
func NewJoinToken(
	id, attendeeID, sessionID, tokenHash string,
	issuedAt, expiresAt time.Time,
) (*JoinToken, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(attendeeID) == "" || strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("%w: ids are required", ErrInvalidToken)
	}
	if len(tokenHash) != 64 {
		return nil, fmt.Errorf("%w: token hash must be a sha-256 hex string", ErrInvalidToken)
	}
	if !expiresAt.After(issuedAt) {
		return nil, fmt.Errorf("%w: expiry must be after issue", ErrInvalidToken)
	}
	return &JoinToken{
		ID:         id,
		AttendeeID: attendeeID,
		SessionID:  sessionID,
		TokenHash:  tokenHash,
		IssuedAt:   issuedAt,
		ExpiresAt:  expiresAt,
	}, nil
}

// HydrateJoinToken reconstructs a join token from persistence.
func HydrateJoinToken(
	id, attendeeID, sessionID, tokenHash string,
	issuedAt, expiresAt time.Time,
	revokedAt *time.Time,
) *JoinToken {
	return &JoinToken{
		ID:         id,
		AttendeeID: attendeeID,
		SessionID:  sessionID,
		TokenHash:  tokenHash,
		IssuedAt:   issuedAt,
		ExpiresAt:  expiresAt,
		RevokedAt:  revokedAt,
	}
}

// IsActive reports whether the token is usable at the given time.
func (t *JoinToken) IsActive(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return now.Before(t.ExpiresAt)
}

// Revoke marks the token as revoked. Idempotent.
func (t *JoinToken) Revoke(now time.Time) {
	if t.RevokedAt == nil {
		t.RevokedAt = &now
	}
}

// JoinTokenLifetime computes the expiry for a token issued for a
// session. Tokens expire at session end plus the given grace period.
func JoinTokenLifetime(session *Session, grace time.Duration) time.Time {
	return session.ScheduledEnd.Add(grace)
}