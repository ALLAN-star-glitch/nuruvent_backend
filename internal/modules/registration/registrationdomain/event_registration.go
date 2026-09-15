package registrationdomain

import "time"

// EventRegistration is the event-specific extension of Registration.
type EventRegistration struct {
    Registration  *Registration
    EventID       string
    Selections    []TicketSelection
    Pricing       PricingSnapshot
    AttendedAt    *time.Time // reserved for future attendance
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
        EventID:      eventID,
        Selections:   selections,
        Pricing:      pricing,
    }, nil
}

// TotalQuantity is the sum of quantities across all selections.
func (e *EventRegistration) TotalQuantity() int {
    total := 0
    for _, s := range e.Selections {
        total += s.Quantity
    }
    return total
}