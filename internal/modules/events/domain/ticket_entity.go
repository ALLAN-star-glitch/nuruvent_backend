// internal/modules/events/domain/ticket_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// TICKET - Domain Entity (Database Representation)
// ============================================================

// Ticket represents an event ticket type
type Ticket struct {
	ID                 string
	EventID            string // Foreign key to Event
	TicketTypeID       string // Foreign key to TicketType (seeded)
	Price              float64
	Quantity           int
	MaxPerPerson       *int
	EarlyBirdDeadline  *time.Time
	GroupMinAttendees  *int
	GroupDiscount      *float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewTicket creates a new Ticket entity
func NewTicket(
	eventID, ticketTypeID string,
	price float64,
	quantity int,
	maxPerPerson *int,
	earlyBirdDeadline *time.Time,
	groupMinAttendees *int,
	groupDiscount *float64,
) (*Ticket, error) {
	if eventID == "" {
		return nil, ErrInvalidTicket
	}
	if ticketTypeID == "" {
		return nil, ErrInvalidTicket
	}
	if price < 0 {
		return nil, ErrInvalidTicket
	}
	if quantity < 1 {
		return nil, ErrInvalidTicket
	}

	// Validate ticket type
	if !IsTicketTypeValid(TicketTypeValue(ticketTypeID)) {
		return nil, ErrInvalidTicket
	}

	// Group discount validation
	if groupMinAttendees != nil && groupDiscount != nil {
		if *groupMinAttendees < 2 {
			return nil, ErrInvalidTicket
		}
		if *groupDiscount < 0 {
			return nil, ErrInvalidTicket
		}
	}

	now := time.Now()
	return &Ticket{
		ID:                 generateTicketID(),
		EventID:            eventID,
		TicketTypeID:       ticketTypeID,
		Price:              price,
		Quantity:           quantity,
		MaxPerPerson:       maxPerPerson,
		EarlyBirdDeadline:  earlyBirdDeadline,
		GroupMinAttendees:  groupMinAttendees,
		GroupDiscount:      groupDiscount,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

func (t *Ticket) IsValid() bool {
	return t.ID != "" && t.EventID != "" && t.TicketTypeID != "" && t.Quantity > 0
}

func (t *Ticket) IsFree() bool {
	return t.Price == 0
}

func (t *Ticket) IsEarlyBird() bool {
	return t.EarlyBirdDeadline != nil
}

func (t *Ticket) IsGroupDiscount() bool {
	return t.GroupMinAttendees != nil && t.GroupDiscount != nil
}

func (t *Ticket) UpdatePrice(price float64) error {
	if price < 0 {
		return ErrInvalidTicket
	}
	t.Price = price
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Ticket) UpdateQuantity(quantity int) error {
	if quantity < 1 {
		return ErrInvalidTicket
	}
	t.Quantity = quantity
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Ticket) UpdateMaxPerPerson(maxPerPerson *int) {
	t.MaxPerPerson = maxPerPerson
	t.UpdatedAt = time.Now()
}

func (t *Ticket) UpdateEarlyBirdDeadline(deadline *time.Time) {
	t.EarlyBirdDeadline = deadline
	t.UpdatedAt = time.Now()
}

func (t *Ticket) UpdateGroupDiscount(minAttendees *int, discount *float64) error {
	if minAttendees != nil && discount != nil {
		if *minAttendees < 2 {
			return ErrInvalidTicket
		}
		if *discount < 0 {
			return ErrInvalidTicket
		}
	}
	t.GroupMinAttendees = minAttendees
	t.GroupDiscount = discount
	t.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// HELPERS
// ============================================================

func generateTicketID() string {
	return uuid.New().String()
}