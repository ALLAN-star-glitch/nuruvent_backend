// internal/models/event_status.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStatus struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug        string         `gorm:"type:varchar(50);unique;not null" json:"slug"`          // URL-safe: "published"
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`                // Canonical: "Published"
	DisplayName string         `gorm:"type:varchar(150)" json:"display_name,omitempty"`        // UI: "✅ Published"
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"type:varchar(20)" json:"color"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	
	// Status properties
	IsFinal     bool `gorm:"default:false" json:"is_final"`
	CanEdit     bool `gorm:"default:true" json:"can_edit"`
	CanRegister bool `gorm:"default:true" json:"can_register"`
	
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}