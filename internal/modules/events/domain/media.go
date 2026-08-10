package domain

import "context"

// MediaService defines what the events domain needs for media operations
type MediaService interface {
	UploadFile(ctx context.Context, cmd UploadMediaCommand) (*MediaInfo, error)
	GetMediaByID(ctx context.Context, id string) (*MediaInfo, error)
	DeleteMediaByEntity(ctx context.Context, entityID string) error
}

type UploadMediaCommand struct {
	File          interface{}
	FileHeader    interface{}
	MediaTypeName string
	EntityID      string
	UploadedBy    string
}

type MediaInfo struct {
	ID  string
	URL string
}