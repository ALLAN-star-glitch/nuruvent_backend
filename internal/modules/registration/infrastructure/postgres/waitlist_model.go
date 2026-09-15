package postgres

import (
	"time"

	"gorm.io/gorm"
)

// WaitlistModel maps to the `event_waitlist` table.
type WaitlistModel struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	EventID          string         `gorm:"type:uuid;not null;index"`
	UserID           *string        `gorm:"type:uuid;index"`
	GuestEmail       *string        `gorm:"type:varchar(255);index"`
	GuestName        *string        `gorm:"type:varchar(255)"`
	TicketTypeID     *string        `gorm:"type:uuid;index"`
	Position         int            `gorm:"not null;index"`
	Status           string         `gorm:"type:varchar(50);default:'waiting'"`
	NotifiedAt       *time.Time
	OfferedAt        *time.Time
	OfferedExpiresAt *time.Time
	ConvertedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (WaitlistModel) TableName() string {
	return "event_waitlist"
}