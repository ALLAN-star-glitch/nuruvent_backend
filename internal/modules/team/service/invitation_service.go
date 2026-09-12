// internal/modules/team/service/member_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	accountdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
	types "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types/emails"
)

// ============================================================
// INVITATION OPERATIONS
// ============================================================

// InviteMember creates a pending invitation for the given email to join
// the given team with the given account role.
//
// Authorization:
//   - The inviter must hold `member:invite` in the team's parent account.
//   - To invite as `account_admin`, the inviter must additionally have
//     `member:manage` (i.e. be an account_admin).
//
// The Role is applied at accept time:
//   - If the invitee is not yet an account member, they are added with
//     this role.
//   - If the invitee is already an account member, this role is ignored;
//     their existing role wins.
func (s *teamService) InviteMember(ctx context.Context, cmd InviteMemberCommand) (*teamdomain.Invitation, error) {
	// 1. Validate input.
	if cmd.TeamID == "" {
		return nil, fmt.Errorf("team ID is required")
	}
	if cmd.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if cmd.InvitedBy == "" {
		return nil, fmt.Errorf("invited by is required")
	}
	if cmd.Role == "" {
		return nil, fmt.Errorf("role is required")
	}
	if !teamdomain.IsValidAccountRole(cmd.Role) {
		return nil, fmt.Errorf("invalid role: %q (must be %q or %q)",
			cmd.Role, teamdomain.RoleAccountAdmin, teamdomain.RoleTrainer)
	}

	// 2. Load the team.
	team, err := s.repo.GetTeamByID(ctx, cmd.TeamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, teamdomain.ErrTeamNotFound
	}

	// 3. Authorize: the inviter must have `member:invite` in the team's
	//    parent account domain.
	accountDomain := accountdomain.AccountDomain(team.AccountID)
	allowed, err := s.casbinSvc.CanInviteMember(ctx, cmd.InvitedBy, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, teamdomain.ErrPermissionDenied
	}

	// 4. Guard: only an account_admin may invite another account_admin.
	if cmd.Role == teamdomain.RoleAccountAdmin {
		inviterIsAdmin, err := s.casbinSvc.CanManageMembers(ctx, cmd.InvitedBy, accountDomain)
		if err != nil {
			return nil, fmt.Errorf("permission check failed: %w", err)
		}
		if !inviterIsAdmin {
			return nil, teamdomain.ErrPermissionDenied
		}
	}

	// 5. Resolve inviter name for the email.
	inviter, err := s.authSvc.GetUserByID(ctx, cmd.InvitedBy)
	inviterName := cmd.InvitedBy
	if err == nil && inviter != nil && inviter.DisplayName != "" {
		inviterName = inviter.DisplayName
	}
	log.Printf("[InviteMember] Inviter: %s (%s)", inviterName, cmd.InvitedBy)

		// 6. Check whether the invitee is a registered user, and if so, whether
	//    they're already a team member; also check whether they're already
	//    an account member (which affects the email copy).
	userExists, err := s.authSvc.UserExists(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user: %w", err)
	}

	var user *UserResult
	var isExistingAccountMember bool

	if userExists {
		user, err = s.authSvc.GetUserByEmail(ctx, cmd.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		existing, err := s.repo.GetMemberByTeamAndUser(ctx, cmd.TeamID, user.ID)
		if err == nil && existing != nil && existing.IsActive {
			return nil, teamdomain.ErrMemberAlreadyExists
		}

		// NEW: is the invitee already in the account?
		existingRole, err := s.authSvc.GetUserRoleInAccount(ctx, user.ID, team.AccountID)
		if err != nil {
			log.Printf("⚠️ failed to check account membership for %s: %v", user.ID, err)
		}
		isExistingAccountMember = existingRole != ""
	}

	// 7. Reject if there is already a pending invitation for this email + team.
	existingInvitation, err := s.repo.GetInvitationByEmailAndTeam(ctx, cmd.Email, cmd.TeamID)
	if err == nil && existingInvitation != nil &&
		existingInvitation.Status == teamdomain.InvitationStatusPending {
		return nil, teamdomain.ErrInvitationPending
	}

	// 8. Create the invitation.
	token := teamdomain.GenerateToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	invitation, err := teamdomain.NewInvitation(
		cmd.TeamID,
		cmd.Email,
		cmd.Role,
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

	// 9. Dispatch the invitation email asynchronously.
	go s.processAsyncInvitationEmail(
		cmd.Email,
		token,
		inviterName,
		cmd.InvitedBy,
		team,
		invitation.Role,
		isExistingAccountMember,
		user,
		userExists,
	)

	log.Printf("✅ Invitation created for %s to join team %s as %s (user exists: %v)",
		cmd.Email, team.ID, invitation.Role, userExists)
	return invitation, nil
}

// ResendInvitation refreshes an existing invitation and re-dispatches the
// email.
func (s *teamService) ResendInvitation(ctx context.Context, invitationID string) (*teamdomain.Invitation, error) {
	// 1. Load invitation.
	invitation, err := s.repo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, teamdomain.ErrInvitationNotFound
	}
	if invitation.Status == teamdomain.InvitationStatusAccepted {
		return nil, fmt.Errorf("invitation already accepted")
	}

	// 2. Resolve inviter name for the email.
	inviter, err := s.authSvc.GetUserByID(ctx, invitation.InvitedBy)
	inviterName := invitation.InvitedBy
	if err == nil && inviter != nil && inviter.DisplayName != "" {
		inviterName = inviter.DisplayName
	}
	log.Printf("[ResendInvitation] Inviter: %s (%s)", inviterName, invitation.InvitedBy)

	// 3. Reissue: new token, new expiry, reset to pending.
	newToken := teamdomain.GenerateToken()
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := invitation.Reissue(newToken, newExpiresAt); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

		// 4. Load team.
	team, err := s.repo.GetTeamByID(ctx, invitation.TeamID)
	if err != nil {
		return nil, err
	}

	// 5. Check whether the invitee is a registered user, and if so, whether
	//    they're already an account member.
	userExists, err := s.authSvc.UserExists(ctx, invitation.Email)
	if err != nil {
		log.Printf("⚠️ Failed to check if user exists: %v", err)
	}

	var user *UserResult
	var isExistingAccountMember bool

	if userExists {
		user, err = s.authSvc.GetUserByEmail(ctx, invitation.Email)
		if err != nil {
			log.Printf("⚠️ Failed to get user: %v", err)
		}

		if user != nil && team != nil {
			existingRole, err := s.authSvc.GetUserRoleInAccount(ctx, user.ID, team.AccountID)
			if err != nil {
				log.Printf("⚠️ failed to check account membership for %s: %v", user.ID, err)
			}
			isExistingAccountMember = existingRole != ""
		}
	}

	// 6. Dispatch email asynchronously.
	go s.processAsyncInvitationEmail(
		invitation.Email,
		newToken,
		inviterName,
		invitation.InvitedBy,
		team,
		invitation.Role,
		isExistingAccountMember,   // ← added
		user,
		userExists,
	)

	log.Printf("✅ Invitation reissued for %s as %s", invitation.Email, invitation.Role)
	return invitation, nil
}

// ============================================================
// HELPER WORKERS & VALIDATION
// ============================================================

// processAsyncInvitationEmail handles AI generation and notification
// sending off the HTTP thread.
func (s *teamService) processAsyncInvitationEmail(
	email string,
	token string,
	inviterName string,
	invitedByID string,
	team *teamdomain.Team,
	role string,
	isExistingAccountMember bool,
	user *UserResult,
	userExists bool,
) {
	bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var personalizedContent *types.PersonalizedInvitationContent

	if s.aiSvc != nil {
		log.Printf("[AsyncInviteWorker] 🤖 Generating AI content for invitation to %s", email)

		aiReq := GenerateInvitationRequest{
			RecipientName:    getUserDisplayName(user),
			RecipientEmail:   email,
			IsExistingAccountMember: isExistingAccountMember,
			IsExistingUser:   userExists,
			InviterName:      inviterName,
			InviterRole:      s.getUserRole(bgCtx, invitedByID, team.AccountID),
			TeamName:         team.DisplayName,
			TeamType:         string(team.Type),
			InvitedRole:      role,
			TeamMemberCount:  s.getTeamMemberCount(bgCtx, team.ID),
			TeamEventCount:   s.getTeamEventCount(bgCtx, team.ID),
			RecentEventNames: s.getRecentEventNames(bgCtx, team.ID, 3),
			TeamMemberNames:  s.getTeamMemberNames(bgCtx, team.ID, 5),
		}

		content, err := s.aiSvc.GenerateInvitationContent(bgCtx, aiReq)
		if err != nil {
			log.Printf("⚠️ [AsyncInviteWorker] AI generation failed for %s: %v, proceeding with default template", email, err)
		} else {
			personalizedContent = content
			log.Printf("✅ [AsyncInviteWorker] AI content generated for %s", email)
		}
	}

	if userExists {
		userName := ""
		if user != nil {
			userName = user.Name
		}

		if err := s.notifSvc.SendTeamInviteExistingUser(bgCtx, SendTeamInviteExistingUserRequest{
			To:                  email,
			UserName:            userName,
			InvitedBy:           inviterName,
			TeamName:            team.DisplayName,
			Role:                role,
			TeamID:              team.ID,
			AcceptLink:          fmt.Sprintf("https://nuruvent.com/invitations/accept?token=%s", token),
			ExpiresIn:           "7 days",
			PersonalizedContent: personalizedContent,
		}); err != nil {
			log.Printf("❌ [AsyncInviteWorker] Failed to send invitation email to existing user (%s): %v", email, err)
			return
		}
	} else {
		if err := s.notifSvc.SendTeamInviteRegistration(bgCtx, SendTeamInviteRegistrationRequest{
			To:                  email,
			Name:                "",
			InvitedBy:           inviterName,
			TeamName:            team.DisplayName,
			Role:                role,
			TeamID:              team.ID,
			RegistrationLink:    fmt.Sprintf("https://nuruvent.com/register?token=%s", token),
			ExpiresIn:           "7 days",
			PersonalizedContent: personalizedContent,
		}); err != nil {
			log.Printf("❌ [AsyncInviteWorker] Failed to send invitation email to new user (%s): %v", email, err)
			return
		}
	}

	log.Printf("✅ [AsyncInviteWorker] Invitation email successfully delivered to %s", email)
}

// ValidateInvitationToken validates an invitation token and returns the
// invitation if it is usable.
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

// AcceptInvitation accepts an invitation, adds the user to the account
// (with the invited role, if not already a member), and adds them to the
// team.
//
// Behavior:
//   - If the user is not yet an account member: create account_members
//     row with the invitation's role, write Casbin g rule.
//   - If the user is already an account member: their existing role wins;
//     only the team membership is added.
func (s *teamService) AcceptInvitation(ctx context.Context, token, userID string) (*teamdomain.Member, error) {
	// 1. Load invitation.
	invitation, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, teamdomain.ErrInvitationNotFound
	}

	// 2. Validate state and expiry.
	if invitation.Status != teamdomain.InvitationStatusPending {
		return nil, teamdomain.ErrInvitationAlreadyAccepted
	}
	if invitation.IsExpired() {
		return nil, teamdomain.ErrInvitationExpired
	}

	// 3. Load the user.
	user, err := s.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 4. Email must match (case-insensitive).
	if !strings.EqualFold(user.Email, invitation.Email) {
		return nil, teamdomain.ErrInvitationEmailMismatch
	}

	// 5. Load the team.
	team, err := s.repo.GetTeamByID(ctx, invitation.TeamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, teamdomain.ErrTeamNotFound
	}

	// 6. If already a team member, just mark the invitation accepted.
	existing, err := s.repo.GetMemberByTeamAndUser(ctx, invitation.TeamID, userID)
	if err == nil && existing != nil && existing.IsActive {
		if err := invitation.Accept(); err == nil {
			_ = s.repo.UpdateInvitation(ctx, invitation)
		}
		return existing, nil
	}

	// 7. Execute everything in a transaction.
	var member *teamdomain.Member
	err = s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// 7a. Ensure account membership with the invited role.
		if err := s.ensureAccountMembership(txCtx, userID, team.AccountID, invitation.Role); err != nil {
			return err
		}

		// 7b. Add team membership.
		m, err := s.AddMember(txCtx, invitation.TeamID, userID, invitation.InvitedBy)
		if err != nil {
			return fmt.Errorf("failed to add team member: %w", err)
		}
		member = m

		// 7c. Mark invitation accepted.
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

	// 8. Reload Casbin so the new g rule (if any) is visible immediately.
	if err := s.casbinSvc.ReloadPolicies(ctx); err != nil {
		log.Printf("⚠️ failed to reload casbin policies after invitation accept: %v", err)
	}

	
	// 9. Notify the inviter asynchronously.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Resolve the inviter (user ID → email + display name).
		inviter, err := s.authSvc.GetUserByID(bgCtx, invitation.InvitedBy)
		if err != nil || inviter == nil {
			log.Printf("⚠️ [AcceptInvitation] could not load inviter %s for notification: %v",
				invitation.InvitedBy, err)
			return
		}

		adminName := inviter.DisplayName
		if adminName == "" {
			adminName = inviter.Name
		}
		if adminName == "" {
			adminName = inviter.Email
		}

		if err := s.notifSvc.SendTeamInviteAccepted(bgCtx, SendTeamInviteAcceptedRequest{
			To:        inviter.Email,   // ✅ real email
			AdminName: adminName,       // ✅ real name
			UserName:  user.Name,
			UserEmail: user.Email,
			TeamName:  team.DisplayName,
			TeamID:    team.ID,
		}); err != nil {
			log.Printf("⚠️ Failed to send invitation-accepted notification to %s: %v",
				inviter.Email, err)
		}
	}()

	log.Printf("✅ User %s accepted invitation to team %s as %s",
		user.Email, invitation.TeamID, invitation.Role)
	return member, nil
}

// ensureAccountMembership ensures the user has the given role in the account.
//
// Behavior:
//   - If the user is already an account member: no-op. Their existing
//     role wins; the invitation's role is ignored.
//   - If the user is not yet an account member: create the account_members
//     row with the given role and write the Casbin g rule.
func (s *teamService) ensureAccountMembership(
	ctx context.Context,
	userID, accountID, role string,
) error {
	existingRole, err := s.authSvc.GetUserRoleInAccount(ctx, userID, accountID)
	if err != nil {
		return fmt.Errorf("failed to check account membership: %w", err)
	}
	if existingRole != "" {
		// Already a member. Keep the existing role.
		log.Printf("[AcceptInvitation] User %s is already a member of account %s with role %s; keeping existing role",
			userID, accountID, existingRole)
		return nil
	}

	if err := s.authSvc.AddAccountMember(ctx, accountID, userID, role); err != nil {
		return fmt.Errorf("failed to add account member: %w", err)
	}

	accountDomain := accountdomain.AccountDomain(accountID)
	if err := s.casbinSvc.AssignRole(ctx, accountDomain, userID, role); err != nil {
		return fmt.Errorf("failed to assign casbin role: %w", err)
	}

	log.Printf("[AcceptInvitation] Added user %s to account %s with role %s",
		userID, accountID, role)
	return nil
}

// DeclineInvitation declines an invitation.
func (s *teamService) DeclineInvitation(ctx context.Context, token, userID string) error {
	// 1. Load invitation.
	invitation, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return err
	}
	if invitation == nil {
		return teamdomain.ErrInvitationNotFound
	}

	// 2. Validate.
	if invitation.Status != teamdomain.InvitationStatusPending {
		return teamdomain.ErrInvitationAlreadyAccepted
	}
	if invitation.IsExpired() {
		return teamdomain.ErrInvitationExpired
	}

	// 3. Load user.
	user, err := s.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// 4. Email must match (case-insensitive).
	if !strings.EqualFold(user.Email, invitation.Email) {
		return teamdomain.ErrInvitationEmailMismatch
	}

	// 5. Decline.
	if err := invitation.Decline(); err != nil {
		return err
	}
	if err := s.repo.UpdateInvitation(ctx, invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	log.Printf("✅ User %s declined invitation to team %s", user.Email, invitation.TeamID)
	return nil
}

// GetTeamInvitations retrieves all invitations for a team.
func (s *teamService) GetTeamInvitations(
	ctx context.Context,
	teamID string,
	filters teamdomain.ListInvitationsFilters,
) ([]*teamdomain.Invitation, int64, error) {
	if teamID == "" {
		return nil, 0, fmt.Errorf("team ID is required")
	}
	return s.repo.GetInvitationsByTeam(ctx, teamID, filters)
}