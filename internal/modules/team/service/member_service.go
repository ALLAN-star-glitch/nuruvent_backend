// internal/modules/team/service/member_service.go

package service

import (
    "context"
    "fmt"
    "log"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// MEMBER OPERATIONS
// ============================================================

// AddMember adds a member to a team
// ✅ REMOVED: role parameter - roles are inherited from account
func (s *teamService) AddMember(ctx context.Context, teamID, userID string, addedBy string) (*teamdomain.Member, error) {
    if teamID == "" {
        return nil, fmt.Errorf("team ID is required")
    }
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }
    // ❌ REMOVED: Role validation - roles are inherited from account

    // Check if team exists
    team, err := s.repo.GetTeamByID(ctx, teamID)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, teamdomain.ErrTeamNotFound
    }

    // ✅ Check permission using domain
    domain := NewTeamDomain(team).String()
    isAdmin, err := s.casbinSvc.IsAccountAdmin(ctx, domain, addedBy)
    if err != nil {
        return nil, fmt.Errorf("permission check failed: %w", err)
    }
    if !isAdmin {
        return nil, teamdomain.ErrPermissionDenied
    }

    // Check if user exists
    user, err := s.authSvc.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    if user == nil {
        return nil, fmt.Errorf("user not found")
    }

    // Check if already a member
    existing, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
    if err == nil && existing != nil && existing.IsActive {
        return nil, teamdomain.ErrMemberAlreadyExists
    }

    // Create member
    // ✅ REMOVED: role parameter - roles are inherited from account
    member, err := teamdomain.NewMember(teamID, userID)
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateMember(ctx, member); err != nil {
        return nil, fmt.Errorf("failed to add member: %w", err)
    }

    // ❌ REMOVED: Casbin role assignment - roles are inherited from account
    // Roles will be assigned based on account membership, not team membership

    log.Printf("✅ User %s added to team %s", userID, teamID)
    return member, nil
}

// ❌ REMOVED: UpdateMemberRole - roles cannot be updated at team level
// Roles are managed at the account level through account_members table

// RemoveMember removes a member from a team
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

    // Check if team exists
    team, err := s.repo.GetTeamByID(ctx, teamID)
    if err != nil {
        return err
    }
    if team == nil {
        return teamdomain.ErrTeamNotFound
    }

    // ✅ Check permission using domain
    domain := NewTeamDomain(team).String()
    isAdmin, err := s.casbinSvc.IsAccountAdmin(ctx, domain, removedBy)
    if err != nil {
        return fmt.Errorf("permission check failed: %w", err)
    }
    if !isAdmin {
        return teamdomain.ErrPermissionDenied
    }

    // Get member
    member, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
    if err != nil {
        return err
    }
    if member == nil {
        return teamdomain.ErrMemberNotFound
    }

    // Don't allow removing self
    if userID == removedBy {
        return teamdomain.ErrCannotRemoveSelf
    }

    // For personal teams, don't allow removing
    if team.IsPersonal() {
        return fmt.Errorf("cannot remove members from personal team")
    }

    // ❌ REMOVED: Last admin check - roles are at account level now
    // The concept of "admin" is now at the account level, not team level

    // Remove member
    if err := s.repo.DeleteMember(ctx, teamID, userID); err != nil {
        return fmt.Errorf("failed to remove member: %w", err)
    }

    // ❌ REMOVED: Casbin role removal - roles are inherited from account

    log.Printf("✅ User %s removed from team %s by %s", userID, teamID, removedBy)
    return nil
}

// GetTeamMembers retrieves all members of a team
func (s *teamService) GetTeamMembers(ctx context.Context, teamID string, filters teamdomain.ListMembersFilters) ([]*teamdomain.Member, int64, error) {
    if teamID == "" {
        return nil, 0, fmt.Errorf("team ID is required")
    }

    return s.repo.GetMembersByTeam(ctx, teamID, filters)
}

// GetUserMemberships retrieves all team memberships for a user
func (s *teamService) GetUserMemberships(ctx context.Context, userID string) ([]*teamdomain.Member, error) {
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }

    return s.repo.GetMembersByUser(ctx, userID)
}

// GetUserTeamMemberships gets all team memberships for a user (for Auth module)
// ✅ REMOVED: Role from TeamMemberInfo - roles are inherited from account
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
            // ❌ REMOVED: Role     string(member.Role)
        }
    }
    return result, nil
}

// GetUserPersonalTeamIDs gets all personal team IDs for a user (for Auth module)
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

// GetUserInstitutionTeamIDs gets all institution team IDs for a user (for Auth module)
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

// LeaveTeam allows a user to leave a team
func (s *teamService) LeaveTeam(ctx context.Context, teamID, userID string) error {
    if teamID == "" {
        return fmt.Errorf("team ID is required")
    }
    if userID == "" {
        return fmt.Errorf("user ID is required")
    }

    // Check if team exists
    team, err := s.repo.GetTeamByID(ctx, teamID)
    if err != nil {
        return err
    }
    if team == nil {
        return teamdomain.ErrTeamNotFound
    }

    // Don't allow leaving personal team
    if team.IsPersonal() {
        return fmt.Errorf("cannot leave personal team")
    }

    // Get member
    member, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
    if err != nil {
        return err
    }
    if member == nil {
        return teamdomain.ErrMemberNotFound
    }

    // ❌ REMOVED: Last admin check - roles are at account level now
    // The concept of "admin" is now at the account level, not team level

    // Remove member
    if err := s.repo.DeleteMember(ctx, teamID, userID); err != nil {
        return fmt.Errorf("failed to leave team: %w", err)
    }

    // ❌ REMOVED: Casbin role removal - roles are inherited from account

    log.Printf("✅ User %s left team %s", userID, teamID)
    return nil
}