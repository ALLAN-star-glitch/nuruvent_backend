// internal/models/event_type.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventType struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`          // URL-safe: "workshop"
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`                // Canonical: "Workshop"
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`        // UI: "🛠️ Workshop"
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`                          // graduation-cap, video, users, laptop
	Color       string         `gorm:"type:varchar(20)" json:"color"`                         // #4F46E5, #7C3AED, etc.
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	
	// Event type properties
	SupportsCertificate bool   `gorm:"default:true" json:"supports_certificate"`
	MinDuration         int    `gorm:"default:60" json:"min_duration"`
	MaxDuration         int    `gorm:"default:480" json:"max_duration"`
	
	// SEO and discoverability
	MetaTitle       string `gorm:"type:varchar(150)" json:"meta_title"`
	MetaDescription string `gorm:"type:text" json:"meta_description"`
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}