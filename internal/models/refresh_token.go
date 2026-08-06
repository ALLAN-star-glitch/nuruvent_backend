// internal/models/refresh_token.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	BaseModel
	AccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"account_id"` // Changed from UserID
	Account   Account   `gorm:"foreignKey:AccountID" json:"-"`              // Changed from User
	Token     string    `gorm:"uniqueIndex;not null;size:512" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
	UserAgent string    `json:"user_agent,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
}