// internal/app/adapters/events/media_adapter.go

package events

import (
	"context"

	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// MediaAdapter adapts the media module's Service to the events domain's MediaService
type MediaAdapter struct {
	mediaSvc mediaService.Service
}

// NewMediaAdapter creates a new media adapter for events
func NewMediaAdapter(mediaSvc mediaService.Service) domain.MediaService {
	return &MediaAdapter{
		mediaSvc: mediaSvc,
	}
}

// ============================================================
// UPLOAD
// ============================================================

func (a *MediaAdapter) UploadFile(ctx context.Context, cmd domain.UploadMediaCommand) (*domain.MediaInfo, error) {
	// Convert domain command to media service command
	uploadCmd := mediaService.UploadCommand{
		File:          cmd.File,
		FileName:      cmd.FileName,
		ContentType:   cmd.ContentType,
		MediaTypeName: cmd.MediaTypeName,
		EntityID:      cmd.EntityID,
		UploadedBy:    cmd.UploadedBy,
	}

	media, err := a.mediaSvc.UploadFile(ctx, uploadCmd)
	if err != nil {
		return nil, err
	}

	return &domain.MediaInfo{
		ID:         media.ID,
		URL:        media.URL,
		MediaType:  media.MediaTypeID,
		EntityID:   media.EntityID,
		UploadedBy: media.UploadedBy,
		CreatedAt:  media.CreatedAt.String(),
	}, nil
}

// ============================================================
// GET
// ============================================================

func (a *MediaAdapter) GetMediaByID(ctx context.Context, id string) (*domain.MediaInfo, error) {
	media, err := a.mediaSvc.GetMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, nil
	}

	return &domain.MediaInfo{
		ID:         media.ID,
		URL:        media.URL,
		MediaType:  media.MediaTypeID,
		EntityID:   media.EntityID,
		UploadedBy: media.UploadedBy,
		CreatedAt:  media.CreatedAt.String(),
	}, nil
}

func (a *MediaAdapter) GetMediaByEntity(ctx context.Context, entityID string) ([]*domain.MediaInfo, error) {
	mediaList, _, err := a.mediaSvc.GetMediaByEntity(ctx, entityID, 1, 1000)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.MediaInfo, len(mediaList))
	for i, media := range mediaList {
		result[i] = &domain.MediaInfo{
			ID:         media.ID,
			URL:        media.URL,
			MediaType:  media.MediaTypeID,
			EntityID:   media.EntityID,
			UploadedBy: media.UploadedBy,
			CreatedAt:  media.CreatedAt.String(),
		}
	}
	return result, nil
}

func (a *MediaAdapter) GetMediaTypeByName(ctx context.Context, name string) (*domain.MediaTypeInfo, error) {
	mediaType, err := a.mediaSvc.GetMediaTypeByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, nil
	}

	return &domain.MediaTypeInfo{
		ID:   mediaType.ID,
		Name: mediaType.Name,
		Slug: mediaType.Slug,
	}, nil
}

// ============================================================
// DELETE
// ============================================================

func (a *MediaAdapter) DeleteFile(ctx context.Context, id string) error {
	return a.mediaSvc.DeleteFile(ctx, id)
}

func (a *MediaAdapter) DeleteFilesByEntity(ctx context.Context, entityID string) error {
	return a.mediaSvc.DeleteFilesByEntity(ctx, entityID)
}

func (a *MediaAdapter) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error {
	return a.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, entityID, mediaTypeID)
}