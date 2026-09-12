// internal/modules/team/service/member_service.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// MEMBER OPERATIONS
// ============================================================
//
// POST-REVAMP: team membership is stored ONLY in team_members. Adding or
// removing a member does NOT touch Casbin. Roles are assigned at the
// ACCOUNT level (account_members → Casbin g rules) and apply to every
// team under that account.
//
// PERMISSION MODEL:
//
// Every state-changing operation and every read of another user's data
// performs a permission check against the TEAM'S PARENT ACCOUNT domain:
//
//     accountdomain.AccountDomain(team.AccountID)
//
// The check goes through `casbinSvc.CanX(ctx, userID, accountDomain)`.
// Never pass a team domain — teams are not authz domains.
//
// Self-service reads (the caller's own memberships) do not check
// permissions. They trust the caller to pass their own user ID.

// ============================================================
// WRITE OPERATIONS
// ============================================================

// AddMember adds a user to a team.
//
// Authorization: the adding user must hold `member:create` in the team's
// parent account domain.
//
// Requirements:
//   - The target user must already be a member of the team's parent
//     account ("account membership precedes team membership").
//   - No Casbin writes happen here.
func (s *teamService) AddMember(ctx context.Context, teamID, userID, addedBy string) (*teamdomain.Member, error) {
	if teamID == "" {
		return nil, fmt.Errorf("team ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if addedBy == "" {
		return nil, fmt.Errorf("added by is required")
	}

	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, teamdomain.ErrTeamNotFound
	}

	// Permission: addedBy must be allowed to add members to this account.
	accountDomain := accountdomain.AccountDomain(team.AccountID)
	if accountDomain == "" {
		return nil, fmt.Errorf("team %s has no valid account", teamID)
	}

	allowed, err := s.casbinSvc.CanAddMember(ctx, addedBy, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, teamdomain.ErrPermissionDenied
	}

	user, err := s.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Gate: the target user must already be an account member.
	accountRole, err := s.authSvc.GetUserRoleInAccount(ctx, userID, team.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify account membership for user %s: %w", userID, err)
	}
	if accountRole == "" {
		return nil, fmt.Errorf("user %s is not a member of account %s; add them to the account first",
			userID, team.AccountID)
	}

	existing, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
	if err == nil && existing != nil && existing.IsActive {
		return nil, teamdomain.ErrMemberAlreadyExists
	}

	member, err := teamdomain.NewMember(teamID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// NO Casbin writes. The user's account role already covers this team.
	log.Printf("✅ User %s added to team %s by %s (account role: %s)",
		userID, teamID, addedBy, accountRole)
	return member, nil
}

// RemoveMember removes a member from a team.
//
// Authorization: the remover must hold `member:delete` in the team's
// parent account domain.
//
// No Casbin writes happen here. The removed user's account role is
// unchanged; they simply lose visibility to this specific team.
func (s *teamService) RemoveMember(ctx context.Context, teamID, userID, removedBy string) error {
	if teamID == "" {
		return fmt.Errorf("team ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if removedBy == "" {
		return fmt.Errorf("removed by is required")
	}

	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return teamdomain.ErrTeamNotFound
	}

	// Permission: removedBy must be allowed to remove members in this account.
	accountDomain := accountdomain.AccountDomain(team.AccountID)
	if accountDomain == "" {
		return fmt.Errorf("team %s has no valid account", teamID)
	}

	allowed, err := s.casbinSvc.CanRemoveMember(ctx, removedBy, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return teamdomain.ErrPermissionDenied
	}

	member, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return teamdomain.ErrMemberNotFound
	}

	if userID == removedBy {
		return teamdomain.ErrCannotRemoveSelf
	}

	if team.IsPersonal() {
		return fmt.Errorf("cannot remove members from personal team")
	}

	if err := s.repo.DeleteMember(ctx, teamID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// NO Casbin writes. Account role is unchanged.
	log.Printf("✅ User %s removed from team %s by %s", userID, teamID, removedBy)
	return nil
}

// LeaveTeam allows a user to leave a team.
//
// Authorization: no permission check. The member-existence check below
// ensures the caller can only leave a team they belong to. Removing
// yourself never requires elevated permissions.
//
// No Casbin writes. The user's account role is unchanged.
func (s *teamService) LeaveTeam(ctx context.Context, teamID, userID string) error {
	if teamID == "" {
		return fmt.Errorf("team ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team == nil {
		return teamdomain.ErrTeamNotFound
	}

	if team.IsPersonal() {
		return fmt.Errorf("cannot leave personal team")
	}

	member, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return teamdomain.ErrMemberNotFound
	}

	if err := s.repo.DeleteMember(ctx, teamID, userID); err != nil {
		return fmt.Errorf("failed to leave team: %w", err)
	}

	log.Printf("✅ User %s left team %s", userID, teamID)
	return nil
}

// ============================================================
// READ OPERATIONS
// ============================================================

// GetTeamMembers retrieves all members of a team.
//
// Authorization: the viewer must hold `member:read` in the team's parent
// account domain. Any account member with the read grant on members can
// view the roster.
//
// The viewerUserID parameter is required — it cannot be inferred from the
// team or filters. Callers must pass the authenticated user's ID.
func (s *teamService) GetTeamMembers(
	ctx context.Context,
	viewerUserID, teamID string,
	filters teamdomain.ListMembersFilters,
) ([]*teamdomain.Member, int64, error) {
	if viewerUserID == "" {
		return nil, 0, fmt.Errorf("viewer user ID is required")
	}
	if teamID == "" {
		return nil, 0, fmt.Errorf("team ID is required")
	}

	team, err := s.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, 0, err
	}
	if team == nil {
		return nil, 0, teamdomain.ErrTeamNotFound
	}

	accountDomain := accountdomain.AccountDomain(team.AccountID)
	if accountDomain == "" {
		return nil, 0, fmt.Errorf("team %s has no valid account", teamID)
	}

	allowed, err := s.casbinSvc.CanViewMembers(ctx, viewerUserID, accountDomain)
	if err != nil {
		return nil, 0, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, 0, teamdomain.ErrPermissionDenied
	}

	return s.repo.GetMembersByTeam(ctx, teamID, filters)
}

// GetUserMemberships retrieves all team memberships for a user.
//
// Self-service read: no permission check. The caller must pass their own
// user ID. This method is used for "my teams" endpoints and dashboard
// widgets.
func (s *teamService) GetUserMemberships(ctx context.Context, userID string) ([]*teamdomain.Member, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.repo.GetMembersByUser(ctx, userID)
}

// GetUserTeamMemberships gets all team memberships for a user.
//
// Self-service read: no permission check.
func (s *teamService) GetUserTeamMemberships(ctx context.Context, userID string) ([]*TeamMemberInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	members, err := s.repo.GetMembersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*TeamMemberInfo, len(members))
	for i, member := range members {
		result[i] = &TeamMemberInfo{
			ID:       member.ID,
			TeamID:   member.TeamID,
			UserID:   member.UserID,
			IsActive: member.IsActive,
		}
	}
	return result, nil
}

// GetUserPersonalTeamIDs gets all personal team IDs for a user.
//
// Self-service read: no permission check.
func (s *teamService) GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	members, err := s.repo.GetMembersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var teamIDs []string
	for _, member := range members {
		team, err := s.repo.GetTeamByID(ctx, member.TeamID)
		if err != nil || team == nil {
			continue
		}
		if team.IsPersonal() {
			teamIDs = append(teamIDs, team.ID)
		}
	}
	return teamIDs, nil
}

// GetUserInstitutionTeamIDs gets all institution team IDs for a user.
//
// Self-service read: no permission check.
func (s *teamService) GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	members, err := s.repo.GetMembersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var teamIDs []string
	for _, member := range members {
		team, err := s.repo.GetTeamByID(ctx, member.TeamID)
		if err != nil || team == nil {
			continue
		}
		if team.IsInstitution() {
			teamIDs = append(teamIDs, team.ID)
		}
	}
	return teamIDs, nil
}