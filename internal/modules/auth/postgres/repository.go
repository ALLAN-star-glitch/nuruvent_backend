// internal/modules/auth/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"gorm.io/gorm"
)

// ============================================================
// POSTGRES REPOSITORY - Implements authdomain.Repository
// ============================================================

type PostgresRepository struct {
	db *gorm.DB
}

func (r *PostgresRepository) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Pass the transaction DB handle into context so subsequent repo calls use it
		txCtx := context.WithValue(ctx, "tx_db", tx)
		return fn(txCtx)
	})
}

func NewPostgresRepository(db *gorm.DB) authdomain.Repository {
	return &PostgresRepository{db: db}
}

// ============================================================
// USER OPERATIONS
// ============================================================

func (r *PostgresRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) UserExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*authdomain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainUser(&model), nil
}

func (r *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*authdomain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainUser(&model), nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*authdomain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainUser(&model), nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *authdomain.User) error {
	model := ToUserModel(user)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) UpdateUser(ctx context.Context, user *authdomain.User) error {
	model := ToUserModel(user)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&UserModel{}).
		Where("id = ?", userID).
		Update("is_active", false).Error
}

func (r *PostgresRepository) ReactivateUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&UserModel{}).
		Where("id = ?", userID).
		Update("is_active", true).Error
}

// ============================================================
// ACCOUNT TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetAccountTypeByID(ctx context.Context, id string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainAccountType(&model), nil
}

func (r *PostgresRepository) GetAccountTypeBySlug(ctx context.Context, slug string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.WithContext(ctx).Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainAccountType(&model), nil
}

func (r *PostgresRepository) GetAccountTypeByName(ctx context.Context, name string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainAccountType(&model), nil
}

// ============================================================
// INSTITUTION TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetInstitutionTypeByID(ctx context.Context, id string) (*authdomain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainInstitutionType(&model), nil
}

func (r *PostgresRepository) GetInstitutionTypeBySlug(ctx context.Context, slug string) (*authdomain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.WithContext(ctx).Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainInstitutionType(&model), nil
}

func (r *PostgresRepository) GetInstitutionTypeByName(ctx context.Context, name string) (*authdomain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainInstitutionType(&model), nil
}

func (r *PostgresRepository) ListInstitutionTypes(ctx context.Context) ([]*authdomain.InstitutionType, error) {
	var models []InstitutionTypeModel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	types := make([]*authdomain.InstitutionType, len(models))
	for i, model := range models {
		types[i] = ToAuthDomainInstitutionType(&model)
	}
	return types, nil
}

// ============================================================
// PROFESSIONAL TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetProfessionalTypeByID(ctx context.Context, id string) (*authdomain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainProfessionalType(&model), nil
}

func (r *PostgresRepository) GetProfessionalTypeBySlug(ctx context.Context, slug string) (*authdomain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.WithContext(ctx).Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainProfessionalType(&model), nil
}

func (r *PostgresRepository) GetProfessionalTypeByName(ctx context.Context, name string) (*authdomain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainProfessionalType(&model), nil
}

func (r *PostgresRepository) ListProfessionalTypes(ctx context.Context) ([]*authdomain.ProfessionalType, error) {
	var models []ProfessionalTypeModel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	types := make([]*authdomain.ProfessionalType, len(models))
	for i, model := range models {
		types[i] = ToAuthDomainProfessionalType(&model)
	}
	return types, nil
}

// ============================================================
// REFRESH TOKEN OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, token *authdomain.RefreshToken) error {
	model := ToRefreshTokenModel(token)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) GetRefreshTokenByToken(ctx context.Context, token string) (*authdomain.RefreshToken, error) {
	var model RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainRefreshToken(&model), nil
}

func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

func (r *PostgresRepository) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"revoked":    true,
			"updated_at": time.Now(),
		}).Error
}

func (r *PostgresRepository) UpdateRefreshTokenContext(ctx context.Context, token, userAgent, ipAddress string) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if userAgent != "" {
		updates["user_agent"] = userAgent
	}
	if ipAddress != "" {
		updates["ip_address"] = ipAddress
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Model(&RefreshTokenModel{}).
		Where("token = ?", token).
		Updates(updates).Error
}

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

func (r *PostgresRepository) AccountExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("account ID is required")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&AccountModel{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check account existence: %w", err)
	}

	return count > 0, nil
}

func (r *PostgresRepository) CreateAccount(ctx context.Context, account *authdomain.Account) error {
	if account == nil {
		return fmt.Errorf("account is nil")
	}

	model := AccountModel{}
	model.FromDomain(account)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetAccountByID(ctx context.Context, id string) (*authdomain.Account, error) {
	if id == "" {
		return nil, authdomain.ErrAccountNotFound
	}

	var model AccountModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authdomain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetAccountByEmail(ctx context.Context, email string) (*authdomain.Account, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	var model AccountModel
	if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authdomain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetAccountBySlug(ctx context.Context, slug string) (*authdomain.Account, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	var model AccountModel
	if err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authdomain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by slug: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]*authdomain.Account, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	var models []AccountModel
	if err := r.db.WithContext(ctx).
		Joins("INNER JOIN account_members ON account_members.account_id = accounts.id").
		Where("account_members.user_id = ? AND account_members.deleted_at IS NULL AND accounts.deleted_at IS NULL", userID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get accounts by user: %w", err)
	}

	accounts := make([]*authdomain.Account, len(models))
	for i, model := range models {
		accounts[i] = model.ToDomain()
	}
	return accounts, nil
}

func (r *PostgresRepository) UpdateAccount(ctx context.Context, account *authdomain.Account) error {
	if account == nil {
		return fmt.Errorf("account is nil")
	}
	if account.ID == "" {
		return authdomain.ErrAccountNotFound
	}

	model := AccountModel{}
	model.FromDomain(account)

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteAccount(ctx context.Context, id string) error {
	if id == "" {
		return authdomain.ErrAccountNotFound
	}

	// Soft delete
	result := r.db.WithContext(ctx).Model(&AccountModel{}).Where("id = ?", id).Update("deleted_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to delete account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return authdomain.ErrAccountNotFound
	}

	return nil
}

// ============================================================
// ACCOUNT MEMBER OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateAccountMember(ctx context.Context, member *authdomain.AccountMember) error {
	if member == nil {
		return fmt.Errorf("member is nil")
	}

	model := AccountMemberModel{}
	model.FromDomain(member)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create account member: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetAccountMemberByID(ctx context.Context, id string) (*authdomain.AccountMember, error) {
	if id == "" {
		return nil, fmt.Errorf("member ID is required")
	}

	var model AccountMemberModel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authdomain.ErrAccountMemberNotFound
		}
		return nil, fmt.Errorf("failed to get account member: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetAccountMemberByAccountAndUser(ctx context.Context, accountID string, userID string) (*authdomain.AccountMember, error) {
	if accountID == "" || userID == "" {
		return nil, fmt.Errorf("account ID and user ID are required")
	}

	var model AccountMemberModel
	if err := r.db.WithContext(ctx).
		Where("account_id = ? AND user_id = ? AND deleted_at IS NULL", accountID, userID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, authdomain.ErrAccountMemberNotFound
		}
		return nil, fmt.Errorf("failed to get account member: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *PostgresRepository) GetAccountMembersByAccount(ctx context.Context, accountID string) ([]*authdomain.AccountMember, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	var models []AccountMemberModel
	if err := r.db.WithContext(ctx).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get account members: %w", err)
	}

	members := make([]*authdomain.AccountMember, len(models))
	for i, model := range models {
		members[i] = model.ToDomain()
	}
	return members, nil
}

func (r *PostgresRepository) GetAccountMembersByUser(ctx context.Context, userID string) ([]*authdomain.AccountMember, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	var models []AccountMemberModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get account memberships: %w", err)
	}

	members := make([]*authdomain.AccountMember, len(models))
	for i, model := range models {
		members[i] = model.ToDomain()
	}
	return members, nil
}

func (r *PostgresRepository) UpdateAccountMember(ctx context.Context, member *authdomain.AccountMember) error {
	if member == nil {
		return fmt.Errorf("member is nil")
	}
	if member.ID == "" {
		return authdomain.ErrAccountMemberNotFound
	}

	model := AccountMemberModel{}
	model.FromDomain(member)

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return fmt.Errorf("failed to update account member: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteAccountMember(ctx context.Context, accountID string, userID string) error {
	if accountID == "" || userID == "" {
		return fmt.Errorf("account ID and user ID are required")
	}

	// Soft delete
	result := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
		Where("account_id = ? AND user_id = ?", accountID, userID).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to delete account member: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return authdomain.ErrAccountMemberNotFound
	}

	return nil
}

// ============================================================
// ACCOUNT MEMBER CHECK OPERATIONS
// ============================================================

func (r *PostgresRepository) IsAccountMember(ctx context.Context, accountID string, userID string) (bool, error) {
	if accountID == "" || userID == "" {
		return false, fmt.Errorf("account ID and user ID are required")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
		Where("account_id = ? AND user_id = ? AND deleted_at IS NULL", accountID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check account membership: %w", err)
	}

	return count > 0, nil
}

func (r *PostgresRepository) IsAccountAdmin(ctx context.Context, accountID string, userID string) (bool, error) {
	if accountID == "" || userID == "" {
		return false, fmt.Errorf("account ID and user ID are required")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
		Where("account_id = ? AND user_id = ? AND role = ? AND deleted_at IS NULL", accountID, userID, authdomain.RoleAccountAdmin.String()).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check account admin: %w", err)
	}

	return count > 0, nil
}

func (r *PostgresRepository) IsAccountTrainer(ctx context.Context, accountID string, userID string) (bool, error) {
	if accountID == "" || userID == "" {
		return false, fmt.Errorf("account ID and user ID are required")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
		Where("account_id = ? AND user_id = ? AND role = ? AND deleted_at IS NULL", accountID, userID, authdomain.RoleTrainer.String()).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check account trainer: %w", err)
	}

	return count > 0, nil
}

func (r *PostgresRepository) CountAccountMembers(ctx context.Context, accountID string) (int64, error) {
	if accountID == "" {
		return 0, fmt.Errorf("account ID is required")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count account members: %w", err)
	}

	return count, nil
}

// ============================================================
// PLATFORM ADMIN CHECKS
// ============================================================

func (r *PostgresRepository) IsPlatformAdmin(ctx context.Context, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("casbin_rule").
		Where("ptype = 'g' AND v0 = ? AND v1 = 'admin' AND v2 = 'platform'", userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresRepository) IsSuperAdmin(ctx context.Context, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("casbin_rule").
		Where("ptype = 'g' AND v0 = ? AND v1 = 'super_admin' AND v2 = 'platform'", userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}