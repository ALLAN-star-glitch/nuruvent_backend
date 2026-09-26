// internal/modules/video/videodomain/oauth_state.go

package videodomain

import (
	"fmt"
	"strings"
	"time"
)

// OAuthStateTTL is how long a state remains usable after creation.
// Short enough to prevent stale states from lingering; long enough
// for a user to complete the authorization flow on the platform.
const OAuthStateTTL = 10 * time.Minute

// OAuthState is a single-use CSRF protection record for the OAuth
// flow.
//
// The state is generated before the user leaves Nuruvent for the
// platform's authorize page. When the platform redirects back, the
// state must match an unconsumed, unexpired record. This prevents an
// attacker from forcing a victim's browser to complete an OAuth flow
// they didn't initiate.
type OAuthState struct {
	State     string
	UserID    string
	Platform  Platform
	ReturnURL string
	CreatedAt time.Time
	ExpiresAt time.Time

	// ConsumedAt is set when the state is claimed. Non-nil means the
	// state cannot be used again.
	ConsumedAt *time.Time
}

// NewOAuthState constructs a new state.
func NewOAuthState(
	state, userID string,
	platform Platform,
	returnURL string,
	now time.Time,
) (*OAuthState, error) {
	if strings.TrimSpace(state) == "" {
		return nil, fmt.Errorf("%w: state is required", ErrInvalidOAuthState)
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidOAuthState)
	}
	if !platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform %q", ErrInvalidOAuthState, platform)
	}
	return &OAuthState{
		State:     state,
		UserID:    userID,
		Platform:  platform,
		ReturnURL: strings.TrimSpace(returnURL),
		CreatedAt: now,
		ExpiresAt: now.Add(OAuthStateTTL),
	}, nil
}

// HydrateOAuthState reconstructs a state from persistence.
func HydrateOAuthState(
	state, userID string,
	platform Platform,
	returnURL string,
	createdAt, expiresAt time.Time,
	consumedAt *time.Time,
) *OAuthState {
	return &OAuthState{
		State:      state,
		UserID:     userID,
		Platform:   platform,
		ReturnURL:  returnURL,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
		ConsumedAt: consumedAt,
	}
}

// IsUsable reports whether the state can be consumed at the given
// time. A state is usable if it hasn't been consumed and hasn't
// expired.
func (s *OAuthState) IsUsable(now time.Time) bool {
	if s.ConsumedAt != nil {
		return false
	}
	return now.Before(s.ExpiresAt)
}

// Consume marks the state as used. Idempotent — the caller should
// check IsUsable first if a specific error is needed.
func (s *OAuthState) Consume(now time.Time) {
	if s.ConsumedAt == nil {
		s.ConsumedAt = &now
	}
}

// ValidateFor returns a specific error describing why a state can't
// be used, or nil if it can. Useful for the service to return
// precise errors instead of a generic "invalid".
func (s *OAuthState) ValidateFor(now time.Time) error {
	if s.ConsumedAt != nil {
		return ErrOAuthStateConsumed
	}
	if !now.Before(s.ExpiresAt) {
		return ErrOAuthStateExpired
	}
	return nil
}