package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// MEDIA ENTITY
// ============================================================

type Media struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Bucket      string
	Path        string
	URL         string
	MediaTypeID string
	EntityID    string
	UploadedBy  string
	FileSize    int64
	MimeType    string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// ============================================================
// FACTORY
// ============================================================

func NewMedia(
	name, bucket, path, url, mediaTypeID, entityID, uploadedBy, mimeType string,
	fileSize int64,
) (*Media, error) {
	if name == "" {
		return nil, errors.New("media name is required")
	}
	if bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if path == "" {
		return nil, errors.New("path is required")
	}
	if url == "" {
		return nil, errors.New("url is required")
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
		Slug:        generateSlug(name),
		Name:        name,
		DisplayName: name,
		Bucket:      bucket,
		Path:        path,
		URL:         url,
		MediaTypeID: mediaTypeID,
		EntityID:    entityID,
		UploadedBy:  uploadedBy,
		FileSize:    fileSize,
		MimeType:    mimeType,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ============================================================
// BEHAVIORS
// ============================================================

func (m *Media) IsDeleted() bool {
	return !m.IsActive
}

func (m *Media) SoftDelete() {
	now := time.Now()
	m.IsActive = false
	m.UpdatedAt = now
	m.DeletedAt = &now
}

func (m *Media) Restore() {
	m.IsActive = true
	m.DeletedAt = nil
	m.UpdatedAt = time.Now()
}

func (m *Media) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	m.Name = name
	m.UpdatedAt = time.Now()
	return nil
}

func (m *Media) UpdateDisplayName(displayName string) {
	m.DisplayName = displayName
	m.UpdatedAt = time.Now()
}

// ============================================================
// HELPERS
// ============================================================

func generateSlug(name string) string {
	slug := "media-" + uuid.New().String()[:8]
	return slug
}