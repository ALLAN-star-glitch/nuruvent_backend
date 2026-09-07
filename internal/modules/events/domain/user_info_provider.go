// internal/modules/events/domain/user_info_provider.go

package domain

import "context"

// UserInfoProvider defines how the events domain fetches user and account information
type UserInfoProvider interface {
	// GetUserByID retrieves basic user information by ID
	// Includes: ID, Name, DisplayName, AvatarURL
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
	
	// GetUserByIDWithDetails retrieves full user information including email, phone, etc.
	// Includes: ID, Name, DisplayName, Email, Phone, AccountType, AvatarURL
	// This should only be called when the current user has permission
	GetUserByIDWithDetails(ctx context.Context, userID string) (*UserInfo, error)
	
	// GetAccountByID retrieves account information by ID
	// Includes: ID, Name, DisplayName, Slug, Type, LogoURL
	GetAccountByID(ctx context.Context, accountID string) (*AccountInfo, error)
}