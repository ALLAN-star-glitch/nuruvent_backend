package postgres

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"
	"github.com/google/uuid"
)

// ============================================================
// DOMAIN → DATABASE MODEL MAPPERS
// ============================================================

func ToModelMedia(media *domain.Media) *MediaModel {
	if media == nil {
		return nil
	}

	return &MediaModel{
		ID:          media.ID,
		FileName:    media.FileName,
		FileSize:    media.FileSize,
		MimeType:    media.MimeType,
		Bucket:      media.Bucket,
		Path:        media.Path,
		URL:         media.URL,
		MediaTypeID: media.MediaTypeID,
		EntityID:    media.EntityID,
		UploadedBy:  media.UploadedBy,
		IsActive:    media.IsActive,
		CreatedAt:   media.CreatedAt,
		UpdatedAt:   media.UpdatedAt,
	}
}

func ToModelMediaType(mediaType *domain.MediaType) *MediaTypeModel {
	if mediaType == nil {
		return nil
	}

	return &MediaTypeModel{
		ID:          uuid.MustParse(mediaType.ID),
		Slug:        mediaType.Slug,
		Name:        mediaType.Name,
		DisplayName: mediaType.DisplayName,
		Description: mediaType.Description,
		Bucket:      mediaType.Bucket,
		Icon:        mediaType.Icon,
		SortOrder:   mediaType.SortOrder,
		MaxFileSize: mediaType.MaxFileSize,
		IsActive:    mediaType.IsActive,
	}
}

// ============================================================
// DATABASE MODEL → DOMAIN MAPPERS
// ============================================================

func ToDomainMedia(model *MediaModel) *domain.Media {
	if model == nil {
		return nil
	}

	return &domain.Media{
		ID:          model.ID,
		FileName:    model.FileName,
		FileSize:    model.FileSize,
		MimeType:    model.MimeType,
		Bucket:      model.Bucket,
		Path:        model.Path,
		URL:         model.URL,
		MediaTypeID: model.MediaTypeID,
		EntityID:    model.EntityID,
		UploadedBy:  model.UploadedBy,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func ToDomainMediaType(model *MediaTypeModel) *domain.MediaType {
	if model == nil {
		return nil
	}

	return &domain.MediaType{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Bucket:      model.Bucket,
		Icon:        model.Icon,
		SortOrder:   model.SortOrder,
		MaxFileSize: model.MaxFileSize,
		IsActive:    model.IsActive,
	}
}