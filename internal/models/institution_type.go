// internal/models/institution_type.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InstitutionType struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`          // URL-safe: "training-institute"
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`                // Canonical: "Training Institute"
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`        // UI: "🏫 Training Institute"
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	
	// SEO and discoverability
	MetaTitle       string `gorm:"type:varchar(150)" json:"meta_title"`
	MetaDescription string `gorm:"type:text" json:"meta_description"`
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}