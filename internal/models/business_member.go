// internal/models/business_member.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessMember struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BusinessID uuid.UUID      `gorm:"type:uuid;index;not null" json:"business_id"`
	UserID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Role       string         `gorm:"not null" json:"role"` // host, event_manager, member
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	JoinedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Business Business `gorm:"foreignKey:BusinessID" json:"business,omitempty"`
	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name
func (BusinessMember) TableName() string {
	return "business_members"
}