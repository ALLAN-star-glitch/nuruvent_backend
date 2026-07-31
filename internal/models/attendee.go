// internal/models/attendee.go

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Attendee struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"event_id"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Email      string         `gorm:"not null" json:"email"`
	Name       string         `gorm:"not null" json:"name"`
	Phone      string         `gorm:"type:varchar(20)" json:"phone"`
	Status     string         `gorm:"default:'registered'" json:"status"` // registered, joined, partial, full, confirmed, no-show
	JoinedAt   *time.Time     `json:"joined_at,omitempty"`
	LeftAt     *time.Time     `json:"left_at,omitempty"`
	Duration   int            `json:"duration"` // in minutes
	AttendedAt *time.Time     `json:"attended_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Event Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Attendee) TableName() string {
	return "attendees"
}