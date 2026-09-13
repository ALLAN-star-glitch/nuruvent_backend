// internal/app/adapters/team/auth_adapter.go

package team

import (
	"context"
	"fmt"

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

// ============================================================
// USER QUERIES
// ============================================================

// GetUserByID retrieves a user by ID
func (a *AuthAdapter) GetUserByID(ctx context.Context, userID string) (*teamService.UserResult, error) {
	user, err := a.authRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	// Get account ID from account_members
	accountID, _ := a.getAccountIDByUserID(ctx, userID)

	return &teamService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		AccountID:   accountID,
		IsActive:    user.IsActive,
	}, nil
}

// GetUserByIDWithAccount retrieves a user by ID with their account ID
func (a *AuthAdapter) GetUserByIDWithAccount(ctx context.Context, userID string) (*teamService.UserResult, error) {
	user, err := a.authRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	accountID, err := a.getAccountIDByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &teamService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		AccountID:   accountID,
		IsActive:    user.IsActive,
	}, nil
}

// GetUserByEmail retrieves a user by email
func (a *AuthAdapter) GetUserByEmail(ctx context.Context, email string) (*teamService.UserResult, error) {
	user, err := a.authRepo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, err
	}

	accountID, _ := a.getAccountIDByUserID(ctx, user.ID)

	return &teamService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		AccountID:   accountID,
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

// ============================================================
// ACCOUNT QUERIES
// ============================================================

// GetAccountIDByUserID gets the account ID for a user
func (a *AuthAdapter) GetAccountIDByUserID(ctx context.Context, userID string) (string, error) {
	return a.getAccountIDByUserID(ctx, userID)
}

func (a *AuthAdapter) AddAccountMember(ctx context.Context, accountID, userID, role string) error {
	// Validate inputs.
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if role == "" {
		return fmt.Errorf("role is required")
	}

	
	// Construct the domain entity.
	member, err := authDomain.NewAccountMember(accountID, userID, role, userID)
	if err != nil {
		return fmt.Errorf("failed to construct account member: %w", err)
	}

	// Persist via the auth repository.
	if err := a.authRepo.CreateAccountMember(ctx, member); err != nil {
		return fmt.Errorf("failed to create account member: %w", err)
	}

	return nil
}

// GetUserRoleInAccount gets a user's role in a specific account
// ✅ Added: Required for AddMember to assign Casbin team role
func (a *AuthAdapter) GetUserRoleInAccount(ctx context.Context, userID, accountID string) (string, error) {
	members, err := a.authRepo.GetAccountMembersByUser(ctx, userID)
	if err != nil {
		return "", err
	}

	for _, member := range members {
		if member.AccountID == accountID {
			return member.Role, nil
		}
	}

	return "", nil
}


// ============================================================
// PRIVATE HELPERS
// ============================================================

// getAccountIDByUserID gets the account ID for a user from account_members
func (a *AuthAdapter) getAccountIDByUserID(ctx context.Context, userID string) (string, error) {
	members, err := a.authRepo.GetAccountMembersByUser(ctx, userID)
	if err != nil {
		return "", err
	}
	if len(members) == 0 {
		return "", nil
	}
	// Use the first account (primary)
	return members[0].AccountID, nil
}


