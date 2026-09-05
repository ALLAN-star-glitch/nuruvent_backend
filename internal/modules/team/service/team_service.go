// internal/modules/team/service/team_service.go

package service

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// TeamService implements the Service interface
type teamService struct {
    repo        teamdomain.Repository
    authSvc     AuthService
    casbinSvc   CasbinService
    notifSvc    NotificationService
}

// NewTeamService creates a new team service
func NewTeamService(
    repo teamdomain.Repository,
    authSvc AuthService,
    casbinSvc CasbinService,
    notifSvc NotificationService,
) Service {
    return &teamService{
        repo:      repo,
        authSvc:   authSvc,
        casbinSvc: casbinSvc,
        notifSvc:  notifSvc,
    }
}

// ============================================================
// TEAM OPERATIONS
// ============================================================

// CreatePersonalTeam creates a personal team for a user
func (s *teamService) CreatePersonalTeam(ctx context.Context, userID, userName string) (*teamdomain.Team, error) {
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }
    if userName == "" {
        return nil, fmt.Errorf("user name is required")
    }

    // Check if personal team already exists
    teams, err := s.repo.GetTeamsByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to check existing teams: %w", err)
    }
    for _, team := range teams {
        if team.IsPersonal() {
            return nil, teamdomain.ErrTeamAlreadyExists
        }
    }

    // Create personal team
    team, err := teamdomain.NewPersonalTeam(userID, userName)
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateTeam(ctx, team); err != nil {
        return nil, fmt.Errorf("failed to create personal team: %w", err)
    }

    // Add user as admin of their personal team
    member, err := teamdomain.NewMember(team.ID, userID, teamdomain.RoleAccountAdmin)
    if err != nil {
        return nil, err
    }
    if err := s.repo.CreateMember(ctx, member); err != nil {
        return nil, fmt.Errorf("failed to add user to personal team: %w", err)
    }

    // Add Casbin policies for personal team
    scope := NewPersonalTeamScope(team.ID)
    if err := s.casbinSvc.AddTeamPolicies(ctx, scope); err != nil {
        log.Printf("⚠️ Failed to add personal team policies: %v", err)
    }

    // Assign admin role
    if err := s.casbinSvc.AssignRole(ctx, scope, userID, string(teamdomain.RoleAccountAdmin)); err != nil {
        log.Printf("⚠️ Failed to assign admin role: %v", err)
    }

    log.Printf("✅ Personal team created for user: %s", userID)
    return team, nil
}

// CreateInstitutionTeam creates an institution team
func (s *teamService) CreateInstitutionTeam(ctx context.Context, name, displayName, slug string) (*teamdomain.Team, error) {
    if name == "" {
        return nil, teamdomain.ErrTeamNameRequired
    }
    if slug == "" {
        return nil, teamdomain.ErrTeamSlugRequired
    }

    // Check if team with slug already exists
    existing, err := s.repo.GetTeamBySlug(ctx, slug)
    if err == nil && existing != nil {
        return nil, teamdomain.ErrTeamAlreadyExists
    }

    team, err := teamdomain.NewInstitutionTeam(name, displayName, slug)
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateTeam(ctx, team); err != nil {
        return nil, fmt.Errorf("failed to create institution team: %w", err)
    }

    log.Printf("✅ Institution team created: %s", team.ID)
    return team, nil
}

// GetTeamByID retrieves a team by ID
func (s *teamService) GetTeamByID(ctx context.Context, id string) (*teamdomain.Team, error) {
    if id == "" {
        return nil, fmt.Errorf("team ID is required")
    }

    team, err := s.repo.GetTeamByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, teamdomain.ErrTeamNotFound
    }

    return team, nil
}

// GetUserTeams retrieves all teams a user belongs to
func (s *teamService) GetUserTeams(ctx context.Context, userID string) ([]*teamdomain.Team, error) {
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }

    return s.repo.GetTeamsByUserID(ctx, userID)
}

// UpdateTeam updates a team
func (s *teamService) UpdateTeam(ctx context.Context, id string, updates map[string]interface{}) (*teamdomain.Team, error) {
    if id == "" {
        return nil, fmt.Errorf("team ID is required")
    }

    team, err := s.repo.GetTeamByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, teamdomain.ErrTeamNotFound
    }

    // Apply updates
    if name, ok := updates["name"].(string); ok && name != "" {
        team.Name = name
    }
    if displayName, ok := updates["display_name"].(string); ok {
        team.DisplayName = displayName
    }
    if slug, ok := updates["slug"].(string); ok && slug != "" {
        team.Slug = slug
    }
    team.UpdatedAt = time.Now()

    if err := s.repo.UpdateTeam(ctx, team); err != nil {
        return nil, fmt.Errorf("failed to update team: %w", err)
    }

    return team, nil
}

// DeleteTeam deletes a team
func (s *teamService) DeleteTeam(ctx context.Context, id string) error {
    if id == "" {
        return fmt.Errorf("team ID is required")
    }

    // Get team
    team, err := s.repo.GetTeamByID(ctx, id)
    if err != nil {
        return err
    }
    if team == nil {
        return teamdomain.ErrTeamNotFound
    }

    // Delete team
    if err := s.repo.DeleteTeam(ctx, id); err != nil {
        return fmt.Errorf("failed to delete team: %w", err)
    }

    // Remove Casbin policies
    scope := NewTeamScope(team)
    if err := s.casbinSvc.RemoveTeamPolicies(ctx, scope); err != nil {
        log.Printf("⚠️ Failed to remove team policies: %v", err)
    }

    log.Printf("✅ Team deleted: %s", id)
    return nil
}