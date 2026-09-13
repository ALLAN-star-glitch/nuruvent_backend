// internal/modules/account/accountdomain/media_service.go

package accountdomain

import "context"

// ============================================================
// OUTBOUND PORT: MediaService
// ============================================================
//
// MediaService exposes the media module's upload/delete capabilities to
// the account module. Used for user avatars and account logos.
//
// Implementations live in the composition layer (app/adapters/accounts)
// and delegate to the media module's Service. The account module does not
// import the media module directly.

type MediaService interface {
	UploadFile(ctx context.Context, cmd UploadMediaCommand) (*MediaInfo, error)
	GetMediaTypeByName(ctx context.Context, name string) (*MediaTypeInfo, error)
	DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error
}

// UploadMediaCommand is the account module's own upload command.
type UploadMediaCommand struct {
	File          []byte
	FileName      string
	ContentType   string
	MediaTypeName string
	EntityID      string
	UploadedBy    string
}

// MediaInfo is the projection returned by MediaService.
type MediaInfo struct {
	ID         string
	URL        string
	MediaType  string
	EntityID   string
	UploadedBy string
	CreatedAt  string
}

// MediaTypeInfo is a lightweight projection of a media type.
type MediaTypeInfo struct {
	ID   string
	Name string
	Slug string
}