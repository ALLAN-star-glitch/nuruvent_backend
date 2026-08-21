// internal/modules/events/domain/media.go

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
    DeleteMediaByEntity(ctx context.Context, entityID string) error
    DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error
}

// ============================================================
// COMMANDS - Pure Domain Types (NO external dependencies)
// ============================================================

type UploadMediaCommand struct {
    File          []byte // ✅ Raw file data
    FileName      string // ✅ Original filename
    ContentType   string // ✅ MIME type
    MediaTypeName string
    EntityID      string
    UploadedBy    string
}

// ============================================================
// DTOs - Pure Domain Types (NO external dependencies)
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