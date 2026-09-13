// internal/modules/account/service/service.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
)

// Service defines the account module's business logic interface.
//
// Owns:
//   - Account types (lookup data)
//   - Accounts (personal + institution)
//   - Account members (user ↔ account, with roles)
//   - User info projections (for cross-module consumers like events)
type Service interface {
	// ============================================================
	// ACCOUNT TYPE OPERATIONS
	// ============================================================

	GetAccountTypes(ctx context.Context) ([]*accountdomain.AccountType, error)
	GetAccountTypeByID(ctx context.Context, id string) (*accountdomain.AccountType, error)
	GetAccountTypeBySlug(ctx context.Context, slug string) (*accountdomain.AccountType, error)

	// ============================================================
	// ACCOUNT OPERATIONS
	// ============================================================

	// GetAccountByTeamID resolves the account that owns a given team.
	// Returns (nil, nil) if the team doesn't exist or has no account.
	GetAccountByTeamID(ctx context.Context, teamID string) (*accountdomain.Account, error)

	CreatePersonalAccount(ctx context.Context, cmd CreatePersonalAccountCommand) (*accountdomain.Account, error)
	CreateInstitutionAccount(ctx context.Context, cmd CreateInstitutionAccountCommand) (*accountdomain.Account, error)

	GetAccountByID(ctx context.Context, id string) (*accountdomain.Account, error)
	GetAccountBySlug(ctx context.Context, slug string) (*accountdomain.Account, error)
	GetUserAccounts(ctx context.Context, userID string) ([]*accountdomain.Account, error)

	UpdateAccount(ctx context.Context, id string, updates map[string]interface{}) (*accountdomain.Account, error)
	DeleteAccount(ctx context.Context, id string) error

	// ============================================================
	// ACCOUNT MEMBER OPERATIONS
	// ============================================================

	AddMember(ctx context.Context, cmd AddMemberCommand) (*accountdomain.AccountMember, error)
	RemoveMember(ctx context.Context, accountID, userID, removedBy string) error
	UpdateMemberRole(ctx context.Context, accountID, userID, newRole, updatedBy string) (*accountdomain.AccountMember, error)
	GetAccountMembers(ctx context.Context, accountID string) ([]*accountdomain.AccountMember, error)
	LeaveAccount(ctx context.Context, accountID, userID string) error

	// ============================================================
	// USER INFO PROJECTIONS
	// ============================================================
	//
	// Read-only user lookups for cross-module consumers (events creator
	// display, audit logs, etc.). These replace what the profile module
	// used to provide.
	//
	// The returned UserInfo is a projection — not the full user entity.
	// No permission checks are performed here; callers authorize access
	// at their own boundary.

	GetUserByID(ctx context.Context, userID string) (*accountdomain.UserInfo, error)
	GetUserByIDWithDetails(ctx context.Context, userID string) (*accountdomain.UserInfo, error)
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]*accountdomain.UserInfo, error)

    // ============================================================
    // MEDIA OPERATIONS (avatars, logos)
    // ============================================================

    // UploadUserAvatar stores a new avatar for the user and updates avatar_url.
    // Self-service: the caller may only change their own avatar.
    UploadUserAvatar(ctx context.Context, userID string, file []byte, filename, contentType string) (*accountdomain.UserInfo, error)
    DeleteUserAvatar(ctx context.Context, userID string) error

    // UploadAccountLogo stores a new logo for the account and updates logo_url.
    // Permission: caller must be able to manage the account.
    UploadAccountLogo(ctx context.Context, accountID string, file []byte, filename, contentType string) (*accountdomain.Account, error)
    DeleteAccountLogo(ctx context.Context, accountID string) error

        // ============================================================
    // USER PROFILE OPERATIONS
    // ============================================================

    // GetMyProfile returns the authenticated user's own profile.
    // Self-service: no permission check beyond authentication.
    GetMyProfile(ctx context.Context, userID string) (*accountdomain.UserInfo, error)

    // UpdateMyProfile updates the authenticated user's own profile.
    // Self-service: no permission check beyond authentication.
    UpdateMyProfile(ctx context.Context, userID string, updates ProfileUpdates) (*accountdomain.UserInfo, error)

    // GetPublicProfile returns a user's public-facing profile by ID.
    // No auth required; returns only safe fields.
    GetPublicProfile(ctx context.Context, userID string) (*accountdomain.PublicProfile, error)

    // GetPublicProfileBySlug returns a user's public profile by slug.
    GetPublicProfileBySlug(ctx context.Context, slug string) (*accountdomain.PublicProfile, error)
}

// ============================================================
// COMMANDS
// ============================================================

// CreatePersonalAccountCommand — used when a user creates a personal account.
type CreatePersonalAccountCommand struct {
	Name      string // User's name (e.g., "John Doe")
	Email     string // User's email
	Phone     string // User's phone
	CreatedBy string // User ID of the creator
}

// CreateInstitutionAccountCommand — used when a user creates an institution account.
type CreateInstitutionAccountCommand struct {
	Name        string // Institution name (e.g., "Nuruvent Ltd")
	Email       string // Institution email (must be professional)
	Phone       string // Institution phone
	Website     string // Institution website
	Description string // Institution description
	CreatedBy   string // User ID of the creator
}

// AddMemberCommand — used when adding a member to an account.
type AddMemberCommand struct {
	AccountID string
	UserID    string
	Role      string // "account_admin" or "trainer"
	InvitedBy string
}

// UpdateAccountCommand — used when updating an account.
type UpdateAccountCommand struct {
	Name        string
	Email       string
	Phone       string
	Website     string
	Description string
	LogoURL     string
}

// ProfileUpdates carries optional profile field updates.
// Nil pointers mean "do not change". Empty strings mean "clear".
type ProfileUpdates struct {
	DisplayName *string
	Phone       *string
	Bio         *string
	Location    *string
	Website     *string
	SocialLinks *map[string]string
}