// internal/app/adapters/team/auth_adapter.go

package team

import (
	"context"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// AuthAdapter adapts auth repository to team module's AuthService interface
type AuthAdapter struct {
	authRepo authDomain.Repository
}

// NewAuthAdapter creates a new team auth adapter using auth domain repository directly
func NewAuthAdapter(authRepo authDomain.Repository) teamService.AuthService {
	return &AuthAdapter{authRepo: authRepo}
}

// GetUserByID retrieves a user by ID
func (a *AuthAdapter) GetUserByID(ctx context.Context, userID string) (*teamService.UserResult, error) {
	user, err := a.authRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	return &teamService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

// GetUserByEmail retrieves a user by email
func (a *AuthAdapter) GetUserByEmail(ctx context.Context, email string) (*teamService.UserResult, error) {
	user, err := a.authRepo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, err
	}

	return &teamService.UserResult{
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
	user, err := a.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return user != nil, nil
} 