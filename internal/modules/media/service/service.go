// internal/modules/media/service/service.go

package service

import (
	"context"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/domain"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
	// Upload - Uses pure domain types ([]byte, string)
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
	GetMediaTypeByID(ctx context.Context, id string) (*domain.MediaType, error)
	GetMediaTypeByName(ctx context.Context, name string) (*domain.MediaType, error)
}

// ============================================================
// COMMANDS - Pure Domain Types (NO external dependencies)
// ============================================================

type UploadCommand struct {
	// ✅ Pure Go types - NO multipart.File or *multipart.FileHeader
	File          []byte // Raw file data
	FileName      string // Original filename
	ContentType   string // MIME type (e.g., "image/jpeg")
	MediaTypeName string // "Event" or "Certificate"
	EntityID      string // Event ID
	UploadedBy    string // User ID
}