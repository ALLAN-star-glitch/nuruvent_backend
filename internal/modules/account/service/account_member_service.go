// internal/modules/account/service/account_member_service.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
)

// ============================================================
// ADD MEMBER
// ============================================================

// AddMember adds a member to an account.
// Permission: caller must be able to manage account members.
func (s *accountService) AddMember(ctx context.Context, cmd AddMemberCommand) (*accountdomain.AccountMember, error) {
	// 1. Validate input
	if cmd.AccountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if cmd.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if cmd.Role == "" {
		return nil, fmt.Errorf("role is required")
	}
	if cmd.InvitedBy == "" {
		return nil, fmt.Errorf("invited by is required")
	}

	// 2. Permission check
	if err := s.requireMemberManage(ctx, cmd.AccountID); err != nil {
		return nil, err
	}

	// 3. Check if account exists
	account, err := s.repo.GetAccountByID(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, accountdomain.ErrAccountNotFound
	}

	// 4. Check if user exists
	user, err := s.authSvc.GetUserByID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 5. Check if already a member
	existing, err := s.repo.GetAccountMemberByAccountAndUser(ctx, cmd.AccountID, cmd.UserID)
	if err == nil && existing != nil && existing.IsActive {
		return nil, accountdomain.ErrAccountMemberAlreadyExists
	}

	// 6. Create member
	member, err := accountdomain.NewAccountMember(cmd.AccountID, cmd.UserID, cmd.Role, cmd.InvitedBy)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateAccountMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add account member: %w", err)
	}

	log.Printf("✅ User %s added to account %s as %s", cmd.UserID, cmd.AccountID, cmd.Role)
	return member, nil
}

// ============================================================
// REMOVE MEMBER
// ============================================================

// RemoveMember removes a member from an account.
// Permission: caller must be able to manage account members.
// Guard: cannot remove self; cannot remove the last admin.
func (s *accountService) RemoveMember(ctx context.Context, accountID, userID, removedBy string) error {
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if removedBy == "" {
		return fmt.Errorf("removed by is required")
	}

	// Permission check
	if err := s.requireMemberManage(ctx, accountID); err != nil {
		return err
	}

	// Check if account exists
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return accountdomain.ErrAccountNotFound
	}

	// Get member
	member, err := s.repo.GetAccountMemberByAccountAndUser(ctx, accountID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return accountdomain.ErrAccountMemberNotFound
	}

	// Don't allow removing self
	if userID == removedBy {
		return accountdomain.ErrCannotRemoveSelf
	}

	// Check if this is the last admin
	if err := s.ensureNotLastAdmin(ctx, accountID, member); err != nil {
		return err
	}

	if err := s.repo.DeleteAccountMember(ctx, accountID, userID); err != nil {
		return fmt.Errorf("failed to remove account member: %w", err)
	}

	log.Printf("✅ User %s removed from account %s by %s", userID, accountID, removedBy)
	return nil
}

// ============================================================
// UPDATE MEMBER ROLE
// ============================================================

// UpdateMemberRole changes a member's role.
// Permission: caller must be able to manage account members.
// Guard: cannot change own role.
func (s *accountService) UpdateMemberRole(ctx context.Context, accountID, userID, newRole, updatedBy string) (*accountdomain.AccountMember, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if newRole == "" {
		return nil, fmt.Errorf("new role is required")
	}

	// Permission check
	if err := s.requireMemberManage(ctx, accountID); err != nil {
		return nil, err
	}

	// Check if account exists
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, accountdomain.ErrAccountNotFound
	}

	// Get member
	member, err := s.repo.GetAccountMemberByAccountAndUser(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, accountdomain.ErrAccountMemberNotFound
	}

	// Don't allow changing own role
	if userID == updatedBy {
		return nil, accountdomain.ErrCannotChangeOwnRole
	}

	// Update role
	if err := member.UpdateRole(newRole); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateAccountMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	log.Printf("✅ User %s role updated to %s in account %s", userID, newRole, accountID)
	return member, nil
}

// ============================================================
// LIST MEMBERS
// ============================================================

// GetAccountMembers lists the members of an account.
// Permission: caller must be able to view the account.
func (s *accountService) GetAccountMembers(ctx context.Context, accountID string) ([]*accountdomain.AccountMember, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	// Permission check
	if err := s.requireAccountView(ctx, accountID); err != nil {
		return nil, err
	}

	return s.repo.GetAccountMembersByAccount(ctx, accountID)
}

// ============================================================
// LEAVE ACCOUNT
// ============================================================

// LeaveAccount removes the caller themselves from an account.
// No permission check beyond "you are a member" — the caller can
// always leave.
// Guard: last admin cannot leave.
func (s *accountService) LeaveAccount(ctx context.Context, accountID, userID string) error {
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Check if account exists
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return accountdomain.ErrAccountNotFound
	}

	// Get member
	member, err := s.repo.GetAccountMemberByAccountAndUser(ctx, accountID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return accountdomain.ErrAccountMemberNotFound
	}

	// Check if this is the last admin
	if err := s.ensureNotLastAdmin(ctx, accountID, member); err != nil {
		return err
	}

	if err := s.repo.DeleteAccountMember(ctx, accountID, userID); err != nil {
		return fmt.Errorf("failed to leave account: %w", err)
	}

	log.Printf("✅ User %s left account %s", userID, accountID)
	return nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// requireMemberManage checks that the caller can manage members of an account.
func (s *accountService) requireMemberManage(ctx context.Context, accountID string) error {
	userID := accountdomain.GetUserID(ctx)
	if userID == "" {
		return nil
	}

	domainStr := accountdomain.AccountDomain(accountID)
	allowed, err := s.permChecker.CanManageAccountMembers(ctx, userID, domainStr)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return accountdomain.ErrForbidden
	}
	return nil
}

// ensureNotLastAdmin returns ErrLastAdminCannotLeave if removing the
// member would leave the account with zero admins.
func (s *accountService) ensureNotLastAdmin(ctx context.Context, accountID string, member *accountdomain.AccountMember) error {
	if member.Role != accountdomain.RoleAccountAdmin {
		return nil
	}

	admins, err := s.repo.GetAccountMembersByAccount(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to check admins: %w", err)
	}

	adminCount := 0
	for _, a := range admins {
		if a.Role == accountdomain.RoleAccountAdmin && a.IsActive {
			adminCount++
		}
	}
	if adminCount <= 1 {
		return accountdomain.ErrLastAdminCannotLeave
	}
	return nil
}