// internal/modules/profile/service/account_profile.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// ============================================================
// ACCOUNT PROFILE METHODS
// ============================================================

// GetAccountProfile returns basic account information
func (s *profileService) GetAccountProfile(ctx context.Context, accountID string) (*domain.AccountInfo, error) {
	if accountID == "" {
		return nil, domain.ErrInvalidAccountID
	}

	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	return &domain.AccountInfo{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		LogoURL:     account.LogoURL,
	}, nil
}

// GetAccountProfileWithDetails returns detailed account information with permission checks
func (s *profileService) GetAccountProfileWithDetails(ctx context.Context, accountID string) (*domain.AccountInfo, error) {
	if accountID == "" {
		return nil, domain.ErrInvalidAccountID
	}

	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, domain.ErrPermissionDenied
	}

	// Check if viewer has read permission in the account
	accountDomain := domain.AccountDomain(accountID)
	if accountDomain == "" {
		return nil, domain.ErrInvalidAccountID
	}

	allowed, err := s.permChecker.CanReadProfile(ctx, viewerID, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, domain.ErrPermissionDenied
	}

	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	return &domain.AccountInfo{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		Email:       account.Email,
		Phone:       account.Phone,
		Website:     account.Website,
		Description: account.Description,
		LogoURL:     account.LogoURL,
		Address:     account.Address,
		City:        account.City,
		Country:     account.Country,
		Status:      account.Status,
		KYCStatus:   account.KYCStatus,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}, nil
}

// GetAccountProfiles returns multiple account profiles
func (s *profileService) GetAccountProfiles(ctx context.Context, accountIDs []string) ([]*domain.AccountInfo, error) {
	if len(accountIDs) == 0 {
		return []*domain.AccountInfo{}, nil
	}

	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, domain.ErrPermissionDenied
	}

	accounts, err := s.repo.GetAccountsByIDs(ctx, accountIDs)
	if err != nil {
		return nil, err
	}

	accountInfos := make([]*domain.AccountInfo, len(accounts))
	for i, account := range accounts {
		accountInfos[i] = &domain.AccountInfo{
			ID:          account.ID,
			Name:        account.Name,
			DisplayName: account.DisplayName,
			Slug:        account.Slug,
			Type:        account.Type,
			LogoURL:     account.LogoURL,
		}
	}
	return accountInfos, nil
}

// UpdateAccountProfile updates an account's profile
func (s *profileService) UpdateAccountProfile(ctx context.Context, accountID string, updates map[string]interface{}) (*domain.AccountInfo, error) {
	if accountID == "" {
		return nil, domain.ErrInvalidAccountID
	}

	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, domain.ErrPermissionDenied
	}

	// Check if viewer has update permission in the account
	accountDomain := domain.AccountDomain(accountID)
	if accountDomain == "" {
		return nil, domain.ErrInvalidAccountID
	}

	allowed, err := s.permChecker.CanUpdateProfile(ctx, viewerID, accountDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, domain.ErrPermissionDenied
	}

	account, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		account.Name = name
	}
	if displayName, ok := updates["display_name"].(string); ok {
		account.DisplayName = displayName
	}
	if email, ok := updates["email"].(string); ok {
		account.Email = email
	}
	if phone, ok := updates["phone"].(string); ok {
		account.Phone = phone
	}
	if website, ok := updates["website"].(string); ok {
		account.Website = website
	}
	if description, ok := updates["description"].(string); ok {
		account.Description = description
	}
	if logoURL, ok := updates["logo_url"].(string); ok {
		account.LogoURL = logoURL
	}
	if address, ok := updates["address"].(string); ok {
		account.Address = address
	}
	if city, ok := updates["city"].(string); ok {
		account.City = city
	}
	if country, ok := updates["country"].(string); ok {
		account.Country = country
	}
	if status, ok := updates["status"].(string); ok {
		account.Status = status
	}
	if kycStatus, ok := updates["kyc_status"].(string); ok {
		account.KYCStatus = kycStatus
	}

	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("✅ Account profile updated: %s", accountID)

	return &domain.AccountInfo{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		Email:       account.Email,
		Phone:       account.Phone,
		Website:     account.Website,
		Description: account.Description,
		LogoURL:     account.LogoURL,
		Address:     account.Address,
		City:        account.City,
		Country:     account.Country,
		Status:      account.Status,
		KYCStatus:   account.KYCStatus,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}, nil
}

// ListAccounts returns a paginated list of accounts
func (s *profileService) ListAccounts(ctx context.Context, filters domain.ListAccountsFilters) ([]*domain.AccountInfo, int64, error) {
	viewerID := s.getUserIDFromContext(ctx)
	if viewerID == "" {
		return nil, 0, domain.ErrPermissionDenied
	}

	// Check if viewer has read_all in their personal account
	viewerPersonalDomain := domain.PersonalTeamDomain(viewerID)
	if viewerPersonalDomain == "" {
		return nil, 0, domain.ErrInvalidUserID
	}

	allowed, err := s.permChecker.CanReadAllProfiles(ctx, viewerID, viewerPersonalDomain)
	if err != nil {
		return nil, 0, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, 0, domain.ErrPermissionDenied
	}

	accounts, total, err := s.repo.ListAccounts(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	accountInfos := make([]*domain.AccountInfo, len(accounts))
	for i, account := range accounts {
		accountInfos[i] = &domain.AccountInfo{
			ID:          account.ID,
			Name:        account.Name,
			DisplayName: account.DisplayName,
			Slug:        account.Slug,
			Type:        account.Type,
			LogoURL:     account.LogoURL,
		}
	}
	return accountInfos, total, nil
}