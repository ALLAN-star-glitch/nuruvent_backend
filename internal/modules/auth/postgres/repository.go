package postgres

import (
	"errors"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
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

func (r *PostgresRepository) GetAccountTypeBySlug(slug string) (*authdomain.AccountType, error) {
    var model AccountTypeModel
    err := r.db.Where("slug = ?", slug).First(&model).Error
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
    return r.db.Model(&RefreshTokenModel{}).
        Where("account_id = ?", accountID).
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

func (r *PostgresRepository) GetInstitutionTypeBySlug(slug string) (*authdomain.InstitutionType, error) {
    var model InstitutionTypeModel
    err := r.db.Where("slug = ?", slug).First(&model).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return ToauthdomainInstitutionType(&model), nil
}

// ============================================================
// TEAM MEMBER OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateTeamMember(member *authdomain.TeamMember) error {
    model := ToTeamMemberModel(member)
    return r.db.Create(model).Error
}

func (r *PostgresRepository) GetTeamMemberByAccountAndInstitution(accountID, institutionID string) (*authdomain.TeamMember, error) {
    var model TeamMemberModel
    err := r.db.
        Where("account_id = ? AND institution_id = ?", accountID, institutionID).
        First(&model).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, err
    }
    return ToauthdomainTeamMember(&model), nil
}