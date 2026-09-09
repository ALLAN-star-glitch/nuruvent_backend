// internal/modules/team/service/invitation_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// ============================================================
// INVITATION OPERATIONS
// ============================================================

func (s *teamService) InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error) {
    // 1. Validate input
    if cmd.TeamID == "" {
        return nil, fmt.Errorf("team ID is required")
    }
    if cmd.Email == "" {
        return nil, fmt.Errorf("email is required")
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

    // 3. Get inviter details
    inviter, err := s.authSvc.GetUserByID(ctx, cmd.InvitedBy)
    inviterName := cmd.InvitedBy // fallback to ID
    if err == nil && inviter != nil && inviter.DisplayName != "" {
        inviterName = inviter.DisplayName
    }
    log.Printf("[InviteMember] Inviter: %s (%s)", inviterName, cmd.InvitedBy)

    // 4. Check if user exists
    userExists, err := s.authSvc.UserExists(ctx, cmd.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to check user: %w", err)
    }

    var user *UserResult
    if userExists {
        user, err = s.authSvc.GetUserByEmail(ctx, cmd.Email)
        if err != nil {
            return nil, fmt.Errorf("failed to get user: %w", err)
        }

        existing, err := s.repo.GetMemberByTeamAndUser(ctx, cmd.TeamID, user.ID)
        if err == nil && existing != nil && existing.IsActive {
            return nil, fmt.Errorf("user is already a member of this team")
        }
    }

    // 5. Check for existing pending invitation
    existingInvitation, err := s.repo.GetInvitationByEmailAndTeam(ctx, cmd.Email, cmd.TeamID)
    if err == nil && existingInvitation != nil && existingInvitation.Status == teamdomain.InvitationStatusPending {
        return nil, fmt.Errorf("an invitation has already been sent to this email")
    }

    // 6. Create invitation for ALL users
    token := teamdomain.GenerateToken()
    expiresAt := time.Now().Add(7 * 24 * time.Hour)

    invitation, err := teamdomain.NewInvitation(
        cmd.TeamID,
        cmd.Email,
        cmd.InvitedBy,
        token,
        expiresAt,
    )
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
        return nil, fmt.Errorf("failed to create invitation: %w", err)
    }

    // 7. Send invitation email based on user type
    if userExists {
        if err := s.notifSvc.SendTeamInviteExistingUser(ctx, SendTeamInviteExistingUserRequest{
            To:         cmd.Email,
            UserName:   user.Name,
            InvitedBy:  inviterName, // ✅ Use display name
            TeamName:   team.DisplayName,
            TeamID:     team.ID,
            AcceptLink: fmt.Sprintf("https://nuruvent.com/invitations/accept?token=%s", token),
            ExpiresIn:  "7 days",
        }); err != nil {
            log.Printf("⚠️ Failed to send invitation email to existing user: %v", err)
        }
    } else {
        if err := s.notifSvc.SendTeamInviteRegistration(ctx, SendTeamInviteRegistrationRequest{
            To:               cmd.Email,
            Name:             "",
            InvitedBy:        inviterName, // ✅ Use display name
            TeamName:         team.DisplayName,
            TeamID:           team.ID,
            RegistrationLink: fmt.Sprintf("https://nuruvent.com/register?token=%s", token),
            ExpiresIn:        "7 days",
        }); err != nil {
            log.Printf("⚠️ Failed to send invitation email to new user: %v", err)
        }
    }

    log.Printf("✅ Invitation created for %s to join team %s (user exists: %v)", cmd.Email, team.ID, userExists)
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
func (s *teamService) AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, error) {
	// 1. Get invitation
	invitation, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, teamdomain.ErrInvitationNotFound
	}

	// 2. Validate state & expiration
	if invitation.Status != teamdomain.InvitationStatusPending {
		return nil, teamdomain.ErrInvitationAlreadyAccepted
	}
	if invitation.IsExpired() {
		return nil, teamdomain.ErrInvitationExpired
	}

	// 3. Get user (read-only)
	user, err := s.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 4. Case-insensitive email comparison
	if !strings.EqualFold(user.Email, invitation.Email) {
		return nil, teamdomain.ErrInvitationEmailMismatch
	}

	// 5. Check if already a member
	existing, err := s.repo.GetMemberByTeamAndUser(ctx, invitation.TeamID, userID)
	if err == nil && existing != nil && existing.IsActive {
		// User is already a member - just update invitation status
		if err := invitation.Accept(); err == nil {
			_ = s.repo.UpdateInvitation(ctx, invitation)
		}
		return existing, nil
	}

	// 6. Execute Membership Addition & Invitation Accept inside an Atomic Transaction
	var member *teamdomain.Member
	err = s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create team membership
		m, err := s.AddMember(txCtx, invitation.TeamID, userID, invitation.InvitedBy)
		if err != nil {
			return fmt.Errorf("failed to add member: %w", err)
		}
		member = m

		// Update invitation state
		if err := invitation.Accept(); err != nil {
			return err
		}
		if err := s.repo.UpdateInvitation(txCtx, invitation); err != nil {
			return fmt.Errorf("failed to update invitation: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 7. Dispatch notification asynchronously
	go func() {
		team, err := s.repo.GetTeamByID(context.Background(), invitation.TeamID)
		if err == nil && team != nil {
			if err := s.notifSvc.SendTeamInviteAccepted(context.Background(), SendTeamInviteAcceptedRequest{
				To:        invitation.InvitedBy,
				AdminName: "Admin",
				UserName:  user.Name,
				UserEmail: user.Email,
				TeamName:  team.DisplayName,
				TeamID:    team.ID,
			}); err != nil {
				log.Printf("⚠️ Failed to send notification: %v", err)
			}
		}
	}()

	log.Printf("✅ User %s accepted invitation to team %s", user.Email, invitation.TeamID)
	return member, nil
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

	// 3. Get inviter name
	inviter, err := s.authSvc.GetUserByID(ctx, invitation.InvitedBy)
	inviterName := invitation.InvitedBy // fallback to ID
	if err == nil && inviter != nil && inviter.DisplayName != "" {
		inviterName = inviter.DisplayName
	}
	log.Printf("[ResendInvitation] Inviter: %s (%s)", inviterName, invitation.InvitedBy)

	// 4. Generate new token
	newToken := teamdomain.GenerateToken()
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	// 5. Update invitation
	invitation.Token = newToken
	invitation.ExpiresAt = newExpiresAt
	invitation.Status = teamdomain.InvitationStatusPending
	invitation.UpdatedAt = time.Now()

	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// 6. Resend email
	team, err := s.repo.GetTeamByID(ctx, invitation.TeamID)
	if err != nil {
		return nil, err
	}

	// Check if user exists to send the right email
	userExists, err := s.authSvc.UserExists(ctx, invitation.Email)
	if err != nil {
		log.Printf("⚠️ Failed to check if user exists: %v", err)
	}

	if userExists {
		user, err := s.authSvc.GetUserByEmail(ctx, invitation.Email)
		if err == nil && user != nil {
			if err := s.notifSvc.SendTeamInviteExistingUser(ctx, SendTeamInviteExistingUserRequest{
				To:         invitation.Email,
				UserName:   user.Name,
				InvitedBy:  inviterName, // ✅ Use display name
				TeamName:   team.DisplayName,
				TeamID:     team.ID,
				AcceptLink: fmt.Sprintf("https://nuruvent.com/invitations/accept?token=%s", newToken),
				ExpiresIn:  "7 days",
			}); err != nil {
				log.Printf("⚠️ Failed to resend invitation to existing user: %v", err)
			}
		}
	} else {
		if err := s.notifSvc.SendTeamInviteRegistration(ctx, SendTeamInviteRegistrationRequest{
			To:               invitation.Email,
			Name:             "",
			InvitedBy:        inviterName, // ✅ Use display name
			TeamName:         team.DisplayName,
			TeamID:           team.ID,
			RegistrationLink: fmt.Sprintf("https://nuruvent.com/register?token=%s", newToken),
			ExpiresIn:        "7 days",
		}); err != nil {
			log.Printf("⚠️ Failed to resend invitation to new user: %v", err)
		}
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