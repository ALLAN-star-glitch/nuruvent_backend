package postgres

import "time"

// EventRegistrationTicketModel maps to the `event_registration_tickets` table.
type EventRegistrationTicketModel struct {
	ID             string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	RegistrationID string    `gorm:"type:uuid;not null;index"`
	TicketTypeID   string    `gorm:"type:uuid;not null;index"`
	Quantity       int       `gorm:"not null"`
	UnitPrice      int64     `gorm:"not null"`
	Discount       int64     `gorm:"not null;default:0"`
	CreatedAt      time.Time

}

func (EventRegistrationTicketModel) TableName() string {
	return "event_registration_tickets"
}