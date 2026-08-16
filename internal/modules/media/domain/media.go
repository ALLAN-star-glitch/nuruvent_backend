// internal/modules/media/domain/media.go

package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID          string
	FileName    string
	FileSize    int64
	MimeType    string
	Bucket      string
	Path        string
	URL         string
	MediaTypeID string
	EntityID    string
	UploadedBy  string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewMedia(
	fileName, bucket, path, url, mediaTypeID, entityID, uploadedBy, mimeType string,
	fileSize int64,
) (*Media, error) {
	if fileName == "" {
		return nil, errors.New("file name is required")
	}
	if bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if path == "" {
		return nil, errors.New("path is required")
	}
	if url == "" {
		return nil, errors.New("URL is required")
	}
	if mediaTypeID == "" {
		return nil, errors.New("media type is required")
	}
	if entityID == "" {
		return nil, errors.New("entity ID is required")
	}
	if uploadedBy == "" {
		return nil, errors.New("uploaded by is required")
	}

	now := time.Now()
	return &Media{
		ID:          uuid.New().String(),
		FileName:    fileName,
		FileSize:    fileSize,
		MimeType:    mimeType,
		Bucket:      bucket,
		Path:        path,
		URL:         url,
		MediaTypeID: mediaTypeID,
		EntityID:    entityID,
		UploadedBy:  uploadedBy,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}