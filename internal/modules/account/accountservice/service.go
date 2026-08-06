// internal/modules/account/accountservice/service.go

package accountservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountrepo"
	"github.com/google/uuid"
)

// Service defines the account service interface
type Service interface {
	// Account operations
	GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*models.Account, error)
	UpdateAccount(ctx context.Context, account *models.Account) error
	DeleteAccount(ctx context.Context, id uuid.UUID) error
	
	// Profile operations
	UpdateProfile(ctx context.Context, id uuid.UUID, name, phone, displayName string) (*models.Account, error)
	
	// Professional profile operations
	UpdateProfessionalType(ctx context.Context, id uuid.UUID, professionalTypeID *uuid.UUID) (*models.Account, error)
	
	// Account type operations
	GetAccountTypeBySlug(ctx context.Context, slug string) (*models.AccountType, error)
}

type service struct {
	repo *accountrepo.Repository
}

// NewService creates a new account service
func NewService(repo *accountrepo.Repository) Service {
	return &service{
		repo: repo,
	}
}

// ================================================
// ACCOUNT OPERATIONS
// ================================================

// GetAccountByID gets an account by ID
func (s *service) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	if id == uuid.Nil {
		return nil, errors.New("account ID is required")
	}
	return s.repo.GetAccountByID(id)
}

// GetAccountByEmail gets an account by email
func (s *service) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	return s.repo.GetAccountByEmail(email)
}

// UpdateAccount updates an existing account
func (s *service) UpdateAccount(ctx context.Context, account *models.Account) error {
	if account == nil {
		return errors.New("account is required")
	}
	if account.ID == uuid.Nil {
		return errors.New("account ID is required")
	}
	return s.repo.UpdateAccount(account)
}

// DeleteAccount soft deletes an account
func (s *service) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("account ID is required")
	}
	return s.repo.DeleteAccount(id)
}

// ================================================
// PROFILE OPERATIONS
// ================================================

// UpdateProfile updates an account's profile information
func (s *service) UpdateProfile(ctx context.Context, id uuid.UUID, name, phone, displayName string) (*models.Account, error) {
	if id == uuid.Nil {
		return nil, errors.New("account ID is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}

	account, err := s.repo.GetAccountByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, errors.New("account not found")
	}

	// Update fields
	if name != "" {
		account.Name = name
	}
	if phone != "" {
		account.Phone = phone
	}
	if displayName != "" {
		account.DisplayName = displayName
	}
	account.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// ================================================
// PROFESSIONAL PROFILE OPERATIONS
// ================================================

// UpdateProfessionalType updates an account's professional type
func (s *service) UpdateProfessionalType(ctx context.Context, id uuid.UUID, professionalTypeID *uuid.UUID) (*models.Account, error) {
	if id == uuid.Nil {
		return nil, errors.New("account ID is required")
	}

	account, err := s.repo.GetAccountByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return nil, errors.New("account not found")
	}

	account.ProfessionalTypeID = professionalTypeID
	account.UpdatedAt = time.Now()

	if err := s.repo.UpdateAccount(account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// ================================================
// ACCOUNT TYPE OPERATIONS
// ================================================

// GetAccountTypeBySlug gets an account type by slug
func (s *service) GetAccountTypeBySlug(ctx context.Context, slug string) (*models.AccountType, error) {
	if slug == "" {
		return nil, errors.New("account type slug is required")
	}
	return s.repo.GetAccountTypeBySlug(slug)
}