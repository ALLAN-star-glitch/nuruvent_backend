package attendance

import (
	"context"
	"time"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
)

type AuthSessionIssuerAdapter struct {
	auth authService.Service
}

func NewAuthSessionIssuerAdapter(
	auth authService.Service,
) attendanceDomain.AuthSessionIssuer {
	return &AuthSessionIssuerAdapter{auth: auth}
}

func (a *AuthSessionIssuerAdapter) IssueSessionForUser(
	ctx context.Context,
	userID string,
	ttl time.Duration,
) (string, time.Time, error) {
	return a.auth.IssueSessionForUser(ctx, userID, ttl)
}