// internal/modules/media/mediarepo/repository.go

package mediarepo

import (
	"errors"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

// ================================================
// CRUD OPERATIONS
// ================================================

// Create creates a new media record
func (r *MediaRepository) Create(media *models.Media) error {
	return r.db.Create(media).Error
}

// GetByID gets a media record by ID
func (r *MediaRepository) GetByID(id uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := r.db.
		Preload("MediaType").
		Where("id = ? AND is_active = ?", id, true).
		First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

// Update updates a media record
func (r *MediaRepository) Update(media *models.Media) error {
	return r.db.Save(media).Error
}

// Delete soft deletes a media record
func (r *MediaRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Media{}, id).Error
}

// HardDelete permanently deletes a media record
func (r *MediaRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.Media{}, id).Error
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetByMediaType gets all media for a specific media type
func (r *MediaRepository) GetByMediaType(mediaTypeID uuid.UUID, limit, offset int) ([]models.Media, int64, error) {
	var media []models.Media
	var total int64

	query := r.db.Model(&models.Media{}).
		Where("media_type_id = ? AND is_active = ?", mediaTypeID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("MediaType").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&media).Error

	return media, total, err
}

// GetByEntity gets all media for a specific entity
func (r *MediaRepository) GetByEntity(entityID uuid.UUID, limit, offset int) ([]models.Media, int64, error) {
	var media []models.Media
	var total int64

	query := r.db.Model(&models.Media{}).
		Where("entity_id = ? AND is_active = ?", entityID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("MediaType").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&media).Error

	return media, total, err
}

// GetByEntityAndMediaType gets media by entity and media type
func (r *MediaRepository) GetByEntityAndMediaType(entityID, mediaTypeID uuid.UUID) ([]models.Media, error) {
	var media []models.Media
	err := r.db.
		Preload("MediaType").
		Where("entity_id = ? AND media_type_id = ? AND is_active = ?", entityID, mediaTypeID, true).
		Order("created_at DESC").
		Find(&media).Error
	return media, err
}

// GetByEntityAndMimeType gets media by entity and mime type
func (r *MediaRepository) GetByEntityAndMimeType(entityID uuid.UUID, mimeType string) ([]models.Media, error) {
	var media []models.Media
	err := r.db.
		Preload("MediaType").
		Where("entity_id = ? AND mime_type LIKE ? AND is_active = ?", entityID, mimeType+"%", true).
		Order("created_at DESC").
		Find(&media).Error
	return media, err
}

// GetByUploader gets all media uploaded by a specific user
func (r *MediaRepository) GetByUploader(userID uuid.UUID, limit, offset int) ([]models.Media, int64, error) {
	var media []models.Media
	var total int64

	query := r.db.Model(&models.Media{}).
		Where("uploaded_by = ? AND is_active = ?", userID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("MediaType").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&media).Error

	return media, total, err
}

// GetByBucket gets all media in a specific bucket
func (r *MediaRepository) GetByBucket(bucket string, limit, offset int) ([]models.Media, int64, error) {
	var media []models.Media
	var total int64

	query := r.db.Model(&models.Media{}).
		Where("bucket = ? AND is_active = ?", bucket, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("MediaType").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&media).Error

	return media, total, err
}

// ================================================
// COUNT OPERATIONS
// ================================================

// CountByEntity counts media for a specific entity
func (r *MediaRepository) CountByEntity(entityID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Media{}).
		Where("entity_id = ? AND is_active = ?", entityID, true).
		Count(&count).Error
	return count, err
}

// CountByMediaType counts media for a specific media type
func (r *MediaRepository) CountByMediaType(mediaTypeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Media{}).
		Where("media_type_id = ? AND is_active = ?", mediaTypeID, true).
		Count(&count).Error
	return count, err
}

// CountByUploader counts media uploaded by a specific user
func (r *MediaRepository) CountByUploader(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Media{}).
		Where("uploaded_by = ? AND is_active = ?", userID, true).
		Count(&count).Error
	return count, err
}

// ================================================
// CHECK OPERATIONS
// ================================================

// Exists checks if a media record exists
func (r *MediaRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Media{}).
		Where("id = ? AND is_active = ?", id, true).
		Count(&count).Error
	return count > 0, err
}

// ExistsByEntityAndMediaType checks if media exists for an entity and media type
func (r *MediaRepository) ExistsByEntityAndMediaType(entityID, mediaTypeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Media{}).
		Where("entity_id = ? AND media_type_id = ? AND is_active = ?", entityID, mediaTypeID, true).
		Count(&count).Error
	return count > 0, err
}

// ================================================
// BULK OPERATIONS
// ================================================

// DeleteByEntity soft deletes all media for an entity
func (r *MediaRepository) DeleteByEntity(entityID uuid.UUID) error {
	return r.db.Model(&models.Media{}).
		Where("entity_id = ?", entityID).
		Update("is_active", false).Error
}

// HardDeleteByEntity permanently deletes all media for an entity
func (r *MediaRepository) HardDeleteByEntity(entityID uuid.UUID) error {
	return r.db.Unscoped().
		Where("entity_id = ?", entityID).
		Delete(&models.Media{}).Error
}

// GetPrimaryMediaForEntity gets the primary media for an entity (first one)
func (r *MediaRepository) GetPrimaryMediaForEntity(entityID uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := r.db.
		Preload("MediaType").
		Where("entity_id = ? AND is_active = ?", entityID, true).
		Order("created_at ASC").
		First(&media).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}