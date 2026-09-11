// internal/modules/account/service/account_service.go

package service

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

type accountService struct {
    repo      accountdomain.Repository
    authSvc   AuthService
    notifSvc  NotificationService
    sanitizer validation.Sanitize
}

func NewAccountService(
    repo accountdomain.Repository,
    authSvc AuthService,
    notifSvc NotificationService,
) Service {
    return &accountService{
        repo:      repo,
        authSvc:   authSvc,
        notifSvc:  notifSvc,
        sanitizer: validation.Sanitize{},
    }
}

// ============================================================
// ACCOUNT TYPE OPERATIONS
// ============================================================

func (s *accountService) GetAccountTypes(ctx context.Context) ([]*accountdomain.AccountType, error) {
    return s.repo.GetAccountTypes(ctx)
}

func (s *accountService) GetAccountTypeByID(ctx context.Context, id string) (*accountdomain.AccountType, error) {
    if id == "" {
        return nil, accountdomain.ErrAccountTypeNotFound
    }
    return s.repo.GetAccountTypeByID(ctx, id)
}

func (s *accountService) GetAccountTypeBySlug(ctx context.Context, slug string) (*accountdomain.AccountType, error) {
    if slug == "" {
        return nil, accountdomain.ErrAccountTypeNotFound
    }
    return s.repo.GetAccountTypeBySlug(ctx, slug)
}

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

func (s *accountService) CreatePersonalAccount(ctx context.Context, cmd CreatePersonalAccountCommand) (*accountdomain.Account, error) {
    // 1. Validate input
    if cmd.Name == "" {
        return nil, fmt.Errorf("name is required")
    }
    if cmd.Email == "" {
        return nil, fmt.Errorf("email is required")
    }
    if cmd.CreatedBy == "" {
        return nil, fmt.Errorf("created by is required")
    }

    // 2. Get account type
    accountType, err := s.repo.GetAccountTypeByName(ctx, types.AccountTypePersonalName)
    if err != nil {
        return nil, fmt.Errorf("failed to get account type: %w", err)
    }

    // 3. Check if user exists
    user, err := s.authSvc.GetUserByID(ctx, cmd.CreatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    if user == nil {
        return nil, fmt.Errorf("user not found")
    }

    // 4. Check if account with email already exists
    existing, err := s.repo.GetAccountByEmail(ctx, cmd.Email)
    if err == nil && existing != nil {
        return nil, accountdomain.ErrAccountAlreadyExists
    }

    // 5. Create account
    account, err := accountdomain.NewPersonalAccount(
        cmd.Name,
        cmd.Email,
        cmd.Phone,
        cmd.CreatedBy,
        accountType.ID,
    )
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateAccount(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to create account: %w", err)
    }

    // 6. Add user as account_admin
    member, err := accountdomain.NewAccountMember(account.ID, cmd.CreatedBy, accountdomain.RoleAccountAdmin, cmd.CreatedBy)
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateAccountMember(ctx, member); err != nil {
        // Rollback account creation
        _ = s.repo.DeleteAccount(ctx, account.ID)
        return nil, fmt.Errorf("failed to add account member: %w", err)
    }

    log.Printf("✅ Personal account created: %s (ID: %s)", account.Name, account.ID)
    return account, nil
}

func (s *accountService) CreateInstitutionAccount(ctx context.Context, cmd CreateInstitutionAccountCommand) (*accountdomain.Account, error) {
    // 1. Validate input
    if cmd.Name == "" {
        return nil, fmt.Errorf("institution name is required")
    }
    if cmd.Email == "" {
        return nil, fmt.Errorf("institution email is required")
    }
    if cmd.CreatedBy == "" {
        return nil, fmt.Errorf("created by is required")
    }

    // 2. Get account type
    accountType, err := s.repo.GetAccountTypeByName(ctx, types.AccountTypeInstitutionName)
    if err != nil {
        return nil, fmt.Errorf("failed to get account type: %w", err)
    }

    // 3. Check if user exists
    user, err := s.authSvc.GetUserByID(ctx, cmd.CreatedBy)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    if user == nil {
        return nil, fmt.Errorf("user not found")
    }

    // 4. Check if account with email already exists
    existing, err := s.repo.GetAccountByEmail(ctx, cmd.Email)
    if err == nil && existing != nil {
        return nil, accountdomain.ErrAccountAlreadyExists
    }

    // 5. Create account
    account, err := accountdomain.NewInstitutionAccount(
        cmd.Name,
        cmd.Email,
        cmd.Phone,
        cmd.Website,
        cmd.Description,
        cmd.CreatedBy,
        accountType.ID,
    )
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateAccount(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to create account: %w", err)
    }

    // 6. Add user as account_admin
    member, err := accountdomain.NewAccountMember(account.ID, cmd.CreatedBy, accountdomain.RoleAccountAdmin, cmd.CreatedBy)
    if err != nil {
        return nil, err
    }

    if err := s.repo.CreateAccountMember(ctx, member); err != nil {
        // Rollback account creation
        _ = s.repo.DeleteAccount(ctx, account.ID)
        return nil, fmt.Errorf("failed to add account member: %w", err)
    }

    log.Printf("✅ Institution account created: %s (ID: %s)", account.Name, account.ID)
    return account, nil
}

func (s *accountService) GetAccountByID(ctx context.Context, id string) (*accountdomain.Account, error) {
    if id == "" {
        return nil, accountdomain.ErrAccountNotFound
    }
    return s.repo.GetAccountByID(ctx, id)
}

func (s *accountService) GetAccountBySlug(ctx context.Context, slug string) (*accountdomain.Account, error) {
    if slug == "" {
        return nil, accountdomain.ErrAccountNotFound
    }
    return s.repo.GetAccountBySlug(ctx, slug)
}

func (s *accountService) GetUserAccounts(ctx context.Context, userID string) ([]*accountdomain.Account, error) {
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }
    return s.repo.GetAccountsByUserID(ctx, userID)
}

func (s *accountService) UpdateAccount(ctx context.Context, id string, updates map[string]interface{}) (*accountdomain.Account, error) {
    if id == "" {
        return nil, accountdomain.ErrAccountNotFound
    }

    account, err := s.repo.GetAccountByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if account == nil {
        return nil, accountdomain.ErrAccountNotFound
    }

    // Apply updates
    if name, ok := updates["name"].(string); ok && name != "" {
        account.Name = s.sanitizer.DisplayName(name)
        account.DisplayName = account.Name
        account.Slug = s.sanitizer.GenerateSlugFromName(account.Name)
    }
    if email, ok := updates["email"].(string); ok && email != "" {
        account.Email = s.sanitizer.Identifier(email)
    }
    if phone, ok := updates["phone"].(string); ok {
        account.Phone = phone
    }
    if website, ok := updates["website"].(string); ok {
        account.Website = s.sanitizer.Identifier(website)
    }
    if description, ok := updates["description"].(string); ok {
        account.Description = s.sanitizer.Description(description)
    }
    if logoURL, ok := updates["logo_url"].(string); ok {
        account.LogoURL = logoURL
    }
    account.UpdatedAt = time.Now()

    if err := s.repo.UpdateAccount(ctx, account); err != nil {
        return nil, fmt.Errorf("failed to update account: %w", err)
    }

    log.Printf("✅ Account updated: %s (ID: %s)", account.Name, account.ID)
    return account, nil
}

func (s *accountService) DeleteAccount(ctx context.Context, id string) error {
    if id == "" {
        return accountdomain.ErrAccountNotFound
    }

    // Check if account exists
    account, err := s.repo.GetAccountByID(ctx, id)
    if err != nil {
        return err
    }
    if account == nil {
        return accountdomain.ErrAccountNotFound
    }

    // Delete account (soft delete)
    if err := s.repo.DeleteAccount(ctx, id); err != nil {
        return fmt.Errorf("failed to delete account: %w", err)
    }

    log.Printf("✅ Account deleted: %s (ID: %s)", account.Name, account.ID)
    return nil
}

// ============================================================
// ACCOUNT MEMBER OPERATIONS
// ============================================================

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

    // 2. Check if account exists
    account, err := s.repo.GetAccountByID(ctx, cmd.AccountID)
    if err != nil {
        return nil, err
    }
    if account == nil {
        return nil, accountdomain.ErrAccountNotFound
    }

    // 3. Check if user exists
    user, err := s.authSvc.GetUserByID(ctx, cmd.UserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    if user == nil {
        return nil, fmt.Errorf("user not found")
    }

    // 4. Check if already a member
    existing, err := s.repo.GetAccountMemberByAccountAndUser(ctx, cmd.AccountID, cmd.UserID)
    if err == nil && existing != nil && existing.IsActive {
        return nil, accountdomain.ErrAccountMemberAlreadyExists
    }

    // 5. Create member
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
    if member.Role == accountdomain.RoleAccountAdmin {
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
    }

    // Delete member
    if err := s.repo.DeleteAccountMember(ctx, accountID, userID); err != nil {
        return fmt.Errorf("failed to remove account member: %w", err)
    }

    log.Printf("✅ User %s removed from account %s by %s", userID, accountID, removedBy)
    return nil
}

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

func (s *accountService) GetAccountMembers(ctx context.Context, accountID string) ([]*accountdomain.AccountMember, error) {
    if accountID == "" {
        return nil, fmt.Errorf("account ID is required")
    }

    return s.repo.GetAccountMembersByAccount(ctx, accountID)
}

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
    if member.Role == accountdomain.RoleAccountAdmin {
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
    }

    // Delete member
    if err := s.repo.DeleteAccountMember(ctx, accountID, userID); err != nil {
        return fmt.Errorf("failed to leave account: %w", err)
    }

    log.Printf("✅ User %s left account %s", userID, accountID)
    return nil
}

func (s *accountService) GetAccountByTeamID(ctx context.Context, teamID string) (*accountdomain.Account, error) {
	// resolve team → account_id
	accountID, err := s.repo.GetAccountIDByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if accountID == "" {
		return nil, nil
	}
	// resolve account
	return s.repo.GetAccountByID(ctx, accountID)
}