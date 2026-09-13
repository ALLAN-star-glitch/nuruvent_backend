// internal/modules/auth/service/team_service.go

package service

import "context"

// TeamService defines the team operations needed by the auth module.
type TeamService interface {
	// ============================================================
	// TEAM QUERIES
	// ============================================================
	GetTeamByID(ctx context.Context, teamID string) (*TeamInfo, error)
	GetPersonalTeamByUserID(ctx context.Context, userID string) (*TeamInfo, error)
	GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*TeamInfo, error)
	GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error)

	// ============================================================
	// TEAM CREATION
	// ============================================================

	// CreatePersonalTeam creates a personal team for the given user.
	//
	// No role parameter: roles are assigned at the ACCOUNT level and
	// inherited by every team under that account. The team service does
	// not touch Casbin.
	CreatePersonalTeam(ctx context.Context, userID, userName string) (*TeamInfo, error)

	// CreateInstitutionTeam creates an institution team under the given account.
	//
	// No role parameter: same reasoning as above.
	CreateInstitutionTeam(ctx context.Context, accountID, name, displayName, slug, createdBy string) (*TeamInfo, error)

	// ============================================================
	// MEMBERSHIP QUERIES
	// ============================================================
	GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error)
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)

	// ============================================================
	// INVITATION LIFECYCLE
	// ============================================================

	// ValidateInvitationToken returns the invitation for the given token
	// if and only if it is usable: exists, status = pending, not expired.
	//
	// Used by RegisterWithInvitation to prove email ownership before the
	// user is created. The returned InvitationInfo.Email is the email the
	// user must be created with.
	//
	// Returns an error (not nil, nil) if the token is invalid, expired,
	// or already accepted — the caller is expected to surface that error
	// directly to the client.
	ValidateInvitationToken(ctx context.Context, token string) (*InvitationInfo, error)

	// AcceptInvitation accepts the invitation identified by token on behalf
	// of the given user. It:
	//   - creates the account membership (if the user isn't already a member)
	//   - creates the team membership
	//   - writes the Casbin `g` rule for the invited role
	//   - marks the invitation accepted
	//
	// It does NOT create a personal account or personal team. Invited users
	// join the inviter's account only.
	//
	// Returns the InvitationInfo so the caller can include AccountID/TeamID
	// in its response payload without a second round-trip.
	AcceptInvitation(ctx context.Context, token string, userID string) (*InvitationInfo, error)
}

// ============================================================
// VALUE OBJECTS
// ============================================================

// TeamInfo represents team information.
type TeamInfo struct {
	ID          string
	AccountID   string
	Name        string
	DisplayName string
	Slug        string
	Type        string // "personal" or "institution"
	IsActive    bool
}

// AccountInfo represents account information.
type AccountInfo struct {
	ID      string
	Name    string
	Email   string
	Phone   string
	Website string
}

// TeamMemberInfo represents team member information.
//
// Roles are not modeled here — they belong to the account, not the team.
type TeamMemberInfo struct {
	ID       string
	TeamID   string
	UserID   string
	IsActive bool
}

// InvitationInfo is the read-only projection of an invitation as the auth
// module needs it. It carries just enough to:
//   - create the user with the right email (Email)
//   - decide whether the invitation is still usable (Status, ExpiresAt)
//   - build a response with the joined account/team (AccountID, TeamID)
//
// It intentionally does NOT expose the inviter, role, or timestamps —
// the auth module has no business reading those.
type InvitationInfo struct {
	Token     string
	Email     string // the invitee's email; must match the user being created
	TeamID    string
	AccountID string
	Role      string // invited account-level role (e.g. "trainer")
	Status    string // "pending", "accepted", "declined"
	ExpiresAt string // ISO-8601, for logging only
}

// WorkspaceContext aggregates primary workspace context for JWT claims.
type WorkspaceContext struct {
	TeamID    string
	AccountID string
	Role      string
	TeamType  string // "personal" or "institution"
}