// internal/modules/profile/domain/repository.go

package domain

import "context"

// ============================================================
// TEAM FILTER
// ============================================================

// TeamFilter identifies which team to filter by
type TeamFilter struct {
	// ID is the team identifier
	// - For personal teams: user_id
	// - For institution teams: account_id
	ID string

	// Type indicates the team type
	// Valid values: "personal" or "institution"
	// If empty, no team filter is applied
	Type string
}

// ============================================================
// REPOSITORY INTERFACE
// ============================================================

// Repository defines the data access interface for the profile module
type Repository interface {
	// ============================================================
	// USER CRUD OPERATIONS
	// ============================================================

	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*User, error)
	UpdateUser(ctx context.Context, user *User) error

	// ============================================================
	// ACCOUNT CRUD OPERATIONS
	// ============================================================

	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccountBySlug(ctx context.Context, slug string) (*Account, error)
	GetAccountsByIDs(ctx context.Context, ids []string) ([]*Account, error)
	UpdateAccount(ctx context.Context, account *Account) error

	// ============================================================
	// QUERY OPERATIONS
	// ============================================================

	// ListUsers returns a paginated list of users with flexible filtering
	// TeamFilter determines which team's users to return:
	//   - Type="personal", ID=userID → personal team users (just the user themselves)
	//   - Type="institution", ID=accountID → institution team users (members)
	//   - Type="" → no team filter (all users)
	ListUsers(ctx context.Context, filters ListUsersFilters) ([]*User, int64, error)

	// ListAccounts returns a paginated list of accounts with flexible filtering
	ListAccounts(ctx context.Context, filters ListAccountsFilters) ([]*Account, int64, error)
}

// ============================================================
// FILTER STRUCTS
// ============================================================

// ListUsersFilters provides comprehensive filtering for ListUsers
type ListUsersFilters struct {
	// TeamFilter filters users by team (personal or institution)
	Team TeamFilter

	// UserID filters by specific user ID
	UserID string

	// Search query for name or email
	Search string

	// IncludeDeleted controls whether soft-deleted users are included
	IncludeDeleted bool

	// OnlyDeleted controls whether ONLY soft-deleted users are returned
	OnlyDeleted bool

	// Limit controls the maximum number of users returned
	Limit int

	// Offset controls pagination offset
	Offset int

	// SortBy specifies the field to sort by
	SortBy string

	// SortOrder specifies the sort direction
	SortOrder string
}

// ListAccountsFilters provides comprehensive filtering for ListAccounts
type ListAccountsFilters struct {
	// Type filters by account type: "personal" or "institution"
	Type string

	// AccountID filters by specific account ID
	AccountID string

	// Search query for name or email
	Search string

	// Status filters by account status
	Status string

	// KYCStatus filters by KYC status
	KYCStatus string

	// IncludeDeleted controls whether soft-deleted accounts are included
	IncludeDeleted bool

	// OnlyDeleted controls whether ONLY soft-deleted accounts are returned
	OnlyDeleted bool

	// Limit controls the maximum number of accounts returned
	Limit int

	// Offset controls pagination offset
	Offset int

	// SortBy specifies the field to sort by
	SortBy string

	// SortOrder specifies the sort direction
	SortOrder string
}