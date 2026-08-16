package postgres

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
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FileName    string         `gorm:"not null;size:255"`
	FileSize    int64          `gorm:"not null"`
	MimeType    string         `gorm:"not null;size:100"`
	Bucket      string         `gorm:"not null;size:50;index:idx_media_bucket"`
	Path        string         `gorm:"not null;size:500"`
	URL         string         `gorm:"not null;size:500"`
	MediaTypeID string         `gorm:"type:uuid;not null;index:idx_media_media_type_id"`
	EntityID    string         `gorm:"type:uuid;not null;index:idx_media_entity"`
	UploadedBy  string         `gorm:"type:uuid;not null;index:idx_media_uploaded_by"`
	IsActive    bool           `gorm:"default:true;index:idx_media_is_active"`
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
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