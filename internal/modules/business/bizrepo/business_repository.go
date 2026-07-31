package bizrepo

import (
	"errors"
	"log"
	"strings"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessRepository struct {
	db *gorm.DB
}

func NewBusinessRepository(db *gorm.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}

// ================================================
// GENERIC QUERY HELPER
// ================================================

// GetInstances fetches a paginated list of any model type T
func GetInstances[T any](
	db *gorm.DB,
	condition string,
	preloads []string,
	limit, offset int,
	order string,
	args ...any,
) ([]T, int64, error) {
	var results []T
	var total int64

	// Base query using zero-value of T to infer table/model
	var model T
	q := db.Model(&model)

	if condition != "" {
		q = q.Where(condition, args...)
	}

	// 1. Count total records
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. Apply dynamic preloads
	for _, p := range preloads {
		if p != "" {
			q = q.Preload(p)
		}
	}

	// 3. Apply pagination and ordering
	err := q.Order(order).
		Limit(limit).
		Offset(offset).
		Find(&results).Error

	return results, total, err
}

// ================================================
// CRUD OPERATIONS
// ================================================

// Create creates a new business
func (r *BusinessRepository) Create(business *models.Business) error {

	// Generate slug if not provided
	if business.Slug == "" {
		business.Slug = r.generateSlug(business.Name)
	}

	if err := r.db.Create(business).Error; err != nil {
		log.Printf("Repository Create error: %v", err)
		return err
	}

	return nil
}

// GetByID gets a business by ID with relations
func (r *BusinessRepository) GetByID(id uuid.UUID) (*models.Business, error) {
	var business models.Business
	err := r.db.
		Preload("BusinessType").
		Preload("Members").
		Preload("Members.User").
		Where("id = ? AND is_active = ?", id, true).
		First(&business).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &business, nil
}

// GetBySlug gets a business by slug
func (r *BusinessRepository) GetBySlug(slug string) (*models.Business, error) {
	var business models.Business
	err := r.db.
		Preload("BusinessType").
		Where("slug = ? AND is_active = ?", slug, true).
		First(&business).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &business, nil
}

// GetByEmail gets a business by email
func (r *BusinessRepository) GetByEmail(email string) (*models.Business, error) {
	var business models.Business
	err := r.db.
		Preload("BusinessType").
		Where("email = ? AND is_active = ?", email, true).
		First(&business).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &business, nil
}

// Update updates a business
func (r *BusinessRepository) Update(business *models.Business) error {
	return r.db.Save(business).Error
}

// Delete soft deletes a business
func (r *BusinessRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Business{}, id).Error
}

// ================================================
// QUERY OPERATIONS
// ================================================

// Search searches businesses by name or description
func (r *BusinessRepository) Search(query string, limit, offset int) ([]models.Business, int64, error) {
	condition := "is_active = ?"
	args := []any{true}

	if query != "" {
		condition += " AND (name ILIKE ? OR description ILIKE ?)"
		args =   append(args, "%"+query+"%", "%"+query+"%")
	}

	preloads := []string{
		"BusinessType",
		"Members",
		"Members.User",
	}

	return GetInstances[models.Business](
		r.db,
		condition,
	
		preloads,
		limit,
		offset,
		"name ASC",
		args...,
	)
}

// GetBusinessesByUser gets all businesses a user belongs to
func (r *BusinessRepository) GetBusinessesByUser(userID uuid.UUID) ([]models.Business, error) {
	var businesses []models.Business

	// Subquery to get business IDs where user is a member
	subQuery := r.db.Table("business_members").
		Select("business_id").
		Where("user_id = ? AND is_active = ?", userID, true)

	err := r.db.
		Preload("BusinessType").
		Preload("Members").
		Preload("Members.User").
		Where("id IN (?) AND is_active = ?", subQuery, true).
		Order("name ASC").
		Find(&businesses).Error

	return businesses, err
}

// GetAllBusinesses gets all active businesses with pagination
func (r *BusinessRepository) GetAllBusinesses(limit, offset int) ([]models.Business, int64, error) {
	preloads := []string{
		"BusinessType",
		"Members",
		"Members.User",
	}

	return GetInstances[models.Business](
		r.db,
		"is_active = ?",
		preloads,
		limit,
		offset,
		"created_at DESC",
		true,
	)
}

// ================================================
// BUSINESS TYPE OPERATIONS
// ================================================

// GetAllBusinessTypes gets all business types
func (r *BusinessRepository) GetAllBusinessTypes() ([]models.BusinessType, error) {
	types, _, err := GetInstances[models.BusinessType](
		r.db,
		"is_active = ?",
		nil,
		-1, // -1 in GORM disables Limit
		-1, // -1 in GORM disables Offset
		"sort_order ASC, display_name ASC",
		true,
	)
	return types, err
}

// GetBusinessTypeByID gets a business type by ID
func (r *BusinessRepository) GetBusinessTypeByID(id uuid.UUID) (*models.BusinessType, error) {
	var businessType models.BusinessType
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&businessType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &businessType, nil
}

// GetBusinessTypeByName gets a business type by name
func (r *BusinessRepository) GetBusinessTypeByName(name string) (*models.BusinessType, error) {
	var businessType models.BusinessType
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&businessType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &businessType, nil
}

// GetBusinessTypeByDisplayName gets a business type by display name
func (r *BusinessRepository) GetBusinessTypeByDisplayName(displayName string) (*models.BusinessType, error) {
	var businessType models.BusinessType
	err := r.db.Where("display_name = ? AND is_active = ?", displayName, true).First(&businessType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &businessType, nil
}

// ================================================
// STATS OPERATIONS
// ================================================

// GetBusinessStats gets statistics for a business
func (r *BusinessRepository) GetBusinessStats(businessID uuid.UUID) (map[string]any, error) {
	var eventCount int64
	var attendeeCount int64
	var revenue float64
	var memberCount int64

	// Count events
	if err := r.db.Model(&models.Event{}).
		Where("business_id = ? AND is_active = ?", businessID, true).
		Count(&eventCount).Error; err != nil {
		return nil, err
	}

	// Count attendees (through events)
	if err := r.db.Table("attendees").
		Joins("JOIN events ON events.id = attendees.event_id").
		Where("events.business_id = ? AND events.is_active = ?", businessID, true).
		Count(&attendeeCount).Error; err != nil {
		return nil, err
	}

	// Calculate revenue (from payments)
	if err := r.db.Model(&models.Payment{}).
		Joins("JOIN events ON events.id = payments.event_id").
		Where("events.business_id = ? AND payments.status = ?", businessID, "completed").
		Select("COALESCE(SUM(payments.amount), 0)").
		Scan(&revenue).Error; err != nil {
		return nil, err
	}

	// Count members
	if err := r.db.Model(&models.BusinessMember{}).
		Where("business_id = ? AND is_active = ?", businessID, true).
		Count(&memberCount).Error; err != nil {
		return nil, err
	}

	stats := map[string]any{
		"event_count":    eventCount,
		"attendee_count": attendeeCount,
		"revenue":        revenue,
		"member_count":   memberCount,
	}

	return stats, nil
}

// ================================================
// HELPER METHODS
// ================================================

// generateSlug generates a unique slug from a name
func (r *BusinessRepository) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "&", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Check if slug exists, if so, append a random string
	var count int64
	r.db.Model(&models.Business{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	return slug
}

// ExistsByEmail checks if a business exists with the given email
func (r *BusinessRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Business{}).
		Where("email = ? AND is_active = ?", email, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByName checks if a business exists with the given name
func (r *BusinessRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Business{}).
		Where("name = ? AND is_active = ?", name, true).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}