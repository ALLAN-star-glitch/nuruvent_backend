// internal/app/adapters/auth/team_adapter.go

package auth

import (
	"context"
	"errors"
	"time"

	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	authDomain  "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	teamDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
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

// CreateInstitutionTeam creates an institution team with the creator as member
func (a *TeamAdapter) CreateInstitutionTeam(ctx context.Context, accountID string, name string, displayName string, slug string, createdBy string) (*authService.TeamInfo, error) {
	team, err := a.teamSvc.CreateInstitutionTeam(ctx, accountID, name, displayName, slug, createdBy)
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
// INVITATION OPERATIONS
// ============================================================

// ValidateInvitationToken returns the invitation projection for a token
// if and only if it is usable (exists, pending, not expired).
//
// Used by RegisterWithInvitation to:
//   - read the invitee's email (the user is created with it)
//   - reject expired/already-accepted tokens before creating a user
//   - populate the response payload with AccountID and TeamID
func (a *TeamAdapter) ValidateInvitationToken(ctx context.Context, token string) (*authService.InvitationInfo, error) {
	inv, err := a.teamSvc.ValidateInvitationToken(ctx, token)
	if err != nil {
		return nil, mapInvitationError(err)
	}
	if inv == nil {
		return nil, authDomain.ErrInvitationNotFound
	}

	return a.toInvitationInfo(ctx, inv)
}


// AcceptInvitation accepts an invitation and adds the user to the team.
//
// Returns the invitation projection so the caller can include the joined
// AccountID/TeamID in its response without a second round-trip.
func (a *TeamAdapter) AcceptInvitation(ctx context.Context, token string, userID string) (*authService.InvitationInfo, error) {
	// Validate first — we need the invitation details regardless, and it
	// lets us distinguish "not found" from "failed to accept" errors.
	inv, err := a.teamSvc.ValidateInvitationToken(ctx, token)
	if err != nil {
		return nil, mapInvitationError(err)
	}

	if _, err := a.teamSvc.AcceptInvitation(ctx, token, userID); err != nil {
		return nil, mapInvitationError(err)
	}

	info, err := a.toInvitationInfo(ctx, inv)
	if err != nil {
		return nil, err
	}
	// The invitation is now accepted — reflect that in the projection.
	info.Status = "accepted"
	return info, nil
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

// ============================================================
// PRIVATE HELPERS
// ============================================================

// toInvitationInfo projects a team-domain invitation into the auth module's
// InvitationInfo. It resolves AccountID via a team lookup, since AccountID
// lives on the team row, not on the invitation.
func (a *TeamAdapter) toInvitationInfo(ctx context.Context, inv *teamDomain.Invitation) (*authService.InvitationInfo, error) {
	// AccountID is not on the invitation domain type — resolve it via the team.
	team, err := a.teamSvc.GetTeamByID(ctx, inv.TeamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, authDomain.ErrInvitationNotFound
	}

	expiresAt := ""
	if !inv.ExpiresAt.IsZero() {
		expiresAt = inv.ExpiresAt.Format(time.RFC3339)
	}

	return &authService.InvitationInfo{
		Token:     inv.Token,
		Email:     inv.Email,
		TeamID:    inv.TeamID,
		AccountID: team.AccountID,
		Role:      inv.Role,
		Status:    string(inv.Status),
		ExpiresAt: expiresAt,
	}, nil
}

// mapInvitationError translates team-domain errors into auth-domain
// sentinels so the auth delivery layer doesn't need to import the team
// module's error package.
func mapInvitationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, teamDomain.ErrInvitationNotFound):
		return authDomain.ErrInvitationNotFound
	case errors.Is(err, teamDomain.ErrInvitationExpired):
		return authDomain.ErrInvitationExpired
	case errors.Is(err, teamDomain.ErrInvitationAlreadyAccepted):
		return authDomain.ErrInvitationAlreadyUsed
	default:
		return err
	}
}