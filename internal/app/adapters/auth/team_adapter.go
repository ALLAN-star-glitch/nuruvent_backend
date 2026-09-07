// internal/app/adapters/auth/team_adapter.go

package auth

import (
	"context"

	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// TeamAdapter adapts team service to auth module's TeamService interface
type TeamAdapter struct {
	teamSvc teamService.Service
}

// NewTeamAdapter creates a new team adapter for auth module
func NewTeamAdapter(teamSvc teamService.Service) authService.TeamService {
	return &TeamAdapter{
		teamSvc: teamSvc,
	}
}

// ============================================================
// TEAM OPERATIONS
// ============================================================

// GetTeamByID retrieves a team by ID
func (a *TeamAdapter) GetTeamByID(ctx context.Context, teamID string) (*authService.TeamInfo, error) {
	team, err := a.teamSvc.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	return &authService.TeamInfo{
		ID:          team.ID,
		AccountID:   team.AccountID,
		Name:        team.Name,
		DisplayName: team.DisplayName,
		Slug:        team.Slug,
		Type:        string(team.Type),
		IsActive:    team.IsActive,
	}, nil
}

// GetPersonalTeamByUserID gets a user's personal team
func (a *TeamAdapter) GetPersonalTeamByUserID(ctx context.Context, userID string) (*authService.TeamInfo, error) {
	team, err := a.teamSvc.GetPersonalTeamByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	return &authService.TeamInfo{
		ID:          team.ID,
		AccountID:   team.AccountID,
		Name:        team.Name,
		DisplayName: team.DisplayName,
		Slug:        team.Slug,
		Type:        string(team.Type),
		IsActive:    team.IsActive,
	}, nil
}

// GetInstitutionTeamByInstitutionID gets an institution's team
func (a *TeamAdapter) GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*authService.TeamInfo, error) {
	team, err := a.teamSvc.GetInstitutionTeamByInstitutionID(ctx, institutionID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	return &authService.TeamInfo{
		ID:          team.ID,
		AccountID:   team.AccountID,
		Name:        team.Name,
		DisplayName: team.DisplayName,
		Slug:        team.Slug,
		Type:        string(team.Type),
		IsActive:    team.IsActive,
	}, nil
}

// GetAccountByTeamID gets the account for a team
func (a *TeamAdapter) GetAccountByTeamID(ctx context.Context, teamID string) (*authService.AccountInfo, error) {
	account, err := a.teamSvc.GetAccountByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}

	return &authService.AccountInfo{
		ID:      account.ID,
		Name:    account.Name,
		Email:   account.Email,
		Phone:   account.Phone,
		Website: account.Website,
	}, nil
}

// CreatePersonalTeam creates a personal team for a user
func (a *TeamAdapter) CreatePersonalTeam(ctx context.Context, userID string, userName string) (*authService.TeamInfo, error) {
	team, err := a.teamSvc.CreatePersonalTeam(ctx, userID, userName)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	return &authService.TeamInfo{
		ID:          team.ID,
		AccountID:   team.AccountID,
		Name:        team.Name,
		DisplayName: team.DisplayName,
		Slug:        team.Slug,
		Type:        string(team.Type),
		IsActive:    team.IsActive,
	}, nil
}

// CreateInstitutionTeam creates an institution team for an account
func (a *TeamAdapter) CreateInstitutionTeam(ctx context.Context, accountID string, name string, displayName string, slug string) (*authService.TeamInfo, error) {
	team, err := a.teamSvc.CreateInstitutionTeam(ctx, accountID, name, displayName, slug)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	return &authService.TeamInfo{
		ID:          team.ID,
		AccountID:   team.AccountID,
		Name:        team.Name,
		DisplayName: team.DisplayName,
		Slug:        team.Slug,
		Type:        string(team.Type),
		IsActive:    team.IsActive,
	}, nil
}

// AcceptInvitation accepts an invitation and adds the user to the team
func (a *TeamAdapter) AcceptInvitation(ctx context.Context, token string, userID string) (*authService.TeamInfo, error) {
	member, err := a.teamSvc.AcceptInvitation(ctx, token, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}

	// Get the team info for the member
	team, err := a.teamSvc.GetTeamByID(ctx, member.TeamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	return &authService.TeamInfo{
		ID:          team.ID,
		AccountID:   team.AccountID,
		Name:        team.Name,
		DisplayName: team.DisplayName,
		Slug:        team.Slug,
		Type:        string(team.Type),
		IsActive:    team.IsActive,
	}, nil
}

// ============================================================
// TEAM MEMBER OPERATIONS
// ============================================================

// GetUserTeamMemberships gets all team memberships for a user
func (a *TeamAdapter) GetUserTeamMemberships(ctx context.Context, userID string) ([]*authService.TeamMemberInfo, error) {
	members, err := a.teamSvc.GetUserTeamMemberships(ctx, userID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		return []*authService.TeamMemberInfo{}, nil
	}

	result := make([]*authService.TeamMemberInfo, len(members))
	for i, m := range members {
		result[i] = &authService.TeamMemberInfo{
			ID:       m.ID,
			TeamID:   m.TeamID,
			UserID:   m.UserID,
			Role:     string(m.Role),
			IsActive: m.IsActive,
		}
	}
	return result, nil
}

// GetUserPersonalTeamIDs gets all personal team IDs for a user
func (a *TeamAdapter) GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error) {
	return a.teamSvc.GetUserPersonalTeamIDs(ctx, userID)
}

// GetUserInstitutionTeamIDs gets all institution team IDs for a user
func (a *TeamAdapter) GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error) {
	return a.teamSvc.GetUserInstitutionTeamIDs(ctx, userID)
}