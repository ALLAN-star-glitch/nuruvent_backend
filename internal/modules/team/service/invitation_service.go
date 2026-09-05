// internal/modules/team/service/invitation_service.go

package service

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// INVITATION OPERATIONS
// ============================================================

// InviteMember invites a user to join a team
func (s *teamService) InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error) {
    // 1. Validate input
    if cmd.TeamID == "" {
        return nil, fmt.Errorf("team ID is required")
    }
    if cmd.Email == "" {
        return nil, fmt.Errorf("email is required")
    }
    if cmd.Role == "" {
        return nil, fmt.Errorf("role is required")
    }
    if cmd.InvitedBy == "" {
        return nil, fmt.Errorf("invited by is required")
    }

    // 2. Check if team exists
    team, err := s.repo.GetTeamByID(ctx, cmd.TeamID)
    if err != nil {
        return nil, err
    }
    if team == nil {
        return nil, teamdomain.ErrTeamNotFound
    }

    // 3. Check if user exists
    userExists, err := s.authSvc.UserExists(ctx, cmd.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to check user: %w", err)
    }

    // 4. If user exists, check if already a member
    if userExists {
        user, err := s.authSvc.GetUserByEmail(ctx, cmd.Email)
        if err != nil {
            return nil, fmt.Errorf("failed to get user: %w", err)
        }

        existing, err := s.repo.GetMemberByTeamAndUser(ctx, cmd.TeamID, user.ID)
        if err == nil && existing != nil && existing.IsActive {
            return nil, teamdomain.ErrMemberAlreadyExists
        }

        // User exists - add directly
        _, err = s.AddMember(ctx, cmd.TeamID, user.ID, cmd.Role, cmd.InvitedBy)
        if err != nil {
            return nil, err
        }

        // Send notification
        if err := s.notifSvc.SendTeamInvite(ctx, SendTeamInviteRequest{
            To:            cmd.Email,
            UserName:      user.Name,
            InvitedBy:     cmd.InvitedBy,
            TeamName:      team.DisplayName,
            TeamID:        team.ID,
            Role:          string(cmd.Role),
            InviteLink:    "https://nuruvent.com/dashboard/teams",
            ExpiresIn:     "N/A",
        }); err != nil {
            log.Printf("⚠️ Failed to send notification: %v", err)
        }

        return nil, nil // Already added
    }

    // 5. User doesn't exist - create invitation
    token := teamdomain.GenerateToken()
    expiresAt := time.Now().Add(7 * 24 * time.Hour)

    invitation, err := teamdomain.NewInvitation(
        cmd.TeamID,
        cmd.Email,
        cmd.InvitedBy,
        cmd.Role,
        token,
        expiresAt,
    )
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
        return nil, fmt.Errorf("failed to create invitation: %w", err)
    }

    // Send invitation email (NO OTP)
    if err := s.notifSvc.SendTeamInviteRegistration(ctx, SendTeamInviteRegistrationRequest{
        To:              cmd.Email,
        Name:            cmd.Email,
        InvitedBy:       cmd.InvitedBy,
        TeamName:        team.DisplayName,
        Role:            string(cmd.Role),
        RegistrationLink: fmt.Sprintf("https://nuruvent.com/register?token=%s", token),
    }); err != nil {
        log.Printf("⚠️ Failed to send invitation email: %v", err)
    }

    log.Printf("✅ Invitation created for %s to join team %s", cmd.Email, team.ID)
    return invitation, nil
}

// ValidateInvitationToken validates an invitation token
func (s *teamService) ValidateInvitationToken(ctx context.Context, token string) (*teamdomain.Invitation, error) {
    if token == "" {
        return nil, fmt.Errorf("token is required")
    }

    invitation, err := s.repo.GetInvitationByToken(ctx, token)
    if err != nil {
        return nil, err
    }
    if invitation == nil {
        return nil, teamdomain.ErrInvitationNotFound
    }

    if invitation.Status != teamdomain.InvitationStatusPending {
        return nil, teamdomain.ErrInvitationAlreadyAccepted
    }

    if invitation.IsExpired() {
        return nil, teamdomain.ErrInvitationExpired
    }

    return invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the team
func (s *teamService) AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, string, string, error) {
    // 1. Get invitation
    invitation, err := s.repo.GetInvitationByToken(ctx, token)
    if err != nil {
        return nil, "", "", err
    }
    if invitation == nil {
        return nil, "", "", teamdomain.ErrInvitationNotFound
    }

    // 2. Validate
    if invitation.Status != teamdomain.InvitationStatusPending {
        return nil, "", "", teamdomain.ErrInvitationAlreadyAccepted
    }
    if invitation.IsExpired() {
        return nil, "", "", teamdomain.ErrInvitationExpired
    }

    // 3. Get user
    user, err := s.authSvc.GetUserByID(ctx, userID)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to get user: %w", err)
    }
    if user == nil {
        return nil, "", "", fmt.Errorf("user not found")
    }

    // 4. Check if email matches
    if user.Email != invitation.Email {
        return nil, "", "", teamdomain.ErrInvitationEmailMismatch
    }

    // 5. Check if already a member
    existing, err := s.repo.GetMemberByTeamAndUser(ctx, invitation.TeamID, userID)
    if err == nil && existing != nil && existing.IsActive {
        // Already a member, mark invitation as accepted
        if err := invitation.Accept(); err != nil {
            return nil, "", "", err
        }
        if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
            log.Printf("⚠️ Failed to update invitation: %v", err)
        }
        return existing, "", "", nil
    }

    // 6. Add to team
    member, err := s.AddMember(ctx, invitation.TeamID, userID, invitation.Role, invitation.InvitedBy)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to add member: %w", err)
    }

    // 7. Mark invitation as accepted
    if err := invitation.Accept(); err != nil {
        return nil, "", "", err
    }
    if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
        log.Printf("⚠️ Failed to update invitation: %v", err)
    }

    // 8. Get team for notification
    team, err := s.repo.GetTeamByID(ctx, invitation.TeamID)
    if err == nil && team != nil {
        // Send notification to admin
        if err := s.notifSvc.SendTeamInviteAccepted(ctx, SendTeamInviteAcceptedRequest{
            To:        "admin@example.com", // Should get from invited_by
            AdminName: "Admin",
            UserName:  user.Name,
            UserEmail: user.Email,
            TeamName:  team.DisplayName,
        }); err != nil {
            log.Printf("⚠️ Failed to send notification: %v", err)
        }
    }

    log.Printf("✅ User %s accepted invitation to team %s", user.Email, invitation.TeamID)
    return member, user.AccessToken, user.RefreshToken, nil
}



// RegisterAndAcceptInvitation registers a new user and accepts the invitation
func (s *teamService) RegisterAndAcceptInvitation(ctx context.Context, cmd RegisterAndAcceptInvitationCommand) (*teamdomain.Member, string, string, error) {
    // 1. Validate input
    if cmd.Token == "" {
        return nil, "", "", fmt.Errorf("token is required")
    }
    if cmd.Name == "" {
        return nil, "", "", fmt.Errorf("name is required")
    }
    if cmd.Password == "" {
        return nil, "", "", fmt.Errorf("password is required")
    }

    // 2. Get invitation
    invitation, err := s.repo.GetInvitationByToken(ctx, cmd.Token)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to get invitation: %w", err)
    }
    if invitation == nil {
        return nil, "", "", teamdomain.ErrInvitationNotFound
    }

    // 3. Validate invitation status
    if invitation.Status != teamdomain.InvitationStatusPending {
        return nil, "", "", teamdomain.ErrInvitationAlreadyAccepted
    }

    // 4. Check if invitation is expired
    if invitation.IsExpired() {
        return nil, "", "", teamdomain.ErrInvitationExpired
    }

    // 5. Get the team
    team, err := s.repo.GetTeamByID(ctx, invitation.TeamID)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to get team: %w", err)
    }
    if team == nil {
        return nil, "", "", teamdomain.ErrTeamNotFound
    }

    // 6. Check if user already exists
    userExists, err := s.authSvc.UserExists(ctx, invitation.Email)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to check user: %w", err)
    }

    var userID string
    var accessToken string
    var refreshToken string

    if userExists {
        // ✅ User already exists - just accept invitation
        // DO NOT change account_type_id - keep whatever they had
        user, err := s.authSvc.GetUserByEmail(ctx, invitation.Email)
        if err != nil {
            return nil, "", "", fmt.Errorf("failed to get user: %w", err)
        }
        userID = user.ID
        accessToken = user.AccessToken
        refreshToken = user.RefreshToken

        // Accept invitation for existing user
        member, _, _, err := s.AcceptInvitation(ctx, cmd.Token, userID)
        if err != nil {
            return nil, "", "", fmt.Errorf("failed to accept invitation: %w", err)
        }
        return member, accessToken, refreshToken, nil
    }

    // 7. ✅ User doesn't exist - Determine account type based on TEAM TYPE only
    var accountTypeName string
    
    // ✅ If the team is an INSTITUTION team → Institution Account
    if team.IsInstitution() {
        accountTypeName = types.AccountTypeInstitution.GetName() // "account_type_institution"
        log.Printf("📧 User %s joining institution team → Institution Account", invitation.Email)
    } else {
        // ❌ If the team is a PERSONAL team → Personal Account
        accountTypeName = types.AccountTypePersonal.GetName() // "account_type_personal"
        log.Printf("📧 User %s joining personal team → Personal Account", invitation.Email)
    }

    // 8. Create user via Auth Service with the determined account type
    userResult, err := s.authSvc.CreateUser(ctx, CreateUserRequest{
        Email:       invitation.Email,
        Password:    cmd.Password,
        Name:        cmd.Name,
        Phone:       cmd.Phone,
        AccountType: accountTypeName, // "account_type_institution" OR "account_type_personal"
    })
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to create user: %w", err)
    }

    userID = userResult.ID
    accessToken = userResult.AccessToken
    refreshToken = userResult.RefreshToken

    log.Printf("✅ User created with ID: %s, Account Type: %s", userID, accountTypeName)

    // 9. Check if user is already a member (shouldn't happen for new user)
    existing, err := s.repo.GetMemberByTeamAndUser(ctx, invitation.TeamID, userID)
    if err == nil && existing != nil && existing.IsActive {
        // Already a member, just mark invitation as accepted
        if err := invitation.Accept(); err != nil {
            return nil, "", "", err
        }
        if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
            log.Printf("⚠️ Failed to update invitation: %v", err)
        }
        return existing, accessToken, refreshToken, nil
    }

    // 10. Add user to team
    member, err := s.AddMember(ctx, invitation.TeamID, userID, invitation.Role, invitation.InvitedBy)
    if err != nil {
        return nil, "", "", fmt.Errorf("failed to add member to team: %w", err)
    }

    // 11. Mark invitation as accepted
    if err := invitation.Accept(); err != nil {
        return nil, "", "", err
    }
    if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
        log.Printf("⚠️ Failed to update invitation: %v", err)
    }

    // 12. Add Casbin role for the user in the team scope
    scope := NewTeamScope(team)
    if err := s.casbinSvc.AssignRole(ctx, scope, userID, string(invitation.Role)); err != nil {
        log.Printf("⚠️ Failed to assign Casbin role: %v", err)
    }

    // 13. Get inviter info for notification
    inviter, err := s.authSvc.GetUserByID(ctx, invitation.InvitedBy)
    if err != nil {
        log.Printf("⚠️ Failed to get inviter info: %v", err)
    }

    // 14. Send notification to admin
    if err := s.notifSvc.SendTeamInviteAccepted(ctx, SendTeamInviteAcceptedRequest{
        To:         invitation.InvitedBy,
        AdminName: func() string {
            if inviter != nil {
                return inviter.Name
            }
            return "Admin"
        }(),
        UserName:  cmd.Name,
        UserEmail: invitation.Email,
        TeamName:  team.DisplayName,
    }); err != nil {
        log.Printf("⚠️ Failed to send notification: %v", err)
    }

    log.Printf("✅ User %s registered and added to team %s", invitation.Email, team.DisplayName)
    return member, accessToken, refreshToken, nil
}


// DeclineInvitation declines an invitation
func (s *teamService) DeclineInvitation(ctx context.Context, token, userID string) error {
    // 1. Get invitation
    invitation, err := s.repo.GetInvitationByToken(ctx, token)
    if err != nil {
        return err
    }
    if invitation == nil {
        return teamdomain.ErrInvitationNotFound
    }

    // 2. Validate
    if invitation.Status != teamdomain.InvitationStatusPending {
        return teamdomain.ErrInvitationAlreadyAccepted
    }
    if invitation.IsExpired() {
        return teamdomain.ErrInvitationExpired
    }

    // 3. Get user
    user, err := s.authSvc.GetUserByID(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }
    if user == nil {
        return fmt.Errorf("user not found")
    }

    // 4. Check if email matches
    if user.Email != invitation.Email {
        return teamdomain.ErrInvitationEmailMismatch
    }

    // 5. Decline
    if err := invitation.Decline(); err != nil {
        return err
    }
    if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
        return fmt.Errorf("failed to update invitation: %w", err)
    }

    log.Printf("✅ User %s declined invitation to team %s", user.Email, invitation.TeamID)
    return nil
}

// ResendInvitation resends an invitation
func (s *teamService) ResendInvitation(ctx context.Context, invitationID string) (*teamdomain.Invitation, error) {
    // 1. Get invitation
    invitation, err := s.repo.GetInvitationByID(ctx, invitationID)
    if err != nil {
        return nil, err
    }
    if invitation == nil {
        return nil, teamdomain.ErrInvitationNotFound
    }

    // 2. Check if already accepted
    if invitation.Status == teamdomain.InvitationStatusAccepted {
        return nil, fmt.Errorf("invitation already accepted")
    }

    // 3. Generate new token
    newToken := teamdomain.GenerateToken()
    newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

    // 4. Update invitation
    invitation.Token = newToken
    invitation.ExpiresAt = newExpiresAt
    invitation.Status = teamdomain.InvitationStatusPending
    invitation.UpdatedAt = time.Now()

    if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
        return nil, fmt.Errorf("failed to update invitation: %w", err)
    }

    // 5. Resend email
    team, err := s.repo.GetTeamByID(ctx, invitation.TeamID)
    if err != nil {
        return nil, err
    }

    if err := s.notifSvc.SendTeamInviteRegistration(ctx, SendTeamInviteRegistrationRequest{
        To:               invitation.Email,
        Name:             invitation.Email,
        InvitedBy:        invitation.InvitedBy,
        TeamName:         team.DisplayName,
        Role:             string(invitation.Role),
        RegistrationLink: fmt.Sprintf("https://nuruvent.com/register?token=%s", newToken),
    }); err != nil {
        log.Printf("⚠️ Failed to resend invitation email: %v", err)
    }

    log.Printf("✅ Invitation resent for %s", invitation.Email)
    return invitation, nil
}

// GetTeamInvitations retrieves all invitations for a team
func (s *teamService) GetTeamInvitations(ctx context.Context, teamID string, filters teamdomain.ListInvitationsFilters) ([]*teamdomain.Invitation, int64, error) {
    if teamID == "" {
        return nil, 0, fmt.Errorf("team ID is required")
    }

    return s.repo.GetInvitationsByTeam(ctx, teamID, filters)
}