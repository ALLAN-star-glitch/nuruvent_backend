// internal/models/certificate.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Certificate struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID           uuid.UUID      `gorm:"type:uuid;index;not null" json:"event_id"`
	AttendeeID        uuid.UUID      `gorm:"type:uuid;index;not null" json:"attendee_id"`
	AccountID         *uuid.UUID     `gorm:"type:uuid;index" json:"account_id,omitempty"` // Changed from UserID
	CertificateNumber string         `gorm:"uniqueIndex;not null" json:"certificate_number"`
	TemplateURL       string         `gorm:"type:varchar(500)" json:"template_url,omitempty"`
	QRCode            string         `gorm:"uniqueIndex;type:varchar(255)" json:"qr_code,omitempty"`
	IsVerified        bool           `gorm:"default:false" json:"is_verified"`
	IssuedAt          time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"issued_at"`
	VerifiedAt        *time.Time     `json:"verified_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Event    Event    `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Attendee Attendee `gorm:"foreignKey:AttendeeID" json:"attendee,omitempty"`
	Account  Account  `gorm:"foreignKey:AccountID" json:"account,omitempty"` // Changed from User
}

