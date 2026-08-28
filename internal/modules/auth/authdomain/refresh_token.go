package authdomain

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

// RefreshToken entity (formerly used AccountID, now UserID)
type RefreshToken struct {
    ID         string
    UserID     string   // formerly AccountID - the user who owns this token
    Token      string
    ExpiresAt  time.Time
    Revoked    bool
    UserAgent  string
    IPAddress  string
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

// NewRefreshToken creates a new refresh token for a user
func NewRefreshToken(userID, token, userAgent, ipAddress string, expiresAt time.Time) (*RefreshToken, error) {
    if userID == "" {
        return nil, errors.New("user ID is required")
    }
    if token == "" {
        return nil, errors.New("token is required")
    }
    if expiresAt.IsZero() {
        return nil, errors.New("expiration time is required")
    }

    now := time.Now()
    return &RefreshToken{
        ID:         uuid.New().String(),
        UserID:     userID,
        Token:      token,
        ExpiresAt:  expiresAt,
        Revoked:    false,
        UserAgent:  userAgent,
        IPAddress:  ipAddress,
        CreatedAt:  now,
        UpdatedAt:  now,
    }, nil
}

// IsValid checks if the token is not expired and not revoked
func (rt *RefreshToken) IsValid() bool {
    return !rt.IsExpired() && !rt.IsRevoked()
}

// IsExpired checks if the token has expired
func (rt *RefreshToken) IsExpired() bool {
    return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
    return rt.Revoked
}

// Revoke revokes the token
func (rt *RefreshToken) Revoke() {
    rt.Revoked = true
    rt.UpdatedAt = time.Now()
}

// Extend extends the token expiration
func (rt *RefreshToken) Extend(newExpiry time.Time) {
    rt.ExpiresAt = newExpiry
    rt.UpdatedAt = time.Now()
}

// UpdateUserAgent updates the user agent
func (rt *RefreshToken) UpdateUserAgent(userAgent string) {
    rt.UserAgent = userAgent
    rt.UpdatedAt = time.Now()
}

// UpdateIPAddress updates the IP address
func (rt *RefreshToken) UpdateIPAddress(ipAddress string) {
    rt.IPAddress = ipAddress
    rt.UpdatedAt = time.Now()
}