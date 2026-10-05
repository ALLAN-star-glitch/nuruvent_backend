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

	// ------------------------------------------------------------
	// Billing — frozen at order creation.
	//
	// BilledAccountID is the account that receives the net proceeds
	// of any payment for this order. Resolved once from the event's
	// ownership chain (registration → event → team → account) and
	// never updated, so later changes to event ownership do not
	// rewrite historical attribution.
	//
	// PlatformFeeRate is Nuruvent's cut, snapshotted at creation so
	// historical reports are immune to later pricing changes.
	//
	// SettledAt and PayoutRef are written only by the payout workflow
	// (manual today; automated later). They stay nil/empty until the
	// organizer's net earnings have been disbursed.
	// ------------------------------------------------------------
	BilledAccountID string
	PlatformFeeRate float64
	SettledAt       *time.Time
	PayoutRef       string
}

// NewOrder constructs an order from a set of items. Status is always
// OrderStatusPending; the caller sets ExpiresAt via ttl.
//
// billedAccountID and platformFeeRate are snapshotted onto the order
// here and never change afterwards.
func NewOrder(
	id string,
	registrationID string,
	userID string,
	guestEmail string,
	currency string,
	items []OrderItem,
	ttl time.Duration,
	now time.Time,
	billedAccountID string,
	platformFeeRate float64,
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

	// Billing invariants.
	if billedAccountID == "" {
		return nil, fmt.Errorf("billed account id is required")
	}
	if platformFeeRate < 0 || platformFeeRate > 1 {
		return nil, fmt.Errorf(
			"platform fee rate must be in [0, 1], got %v",
			platformFeeRate,
		)
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
		ID:              id,
		RegistrationID:  registrationID,
		UserID:          userID,
		GuestEmail:      guestEmail,
		Currency:        currency,
		Items:           items,
		Subtotal:        subtotal,
		DiscountTotal:   discount,
		TotalAmount:     total,
		Status:          OrderStatusPending,
		ExpiresAt:       now.Add(ttl),
		CreatedAt:       now,
		UpdatedAt:       now,
		BilledAccountID: billedAccountID,
		PlatformFeeRate: platformFeeRate,
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
	billedAccountID string,
	platformFeeRate float64,
	settledAt *time.Time,
	payoutRef string,
) *Order {
	return &Order{
		ID:              id,
		RegistrationID:  registrationID,
		UserID:          userID,
		GuestEmail:      guestEmail,
		Currency:        currency,
		Items:           items,
		Subtotal:        subtotal,
		DiscountTotal:   discountTotal,
		TotalAmount:     totalAmount,
		Status:          status,
		ExpiresAt:       expiresAt,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		PaidAt:          paidAt,
		CancelledAt:     cancelledAt,
		BilledAccountID: billedAccountID,
		PlatformFeeRate: platformFeeRate,
		SettledAt:       settledAt,
		PayoutRef:       payoutRef,
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
// BILLING QUERIES
// ============================================================

// PlatformFee returns Nuruvent's cut of the order total, in minor
// units, using the rate snapshotted at creation.
//
// The multiplication promotes to float64, and the conversion back to
// int64 truncates toward zero. That matches how the SQL aggregates
// compute the same value — no drift between in-memory and DB.
func (o *Order) PlatformFee() int64 {
	return int64(float64(o.TotalAmount) * o.PlatformFeeRate)
}

// NetToOrganizer returns the amount owed to the billed account before
// payment-processing fees, in minor units. Processing fees are
// recorded on the payment (they depend on the method chosen at
// checkout) and are subtracted separately when computing the final
// disbursement.
func (o *Order) NetToOrganizer() int64 {
	return o.TotalAmount - o.PlatformFee()
}

// IsSettled reports whether the organizer's net earnings for this
// order have been disbursed.
func (o *Order) IsSettled() bool {
	return o.SettledAt != nil
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