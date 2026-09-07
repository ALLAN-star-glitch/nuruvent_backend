// internal/app/adapters/accounts/auth_adapter.go

package accounts

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
)

// AuthAdapter adapts the auth module's service to the account module's AuthService interface
type AuthAdapter struct {
	authSvc authService.Service
}

// NewAuthAdapter creates a new auth adapter for the account module
func NewAuthAdapter(authSvc authService.Service) service.AuthService {
	return &AuthAdapter{
		authSvc: authSvc,
	}
}

// GetUserByID retrieves a user by ID
func (a *AuthAdapter) GetUserByID(ctx context.Context, userID string) (*service.UserResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := a.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	return &service.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

// GetUserByEmail retrieves a user by email
func (a *AuthAdapter) GetUserByEmail(ctx context.Context, email string) (*service.UserResult, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	user, err := a.authSvc.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	return &service.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

// UserExists checks if a user exists by email
func (a *AuthAdapter) UserExists(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("email is required")
	}

	exists, err := a.authSvc.UserExists(ctx, email)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}