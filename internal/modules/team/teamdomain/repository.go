// internal/modules/team/teamdomain/repository.go

package teamdomain

import "context"

// ============================================================
// REPOSITORY INTERFACE
// ============================================================

// Repository defines the data access interface for the team module.
type Repository interface {

	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error

	// ============================================================
	// TEAM OPERATIONS
	// ============================================================

	CreateTeam(ctx context.Context, team *Team) error
	GetTeamByID(ctx context.Context, id string) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug string) (*Team, error)
	GetTeamsByUserID(ctx context.Context, userID string) ([]*Team, error)
	UpdateTeam(ctx context.Context, team *Team) error
	DeleteTeam(ctx context.Context, id string) error

	// AccountIDForTeam resolves a team ID to its parent account ID.
	//
	// Under the post-revamp authorization model, every team belongs to an
	// account (teams.account_id NOT NULL). This method is the canonical way
	// to translate a team-scoped request into an account-scoped permission
	// check: build the domain with accountdomain.AccountDomain(accountID).
	//
	// Returns an error if the team does not exist. Returns an empty string
	// if the team exists but has no account (which should never happen
	// given the NOT NULL constraint, but is handled defensively).
	AccountIDForTeam(ctx context.Context, teamID string) (string, error)

	// ============================================================
	// MEMBER OPERATIONS
	// ============================================================

	CreateMember(ctx context.Context, member *Member) error
	GetMemberByID(ctx context.Context, id string) (*Member, error)
	GetMemberByTeamAndUser(ctx context.Context, teamID, userID string) (*Member, error)
	GetMembersByTeam(ctx context.Context, teamID string, filters ListMembersFilters) ([]*Member, int64, error)
	GetMembersByUser(ctx context.Context, userID string) ([]*Member, error)
	UpdateMember(ctx context.Context, member *Member) error
	DeleteMember(ctx context.Context, teamID, userID string) error
	CountMembersByTeam(ctx context.Context, teamID string) (int64, error)

	// ============================================================
	// INVITATION OPERATIONS
	// ============================================================

	CreateInvitation(ctx context.Context, invitation *Invitation) error
	GetInvitationByID(ctx context.Context, id string) (*Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*Invitation, error)
	GetInvitationByEmailAndTeam(ctx context.Context, email, teamID string) (*Invitation, error)
	GetInvitationsByTeam(ctx context.Context, teamID string, filters ListInvitationsFilters) ([]*Invitation, int64, error)
	UpdateInvitation(ctx context.Context, invitation *Invitation) error
	DeleteInvitation(ctx context.Context, id string) error
	GetPendingInvitations(ctx context.Context) ([]*Invitation, error)
}

// ============================================================
// FILTER STRUCTS
// ============================================================

// ListMembersFilters provides filtering for listing members.
type ListMembersFilters struct {
	Search    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// ListInvitationsFilters provides filtering for listing invitations.
type ListInvitationsFilters struct {
	Email     string
	Status    InvitationStatus
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}