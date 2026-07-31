// internal/models/refresh_token.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Token     string    `gorm:"uniqueIndex;not null;size:512" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
	UserAgent string    `json:"user_agent,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
}