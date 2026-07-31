package bizrepo

import (
	"errors"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessMemberRepository struct {
	db *gorm.DB
}

func NewBusinessMemberRepository(db *gorm.DB) *BusinessMemberRepository {
	return &BusinessMemberRepository{db: db}
}

// ================================================
// CRUD OPERATIONS
// ================================================

// Create creates a new business member
func (r *BusinessMemberRepository) Create(member *models.BusinessMember) error {
	return r.db.Create(member).Error
}

// GetByID gets a business member by ID
func (r *BusinessMemberRepository) GetByID(id uuid.UUID) (*models.BusinessMember, error) {
	var member models.BusinessMember
	err := r.db.
		Preload("User").
		Preload("Business").
		Where("id = ? AND is_active = ?", id, true).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// GetByUserAndBusiness gets a business member by user and business
func (r *BusinessMemberRepository) GetByUserAndBusiness(userID, businessID uuid.UUID) (*models.BusinessMember, error) {
	var member models.BusinessMember
	err := r.db.
		Preload("User").
		Preload("Business").
		Where("user_id = ? AND business_id = ? AND is_active = ?", userID, businessID, true).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// ================================================
// LIST OPERATIONS
// ================================================

// GetMembersByBusiness gets all members of a business
func (r *BusinessMemberRepository) GetMembersByBusiness(businessID uuid.UUID) ([]models.BusinessMember, error) {
	var members []models.BusinessMember
	err := r.db.
		Preload("User").
		Where("business_id = ? AND is_active = ?", businessID, true).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// GetMembersByBusinessWithRole gets all members of a business with a specific role
func (r *BusinessMemberRepository) GetMembersByBusinessWithRole(businessID uuid.UUID, role string) ([]models.BusinessMember, error) {
	var members []models.BusinessMember
	err := r.db.
		Preload("User").
		Where("business_id = ? AND role = ? AND is_active = ?", businessID, role, true).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// GetBusinessesByUser gets all businesses a user belongs to
func (r *BusinessMemberRepository) GetBusinessesByUser(userID uuid.UUID) ([]models.BusinessMember, error) {
	var members []models.BusinessMember
	err := r.db.
		Preload("Business").
		Preload("Business.BusinessType").
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

// ================================================
// UPDATE OPERATIONS
// ================================================

// Update updates a business member
func (r *BusinessMemberRepository) Update(member *models.BusinessMember) error {
	return r.db.Save(member).Error
}

// UpdateRole updates a member's role
func (r *BusinessMemberRepository) UpdateRole(userID, businessID uuid.UUID, role string) error {
	return r.db.Model(&models.BusinessMember{}).
		Where("user_id = ? AND business_id = ?", userID, businessID).
		Update("role", role).Error
}

// Deactivate deactivates a business member
func (r *BusinessMemberRepository) Deactivate(userID, businessID uuid.UUID) error {
	return r.db.Model(&models.BusinessMember{}).
		Where("user_id = ? AND business_id = ?", userID, businessID).
		Update("is_active", false).Error
}

// Activate activates a business member
func (r *BusinessMemberRepository) Activate(userID, businessID uuid.UUID) error {
	return r.db.Model(&models.BusinessMember{}).
		Where("user_id = ? AND business_id = ?", userID, businessID).
		Update("is_active", true).Error
}

// ================================================
// DELETE OPERATIONS
// ================================================

// Delete soft deletes a business member
func (r *BusinessMemberRepository) Delete(userID, businessID uuid.UUID) error {
	return r.db.Where("user_id = ? AND business_id = ?", userID, businessID).
		Delete(&models.BusinessMember{}).Error
}

// DeleteByBusiness deletes all members of a business
func (r *BusinessMemberRepository) DeleteByBusiness(businessID uuid.UUID) error {
	return r.db.Where("business_id = ?", businessID).
		Delete(&models.BusinessMember{}).Error
}

// ================================================
// CHECK OPERATIONS
// ================================================

// Exists checks if a business member exists
func (r *BusinessMemberRepository) Exists(userID, businessID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.BusinessMember{}).
		Where("user_id = ? AND business_id = ? AND is_active = ?", userID, businessID, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasRole checks if a user has a specific role in a business
func (r *BusinessMemberRepository) HasRole(userID, businessID uuid.UUID, role string) (bool, error) {
	var count int64
	err := r.db.Model(&models.BusinessMember{}).
		Where("user_id = ? AND business_id = ? AND role = ? AND is_active = ?",
			userID, businessID, role, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsHost checks if a user is a host of a business
func (r *BusinessMemberRepository) IsHost(userID, businessID uuid.UUID) (bool, error) {
	return r.HasRole(userID, businessID, "host")
}

// IsEventManager checks if a user is an event manager of a business
func (r *BusinessMemberRepository) IsEventManager(userID, businessID uuid.UUID) (bool, error) {
	return r.HasRole(userID, businessID, "event_manager")
}

// IsMember checks if a user is a member of a business
func (r *BusinessMemberRepository) IsMember(userID, businessID uuid.UUID) (bool, error) {
	return r.HasRole(userID, businessID, "member")
}

// ================================================
// COUNT OPERATIONS
// ================================================

// CountMembers counts all members of a business
func (r *BusinessMemberRepository) CountMembers(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.BusinessMember{}).
		Where("business_id = ? AND is_active = ?", businessID, true).
		Count(&count).Error
	return count, err
}

// CountByRole counts members of a business by role
func (r *BusinessMemberRepository) CountByRole(businessID uuid.UUID, role string) (int64, error) {
	var count int64
	err := r.db.Model(&models.BusinessMember{}).
		Where("business_id = ? AND role = ? AND is_active = ?", businessID, role, true).
		Count(&count).Error
	return count, err
}

// CountAdmins counts business admins (hosts)
func (r *BusinessMemberRepository) CountAdmins(businessID uuid.UUID) (int64, error) {
	return r.CountByRole(businessID, "host")
}

// CountManagers counts all managers (event managers)
func (r *BusinessMemberRepository) CountManagers(businessID uuid.UUID) (int64, error) {
	return r.CountByRole(businessID, "event_manager")
}