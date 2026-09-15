package postgres

import (
	"time"

	"gorm.io/gorm"
)

// RegistrationModel maps to the `registrations` table.
type RegistrationModel struct {
	ID                 string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID             *string        `gorm:"type:uuid;index"`
	GuestEmail         *string        `gorm:"type:varchar(255);index"`
	GuestName          *string        `gorm:"type:varchar(255)"`
	GuestPhone         *string        `gorm:"type:varchar(20)"`
	RegistrationNumber string         `gorm:"type:varchar(100);not null;uniqueIndex"`
	StatusID           string         `gorm:"type:uuid;not null;index"`
	Currency           string         `gorm:"type:varchar(3);not null;default:'KES'"`
	Subtotal           int64          `gorm:"not null;default:0"`
	DiscountTotal      int64          `gorm:"not null;default:0"`
	TotalAmount        int64          `gorm:"not null;default:0"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ConfirmedAt        *time.Time
	CancelledAt        *time.Time
	CancelledBy        *string        `gorm:"type:uuid"`
	CancellationReason *string        `gorm:"type:text"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`

	// Relations
	Status RegistrationStatusModel `gorm:"foreignKey:StatusID"`
}

func (RegistrationModel) TableName() string {
	return "registrations"
}