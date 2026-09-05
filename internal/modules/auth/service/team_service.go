// internal/modules/auth/service/team_service.go

package service

import "context"

// TeamService defines the team operations needed by the auth module
type TeamService interface {
	// Team operations
	GetTeamByID(ctx context.Context, teamID string) (*TeamInfo, error)
	GetPersonalTeamByUserID(ctx context.Context, userID string) (*TeamInfo, error)
	GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*TeamInfo, error)
	GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error)

	// Team member operations
	GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error) // ✅ Added
	GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error)
	GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error)
}

// TeamInfo represents team information
type TeamInfo struct {
	ID          string
	AccountID   string
	Name        string
	DisplayName string
	Slug        string
	Type        string // "personal" or "institution"
	IsActive    bool
}

// AccountInfo represents account information
type AccountInfo struct {
	ID      string
	Name    string
	Email   string
	Phone   string
	Website string
}

// TeamMemberInfo represents team member information
type TeamMemberInfo struct {
	ID       string
	TeamID   string
	UserID   string
	Role     string
	IsActive bool
}