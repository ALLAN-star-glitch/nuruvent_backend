// internal/modules/account/service/account_crud_service.go

package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// CREATE
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

// ============================================================
// READ
// ============================================================

// GetAccountByTeamID resolves the account that owns a given team.
// Permission: caller must be able to view the account.
// Returns (nil, nil) if the team doesn't exist or has no account.
func (s *accountService) GetAccountByTeamID(ctx context.Context, teamID string) (*accountdomain.Account, error) {
	if teamID == "" {
		return nil, fmt.Errorf("team ID is required")
	}

	accountID, err := s.repo.GetAccountIDByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve team %s: %w", teamID, err)
	}
	if accountID == "" {
		return nil, nil
	}

	// Permission check: caller must be able to view this account
	if err := s.requireAccountView(ctx, accountID); err != nil {
		return nil, err
	}

	return s.repo.GetAccountByID(ctx, accountID)
}

// GetAccountByID retrieves an account by ID.
// Permission: caller must be able to view the account.
func (s *accountService) GetAccountByID(ctx context.Context, id string) (*accountdomain.Account, error) {
	if id == "" {
		return nil, accountdomain.ErrAccountNotFound
	}

	// Permission check
	if err := s.requireAccountView(ctx, id); err != nil {
		return nil, err
	}

	return s.repo.GetAccountByID(ctx, id)
}

// GetAccountBySlug retrieves an account by slug.
// Permission: caller must be able to view the account.
func (s *accountService) GetAccountBySlug(ctx context.Context, slug string) (*accountdomain.Account, error) {
	if slug == "" {
		return nil, accountdomain.ErrAccountNotFound
	}

	account, err := s.repo.GetAccountBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}

	// Permission check
	if err := s.requireAccountView(ctx, account.ID); err != nil {
		return nil, err
	}

	return account, nil
}

// GetUserAccounts lists accounts the given user belongs to.
// No additional permission check: a user can always see their own accounts.
// (The caller is expected to pass the authenticated user's ID.)
func (s *accountService) GetUserAccounts(ctx context.Context, userID string) ([]*accountdomain.Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.repo.GetAccountsByUserID(ctx, userID)
}

// ============================================================
// UPDATE
// ============================================================

// UpdateAccount updates an account.
// Permission: caller must be able to manage the account.
func (s *accountService) UpdateAccount(ctx context.Context, id string, updates map[string]interface{}) (*accountdomain.Account, error) {
	if id == "" {
		return nil, accountdomain.ErrAccountNotFound
	}

	// Permission check
	if err := s.requireAccountManage(ctx, id); err != nil {
		return nil, err
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

// ============================================================
// DELETE
// ============================================================

// DeleteAccount soft-deletes an account.
// Permission: caller must be able to manage the account.
func (s *accountService) DeleteAccount(ctx context.Context, id string) error {
	if id == "" {
		return accountdomain.ErrAccountNotFound
	}

	// Permission check
	if err := s.requireAccountManage(ctx, id); err != nil {
		return err
	}

	account, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return err
	}
	if account == nil {
		return accountdomain.ErrAccountNotFound
	}

	if err := s.repo.DeleteAccount(ctx, id); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	log.Printf("✅ Account deleted: %s (ID: %s)", account.Name, account.ID)
	return nil
}

// ============================================================
// PERMISSION HELPERS
// ============================================================

// requireAccountView checks that the caller can view the account.
// Skips the check if no user is in the context (system callers / migrations).
// Returns accountdomain.ErrForbidden if denied.
func (s *accountService) requireAccountView(ctx context.Context, accountID string) error {
	userID := accountdomain.GetUserID(ctx)
	if userID == "" {
		return nil // system caller — no user context
	}

	domainStr := accountdomain.AccountDomain(accountID)
	allowed, err := s.permChecker.CanViewAccount(ctx, userID, domainStr)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return accountdomain.ErrForbidden
	}
	return nil
}

// requireAccountManage checks that the caller can manage the account.
func (s *accountService) requireAccountManage(ctx context.Context, accountID string) error {
	userID := accountdomain.GetUserID(ctx)
	if userID == "" {
		return nil
	}

	domainStr := accountdomain.AccountDomain(accountID)
	allowed, err := s.permChecker.CanManageAccount(ctx, userID, domainStr)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return accountdomain.ErrForbidden
	}
	return nil
}


// ============================================================
// ACCOUNT LOGO
// ============================================================

// UploadAccountLogo stores a new logo for the account and updates logo_url.
// Permission: caller must be able to manage the account.
func (s *accountService) UploadAccountLogo(ctx context.Context, accountID string, file []byte, filename, contentType string) (*accountdomain.Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if len(file) == 0 {
		return nil, fmt.Errorf("logo file is required")
	}

	// Permission check
	if err := s.requireAccountManage(ctx, accountID); err != nil {
		return nil, err
	}

	// 1. Upload via media port (media_type_business)
	media, err := s.mediaSvc.UploadFile(ctx, accountdomain.UploadMediaCommand{
		File:          file,
		FileName:      filename,
		ContentType:   contentType,
		MediaTypeName: types.MediaTypeBusinessName,
		EntityID:      accountID,
		UploadedBy:    accountdomain.GetUserID(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload logo: %w", err)
	}

	// 2. Persist new URL on the account
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to load account: %w", err)
	}
	if account == nil {
		return nil, accountdomain.ErrAccountNotFound
	}
	account.LogoURL = media.URL
	account.UpdatedAt = time.Now()
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("✅ Account logo uploaded: %s (ID: %s)", account.Name, account.ID)
	return account, nil
}

// DeleteAccountLogo removes the account's logo from storage and clears logo_url.
// Permission: caller must be able to manage the account.
func (s *accountService) DeleteAccountLogo(ctx context.Context, accountID string) error {
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}

	// Permission check
	if err := s.requireAccountManage(ctx, accountID); err != nil {
		return err
	}

	// 1. Resolve media type
	mediaType, err := s.mediaSvc.GetMediaTypeByName(ctx, types.MediaTypeBusinessName)
	if err != nil {
		return fmt.Errorf("failed to resolve media type: %w", err)
	}
	if mediaType == nil {
		return fmt.Errorf("media type %q not found", types.MediaTypeBusinessName)
	}

	// 2. Delete files
	if err := s.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, accountID, mediaType.ID); err != nil {
		return fmt.Errorf("failed to delete logo files: %w", err)
	}

	// 3. Clear logo_url
	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to load account: %w", err)
	}
	if account == nil {
		return accountdomain.ErrAccountNotFound
	}
	account.LogoURL = ""
	account.UpdatedAt = time.Now()
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("✅ Account logo deleted: %s (ID: %s)", account.Name, account.ID)
	return nil
}