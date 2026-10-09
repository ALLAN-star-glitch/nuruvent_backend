package registrationdomain

import (
	"context"
	"time"
)

// TicketPrice is the server-authoritative price for a single ticket type.
type TicketPrice struct {
    UnitPrice int64 // minor units
    Discount  int64 // minor units, may be 0
    Remaining int   // remaining quantity available
}

// Registrable is the port implemented by any entity a user can register for
// (event today, course later). The registration module owns this interface;
// the events module provides an adapter that satisfies it.
type Registrable interface {
    // Identity
    ID() string
    Type() string // "event" | "course"

    DisplayName() string   // human-readable name, for emails

    // Capacity
    Capacity() int
    CurrentRegistrations() int

    // Lifecycle
    IsRegistrationOpen() bool
    RequiresPayment() bool

    // Counter maintenance — the registrable owns its cached count
    AdjustCount(delta int) error

    // Ticket availability & pricing, keyed by ticket_type_id
    TicketAvailability() (map[string]int, error)
    TicketPricing() (map[string]TicketPrice, error)

     // Slug returns the URL-friendly identifier of the underlying
    // entity, for building public links.
    Slug() string

    // StartDate returns the scheduled start of the underlying entity.
    StartDate() time.Time
}

// RegistrableResolver locates a Registrable by type and ID. Implemented at
// the composition root so registration doesn't import events directly.
type RegistrableResolver interface {
    Resolve(ctx context.Context, typeName string, id string) (Registrable, error)
}

