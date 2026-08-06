// internal/modules/account/accountrepo/repository.go

package accountrepo

import (
	"errors"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateAccount creates a new account
func (r *Repository) CreateAccount(account *models.Account) error {
	return r.db.Create(account).Error
}

// GetAccountByEmail finds an account by email
func (r *Repository) GetAccountByEmail(email string) (*models.Account, error) {
	var account models.Account
	err := r.db.
		Preload("AccountType").
		Preload("ProfessionalType").
		Preload("Institution").
		Where("email = ?", email).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// GetAccountByID finds an account by ID
func (r *Repository) GetAccountByID(id uuid.UUID) (*models.Account, error) {
	var account models.Account
	err := r.db.
		Preload("AccountType").
		Preload("ProfessionalType").
		Preload("Institution").
		Where("id = ?", id).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// UpdateAccount updates an existing account
func (r *Repository) UpdateAccount(account *models.Account) error {
	return r.db.Save(account).Error
}

// DeleteAccount soft deletes an account
func (r *Repository) DeleteAccount(id uuid.UUID) error {
	return r.db.Delete(&models.Account{}, "id = ?", id).Error
}

// GetAccountTypeBySlug gets account type by slug
func (r *Repository) GetAccountTypeBySlug(slug string) (*models.AccountType, error) {
	var accountType models.AccountType
	err := r.db.Where("slug = ?", slug).First(&accountType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &accountType, nil
}