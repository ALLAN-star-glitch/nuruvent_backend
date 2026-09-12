// internal/modules/auth/service/team_service.go

package service

import "context"

// TeamService defines the team operations needed by the auth module.
type TeamService interface {
	// Team operations
	GetTeamByID(ctx context.Context, teamID string) (*TeamInfo, error)
	GetPersonalTeamByUserID(ctx context.Context, userID string) (*TeamInfo, error)
	GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*TeamInfo, error)
	GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error)

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

	// Member & Invitation operations
	GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error)
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)

	// Process team invitation during user onboarding.
	AcceptInvitation(ctx context.Context, token string, userID string) (*TeamInfo, error)
}

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

// WorkspaceContext aggregates primary workspace context for JWT claims.
type WorkspaceContext struct {
	TeamID    string
	AccountID string
	Role      string
	TeamType  string // "personal" or "institution"
}