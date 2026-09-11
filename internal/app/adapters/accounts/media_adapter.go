// internal/app/adapters/accounts/media_adapter.go

package accounts

import (
	"context"

	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"
)

// MediaAdapter adapts the media module's Service to the account domain's
// MediaService port. Used for user avatars and account logos.
type MediaAdapter struct {
	mediaSvc mediaService.Service
}

func NewMediaAdapter(mediaSvc mediaService.Service) accountDomain.MediaService {
	return &MediaAdapter{mediaSvc: mediaSvc}
}

// ============================================================
// UPLOAD
// ============================================================

func (a *MediaAdapter) UploadFile(ctx context.Context, cmd accountDomain.UploadMediaCommand) (*accountDomain.MediaInfo, error) {
	media, err := a.mediaSvc.UploadFile(ctx, mediaService.UploadCommand{
		File:          cmd.File,
		FileName:      cmd.FileName,
		ContentType:   cmd.ContentType,
		MediaTypeName: cmd.MediaTypeName,
		EntityID:      cmd.EntityID,
		UploadedBy:    cmd.UploadedBy,
	})
	if err != nil {
		return nil, err
	}

	return &accountDomain.MediaInfo{
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

func (a *MediaAdapter) GetMediaTypeByName(ctx context.Context, name string) (*accountDomain.MediaTypeInfo, error) {
	mediaType, err := a.mediaSvc.GetMediaTypeByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, nil
	}

	return &accountDomain.MediaTypeInfo{
		ID:   mediaType.ID,
		Name: mediaType.Name,
		Slug: mediaType.Slug,
	}, nil
}

// ============================================================
// DELETE
// ============================================================

func (a *MediaAdapter) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error {
	return a.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, entityID, mediaTypeID)
}