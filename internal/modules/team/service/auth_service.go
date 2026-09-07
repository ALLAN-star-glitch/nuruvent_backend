// internal/modules/team/service/auth_service.go

package service

import "context"

// AuthService defines the authentication operations needed by the team module
type AuthService interface {
	GetUserByID(ctx context.Context, userID string) (*UserResult, error)
	GetUserByEmail(ctx context.Context, email string) (*UserResult, error)
	UserExists(ctx context.Context, email string) (bool, error)
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