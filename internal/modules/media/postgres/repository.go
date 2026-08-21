// internal/modules/media/postgres/repository.go

package postgres

import (
	"context"
	"errors"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"

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
// MEDIA CRUD OPERATIONS
// ============================================================

func (r *PostgresRepository) CreateMedia(ctx context.Context, media *domain.Media) error {
	model := ToModelMedia(media)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) GetMediaByID(ctx context.Context, id string) (*domain.Media, error) {
	var model MediaModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", id, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainMedia(&model), nil
}


// GetMediaByEntity - Get all media for an entity
// When called for deletion (permanent delete), we need to include ALL media
func (r *PostgresRepository) GetMediaByEntity(ctx context.Context, entityID string, limit, offset int) ([]*domain.Media, int64, error) {
	var models []MediaModel
	var total int64

	// ✅ For deletion, we need to get ALL media regardless of is_active
	// Use Unscoped() to include soft-deleted records
	query := r.db.WithContext(ctx).
		Unscoped().
		Model(&MediaModel{}).
		Where("entity_id = ?", entityID)  // ✅ Remove is_active filter

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	media := make([]*domain.Media, len(models))
	for i, m := range models {
		media[i] = ToDomainMedia(&m)
	}
	return media, total, nil
}

func (r *PostgresRepository) GetMediaByEntityAndType(ctx context.Context, entityID, mediaTypeID string) ([]*domain.Media, error) {
	var models []MediaModel
	err := r.db.WithContext(ctx).
		Where("entity_id = ? AND media_type_id = ? AND is_active = ?", entityID, mediaTypeID, true).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	media := make([]*domain.Media, len(models))
	for i, m := range models {
		media[i] = ToDomainMedia(&m)
	}
	return media, nil
}

func (r *PostgresRepository) GetMediaByUploader(ctx context.Context, userID string, limit, offset int) ([]*domain.Media, int64, error) {
	var models []MediaModel
	var total int64

	query := r.db.WithContext(ctx).Model(&MediaModel{}).
		Where("uploaded_by = ? AND is_active = ?", userID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	media := make([]*domain.Media, len(models))
	for i, m := range models {
		media[i] = ToDomainMedia(&m)
	}
	return media, total, nil
}

func (r *PostgresRepository) GetMediaByType(ctx context.Context, mediaTypeID string, limit, offset int) ([]*domain.Media, int64, error) {
	var models []MediaModel
	var total int64

	query := r.db.WithContext(ctx).Model(&MediaModel{}).
		Where("media_type_id = ? AND is_active = ?", mediaTypeID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	media := make([]*domain.Media, len(models))
	for i, m := range models {
		media[i] = ToDomainMedia(&m)
	}
	return media, total, nil
}

func (r *PostgresRepository) GetPrimaryMediaForEntity(ctx context.Context, entityID string) (*domain.Media, error) {
	var model MediaModel
	err := r.db.WithContext(ctx).
		Where("entity_id = ? AND is_active = ?", entityID, true).
		Order("created_at ASC").
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainMedia(&model), nil
}

func (r *PostgresRepository) UpdateMedia(ctx context.Context, media *domain.Media) error {
	model := ToModelMedia(media)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) DeleteMedia(ctx context.Context, id string) error {
	// ✅ Use Unscoped() to permanently delete even if soft-deleted
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&MediaModel{}, "id = ?", id).Error
}

func (r *PostgresRepository) DeleteMediaByEntity(ctx context.Context, entityID string) error {
	// ✅ Use Unscoped() to permanently delete even if soft-deleted
	return r.db.WithContext(ctx).
		Unscoped().
		Where("entity_id = ?", entityID).
		Delete(&MediaModel{}).Error
}

// ============================================================
// MEDIA TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetMediaTypeByID(ctx context.Context, id string) (*domain.MediaType, error) {
	var model MediaTypeModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_active = ?", id, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainMediaType(&model), nil
}

func (r *PostgresRepository) GetMediaTypeByName(ctx context.Context, name string) (*domain.MediaType, error) {
	var model MediaTypeModel
	err := r.db.WithContext(ctx).
		Where("name = ? AND is_active = ?", name, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainMediaType(&model), nil
}

func (r *PostgresRepository) GetAllMediaTypes(ctx context.Context) ([]*domain.MediaType, error) {
	var models []MediaTypeModel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	types := make([]*domain.MediaType, len(models))
	for i, m := range models {
		types[i] = ToDomainMediaType(&m)
	}
	return types, nil
}