// internal/modules/team/service/team_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// TeamService implements the Service interface.
type teamService struct {
	repo      teamdomain.Repository
	authSvc   AuthService
	casbinSvc CasbinService
	notifSvc  NotificationService
	aiSvc     AIService
}

// NewTeamService creates a new team service.
func NewTeamService(
	repo teamdomain.Repository,
	authSvc AuthService,
	casbinSvc CasbinService,
	notifSvc NotificationService,
	aiSvc AIService,
) Service {
	return &teamService{
		repo:      repo,
		authSvc:   authSvc,
		casbinSvc: casbinSvc,
		notifSvc:  notifSvc,
		aiSvc:     aiSvc,
	}
}

// ============================================================
// ============================================================
// TEAM OPERATIONS
// ============================================================
//
// POST-REVAMP: teams are NOT authorization domains. Every team-scoped
// permission check is performed against the team's parent ACCOUNT domain:
//
//     accountdomain.AccountDomain(team.AccountID)
//
// Team membership is stored as data (team_members). We do not write to
// Casbin for team membership changes.
//
// Creating a team does not touch Casbin. The creator's account role
// already covers the new team.

// CreatePersonalTeam creates a personal team for a user.
//
// No permission check: users implicitly have access to their own personal
// account and its teams. The caller is the user themselves.
//
// Roles are not assigned at team creation. The user's account role
// (assigned during registration) already covers this team. No Casbin
// writes happen here.
func (s *teamService) CreatePersonalTeam(ctx context.Context, userID, userName string) (*teamdomain.Team, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if userName == "" {
		return nil, fmt.Errorf("user name is required")
	}

	log.Printf("[CreatePersonalTeam] Creating personal team for user: %s", userID)

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

	user, err := s.authSvc.GetUserByIDWithAccount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if user.AccountID == "" {
		return nil, fmt.Errorf("user has no account ID")
	}

	team, err := teamdomain.NewPersonalTeamWithAccount(userID, userName, user.AccountID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to create personal team: %w", err)
	}

	member, err := teamdomain.NewMember(team.ID, userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add user to personal team: %w", err)
	}

	// NO Casbin writes. The user's account role already covers this team.

	log.Printf("✅ Personal team created for user: %s (Team ID: %s, AccountID: %s)",
		userID, team.ID, team.AccountID)
	return team, nil
}

// CreateInstitutionTeam creates an institution team and adds the creator
// as its first member.
//
// Authorization: the creator must hold `team:create` in the account.
//
// Roles are not assigned at team creation. The creator's account role
// already covers this team. No Casbin writes happen here.
func (s *teamService) CreateInstitutionTeam(ctx context.Context, accountID, name, displayName, slug, createdBy string) (*teamdomain.Team, error) {
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

	log.Printf("[CreateInstitutionTeam] account=%s createdBy=%s", accountID, createdBy)

	// Permission: createdBy must be allowed to create teams in this account.
	accountDomain := accountdomain.AccountDomain(accountID)
	if accountDomain == "" {
		return nil, fmt.Errorf("invalid account ID: %q", accountID)
	}

	allowed, err := s.casbinSvc.CanCreateTeam(ctx, createdBy, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, teamdomain.ErrPermissionDenied
	}

	// Check slug uniqueness
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

	member, err := teamdomain.NewMember(team.ID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}
	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add user to institution team: %w", err)
	}

	// NO Casbin writes. The creator's account role already covers this team.

	log.Printf("✅ Institution team created: %s (Team ID: %s)", team.Name, team.ID)
	return team, nil
}

// GetTeamByID retrieves a team by ID.
//
// No permission check — callers that need authorization should verify
// against the returned team's parent account. This is a lookup primitive.
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

// GetUserTeams retrieves all teams a user belongs to.
//
// Self-service read: no permission check. Returns teams where the user
// is a member.
func (s *teamService) GetUserTeams(ctx context.Context, userID string) ([]*teamdomain.Team, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.repo.GetTeamsByUserID(ctx, userID)
}

// UpdateTeam updates a team.
//
// Authorization: the updater must hold `team:update` in the team's parent
// account.
//
// NOTE: the previous version read userID from ctx.Value("user_id"). That's
// fragile (contexts should be typed) and no longer supported. Callers must
// pass the actor's user ID explicitly. See the signature change note.
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

	userID := actorFromContext(ctx)
	if userID == "" {
		return nil, teamdomain.ErrPermissionDenied
	}

	accountDomain := accountdomain.AccountDomain(team.AccountID)
	if accountDomain == "" {
		return nil, fmt.Errorf("team %s has no valid account", id)
	}

	allowed, err := s.casbinSvc.CanManageTeam(ctx, userID, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
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

// DeleteTeam deletes a team.
//
// Authorization: the deleter must hold `team:delete` in the team's parent
// account.
//
// No Casbin writes happen here — team membership isn't in Casbin, and no
// per-team policies exist.
func (s *teamService) DeleteTeam(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("team ID is required")
	}

	team, err := s.repo.GetTeamByID(ctx, id)
	if err != nil {
		return err
	}
	if team == nil {
		return teamdomain.ErrTeamNotFound
	}

	userID := actorFromContext(ctx)
	if userID == "" {
		return teamdomain.ErrPermissionDenied
	}

	accountDomain := accountdomain.AccountDomain(team.AccountID)
	if accountDomain == "" {
		return fmt.Errorf("team %s has no valid account", id)
	}

	allowed, err := s.casbinSvc.CanDeleteTeam(ctx, userID, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return teamdomain.ErrPermissionDenied
	}

	if err := s.repo.DeleteTeam(ctx, id); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	log.Printf("✅ Team deleted: %s (ID: %s)", team.Name, team.ID)
	return nil
}

// ============================================================
// TEAM MEMBERSHIP QUERIES (for Auth module and handlers)
// ============================================================

// GetPersonalTeamByUserID gets a user's personal team.
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

// GetInstitutionTeamByInstitutionID gets an institution's team.
//
// Deprecated: this treated institutionID as a teamID, which is no longer
// the model. Use GetTeamByID with an actual team ID instead.
func (s *teamService) GetInstitutionTeamByInstitutionID(ctx context.Context, institutionID string) (*teamdomain.Team, error) {
	if institutionID == "" {
		return nil, fmt.Errorf("institution ID is required")
	}
	return s.repo.GetTeamByID(ctx, institutionID)
}

// GetAccountByTeamID gets the account for a team.
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
		Name: "", // populate from account service if needed
	}, nil
}

// GetTeamDomain returns the authz domain for a team.
//
// POST-REVAMP: teams are not authz domains, so this returns the team's
// PARENT ACCOUNT domain: "account:<team.account_id>".
//
// The name is retained for backward compatibility, but the semantics have
// changed. If you're writing new code, use GetAccountDomainForTeam.
func (s *teamService) GetTeamDomain(ctx context.Context, teamID string) (string, error) {
	return s.GetAccountDomainForTeam(ctx, teamID)
}

// GetAccountDomainForTeam returns the account domain for the team with
// the given ID.
func (s *teamService) GetAccountDomainForTeam(ctx context.Context, teamID string) (string, error) {
	if teamID == "" {
		return "", fmt.Errorf("team ID is required")
	}

	accountID, err := s.repo.AccountIDForTeam(ctx, teamID)
	if err != nil {
		return "", err
	}
	if accountID == "" {
		return "", teamdomain.ErrTeamNotFound
	}

	return accountdomain.AccountDomain(accountID), nil
}

// GetAccountDomain returns the domain for an account.
func (s *teamService) GetAccountDomain(ctx context.Context, accountID string) (string, error) {
	if accountID == "" {
		return "", fmt.Errorf("account ID is required")
	}
	return accountdomain.AccountDomain(accountID), nil
}

// GetUserPersonalTeamDomain returns the account domain for a user's
// personal team.
//
// POST-REVAMP: the "personal team domain" no longer exists as a distinct
// concept. This returns the personal team's parent account domain, which
// for a personal team is the user's personal account.
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

	return accountdomain.AccountDomain(team.AccountID), nil
}

// GetUserInstitutionTeamDomains returns the account domains for every
// institution team a user belongs to.
//
// POST-REVAMP: each entry is the account domain of the team's parent
// account, deduplicated (a user in 3 teams under the same account gets
// one entry, not three).
func (s *teamService) GetUserInstitutionTeamDomains(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	teams, err := s.repo.GetTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var domains []string
	for _, team := range teams {
		if !team.IsInstitution() {
			continue
		}
		domain := accountdomain.AccountDomain(team.AccountID)
		if domain == "" || seen[domain] {
			continue
		}
		seen[domain] = true
		domains = append(domains, domain)
	}
	return domains, nil
}

// ============================================================
// AI CONTEXT HELPERS
// ============================================================

func (s *teamService) getUserRole(ctx context.Context, userID, accountID string) string {
	role, err := s.authSvc.GetUserRoleInAccount(ctx, userID, accountID)
	if err != nil || role == "" {
		return "member"
	}
	return role
}

func getUserDisplayName(user *UserResult) string {
	if user == nil {
		return ""
	}
	if user.DisplayName != "" {
		return user.DisplayName
	}
	return user.Name
}

func (s *teamService) getTeamMemberCount(ctx context.Context, teamID string) int {
	count, err := s.repo.CountMembersByTeam(ctx, teamID)
	if err != nil {
		return 0
	}
	return int(count)
}

func (s *teamService) getTeamEventCount(ctx context.Context, teamID string) int {
	// TODO: Implement when event module is ready
	return 0
}

func (s *teamService) getRecentEventNames(ctx context.Context, teamID string, limit int) []string {
	// TODO: Implement when event module is ready
	return []string{}
}

func (s *teamService) getTeamMemberNames(ctx context.Context, teamID string, limit int) []string {
	members, _, err := s.repo.GetMembersByTeam(ctx, teamID, teamdomain.ListMembersFilters{Limit: limit})
	if err != nil || len(members) == 0 {
		return []string{}
	}

	names := make([]string, 0, len(members))
	for _, m := range members {
		user, err := s.authSvc.GetUserByID(ctx, m.UserID)
		if err == nil && user != nil {
			if user.DisplayName != "" {
				names = append(names, user.DisplayName)
			} else {
				names = append(names, user.Name)
			}
		}
	}
	return names
}