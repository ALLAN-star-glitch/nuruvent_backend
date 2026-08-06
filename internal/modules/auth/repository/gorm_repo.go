// internal/modules/auth/repository/gorm_repo.go

package repository

import (
	"errors"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Ensure GormRepo implements domain.Repository
var _ domain.Repository = (*GormRepo)(nil)

type GormRepo struct {
	db *gorm.DB
}

func NewGormRepo(db *gorm.DB) domain.Repository {
	return &GormRepo{db: db}
}

// ================================================
// ACCOUNT OPERATIONS
// ================================================

func (r *GormRepo) AccountExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Account{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormRepo) AccountExistsByPhone(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Account{}).Where("phone = ?", phone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormRepo) GetAccountByEmail(email string) (*models.Account, error) {
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

func (r *GormRepo) GetAccountByPhone(phone string) (*models.Account, error) {
	var account models.Account
	err := r.db.
		Preload("AccountType").
		Preload("ProfessionalType").
		Preload("Institution").
		Where("phone = ?", phone).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *GormRepo) GetAccountByID(id uuid.UUID) (*models.Account, error) {
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

func (r *GormRepo) CreateAccount(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *GormRepo) UpdateAccount(account *models.Account) error {
	return r.db.Save(account).Error
}

func (r *GormRepo) GetAccountTypeBySlug(slug string) (*models.AccountType, error) {
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

// ================================================
// REFRESH TOKEN OPERATIONS
// ================================================

func (r *GormRepo) CreateRefreshToken(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *GormRepo) GetRefreshTokenByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.
		Preload("Account").
		Where("token = ?", token).
		First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *GormRepo) RevokeRefreshToken(token string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

func (r *GormRepo) RevokeAllAccountRefreshTokens(accountID string) error {
	id, err := uuid.Parse(accountID)
	if err != nil {
		return err
	}
	return r.db.Model(&models.RefreshToken{}).
		Where("account_id = ?", id).
		Update("revoked", true).Error
}

// ================================================
// INSTITUTION OPERATIONS
// ================================================

func (r *GormRepo) CreateInstitution(institution *models.Institution) error {
	return r.db.Create(institution).Error
}

func (r *GormRepo) GetInstitutionTypeBySlug(slug string) (*models.InstitutionType, error) {
	var institutionType models.InstitutionType
	err := r.db.Where("slug = ?", slug).First(&institutionType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &institutionType, nil
}

// ================================================
// TEAM MEMBER OPERATIONS
// ================================================

func (r *GormRepo) CreateTeamMember(member *models.TeamMember) error {
	return r.db.Create(member).Error
}