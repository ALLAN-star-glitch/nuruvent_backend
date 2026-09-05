// internal/modules/team/service/auth_service.go

package service

import "context"

// ============================================================
// AUTH SERVICE INTERFACE (Outbound Port)
// ============================================================

// AuthService defines the authentication operations needed by the team module
type AuthService interface {
    // CreateUser creates a user account
    CreateUser(ctx context.Context, req CreateUserRequest) (*UserResult, error)

    // GetUserByID retrieves a user by ID
    GetUserByID(ctx context.Context, userID string) (*UserResult, error)

    // GetUserByEmail retrieves a user by email
    GetUserByEmail(ctx context.Context, email string) (*UserResult, error)

    // UserExists checks if a user exists by email
    UserExists(ctx context.Context, email string) (bool, error)
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
    Email     string
    Password  string
    Name      string
    Phone     string
    AccountType string
}

// UserResult represents a user returned from auth service
type UserResult struct {
    ID           string
    Email        string
    Name         string
    Phone        string
    DisplayName  string
    AccountType  string
    IsActive     bool
    AccessToken  string
    RefreshToken string
}