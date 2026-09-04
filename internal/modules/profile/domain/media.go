// internal/modules/profile/domain/media.go

package domain

import "context"

// ============================================================
// MEDIA SERVICE - Outbound Port for Media Operations
// ============================================================

type MediaService interface {
    // Upload
    UploadFile(ctx context.Context, cmd UploadMediaCommand) (*MediaInfo, error)

    // Get
    GetMediaByID(ctx context.Context, id string) (*MediaInfo, error)
    GetMediaByEntity(ctx context.Context, entityID string) ([]*MediaInfo, error)
    GetMediaTypeByName(ctx context.Context, name string) (*MediaTypeInfo, error)

    // Delete
    DeleteFile(ctx context.Context, id string) error
    DeleteFilesByEntity(ctx context.Context, entityID string) error
    DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error
}

// ============================================================
// COMMANDS - Pure Domain Types
// ============================================================

type UploadMediaCommand struct {
    File          []byte
    FileName      string
    ContentType   string
    MediaTypeName string
    EntityID      string
    UploadedBy    string
}

// ============================================================
// DTOs - Pure Domain Types
// ============================================================

type MediaInfo struct {
    ID         string
    URL        string
    MediaType  string
    EntityID   string
    UploadedBy string
    CreatedAt  string
}

type MediaTypeInfo struct {
    ID   string
    Name string
    Slug string
}