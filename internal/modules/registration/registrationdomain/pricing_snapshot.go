package registrationdomain

import (
    "fmt"
    "time"
)

// PricingSnapshot freezes pricing at registration time so later edits to
// event/ticket prices don't rewrite history.
type PricingSnapshot struct {
    Currency      string
    Subtotal      int64
    DiscountTotal int64
    Total         int64
    SnapshotAt    time.Time
}

// NewPricingSnapshot computes totals from a set of selections.
func NewPricingSnapshot(currency string, selections []TicketSelection, now time.Time) (PricingSnapshot, error) {
    if currency == "" {
        return PricingSnapshot{}, fmt.Errorf("%w: currency is required", ErrPricingMismatch)
    }

    var subtotal, discount int64
    for _, s := range selections {
        if err := s.Validate(); err != nil {
            return PricingSnapshot{}, err
        }
        subtotal += s.Subtotal()
        discount += s.Discount
    }

    total := subtotal - discount
    if total < 0 {
        return PricingSnapshot{}, fmt.Errorf("%w: total is negative", ErrPricingMismatch)
    }

    return PricingSnapshot{
        Currency:      currency,
        Subtotal:      subtotal,
        DiscountTotal: discount,
        Total:         total,
        SnapshotAt:    now,
    }, nil
}