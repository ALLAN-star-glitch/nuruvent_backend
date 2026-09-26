package registrationdomain

import "time"

// EventRegistration is the event-specific extension of Registration.
type EventRegistration struct {
	Registration *Registration
	IsActive     bool
	EventID      string
	Selections   []TicketSelection
	Pricing      PricingSnapshot
	AttendedAt   *time.Time // reserved for future attendance

	// JoinLinks is populated in memory during confirmation and read by
	// the notifier. Not persisted — the tokens themselves live in the
	// attendance module's join_tokens table.
	JoinLinks []JoinLink
}

// NewEventRegistration composes a base registration with event-specific data.
func NewEventRegistration(
    reg *Registration,
    eventID string,
    selections []TicketSelection,
    pricing PricingSnapshot,
) (*EventRegistration, error) {
    if reg == nil {
        return nil, ErrRegistrationNotFound
    }
    if eventID == "" {
        return nil, ErrRegistrableTypeUnknown
    }
    if len(selections) == 0 {
        return nil, ErrTicketUnavailable
    }
    for _, s := range selections {
        if err := s.Validate(); err != nil {
            return nil, err
        }
    }
    return &EventRegistration{
        Registration: reg,
        IsActive:     true, 
        EventID:      eventID,
        Selections:   selections,
        Pricing:      pricing,
    }, nil
}

// HydrateEventRegistration reconstructs an EventRegistration from
// persisted state. Used by the repository when reading from the DB.
//
// Unlike NewEventRegistration (which assumes a fresh creation and
// sets IsActive = true), this takes the is_active flag as given — it
// reflects whatever was persisted, not a hardcoded default.
//
// No validation runs here. The DB row is the source of truth.
func HydrateEventRegistration(
	reg *Registration,
	eventID string,
	isActive bool,
	selections []TicketSelection,
) *EventRegistration {
	return &EventRegistration{
		Registration: reg,
		EventID:      eventID,
		IsActive:     isActive,
		Selections:   selections,
	}
}

// TotalQuantity is the sum of quantities across all selections.
func (e *EventRegistration) TotalQuantity() int {
    total := 0
    for _, s := range e.Selections {
        total += s.Quantity
    }
    return total
}