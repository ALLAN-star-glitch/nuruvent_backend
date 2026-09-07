// internal/modules/profile/service/service.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// Service defines the profile service interface
type Service interface {
	// ============================================================
	// USER PROFILE METHODS
	// ============================================================

	GetUserProfile(ctx context.Context, userID string) (*domain.UserInfo, error)
	GetUserProfileWithDetails(ctx context.Context, userID string) (*domain.UserInfo, error)
	GetUserProfiles(ctx context.Context, userIDs []string) ([]*domain.UserInfo, error)
	UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*domain.UserInfo, error)
	ListUsers(ctx context.Context, filters domain.ListUsersFilters) ([]*domain.UserInfo, int64, error)

	// ============================================================
	// ACCOUNT PROFILE METHODS
	// ============================================================

	GetAccountProfile(ctx context.Context, accountID string) (*domain.AccountInfo, error)
	GetAccountProfileWithDetails(ctx context.Context, accountID string) (*domain.AccountInfo, error)
	GetAccountProfiles(ctx context.Context, accountIDs []string) ([]*domain.AccountInfo, error)
	UpdateAccountProfile(ctx context.Context, accountID string, updates map[string]interface{}) (*domain.AccountInfo, error)
	ListAccounts(ctx context.Context, filters domain.ListAccountsFilters) ([]*domain.AccountInfo, int64, error)

	// ============================================================
	// ORGANIZER INFO (FOR EVENTS MODULE)
	// ============================================================

	GetOrganizerInfo(ctx context.Context, organizerType, organizerID string) (*domain.OrganizerInfo, error)

	// ============================================================
	// MEDIA UPLOAD METHODS
	// ============================================================

	UploadUserAvatar(ctx context.Context, userID string, file []byte, filename, contentType string) (*domain.UserInfo, error)
	UploadAccountLogo(ctx context.Context, accountID string, file []byte, filename, contentType string) (*domain.AccountInfo, error)
	DeleteUserAvatar(ctx context.Context, userID string) error
	DeleteAccountLogo(ctx context.Context, accountID string) error
}



