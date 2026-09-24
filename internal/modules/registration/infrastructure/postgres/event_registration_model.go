package postgres

import "time"

// EventRegistrationModel maps to the `event_registrations` table.
type EventRegistrationModel struct {
	RegistrationID string    `gorm:"primaryKey;type:uuid"`
    IsActive       bool   `gorm:"not null;default:true"`
	EventID        string    `gorm:"type:uuid;not null;index"`
	UserID         *string   `gorm:"type:uuid;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations
	Registration *RegistrationModel `gorm:"foreignKey:RegistrationID"`
}

func (EventRegistrationModel) TableName() string {
	return "event_registrations"
}