// internal/modules/auth/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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
// USER OPERATIONS (formerly Account)
// ============================================================

func (r *PostgresRepository) UserExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) UserExistsByPhone(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("phone = ?", phone).Count(&count).Error
	return count > 0, err
}

func (r *PostgresRepository) GetUserByEmail(email string) (*authdomain.User, error) {
	var model UserModel
	err := r.db.Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainUser(&model), nil
}

func (r *PostgresRepository) GetUserByPhone(phone string) (*authdomain.User, error) {
	var model UserModel
	err := r.db.Where("phone = ?", phone).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainUser(&model), nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context,id string) (*authdomain.User, error) {
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

func (r *PostgresRepository) CreateUser(user *authdomain.User) error {
	model := ToUserModel(user)
	return r.db.Create(model).Error
}

func (r *PostgresRepository) UpdateUser(user *authdomain.User) error {
	model := ToUserModel(user)
	return r.db.Save(model).Error
}

func (r *PostgresRepository) UpdateUserInstitutionID(userID string, institutionID *string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	var institutionUUID *uuid.UUID
	if institutionID != nil {
		id, err := uuid.Parse(*institutionID)
		if err != nil {
			return err
		}
		institutionUUID = &id
	}

	return r.db.Model(&UserModel{}).
		Where("id = ?", userUUID).
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
	return ToAuthDomainAccountType(&model), nil
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
	return ToAuthDomainAccountType(&model), nil
}

func (r *PostgresRepository) GetAccountTypeByName(name string) (*authdomain.AccountType, error) {
	var model AccountTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainAccountType(&model), nil
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
	return ToAuthDomainRefreshToken(&model), nil
}

func (r *PostgresRepository) RevokeRefreshToken(token string) error {
	return r.db.Model(&RefreshTokenModel{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

func (r *PostgresRepository) RevokeAllUserRefreshTokens(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return r.db.Model(&RefreshTokenModel{}).
		Where("user_id = ?", userUUID).
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
	return ToAuthDomainInstitution(&model), nil
}

func (r *PostgresRepository) GetInstitutionByUserID(userID string) (*authdomain.Institution, error) {
	var model InstitutionModel
	err := r.db.Where("user_id = ?", userID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainInstitution(&model), nil
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
	return ToAuthDomainInstitutionType(&model), nil
}

func (r *PostgresRepository) GetInstitutionTypeByName(name string) (*authdomain.InstitutionType, error) {
	var model InstitutionTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainInstitutionType(&model), nil
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
		institutions[i] = ToAuthDomainInstitution(&model)
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
	return ToAuthDomainTeamMember(&model), nil
}

func (r *PostgresRepository) GetTeamMemberByMemberAndInstitution(memberID, institutionID string) (*authdomain.TeamMember, error) {
	var model TeamMemberModel
	err := r.db.
		Where("member_id = ? AND institution_id = ? AND is_active = ?", memberID, institutionID, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainTeamMember(&model), nil
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
		members[i] = ToAuthDomainTeamMember(&model)
	}
	return members, nil
}

func (r *PostgresRepository) GetTeamMembersByMember(memberID string) ([]*authdomain.TeamMember, error) {
	var models []TeamMemberModel
	err := r.db.
		Where("member_id = ? AND is_active = ?", memberID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	members := make([]*authdomain.TeamMember, len(models))
	for i, model := range models {
		members[i] = ToAuthDomainTeamMember(&model)
	}
	return members, nil
}

func (r *PostgresRepository) GetTeamMembersByTeamType(teamTypeID string) ([]*authdomain.TeamMember, error) {
	var models []TeamMemberModel
	err := r.db.
		Where("team_type_id = ? AND is_active = ?", teamTypeID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	members := make([]*authdomain.TeamMember, len(models))
	for i, model := range models {
		members[i] = ToAuthDomainTeamMember(&model)
	}
	return members, nil
}

func (r *PostgresRepository) IsMemberOfInstitution(ctx context.Context, memberID, institutionID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Where("member_id = ? AND institution_id = ? AND is_active = ?",
			memberID, institutionID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ============================================================
// TEAM TYPE OPERATIONS (NEW)
// ============================================================

func (r *PostgresRepository) GetTeamTypeByID(id string) (*authdomain.TeamType, error) {
	var model TeamTypeModel
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainTeamType(&model), nil
}

func (r *PostgresRepository) GetTeamTypeBySlug(slug string) (*authdomain.TeamType, error) {
	var model TeamTypeModel
	err := r.db.Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainTeamType(&model), nil
}

func (r *PostgresRepository) GetTeamTypeByName(name string) (*authdomain.TeamType, error) {
	var model TeamTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainTeamType(&model), nil
}

func (r *PostgresRepository) CreateTeamType(teamType *authdomain.TeamType) error {
	model := ToTeamTypeModel(teamType)
	return r.db.Create(model).Error
}

func (r *PostgresRepository) UpdateTeamType(teamType *authdomain.TeamType) error {
	model := ToTeamTypeModel(teamType)
	return r.db.Save(model).Error
}

func (r *PostgresRepository) DeleteTeamType(id string) error {
	return r.db.Model(&TeamTypeModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *PostgresRepository) ListTeamTypes(ctx context.Context) ([]*authdomain.TeamType, error) {
	var models []TeamTypeModel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	teamTypes := make([]*authdomain.TeamType, len(models))
	for i, model := range models {
		teamTypes[i] = ToAuthDomainTeamType(&model)
	}
	return teamTypes, nil
}

// ============================================================
// INSTITUTION ADMIN CHECKS
// ============================================================

func (r *PostgresRepository) IsInstitutionAdmin(ctx context.Context, memberID, institutionID string) (bool, error) {
	// Check if user has admin role via team_members
	// Note: Role is managed by Casbin, but we keep this for backward compatibility
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Where("member_id = ? AND institution_id = ? AND is_active = ?",
			memberID, institutionID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ============================================================
// PLATFORM ADMIN CHECKS
// ============================================================

func (r *PostgresRepository) IsPlatformAdmin(ctx context.Context, userID string) (bool, error) {
	// Check if user has admin role in Casbin
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
	// Check if user has super_admin role in Casbin
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

// internal/modules/auth/postgres/repository.go

// ============================================================
// PROFESSIONAL TYPE OPERATIONS (NEW)
// ============================================================

func (r *PostgresRepository) GetProfessionalTypeByID(id string) (*authdomain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainProfessionalType(&model), nil
}

func (r *PostgresRepository) GetProfessionalTypeBySlug(slug string) (*authdomain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.Where("slug = ? AND is_active = ?", slug, true).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainProfessionalType(&model), nil
}

func (r *PostgresRepository) GetProfessionalTypeByName(name string) (*authdomain.ProfessionalType, error) {
	var model ProfessionalTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&model).Error
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


// RevokeAllRefreshTokensForUser revokes all refresh tokens for a user
func (r *PostgresRepository) RevokeAllRefreshTokensForUser(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return r.db.Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked = ?", userUUID, false).
		Updates(map[string]interface{}{
			"revoked":    true,
			"updated_at": time.Now(),
		}).Error
}

// UpdateRefreshTokenContext updates the user agent and IP for a refresh token
func (r *PostgresRepository) UpdateRefreshTokenContext(token, userAgent, ipAddress string) error {
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

	return r.db.Model(&RefreshTokenModel{}).
		Where("token = ?", token).
		Updates(updates).Error
}

func (r *PostgresRepository) GetInstitutionByEmail(email string) (*authdomain.Institution, error) {
	var model InstitutionModel
	err := r.db.Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToAuthDomainInstitution(&model), nil
}

func (r *PostgresRepository) DeleteUser(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return r.db.Model(&UserModel{}).
		Where("id = ?", userUUID).
		Update("is_active", false).Error
}

func (r *PostgresRepository) ReactivateUser(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return r.db.Model(&UserModel{}).
		Where("id = ?", userUUID).
		Update("is_active", true).Error
}	

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

// AccountExists implements authdomain.Repository.
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

// CreateAccount implements authdomain.Repository.
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

// GetAccountByID implements authdomain.Repository.
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

// GetAccountByEmail implements authdomain.Repository.
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

// GetAccountBySlug implements authdomain.Repository.
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

// GetAccountsByUserID implements authdomain.Repository.
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

// UpdateAccount implements authdomain.Repository.
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

// DeleteAccount implements authdomain.Repository.
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

// CreateAccountMember implements authdomain.Repository.
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

// GetAccountMemberByID implements authdomain.Repository.
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

// GetAccountMemberByAccountAndUser implements authdomain.Repository.
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

// GetAccountMembersByAccount implements authdomain.Repository.
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

// GetAccountMembersByUser implements authdomain.Repository.
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

// UpdateAccountMember implements authdomain.Repository.
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

// DeleteAccountMember implements authdomain.Repository.
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

// IsAccountMember implements authdomain.Repository.
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

// IsAccountAdmin implements authdomain.Repository.
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

// IsAccountTrainer implements authdomain.Repository.
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

// CountAccountMembers implements authdomain.Repository.
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