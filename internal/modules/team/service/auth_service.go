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

	// GetUsersByIDs returns a map of userID → UserResult for the given
	// IDs. Users that don't exist are omitted from the map.
	//
	// Used by GetTeamMembersWithUsers to enrich team member rows with
	// display identity in one batch instead of N single lookups.
	GetUsersByIDs(ctx context.Context, userIDs []string) (map[string]*UserResult, error)

	// ============================================================
	// ACCOUNT QUERIES
	// ============================================================

	// GetAccountIDByUserID gets the account ID for a user
	GetAccountIDByUserID(ctx context.Context, userID string) (string, error)

	GetUserRoleInAccount(ctx context.Context, userID, accountID string) (string, error)

	AddAccountMember(ctx context.Context, accountID, userID, role string) error
}

// UserResult represents a user returned from auth service
type UserResult struct {
	ID           string
	Email        string
	Name         string
	Phone        string
	DisplayName  string
	AccountType  string
	AccountID    string
	IsActive     bool
	AccessToken  string
	RefreshToken string
	AvatarURL    string // added for team member enrichment
}