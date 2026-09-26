// internal/modules/video/videodomain/connection.go

package videodomain

import (
	"fmt"
	"strings"
	"time"
)

// Connection is a host's authorized link to their account on a video
// platform.
//
// Exactly one active connection per (UserID, Platform) is allowed.
// Tokens are stored encrypted; the domain treats them as opaque
// strings and never inspects their contents.
type Connection struct {
	ID     string
	UserID string

	// Platform identifies which video platform this connection is for.
	Platform Platform

	// Identity as returned by the platform.
	ExternalUserID  string // Zoom user ID, Google user ID, etc.
	ExternalEmail   string
	ExternalOrgID   string // Zoom account_id, Google workspace domain, etc.

	// OAuth tokens. Encrypted at rest by the repository; the domain
	// holds plaintext values.
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt time.Time

	// Scopes granted by the platform, space-separated.
	Scopes string

	// Lifecycle.
	ConnectedAt time.Time
	RevokedAt   *time.Time
}

// NewConnection constructs a new connection with validation.
func NewConnection(
	id, userID string,
	platform Platform,
	externalUserID, externalEmail, externalOrgID string,
	accessToken, refreshToken string,
	tokenExpiresAt time.Time,
	scopes string,
	now time.Time,
) (*Connection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidConnection)
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidConnection)
	}
	if !platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform %q", ErrInvalidConnection, platform)
	}
	if strings.TrimSpace(externalUserID) == "" {
		return nil, fmt.Errorf("%w: external user id is required", ErrInvalidConnection)
	}
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("%w: access token is required", ErrInvalidConnection)
	}
	if tokenExpiresAt.IsZero() {
		return nil, fmt.Errorf("%w: token expiry is required", ErrInvalidConnection)
	}

	return &Connection{
		ID:              id,
		UserID:          userID,
		Platform:        platform,
		ExternalUserID:  externalUserID,
		ExternalEmail:   strings.ToLower(strings.TrimSpace(externalEmail)),
		ExternalOrgID:   externalOrgID,
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		TokenExpiresAt:  tokenExpiresAt,
		Scopes:          strings.TrimSpace(scopes),
		ConnectedAt:     now,
	}, nil
}

// HydrateConnection reconstructs a connection from persistence without
// re-validating.
func HydrateConnection(
	id, userID string,
	platform Platform,
	externalUserID, externalEmail, externalOrgID string,
	accessToken, refreshToken string,
	tokenExpiresAt time.Time,
	scopes string,
	connectedAt time.Time,
	revokedAt *time.Time,
) *Connection {
	return &Connection{
		ID:              id,
		UserID:          userID,
		Platform:        platform,
		ExternalUserID:  externalUserID,
		ExternalEmail:   externalEmail,
		ExternalOrgID:   externalOrgID,
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		TokenExpiresAt:  tokenExpiresAt,
		Scopes:          scopes,
		ConnectedAt:     connectedAt,
		RevokedAt:       revokedAt,
	}
}

// IsActive reports whether the connection is usable at the given time.
func (c *Connection) IsActive(now time.Time) bool {
	if c.RevokedAt != nil {
		return false
	}
	return now.After(c.ConnectedAt) || now.Equal(c.ConnectedAt)
}

// IsTokenExpired reports whether the access token needs refreshing.
// Treats tokens expiring within the next 5 minutes as expired, so
// callers refresh proactively.
func (c *Connection) IsTokenExpired(now time.Time) bool {
	if c.TokenExpiresAt.IsZero() {
		return true
	}
	return now.Add(5 * time.Minute).After(c.TokenExpiresAt)
}

// Refresh updates the tokens after a successful refresh. Called by the
// service after the provider client returns new credentials.
func (c *Connection) Refresh(
	accessToken, refreshToken string,
	expiresAt time.Time,
) error {
	if strings.TrimSpace(accessToken) == "" {
		return fmt.Errorf("%w: access token is required", ErrInvalidConnection)
	}
	if expiresAt.IsZero() {
		return fmt.Errorf("%w: token expiry is required", ErrInvalidConnection)
	}
	c.AccessToken = accessToken
	if strings.TrimSpace(refreshToken) != "" {
		// Some providers return a new refresh token; some don't.
		c.RefreshToken = refreshToken
	}
	c.TokenExpiresAt = expiresAt
	return nil
}

// Revoke marks the connection as revoked. Idempotent.
func (c *Connection) Revoke(now time.Time) {
	if c.RevokedAt == nil {
		c.RevokedAt = &now
	}
}