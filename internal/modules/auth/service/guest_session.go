package service

import (
	"context"
	"fmt"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// IssueSessionForUser mints a session for an already-identified user.
// Used by the join handler after a join token has been redeemed,
// so guests land on the meeting page authenticated.
//
// Differs from the standard login flow in that it:
//   - skips password verification (the caller already knows the user)
//   - skips 2FA OTP issuance
//   - honors an explicit TTL (defaults to 4 hours)
//   - issues only an access token, no refresh token
//
// The returned expiresAt is the JWT's actual exp claim. Previously
// this returned `now + ttl` while the token itself was minted with
// the default 15-minute TTL — the cookie then outlived the token,
// and every request past 15 minutes 401'd with "Invalid or expired
// token". This version mints the token with the requested TTL so the
// two stay in sync.
func (s *service) IssueSessionForUser(
	ctx context.Context,
	userID string,
	ttl time.Duration,
) (token string, expiresAt time.Time, err error) {
	if userID == "" {
		return "", time.Time{}, fmt.Errorf("user_id is required")
	}
	if ttl <= 0 {
		ttl = 4 * time.Hour
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("load user: %w", err)
	}
	if user == nil {
		return "", time.Time{}, fmt.Errorf("user not found: %s", userID)
	}
	if !user.IsActiveUser() {
		return "", time.Time{}, authdomain.ErrUserInactive
	}

	// Build the TokenContext the same way login does — resolves role,
	// account, and team from memberships. Shared with GenerateTokens so
	// guest sessions and login sessions carry identical context.
	tokenCtx, err := s.GetTokenContext(ctx, user)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("build token context: %w", err)
	}

	// Mint with the requested TTL. The JWT's exp claim is `now + ttl`,
	// matching the value returned below.
	accessToken, err := s.tokenSvc.GenerateAccessTokenWithTTL(tokenCtx, ttl)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate access token: %w", err)
	}

	return accessToken, time.Now().Add(ttl), nil
}