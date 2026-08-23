// internal/modules/auth/postgres/repository.go

package postgres

import (
	"context"
	"errors"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================
// POSTGRES REPOSITORY - Implements authdomain.Repository
// ============================================================

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) authdomain.Repository {
	return &PostgresRepository{db: db}
}

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

func (r *PostgresRepository) AccountExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&AccountModel{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) AccountExistsByPhone(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&AccountModel{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) GetAccountByEmail(email string) (*authdomain.Account, error) {
	var model AccountModel
	err := r.db.Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainAccount(&model), nil
}

func (r *PostgresRepository) GetAccountByPhone(phone string) (*authdomain.Account, error) {
	var model AccountModel
	err := r.db.Where("phone = ?", phone).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainAccount(&model), nil
}

func (r *PostgresRepository) GetAccountByID(id string) (*authdomain.Account, error) {
	var model AccountModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainAccount(&model), nil
}

func (r *PostgresRepository) CreateAccount(account *authdomain.Account) error {
	model := ToAccountModel(account)
	return r.db.Create(model).Error
}

func (r *PostgresRepository) UpdateAccount(account *authdomain.Account) error {
	model := ToAccountModel(account)
	return r.db.Save(model).Error
}

// UpdateAccountInstitutionID updates the institution_id for an account
func (r *PostgresRepository) UpdateAccountInstitutionID(accountID string, institutionID string) error {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return err
	}
	institutionUUID, err := uuid.Parse(institutionID)
	if err != nil {
		return err
	}
	return r.db.Model(&AccountModel{}).
		Where("id = ?", accountUUID).
		Update("institution_id", institutionUUID).Error
}

func (r *PostgresRepository) GetAccountTypeByID(id string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainAccountType(&model), nil
}

func (r *PostgresRepository) GetAccountTypeBySlug(slug string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainAccountType(&model), nil
}

// ============================================================
// REFRESH TOKEN OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateRefreshToken(token *authdomain.RefreshToken) error {
	model := ToRefreshTokenModel(token)
	return r.db.Create(model).Error
}

func (r *PostgresRepository) GetRefreshTokenByToken(token string) (*authdomain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainRefreshToken(&model), nil
}

func (r *PostgresRepository) RevokeRefreshToken(token string) error {
	return r.db.Model(&RefreshTokenModel{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

func (r *PostgresRepository) RevokeAllAccountRefreshTokens(accountID string) error {
	accountUUID, err := uuid.Parse(accountID)
	if err != nil {
		return err
	}
	return r.db.Model(&RefreshTokenModel{}).
		Where("account_id = ?", accountUUID).
		Update("revoked", true).Error
}

// ============================================================
// INSTITUTION OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateInstitution(institution *authdomain.Institution) error {
	model := ToInstitutionModel(institution)
	return r.db.Create(model).Error
}

func (r *PostgresRepository) GetInstitutionByID(id string) (*authdomain.Institution, error) {
	var model InstitutionModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainInstitution(&model), nil
}

func (r *PostgresRepository) GetInstitutionByAccountID(accountID string) (*authdomain.Institution, error) {
	var model InstitutionModel
	err := r.db.Where("account_id = ?", accountID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainInstitution(&model), nil
}

func (r *PostgresRepository) GetInstitutionTypeBySlug(slug string) (*authdomain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainInstitutionType(&model), nil
}

func (r *PostgresRepository) UpdateInstitution(institution *authdomain.Institution) error {
	model := ToInstitutionModel(institution)
	return r.db.Save(model).Error
}

func (r *PostgresRepository) InstitutionExists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&InstitutionModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) GetInstitutionsByType(institutionTypeID string) ([]*authdomain.Institution, error) {
	var models []InstitutionModel
	err := r.db.Where("institution_type_id = ? AND is_active = ?", institutionTypeID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	institutions := make([]*authdomain.Institution, len(models))
	for i, model := range models {
		institutions[i] = ToauthdomainInstitution(&model)
	}
	return institutions, nil
}

// ============================================================
// TEAM MEMBER OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateTeamMember(member *authdomain.TeamMember) error {
	model := ToTeamMemberModel(member)
	return r.db.Create(model).Error
}

func (r *PostgresRepository) GetTeamMemberByID(id string) (*authdomain.TeamMember, error) {
	var model TeamMemberModel
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainTeamMember(&model), nil
}

func (r *PostgresRepository) GetTeamMemberByAccountAndInstitution(accountID, institutionID string) (*authdomain.TeamMember, error) {
	var model TeamMemberModel
	err := r.db.
		Where("account_id = ? AND institution_id = ? AND is_active = ?", accountID, institutionID, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainTeamMember(&model), nil
}

func (r *PostgresRepository) UpdateTeamMember(member *authdomain.TeamMember) error {
	model := ToTeamMemberModel(member)
	return r.db.Save(model).Error
}

func (r *PostgresRepository) DeleteTeamMember(id string) error {
	return r.db.Model(&TeamMemberModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *PostgresRepository) GetTeamMembersByInstitution(institutionID string) ([]*authdomain.TeamMember, error) {
	var models []TeamMemberModel
	err := r.db.
		Where("institution_id = ? AND is_active = ?", institutionID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	members := make([]*authdomain.TeamMember, len(models))
	for i, model := range models {
		members[i] = ToauthdomainTeamMember(&model)
	}
	return members, nil
}

func (r *PostgresRepository) GetTeamMembersByAccount(accountID string) ([]*authdomain.TeamMember, error) {
	var models []TeamMemberModel
	err := r.db.
		Where("account_id = ? AND is_active = ?", accountID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	members := make([]*authdomain.TeamMember, len(models))
	for i, model := range models {
		members[i] = ToauthdomainTeamMember(&model)
	}
	return members, nil
}

func (r *PostgresRepository) IsMemberOfInstitution(ctx context.Context, accountID, institutionID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Where("account_id = ? AND institution_id = ? AND is_active = ?",
			accountID, institutionID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresRepository) GetMemberRole(ctx context.Context, accountID, institutionID string) (string, error) {
	var model TeamMemberModel
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND institution_id = ? AND is_active = ?",
			accountID, institutionID, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return string(model.Role), nil
}

// ============================================================
// INSTITUTION ADMIN CHECKS
// ============================================================

func (r *PostgresRepository) IsInstitutionAdmin(ctx context.Context, accountID, institutionID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Where("account_id = ? AND institution_id = ? AND role = ? AND is_active = ?",
			accountID, institutionID, TeamMemberRoleAdmin, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ============================================================
// PLATFORM ADMIN CHECKS
// ============================================================

// IsPlatformAdmin checks if a user is a platform admin
func (r *PostgresRepository) IsPlatformAdmin(ctx context.Context, accountID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ? AND role = ? AND is_active = ?", accountID, "admin", true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsSuperAdmin checks if a user is a super admin
func (r *PostgresRepository) IsSuperAdmin(ctx context.Context, accountID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ? AND role = ? AND is_active = ?", accountID, "super_admin", true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// internal/modules/auth/postgres/repository.go

// Add after GetAccountTypeBySlug method

// GetAccountTypeByName gets account type by NAME (with underscores) - for internal lookups
func (r *PostgresRepository) GetAccountTypeByName(name string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainAccountType(&model), nil
}

// Add after GetInstitutionTypeBySlug method

// GetInstitutionTypeByName gets institution type by NAME (with underscores) - for internal lookups
func (r *PostgresRepository) GetInstitutionTypeByName(name string) (*authdomain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToauthdomainInstitutionType(&model), nil
}