// internal/models/payment.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Payment struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"event_id"`
	AccountID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"account_id"` // Changed from UserID
	Amount     float64        `gorm:"not null" json:"amount"`
	Currency   string         `gorm:"default:'KES'" json:"currency"`
	Method     string         `gorm:"default:'mpesa'" json:"method"` // mpesa, card, bank
	Reference  string         `gorm:"uniqueIndex" json:"reference"`
	Status     string         `gorm:"default:'pending'" json:"status"` // pending, completed, failed, refunded
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	PaidAt     *time.Time     `json:"paid_at,omitempty"`
	RefundedAt *time.Time     `json:"refunded_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Event   Event   `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Account Account `gorm:"foreignKey:AccountID" json:"account,omitempty"` // Changed from User
}
