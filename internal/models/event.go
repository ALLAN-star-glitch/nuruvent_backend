package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title            string         `gorm:"not null" json:"title"`
	Slug             string         `gorm:"uniqueIndex;not null" json:"slug"`
	Description      string         `gorm:"type:text" json:"description"`
	Type             string         `gorm:"not null" json:"type"` // workshop, webinar, bootcamp, meetup
	Date             time.Time      `gorm:"not null" json:"date"`
	Time             string         `gorm:"not null" json:"time"`
	Duration         int            `json:"duration"` // in minutes
	Price            float64        `gorm:"default:0" json:"price"`
	CertificatePrice float64        `gorm:"default:0" json:"certificate_price"`
	Location         string         `gorm:"type:varchar(255)" json:"location"`
	IsVirtual        bool           `gorm:"default:true" json:"is_virtual"`
	ZoomLink         string         `gorm:"type:varchar(500)" json:"zoom_link,omitempty"`
	MeetLink         string         `gorm:"type:varchar(500)" json:"meet_link,omitempty"`
	MaxAttendees     int            `json:"max_attendees"`
	CurrentAttendees int            `gorm:"default:0" json:"current_attendees"`
	Status           string         `gorm:"default:'draft'" json:"status"` // draft, published, cancelled, completed
	BusinessID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"business_id"`
	CreatedBy        uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Business  Business    `gorm:"foreignKey:BusinessID" json:"business,omitempty"`
	Attendees []Attendee  `gorm:"foreignKey:EventID" json:"attendees,omitempty"`
}

// TableName returns the table name
func (Event) TableName() string {
	return "events"
}