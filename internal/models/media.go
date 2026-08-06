// internal/models/media.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Media represents a media file stored in Supabase Storage
type Media struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FileName    string         `gorm:"not null;size:255" json:"file_name"`
	FileSize    int64          `gorm:"not null" json:"file_size"`
	MimeType    string         `gorm:"not null;size:100" json:"mime_type"`
	Bucket      string         `gorm:"not null;size:50" json:"bucket"`
	Path        string         `gorm:"not null;size:500" json:"path"`
	URL         string         `gorm:"not null;size:500" json:"url"`
	
	// Foreign keys
	MediaTypeID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"media_type_id"`
	EntityID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"entity_id"`
	UploadedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"uploaded_by"`
	
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	MediaType   MediaType      `gorm:"foreignKey:MediaTypeID" json:"media_type,omitempty"`
}

// TableName returns the table name
func (Media) TableName() string {
	return "media"
}

// IsImage returns true if the media is an image
func (m *Media) IsImage() bool {
	imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"}
	for _, t := range imageTypes {
		if m.MimeType == t {
			return true
		}
	}
	return false
}

// IsVideo returns true if the media is a video
func (m *Media) IsVideo() bool {
	videoTypes := []string{"video/mp4", "video/webm", "video/ogg", "video/quicktime"}
	for _, t := range videoTypes {
		if m.MimeType == t {
			return true
		}
	}
	return false
}

// IsPDF returns true if the media is a PDF
func (m *Media) IsPDF() bool {
	return m.MimeType == "application/pdf"
}

// GetFileExtension returns the file extension
func (m *Media) GetFileExtension() string {
	for i := len(m.FileName) - 1; i >= 0; i-- {
		if m.FileName[i] == '.' {
			return m.FileName[i:]
		}
	}
	return ""
}