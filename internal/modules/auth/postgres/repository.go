// internal/modules/auth/postgres/repository.go

package postgres

import (
	"context"
	"errors"
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

func (r *PostgresRepository) GetUserByID(id string) (*authdomain.User, error) {
	var model UserModel
	err := r.db.Where("id = ?", id).First(&model).Error
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