// internal/modules/team/service/service.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// Service defines the team module's business logic interface.
//
// POST-REVAMP notes:
//   - Teams are not authorization domains. Roles are assigned at the
//     ACCOUNT level and apply to every team under that account.
//   - Team membership changes do not touch Casbin.
//   - Every state-changing or read-of-another-user's-data method performs
//     a permission check against the team's parent account domain.
type Service interface {
	// ============================================================
	// TEAM OPERATIONS
	// ============================================================

	// CreatePersonalTeam creates a personal team for the given user.
	//
	// No role parameter: roles are assigned at the ACCOUNT level and
	// inherited by every team under that account. The team service does
	// not touch Casbin.
	CreatePersonalTeam(ctx context.Context, userID, userName string) (*teamdomain.Team, error)

	// CreateInstitutionTeam creates an institution team under the given account.
	//
	// No role parameter: same reasoning as above. The creator must already
	// be an account_admin in the target account (enforced via
	// casbinSvc.CanCreateTeam).
	CreateInstitutionTeam(ctx context.Context, accountID, name, displayName, slug, createdBy string) (*teamdomain.Team, error)

	GetTeamByID(ctx context.Context, id string) (*teamdomain.Team, error)
	GetUserTeams(ctx context.Context, userID string) ([]*teamdomain.Team, error)
	UpdateTeam(ctx context.Context, id string, updates map[string]interface{}) (*teamdomain.Team, error)
	DeleteTeam(ctx context.Context, id string) error

	// ============================================================
	// MEMBER OPERATIONS
	// ============================================================
	//
	// Roles are not modeled at the team level. Team members inherit their
	// account role.

	AddMember(ctx context.Context, teamID, userID, addedBy string) (*teamdomain.Member, error)
	RemoveMember(ctx context.Context, teamID, userID, removedBy string) error

	// GetTeamMembers requires the viewer's user ID so it can authorize the
	// request against the team's parent account domain.
	GetTeamMembers(
		ctx context.Context,
		viewerUserID, teamID string,
		filters teamdomain.ListMembersFilters,
	) ([]*teamdomain.Member, int64, error)

	GetUserMemberships(ctx context.Context, userID string) ([]*teamdomain.Member, error)
	LeaveTeam(ctx context.Context, teamID, userID string) error

	// ============================================================
	// INVITATION OPERATIONS
	// ============================================================

	InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error)
	ValidateInvitationToken(ctx context.Context, token string) (*teamdomain.Invitation, error)
	AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, error)
	DeclineInvitation(ctx context.Context, token, userID string) error
	ResendInvitation(ctx context.Context, invitationID string) (*teamdomain.Invitation, error)
	GetTeamInvitations(ctx context.Context, teamID string, filters teamdomain.ListInvitationsFilters) ([]*teamdomain.Invitation, int64, error)

	// ============================================================
	// TEAM MEMBERSHIP QUERIES (for Auth module and handlers)
	// ============================================================

	GetPersonalTeamByUserID(ctx context.Context, userID string) (*teamdomain.Team, error)
	GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*teamdomain.Team, error)
	GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error)
	GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error)
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)

	// ============================================================
	// DOMAIN HELPERS
	// ============================================================

	// GetTeamDomain returns the authz domain for a team.
	// POST-REVAMP: this is the team's parent ACCOUNT domain, not a
	// team-prefixed domain.
	GetTeamDomain(ctx context.Context, teamID string) (string, error)

	// GetAccountDomainForTeam returns the account domain for the team with
	// the given ID. Preferred over GetTeamDomain for new code.
	GetAccountDomainForTeam(ctx context.Context, teamID string) (string, error)

	// GetAccountDomain returns the domain for an account.
	GetAccountDomain(ctx context.Context, accountID string) (string, error)

	// GetUserPersonalTeamDomain returns the account domain for a user's
	// personal team.
	GetUserPersonalTeamDomain(ctx context.Context, userID string) (string, error)

	// GetUserInstitutionTeamDomains returns the account domains for the
	// institution teams the user belongs to (deduplicated).
	GetUserInstitutionTeamDomains(ctx context.Context, userID string) ([]string, error)
}

// ============================================================
// COMMANDS
// ============================================================

// InviteMemberCommand represents a request to invite a member.
//
// The Role field describes the invitee's ACCOUNT role. It is applied when
// the invitation is accepted and the invitee is not yet an account member.
//
// If the invitee is already an account member, the invited role is
// ignored; the existing role wins. Role changes require a separate
// operation.
//
// Invitation roles are scoped to the account the team belongs to. Teams
// do not have their own roles.
type InviteMemberCommand struct {
	TeamID    string
	Email     string
	Role      string // "account_admin" or "trainer"
	InvitedBy string
}

// ============================================================
// DISPLAY / TRANSFER TYPES
// ============================================================

// AccountInfo represents account information for a team.
type AccountInfo struct {
	ID      string
	Name    string
	Email   string
	Phone   string
	Website string
}

// TeamMemberInfo represents team membership information.
//
// Roles are not modeled here — they belong to the account, not the team.
type TeamMemberInfo struct {
	ID       string
	TeamID   string
	UserID   string
	IsActive bool
}

// MemberInfo represents member information for display.
type MemberInfo struct {
	ID         string
	TeamID     string
	UserID     string
	UserName   string
	UserEmail  string
	UserAvatar string
	JoinedAt   string
	IsActive   bool
}

// TeamInfo represents team information for display.
type TeamInfo struct {
	ID          string
	AccountID   string
	Name        string
	DisplayName string
	Slug        string
	Type        string
	MemberCount int
	CreatedAt   string
}

// InvitationInfo represents invitation information for display.
type InvitationInfo struct {
	ID            string
	Email         string
	Role          string
	Status        string
	ExpiresAt     string
	CreatedAt     string
	InvitedBy     string
	InvitedByName string
}