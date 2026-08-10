package postgres

import (
	"context"
	"errors"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
	"gorm.io/gorm"
)

// ============================================================
// POSTGRES REPOSITORY - Implements domain.Repository
// ============================================================

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	model := ToModelAccount(account)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainAccount(&model), nil
}

func (r *PostgresRepository) GetAccountByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainAccount(&model), nil
}

func (r *PostgresRepository) GetAccountByPhone(ctx context.Context, phone string) (*domain.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainAccount(&model), nil
}

func (r *PostgresRepository) UpdateAccount(ctx context.Context, account *domain.Account) error {
	model := ToModelAccount(account)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) DeleteAccount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&AccountModel{}, "id = ?", id).Error
}

func (r *PostgresRepository) AccountExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AccountModel{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) AccountExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AccountModel{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

// ============================================================
// ACCOUNT TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetAccountTypeByID(ctx context.Context, id string) (*domain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainAccountTypeEntity(&model), nil
}

func (r *PostgresRepository) GetAccountTypeBySlug(ctx context.Context, slug string) (*domain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainAccountTypeEntity(&model), nil
}

func (r *PostgresRepository) CreateAccountType(ctx context.Context, accountType *domain.AccountType) error {
	model := ToModelAccountTypeEntity(accountType)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) UpdateAccountType(ctx context.Context, accountType *domain.AccountType) error {
	model := ToModelAccountTypeEntity(accountType)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) ListAccountTypes(ctx context.Context) ([]*domain.AccountType, error) {
	var models []AccountTypeModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	accountTypes := make([]*domain.AccountType, len(models))
	for i, m := range models {
		accountTypes[i] = ToDomainAccountTypeEntity(&m)
	}
	return accountTypes, nil
}

// ============================================================
// PROFESSIONAL TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetProfessionalTypeByID(ctx context.Context, id string) (*domain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainProfessionalTypeEntity(&model), nil
}

func (r *PostgresRepository) GetProfessionalTypeBySlug(ctx context.Context, slug string) (*domain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainProfessionalTypeEntity(&model), nil
}

func (r *PostgresRepository) ListProfessionalTypes(ctx context.Context) ([]*domain.ProfessionalType, error) {
	var models []ProfessionalTypeModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	profTypes := make([]*domain.ProfessionalType, len(models))
	for i, m := range models {
		profTypes[i] = ToDomainProfessionalTypeEntity(&m)
	}
	return profTypes, nil
}

// ============================================================
// INSTITUTION OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateInstitution(ctx context.Context, institution *domain.Institution) error {
	model := ToModelInstitution(institution)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) GetInstitutionByID(ctx context.Context, id string) (*domain.Institution, error) {
	var model InstitutionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainInstitution(&model), nil
}

func (r *PostgresRepository) GetInstitutionBySlug(ctx context.Context, slug string) (*domain.Institution, error) {
	var model InstitutionModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainInstitution(&model), nil
}

func (r *PostgresRepository) UpdateInstitution(ctx context.Context, institution *domain.Institution) error {
	model := ToModelInstitution(institution)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) DeleteInstitution(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&InstitutionModel{}, "id = ?", id).Error
}

// ============================================================
// INSTITUTION TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetInstitutionTypeByID(ctx context.Context, id string) (*domain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainInstitutionTypeEntity(&model), nil
}

func (r *PostgresRepository) GetInstitutionTypeBySlug(ctx context.Context, slug string) (*domain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainInstitutionTypeEntity(&model), nil
}

func (r *PostgresRepository) ListInstitutionTypes(ctx context.Context) ([]*domain.InstitutionType, error) {
	var models []InstitutionTypeModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	instTypes := make([]*domain.InstitutionType, len(models))
	for i, m := range models {
		instTypes[i] = ToDomainInstitutionTypeEntity(&m)
	}
	return instTypes, nil
}

// ============================================================
// TEAM MEMBER OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateTeamMember(ctx context.Context, member *domain.TeamMember) error {
	model := ToModelTeamMember(member)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) GetTeamMemberByID(ctx context.Context, id string) (*domain.TeamMember, error) {
	var model TeamMemberModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainTeamMember(&model), nil
}

func (r *PostgresRepository) GetTeamMembersByAccount(ctx context.Context, accountID string) ([]*domain.TeamMember, error) {
	var models []TeamMemberModel
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	members := make([]*domain.TeamMember, len(models))
	for i, m := range models {
		members[i] = ToDomainTeamMember(&m)
	}
	return members, nil
}

func (r *PostgresRepository) GetTeamMemberByAccountAndMember(ctx context.Context, accountID, memberID string) (*domain.TeamMember, error) {
	var model TeamMemberModel
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND member_id = ?", accountID, memberID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainTeamMember(&model), nil
}

func (r *PostgresRepository) UpdateTeamMember(ctx context.Context, member *domain.TeamMember) error {
	model := ToModelTeamMember(member)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) DeleteTeamMember(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TeamMemberModel{}, "id = ?", id).Error
}