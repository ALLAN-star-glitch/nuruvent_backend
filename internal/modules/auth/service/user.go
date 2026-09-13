// internal/modules/auth/service/user.go

package service

import (
	"context"
	"fmt"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, userID string) (*authdomain.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUserByIDWithAccount retrieves a user by ID with their account ID
// Returns: user, accountID, error
func (s *service) GetUserByIDWithAccount(ctx context.Context, userID string) (*authdomain.User, string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", nil
	}

	// Get the user's account ID from account_members
	members, err := s.repo.GetAccountMembersByUser(ctx, userID)
	if err != nil {
		// Return user without account ID if there's an error
		return user, "", nil
	}
	if len(members) == 0 {
		// User has no account yet
		return user, "", nil
	}

	// Use the first account (primary)
	return user, members[0].AccountID, nil
}

// GetUserByEmail retrieves a user by email
func (s *service) GetUserByEmail(ctx context.Context, email string) (*authdomain.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// GetUserByEmailWithAccount retrieves a user by email with their account ID
// Returns: user, accountID, error
func (s *service) GetUserByEmailWithAccount(ctx context.Context, email string) (*authdomain.User, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", nil
	}

	// Get the user's account ID from account_members
	members, err := s.repo.GetAccountMembersByUser(ctx, user.ID)
	if err != nil {
		return user, "", nil
	}
	if len(members) == 0 {
		return user, "", nil
	}

	return user, members[0].AccountID, nil
}

// GetAccountIDByUserID gets the account ID for a user
func (s *service) GetAccountIDByUserID(ctx context.Context, userID string) (string, error) {
	members, err := s.repo.GetAccountMembersByUser(ctx, userID)
	if err != nil {
		return "", err
	}
	if len(members) == 0 {
		return "", fmt.Errorf("user has no account")
	}
	return members[0].AccountID, nil
}

// UserExists checks if a user exists by email
func (s *service) UserExists(ctx context.Context, email string) (bool, error) {
	return s.repo.UserExistsByEmail(ctx, email)
}