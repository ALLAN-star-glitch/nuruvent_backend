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
//   - issues a short-lived token (default 4 hours)
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

	accessToken, _, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate tokens: %w", err)
	}

	return accessToken, time.Now().Add(ttl), nil
}