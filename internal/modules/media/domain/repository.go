package domain

import "context"

// ============================================================
// REPOSITORY INTERFACE (OUTBOUND PORT)
// ============================================================

type Repository interface {
	// Media CRUD
	CreateMedia(ctx context.Context, media *Media) error
	GetMediaByID(ctx context.Context, id string) (*Media, error)
	GetMediaByEntity(ctx context.Context, entityID string, limit, offset int) ([]*Media, int64, error)
	GetMediaByEntityAndType(ctx context.Context, entityID, mediaTypeID string) ([]*Media, error)
	GetMediaByUploader(ctx context.Context, userID string, limit, offset int) ([]*Media, int64, error)
	GetMediaByType(ctx context.Context, mediaTypeID string, limit, offset int) ([]*Media, int64, error)
	GetPrimaryMediaForEntity(ctx context.Context, entityID string) (*Media, error)
	UpdateMedia(ctx context.Context, media *Media) error
	DeleteMedia(ctx context.Context, id string) error
	DeleteMediaByEntity(ctx context.Context, entityID string) error

	// Media Types
	GetMediaTypeByID(ctx context.Context, id string) (*MediaType, error)
	GetMediaTypeByName(ctx context.Context, name string) (*MediaType, error)
	GetAllMediaTypes(ctx context.Context) ([]*MediaType, error)
}

// ============================================================
// MEDIA TYPE REPOSITORY (Optional - Can be part of Repository)
// ============================================================

// If you prefer separate interfaces for media type operations
type MediaTypeRepository interface {
	GetByID(ctx context.Context, id string) (*MediaType, error)
	GetByName(ctx context.Context, name string) (*MediaType, error)
	GetAll(ctx context.Context) ([]*MediaType, error)
}