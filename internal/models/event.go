// internal/models/event.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Event represents a virtual event on the platform
type Event struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Slug             string         `gorm:"uniqueIndex;not null;size:255" json:"slug"`      // URL-safe: "annual-tech-conference-2024"
	Name             string         `gorm:"not null;size:255" json:"name"`                  // Canonical: "Annual Tech Conference 2024"
	DisplayName      string         `gorm:"size:255" json:"display_name,omitempty"`          // UI: "🚀 Annual Tech Conference 2024"
	Description      string         `gorm:"type:text" json:"description"`
	EventTypeID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"event_type_id"`
	EventStatusID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"event_status_id"`
	
	// Images
	ImageURL         string         `gorm:"size:500" json:"image_url,omitempty"`
	ThumbnailURL     string         `gorm:"size:500" json:"thumbnail_url,omitempty"`
	
	// Date & Time
	Date             time.Time      `gorm:"not null" json:"date"`
	Time             string         `gorm:"not null;size:50" json:"time"`
	Duration         int            `json:"duration"`
	
	// Pricing
	Price            float64        `gorm:"default:0" json:"price"`
	CertificatePrice float64        `gorm:"default:0" json:"certificate_price"`
	
	// Location & Virtual
	Location         string         `gorm:"size:255" json:"location"`
	IsVirtual        bool           `gorm:"default:true" json:"is_virtual"`
	ZoomLink         string         `gorm:"size:500" json:"zoom_link,omitempty"`
	MeetLink         string         `gorm:"size:500" json:"meet_link,omitempty"`
	
	// Capacity
	MaxAttendees     int            `json:"max_attendees"`
	CurrentAttendees int            `gorm:"default:0" json:"current_attendees"`
	
	// Active status
	IsActive         bool           `gorm:"default:true;index" json:"is_active"`
	
	// Foreign Keys
	InstitutionID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"institution_id"`
	CreatedBy        uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	
	// Timestamps
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	EventType   EventType    `gorm:"foreignKey:EventTypeID" json:"event_type,omitempty"`
	EventStatus EventStatus  `gorm:"foreignKey:EventStatusID" json:"event_status,omitempty"`
	Institution Institution `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"`
	Attendees   []Attendee   `gorm:"foreignKey:EventID" json:"attendees,omitempty"`
	Creator     Account      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

