// internal/modules/team/service/auth_service.go

package service

import "context"

// AuthService defines the authentication operations needed by the team module
type AuthService interface {
	// ============================================================
	// USER QUERIES
	// ============================================================

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, userID string) (*UserResult, error)

	// GetUserByIDWithAccount retrieves a user by ID with their account ID
	GetUserByIDWithAccount(ctx context.Context, userID string) (*UserResult, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*UserResult, error)

	// UserExists checks if a user exists by email
	UserExists(ctx context.Context, email string) (bool, error)

	// ============================================================
	// ACCOUNT QUERIES
	// ============================================================

	// GetAccountIDByUserID gets the account ID for a user
	GetAccountIDByUserID(ctx context.Context, userID string) (string, error)


	GetUserRoleInAccount(ctx context.Context, userID, accountID string) (string, error)
}

// UserResult represents a user returned from auth service
type UserResult struct {
	ID           string
	Email        string
	Name         string
	Phone        string
	DisplayName  string
	AccountType  string
	AccountID    string // ✅ Added: Account ID from account_members
	IsActive     bool
	AccessToken  string
	RefreshToken string
}