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
func (s *teamService) AddMember(ctx context.Context, teamID, userID string, role teamdomain.MemberRole, addedBy string) (*teamdomain.Member, error) {
    if teamID == "" {
        return nil, fmt.Errorf("team ID is required")
    }
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }
    if role == "" {
        return nil, fmt.Errorf("role is required")
    }

    // Check if team exists
    team, err := s.repo.GetTeamByID(ctx, teamID)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, teamdomain.ErrTeamNotFound
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
    member, err := teamdomain.NewMember(teamID, userID, role)
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateMember(ctx, member); err != nil {
        return nil, fmt.Errorf("failed to add member: %w", err)
    }

    // Assign Casbin role
    scope := NewTeamScope(team)
    if err := s.casbinSvc.AssignRole(ctx, scope, userID, string(role)); err != nil {
        log.Printf("⚠️ Failed to assign role: %v", err)
    }

    log.Printf("✅ User %s added to team %s as %s", userID, teamID, role)
    return member, nil
}

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

    // Check if last admin
    if member.Role == teamdomain.RoleAccountAdmin {
        admins, _, err := s.repo.GetMembersByTeam(ctx, teamID, teamdomain.ListMembersFilters{
            Role: teamdomain.RoleAccountAdmin,
        })
        if err != nil {
            return fmt.Errorf("failed to check admins: %w", err)
        }
        if len(admins) <= 1 {
            return teamdomain.ErrLastAdminCannotLeave
        }
    }

    // Remove member
    if err := s.repo.DeleteMember(ctx, teamID, userID); err != nil {
        return fmt.Errorf("failed to remove member: %w", err)
    }

    // Remove Casbin role
    scope := NewTeamScope(team)
    if err := s.casbinSvc.RemoveRole(ctx, scope, userID, string(member.Role)); err != nil {
        log.Printf("⚠️ Failed to remove role: %v", err)
    }

    log.Printf("✅ User %s removed from team %s by %s", userID, teamID, removedBy)
    return nil
}

// UpdateMemberRole updates a member's role in a team
func (s *teamService) UpdateMemberRole(ctx context.Context, teamID, userID string, newRole teamdomain.MemberRole, updatedBy string) (*teamdomain.Member, error) {
    if teamID == "" {
        return nil, fmt.Errorf("team ID is required")
    }
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }
    if newRole == "" {
        return nil, fmt.Errorf("new role is required")
    }

    // Check if team exists
    team, err := s.repo.GetTeamByID(ctx, teamID)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, teamdomain.ErrTeamNotFound
    }

    // Get member
    member, err := s.repo.GetMemberByTeamAndUser(ctx, teamID, userID)
    if err != nil {
        return nil, err
    }
    if member == nil {
        return nil, teamdomain.ErrMemberNotFound
    }

    // Don't allow changing own role
    if userID == updatedBy {
        return nil, teamdomain.ErrCannotChangeOwnRole
    }

    // For personal teams, don't allow role changes
    if team.IsPersonal() {
        return nil, fmt.Errorf("cannot change roles in personal team")
    }

    // Update role
    if err := member.UpdateRole(newRole); err != nil {
        return nil, err
    }

    if err := s.repo.UpdateMember(ctx, member); err != nil {
        return nil, fmt.Errorf("failed to update member role: %w", err)
    }

    // Update Casbin role
    scope := NewTeamScope(team)
    if err := s.casbinSvc.RemoveRole(ctx, scope, userID, string(member.Role)); err != nil {
        log.Printf("⚠️ Failed to remove old role: %v", err)
    }
    if err := s.casbinSvc.AssignRole(ctx, scope, userID, string(newRole)); err != nil {
        log.Printf("⚠️ Failed to assign new role: %v", err)
    }

    log.Printf("✅ User %s role updated to %s in team %s", userID, newRole, teamID)
    return member, nil
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

    // Check if last admin
    if member.Role == teamdomain.RoleAccountAdmin {
        admins, _, err := s.repo.GetMembersByTeam(ctx, teamID, teamdomain.ListMembersFilters{
            Role: teamdomain.RoleAccountAdmin,
        })
        if err != nil {
            return fmt.Errorf("failed to check admins: %w", err)
        }
        if len(admins) <= 1 {
            return teamdomain.ErrLastAdminCannotLeave
        }
    }

    // Remove member
    if err := s.repo.DeleteMember(ctx, teamID, userID); err != nil {
        return fmt.Errorf("failed to leave team: %w", err)
    }

    // Remove Casbin role
    scope := NewTeamScope(team)
    if err := s.casbinSvc.RemoveRole(ctx, scope, userID, string(member.Role)); err != nil {
        log.Printf("⚠️ Failed to remove role: %v", err)
    }

    log.Printf("✅ User %s left team %s", userID, teamID)
    return nil
}