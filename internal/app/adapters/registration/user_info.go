// internal/app/adapters/registration/user_info.go

package registration

import (
	"context"
	"fmt"
	"strings"

	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// UserInfoAdapter implements the registration module's
// UserInfoProvider port by delegating to the auth service.
type UserInfoAdapter struct {
	auth authService.Service
}

func NewUserInfoAdapter(auth authService.Service) registrationDomain.UserInfoProvider {
	return &UserInfoAdapter{auth: auth}
}

func (a *UserInfoAdapter) GetUserInfo(
	ctx context.Context,
	userID string,
) (displayName, email string, err error) {
	if userID == "" {
		return "", "", fmt.Errorf("user_id is required")
	}
	user, err := a.auth.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("auth.GetUserByID: %w", err)
	}
	if user == nil {
		return "", "", fmt.Errorf("user %s not found", userID)
	}
	return user.Name, user.Email, nil
}

func (a *UserInfoAdapter) GetUsername(
	ctx context.Context,
	userID string,
) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user_id is required")
	}
	user, err := a.auth.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("auth.GetUserByID: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("user %s not found", userID)
	}
	return user.Username, nil
}

// GetPhone returns the user's contact phone, trimmed. Empty when the
// user has none on file. Mirrors GetUsername.
func (a *UserInfoAdapter) GetPhone(
	ctx context.Context,
	userID string,
) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user_id is required")
	}
	user, err := a.auth.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("auth.GetUserByID: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("user %s not found", userID)
	}
	return strings.TrimSpace(user.Phone), nil
}