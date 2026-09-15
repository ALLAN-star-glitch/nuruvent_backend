package registrationdomain

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
}

// RegistrableResolver locates a Registrable by type and ID. Implemented at
// the composition root so registration doesn't import events directly.
type RegistrableResolver interface {
    Resolve(typeName string, id string) (Registrable, error)
}