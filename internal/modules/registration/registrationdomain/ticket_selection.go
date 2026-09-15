package registrationdomain

import (
    "fmt"
    "strings"
)

// TicketSelection is a value object: a ticket type and the quantity chosen.
type TicketSelection struct {
    TicketTypeID string
    Quantity     int
    UnitPrice    int64 // minor units, frozen at registration time
    Discount     int64 // minor units, may be 0
}

// Subtotal is the line amount before discount.
func (t TicketSelection) Subtotal() int64 {
    return t.UnitPrice * int64(t.Quantity)
}

// LineTotal is the line amount after discount.
func (t TicketSelection) LineTotal() int64 {
    return t.Subtotal() - t.Discount
}

// Validate enforces invariants on a single selection.
func (t TicketSelection) Validate() error {
    if strings.TrimSpace(t.TicketTypeID) == "" {
        return fmt.Errorf("%w: ticket type id is empty", ErrTicketUnavailable)
    }
    if t.Quantity < 1 {
        return fmt.Errorf("%w: quantity must be >= 1", ErrTicketUnavailable)
    }
    if t.UnitPrice < 0 {
        return fmt.Errorf("%w: unit price must be >= 0", ErrPricingMismatch)
    }
    if t.Discount < 0 {
        return fmt.Errorf("%w: discount must be >= 0", ErrPricingMismatch)
    }
    if t.Discount > t.Subtotal() {
        return fmt.Errorf("%w: discount exceeds subtotal", ErrPricingMismatch)
    }
    return nil
}