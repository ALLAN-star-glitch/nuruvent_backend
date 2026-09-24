// internal/modules/payment/paymentdomain/order.go

package paymentdomain

import (
	"fmt"
	"time"
)

// Order is a commitment to pay for a set of tickets, tied to one
// registration. An order has at most one successful payment.
//
// All monetary values are in minor units. The currency is fixed at
// construction and must match the linked registration's currency.
//
// Lifecycle: see OrderStatus.
type Order struct {
	ID             string
	RegistrationID string

	// Identity: exactly one of UserID or GuestEmail is populated.
	UserID     string
	GuestEmail string

	Currency string
	Items    []OrderItem

	Subtotal      int64
	DiscountTotal int64
	TotalAmount   int64

	Status     OrderStatus
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	PaidAt     *time.Time
	CancelledAt *time.Time
}

// NewOrder constructs an order from a set of items. Status is always
// OrderStatusPending; the caller sets ExpiresAt via ttl.
func NewOrder(
	id string,
	registrationID string,
	userID string,
	guestEmail string,
	currency string,
	items []OrderItem,
	ttl time.Duration,
	now time.Time,
) (*Order, error) {
	if id == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if registrationID == "" {
		return nil, fmt.Errorf("registration id is required")
	}

	// Identity: exactly one of user / guest.
	if userID == "" && guestEmail == "" {
		return nil, ErrIdentityRequired
	}
	if userID != "" && guestEmail != "" {
		return nil, fmt.Errorf("%w: both user and guest identity set", ErrIdentityRequired)
	}

	if currency == "" {
		return nil, ErrInvalidCurrency
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("%w: at least one item is required", ErrInvalidAmount)
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("order ttl must be positive")
	}

	// Validate items and compute totals.
	var subtotal, discount int64
	for i, it := range items {
		// Each item was already validated by NewOrderItem, but callers
		// can construct an OrderItem literally. Re-validate here so the
		// Order cannot be built from malformed items.
		if err := validateItem(it); err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
		subtotal += it.Subtotal()
		discount += it.Discount
	}

	total := subtotal - discount
	if total < 0 {
		return nil, fmt.Errorf("%w: total cannot be negative", ErrInvalidAmount)
	}

	return &Order{
		ID:             id,
		RegistrationID: registrationID,
		UserID:         userID,
		GuestEmail:     guestEmail,
		Currency:       currency,
		Items:          items,
		Subtotal:       subtotal,
		DiscountTotal:  discount,
		TotalAmount:    total,
		Status:         OrderStatusPending,
		ExpiresAt:      now.Add(ttl),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// HydrateOrder reconstructs an order from persistence without
// re-validating. The DB is assumed to hold valid state.
func HydrateOrder(
	id, registrationID, userID, guestEmail, currency string,
	items []OrderItem,
	subtotal, discountTotal, totalAmount int64,
	status OrderStatus,
	expiresAt, createdAt, updatedAt time.Time,
	paidAt, cancelledAt *time.Time,
) *Order {
	return &Order{
		ID:             id,
		RegistrationID: registrationID,
		UserID:         userID,
		GuestEmail:     guestEmail,
		Currency:       currency,
		Items:          items,
		Subtotal:       subtotal,
		DiscountTotal:  discountTotal,
		TotalAmount:    totalAmount,
		Status:         status,
		ExpiresAt:      expiresAt,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		PaidAt:         paidAt,
		CancelledAt:    cancelledAt,
	}
}

// ============================================================
// LIFECYCLE
// ============================================================

// MarkPaid transitions the order to paid. Idempotent — calling on an
// already-paid order is a no-op and returns nil.
func (o *Order) MarkPaid(now time.Time) error {
	if o.Status == OrderStatusPaid {
		return nil
	}
	if !o.Status.CanTransitionTo(OrderStatusPaid) {
		return fmt.Errorf(
			"%w: order %s → %s",
			ErrInvalidStatusTransition, o.Status, OrderStatusPaid,
		)
	}
	o.Status = OrderStatusPaid
	o.PaidAt = &now
	o.UpdatedAt = now
	return nil
}

// Expire transitions the order to expired. Idempotent.
func (o *Order) Expire(now time.Time) error {
	if o.Status == OrderStatusExpired {
		return nil
	}
	if !o.Status.CanTransitionTo(OrderStatusExpired) {
		return fmt.Errorf(
			"%w: order %s → %s",
			ErrInvalidStatusTransition, o.Status, OrderStatusExpired,
		)
	}
	o.Status = OrderStatusExpired
	o.UpdatedAt = now
	return nil
}

// Cancel transitions the order to cancelled. Idempotent.
func (o *Order) Cancel(now time.Time) error {
	if o.Status == OrderStatusCancelled {
		return nil
	}
	if !o.Status.CanTransitionTo(OrderStatusCancelled) {
		return fmt.Errorf(
			"%w: order %s → %s",
			ErrInvalidStatusTransition, o.Status, OrderStatusCancelled,
		)
	}
	o.Status = OrderStatusCancelled
	o.CancelledAt = &now
	o.UpdatedAt = now
	return nil
}

// ============================================================
// QUERIES
// ============================================================

// IsPending reports whether the order is in pending state.
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsPaid reports whether the order has been paid.
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

// IsExpired reports whether the deadline has passed. Note: this checks
// the wall clock, not the status field. An order can be past its
// deadline without having been formally expired.
func (o *Order) IsExpired(now time.Time) bool {
	return !o.ExpiresAt.IsZero() && now.After(o.ExpiresAt)
}

// IsFinal reports whether the order's status is terminal.
func (o *Order) IsFinal() bool {
	return o.Status.IsFinal()
}

// IsGuest reports whether the order belongs to a guest (no user account).
func (o *Order) IsGuest() bool {
	return o.UserID == ""
}

// ItemCount returns the number of line items.
func (o *Order) ItemCount() int {
	return len(o.Items)
}

// TotalQuantity returns the sum of quantities across all items.
func (o *Order) TotalQuantity() int {
	n := 0
	for _, it := range o.Items {
		n += it.Quantity
	}
	return n
}

// ============================================================
// INTERNAL
// ============================================================

// validateItem re-checks an OrderItem's invariants. Callers can bypass
// NewOrderItem by constructing the struct literally, so we validate
// again when building the Order.
func validateItem(it OrderItem) error {
	if it.TicketTypeID == "" {
		return fmt.Errorf("%w: ticket type id is required", ErrInvalidAmount)
	}
	if it.Quantity < 1 {
		return fmt.Errorf("%w: quantity must be >= 1", ErrInvalidAmount)
	}
	if it.UnitPrice < 0 {
		return fmt.Errorf("%w: unit price must be >= 0", ErrInvalidAmount)
	}
	if it.Discount < 0 {
		return fmt.Errorf("%w: discount must be >= 0", ErrInvalidAmount)
	}
	subtotal := it.UnitPrice * int64(it.Quantity)
	if it.Discount > subtotal {
		return fmt.Errorf(
			"%w: discount (%d) exceeds subtotal (%d)",
			ErrInvalidAmount, it.Discount, subtotal,
		)
	}
	if it.LineTotal != subtotal-it.Discount {
		return fmt.Errorf(
			"%w: line total (%d) does not equal subtotal - discount (%d)",
			ErrInvalidAmount, it.LineTotal, subtotal-it.Discount,
		)
	}
	return nil
}