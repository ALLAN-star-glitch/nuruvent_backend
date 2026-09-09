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
// ✅ Updated: Accepts role parameter for Casbin assignment
func (s *teamService) CreatePersonalTeam(ctx context.Context, userID, userName, role string) (*teamdomain.Team, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if userName == "" {
		return nil, fmt.Errorf("user name is required")
	}

	log.Printf("[CreatePersonalTeam] Creating personal team for user: %s with role: %s", userID, role)

	// Check if personal team already exists
	teams, err := s.repo.GetTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing teams: %w", err)
	}
	for _, team := range teams {
		if team.IsPersonal() {
			log.Printf("[CreatePersonalTeam] Personal team already exists for user: %s", userID)
			return nil, teamdomain.ErrTeamAlreadyExists
		}
	}

	// Get user with account ID
	user, err := s.authSvc.GetUserByIDWithAccount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	log.Printf("[CreatePersonalTeam] User: %s, AccountID: %s", user.ID, user.AccountID)

	if user.AccountID == "" {
		return nil, fmt.Errorf("user has no account ID")
	}

	// Create personal team with the user's AccountID
	team, err := teamdomain.NewPersonalTeamWithAccount(userID, userName, user.AccountID)
	if err != nil {
		return nil, err
	}

	log.Printf("[CreatePersonalTeam] Team created with AccountID: %s, Name: %s", team.AccountID, team.Name)

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to create personal team: %w", err)
	}

	// Add user as member of their personal team
	member, err := teamdomain.NewMember(team.ID, userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add user to personal team: %w", err)
	}

	// Add Casbin policies for personal team using domain
	domain := NewTeamDomain(team).String()
	if err := s.casbinSvc.AddTeamPolicies(ctx, domain); err != nil {
		log.Printf("⚠️ Failed to add personal team policies: %v", err)
	}

	// ✅ Assign the user to the team domain with their account role
	if role != "" {
		if err := s.casbinSvc.AssignRole(ctx, domain, userID, role); err != nil {
			log.Printf("⚠️ Failed to assign user to team domain: %v", err)
		} else {
			log.Printf("✅ User %s assigned role %s in domain %s", userID, role, domain)
		}
	} else {
		log.Printf("⚠️ No role provided for user %s in personal team", userID)
	}

	log.Printf("✅ Personal team created for user: %s (Team ID: %s, AccountID: %s)", userID, team.ID, team.AccountID)
	return team, nil
}

// CreateInstitutionTeam creates an institution team and adds the creator as member
// ✅ Updated: Accepts role parameter for Casbin assignment
func (s *teamService) CreateInstitutionTeam(ctx context.Context, accountID, name, displayName, slug, createdBy, role string) (*teamdomain.Team, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if name == "" {
		return nil, teamdomain.ErrTeamNameRequired
	}
	if slug == "" {
		return nil, teamdomain.ErrTeamSlugRequired
	}
	if createdBy == "" {
		return nil, fmt.Errorf("created by is required")
	}

	log.Printf("[CreateInstitutionTeam] Creating institution team for account: %s, created by: %s, role: %s", accountID, createdBy, role)

	// Check if team with slug already exists
	existing, err := s.repo.GetTeamBySlug(ctx, slug)
	if err == nil && existing != nil {
		return nil, teamdomain.ErrTeamAlreadyExists
	}

	team, err := teamdomain.NewInstitutionTeam(accountID, name, displayName, slug)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to create institution team: %w", err)
	}

	// Add the creator as member of the institution team
	log.Printf("[CreateInstitutionTeam] Adding user %s as member to institution team", createdBy)

	member, err := teamdomain.NewMember(team.ID, createdBy)
	if err != nil {
		log.Printf("[CreateInstitutionTeam] Failed to create member: %v", err)
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	if err := s.repo.CreateMember(ctx, member); err != nil {
		log.Printf("[CreateInstitutionTeam] Failed to add user to institution team: %v", err)
		return nil, fmt.Errorf("failed to add user to institution team: %w", err)
	}
	log.Printf("[CreateInstitutionTeam] ✅ User %s added as member to institution team", createdBy)

	// Add Casbin policies for institution team using domain
	domain := NewTeamDomain(team).String()
	if err := s.casbinSvc.AddTeamPolicies(ctx, domain); err != nil {
		log.Printf("⚠️ Failed to add institution team policies: %v", err)
	}

	// ✅ Assign the user to the team domain with their account role
	if role != "" {
		if err := s.casbinSvc.AssignRole(ctx, domain, createdBy, role); err != nil {
			log.Printf("⚠️ Failed to assign user to team domain: %v", err)
		} else {
			log.Printf("✅ User %s assigned role %s in domain %s", createdBy, role, domain)
		}
	} else {
		log.Printf("⚠️ No role provided for user %s in institution team", createdBy)
	}

	log.Printf("✅ Institution team created: %s (Team ID: %s)", team.Name, team.ID)
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

	// Check permission using domain
	domain := NewTeamDomain(team).String()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, teamdomain.ErrPermissionDenied
	}

	isAdmin, err := s.casbinSvc.IsAccountAdmin(ctx, domain, userID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !isAdmin {
		return nil, teamdomain.ErrPermissionDenied
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

	log.Printf("✅ Team updated: %s (ID: %s)", team.Name, team.ID)
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

	// Check permission using domain
	domain := NewTeamDomain(team).String()
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return teamdomain.ErrPermissionDenied
	}

	isAdmin, err := s.casbinSvc.IsAccountAdmin(ctx, domain, userID)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !isAdmin {
		return teamdomain.ErrPermissionDenied
	}

	// Delete team (soft delete)
	if err := s.repo.DeleteTeam(ctx, id); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	// Remove Casbin policies using domain
	if err := s.casbinSvc.RemoveTeamPolicies(ctx, domain); err != nil {
		log.Printf("⚠️ Failed to remove team policies: %v", err)
	}

	log.Printf("✅ Team deleted: %s (ID: %s)", team.Name, team.ID)
	return nil
}

// ============================================================
// TEAM MEMBERSHIP QUERIES (For Auth Module)
// ============================================================

// GetPersonalTeamByUserID gets a user's personal team
func (s *teamService) GetPersonalTeamByUserID(ctx context.Context, userID string) (*teamdomain.Team, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	teams, err := s.repo.GetTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		if team.IsPersonal() {
			return team, nil
		}
	}

	return nil, nil
}

// GetInstitutionTeamByInstitutionID gets an institution's team
func (s *teamService) GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*teamdomain.Team, error) {
	if institutionID == "" {
		return nil, fmt.Errorf("institution ID is required")
	}

	// Note: In the new design, institutionID is the account ID
	return s.repo.GetTeamByID(ctx, institutionID)
}

// GetAccountByTeamID gets the account for a team
func (s *teamService) GetAccountByTeamID(ctx context.Context, teamID string) (*AccountInfo, error) {
	if teamID == "" {
		return nil, fmt.Errorf("team ID is required")
	}

	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, teamdomain.ErrTeamNotFound
	}

	return &AccountInfo{
		ID:   team.AccountID,
		Name: "", // Would need to fetch from account service
	}, nil
}

// GetTeamDomain returns the domain string for a team
func (s *teamService) GetTeamDomain(ctx context.Context, teamID string) (string, error) {
	if teamID == "" {
		return "", fmt.Errorf("team ID is required")
	}

	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", teamdomain.ErrTeamNotFound
	}

	return NewTeamDomain(team).String(), nil
}

// GetAccountDomain returns the domain string for an account
func (s *teamService) GetAccountDomain(ctx context.Context, accountID string) (string, error) {
	if accountID == "" {
		return "", fmt.Errorf("account ID is required")
	}

	return "account:" + accountID, nil
}

// GetUserPersonalTeamDomain returns the personal team domain for a user
func (s *teamService) GetUserPersonalTeamDomain(ctx context.Context, userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("user ID is required")
	}

	team, err := s.GetPersonalTeamByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", nil
	}

	return NewTeamDomain(team).String(), nil
}

// GetUserInstitutionTeamDomains returns all institution team domains for a user
func (s *teamService) GetUserInstitutionTeamDomains(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	teams, err := s.repo.GetTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var domains []string
	for _, team := range teams {
		if team.IsInstitution() {
			domains = append(domains, NewTeamDomain(team).String())
		}
	}

	return domains, nil
}