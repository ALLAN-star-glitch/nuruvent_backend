package service

import (
	"context"
	"mime/multipart"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
	// Upload
	UploadFile(ctx context.Context, cmd UploadCommand) (*domain.Media, error)

	// Delete
	DeleteFile(ctx context.Context, id string) error
	DeleteFilesByEntity(ctx context.Context, entityID string) error
	DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error

	// Get
	GetMediaByID(ctx context.Context, id string) (*domain.Media, error)
	GetMediaByEntity(ctx context.Context, entityID string, page, pageSize int) ([]*domain.Media, int64, error)
	GetMediaByEntityAndType(ctx context.Context, entityID, mediaTypeID string) ([]*domain.Media, error)
	GetImagesByEntity(ctx context.Context, entityID string) ([]*domain.Media, error)
	GetPrimaryMediaForEntity(ctx context.Context, entityID string) (*domain.Media, error)
	GetMediaByUploader(ctx context.Context, userID string, page, pageSize int) ([]*domain.Media, int64, error)
	GetMediaByType(ctx context.Context, mediaTypeName string, page, pageSize int) ([]*domain.Media, int64, error)

	// Media Types
	GetMediaTypes(ctx context.Context) ([]*domain.MediaType, error)
	GetMediaTypeByName(ctx context.Context, name string) (*domain.MediaType, error)
}

// ============================================================
// COMMANDS
// ============================================================

type UploadCommand struct {
	File          multipart.File
	FileHeader    *multipart.FileHeader
	MediaTypeName string
	EntityID      string
	UploadedBy    string
}