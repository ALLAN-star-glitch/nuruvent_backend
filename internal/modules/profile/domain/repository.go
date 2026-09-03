// internal/modules/profile/domain/repository.go

package domain

import "context"

// ============================================================
// TEAM FILTER
// ============================================================

// TeamFilter identifies which team to filter by
// This is a simple data struct - NO methods
type TeamFilter struct {
    // ID is the team identifier
    // - For personal teams: user_id
    // - For institution teams: institution_id
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
    UpdateUser(ctx context.Context, user *User) error

    // ============================================================
    // INSTITUTION CRUD OPERATIONS
    // ============================================================

    GetInstitutionByID(ctx context.Context, id string) (*Institution, error)
    GetInstitutionBySlug(ctx context.Context, slug string) (*Institution, error)
    UpdateInstitution(ctx context.Context, institution *Institution) error

    // ============================================================
    // QUERY OPERATIONS
    // ============================================================

    // ListUsers returns a paginated list of users with flexible filtering
    // TeamFilter determines which team's users to return:
    //   - Type="personal", ID=userID → personal team users (just the user themselves)
    //   - Type="institution", ID=institutionID → institution team users (members)
    //   - Type="" → no team filter (all users)
    ListUsers(ctx context.Context, filters ListUsersFilters) ([]*User, int64, error)

    // ListInstitutions returns a paginated list of institutions with flexible filtering
    ListInstitutions(ctx context.Context, filters ListInstitutionsFilters) ([]*Institution, int64, error)

    // GetUsersByIDs retrieves multiple users by IDs
    GetUsersByIDs(ctx context.Context, ids []string) ([]*User, error)

    // GetInstitutionsByIDs retrieves multiple institutions by IDs
    GetInstitutionsByIDs(ctx context.Context, ids []string) ([]*Institution, error)
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

// ListInstitutionsFilters provides comprehensive filtering for ListInstitutions
type ListInstitutionsFilters struct {
    // TeamFilter filters institutions by team
    Team TeamFilter

    // InstitutionID filters by specific institution ID
    InstitutionID string

    // Search query for name or email
    Search string

    // IncludeDeleted controls whether soft-deleted institutions are included
    IncludeDeleted bool

    // OnlyDeleted controls whether ONLY soft-deleted institutions are returned
    OnlyDeleted bool

    // Limit controls the maximum number of institutions returned
    Limit int

    // Offset controls pagination offset
    Offset int

    // SortBy specifies the field to sort by
    SortBy string

    // SortOrder specifies the sort direction
    SortOrder string
}