package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================
// GORM MODELS
// ============================================================

// MediaModel represents a file stored in the system
type MediaModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Bucket      string         `gorm:"type:varchar(50);not null" json:"bucket"`
	Path        string         `gorm:"type:varchar(500);not null" json:"path"`
	URL         string         `gorm:"type:varchar(500);not null" json:"url"`
	MediaTypeID uuid.UUID      `gorm:"type:uuid;index;not null" json:"media_type_id"`
	EntityID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"entity_id"`
	UploadedBy  uuid.UUID      `gorm:"type:uuid;index;not null" json:"uploaded_by"`
	FileSize    int64          `gorm:"not null" json:"file_size"`
	MimeType    string         `gorm:"type:varchar(100);not null" json:"mime_type"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationship
	MediaType MediaTypeModel `gorm:"foreignKey:MediaTypeID" json:"media_type,omitempty"`
}

func (MediaModel) TableName() string {
	return "media"
}

// MediaTypeModel represents a type of media (event, business, profile, etc.)
type MediaTypeModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`
	Description string         `gorm:"type:text" json:"description"`
	Bucket      string         `gorm:"type:varchar(50);not null" json:"bucket"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	MaxFileSize int64          `gorm:"default:10485760" json:"max_file_size"` // 10MB default
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationship
	Media []MediaModel `gorm:"foreignKey:MediaTypeID" json:"media,omitempty"`
}

func (MediaTypeModel) TableName() string {
	return "media_types"
}