package authdomain

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

type RefreshToken struct {
    ID         string
    AccountID  string
    Token      string
    ExpiresAt  time.Time
    Revoked    bool
    UserAgent  string
    IPAddress  string
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

func NewRefreshToken(accountID, token, userAgent, ipAddress string, expiresAt time.Time) (*RefreshToken, error) {
    if accountID == "" {
        return nil, errors.New("account ID is required")
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
        AccountID:  accountID,
        Token:      token,
        ExpiresAt:  expiresAt,
        Revoked:    false,
        UserAgent:  userAgent,
        IPAddress:  ipAddress,
        CreatedAt:  now,
        UpdatedAt:  now,
    }, nil
}

func (rt *RefreshToken) IsValid() bool {
    return !rt.IsExpired() && !rt.IsRevoked()
}

func (rt *RefreshToken) IsExpired() bool {
    return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsRevoked() bool {
    return rt.Revoked
}

func (rt *RefreshToken) Revoke() {
    rt.Revoked = true
    rt.UpdatedAt = time.Now()
}