package storage

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================
// MEDIA TYPE REPOSITORY
// ============================================================

type MediaTypeRepository struct {
	db *gorm.DB
}

func NewMediaTypeRepository(db *gorm.DB) *MediaTypeRepository {
	return &MediaTypeRepository{db: db}
}

// ============================================================
// CRUD OPERATIONS
// ============================================================

// GetByID gets a media type by ID
func (r *MediaTypeRepository) GetByID(id uuid.UUID) (*MediaTypeModel, error) {
	var mediaType MediaTypeModel
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&mediaType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &mediaType, nil
}

// GetByName gets a media type by name
func (r *MediaTypeRepository) GetByName(name string) (*MediaTypeModel, error) {
	var mediaType MediaTypeModel
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&mediaType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &mediaType, nil
}

// GetAll gets all media types
func (r *MediaTypeRepository) GetAll() ([]MediaTypeModel, error) {
	var mediaTypes []MediaTypeModel
	err := r.db.
		Where("is_active = ?", true).
		Order("sort_order ASC, display_name ASC").
		Find(&mediaTypes).Error
	return mediaTypes, err
}

// GetByBucket gets a media type by bucket name
func (r *MediaTypeRepository) GetByBucket(bucket string) (*MediaTypeModel, error) {
	var mediaType MediaTypeModel
	err := r.db.Where("bucket = ? AND is_active = ?", bucket, true).First(&mediaType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &mediaType, nil
}