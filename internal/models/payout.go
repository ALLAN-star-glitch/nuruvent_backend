// internal/models/payout.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payout struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InstitutionID uuid.UUID      `gorm:"type:uuid;index;not null" json:"institution_id"` // Changed from BusinessID
	Amount        float64        `gorm:"not null" json:"amount"`
	Currency      string         `gorm:"default:'KES'" json:"currency"`
	Method        string         `gorm:"default:'mpesa'" json:"method"`
	Reference     string         `gorm:"uniqueIndex" json:"reference"`
	Status        string         `gorm:"default:'pending'" json:"status"` // pending, processed, failed
	PaidAt        *time.Time     `json:"paid_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Institution Institution `gorm:"foreignKey:InstitutionID" json:"institution,omitempty"` // Changed from Business
}

