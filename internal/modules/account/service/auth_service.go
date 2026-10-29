// internal/modules/account/service/auth_service.go ----- OUTBOUND PORT

package service

import "context"

// AuthService defines the authentication operations needed by the account module
type AuthService interface {
    // GetUserByID retrieves a user by ID
    GetUserByID(ctx context.Context, userID string) (*UserResult, error)

    // GetUserByEmail retrieves a user by email
    GetUserByEmail(ctx context.Context, email string) (*UserResult, error)

    // UserExists checks if a user exists by email
    UserExists(ctx context.Context, email string) (bool, error)


    // NormalizeAndCheckPhone validates + normalizes a phone number to E.164
	// and rejects numbers already held by another active user.
	// Returns "" when raw is empty (phone is optional).
	NormalizeAndCheckPhone(ctx context.Context, raw string) (string, error)
}

// UserResult represents a user from auth service
type UserResult struct {
    ID          string
    Email       string
    Name        string
    Phone       string
    DisplayName string
    IsActive    bool
}