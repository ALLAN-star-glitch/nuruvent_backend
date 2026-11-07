// internal/modules/payment/paymentdomain/order.go

package paymentdomain

import (
	"fmt"
	"time"
)

// OrderKind discriminates between the two kinds of orders the payment
// module handles.
type OrderKind string

const (
	OrderKindRegistration     OrderKind = "registration"
	OrderKindCertificate      OrderKind = "certificate"       // legacy — attendee-pays, dormant
	OrderKindCertificateBatch OrderKind = "certificate_batch" // NEW — organizer-pays
)

func (k OrderKind) IsValid() bool {
	switch k {
	case OrderKindRegistration, OrderKindCertificate, OrderKindCertificateBatch:
		return true
	default:
		return false
	}
}

// String returns the raw kind.
func (k OrderKind) String() string { return string(k) }

// Order is a commitment to pay for a set of tickets, tied to one
// registration. An order has at most one successful payment.
//
// All monetary values are in minor units. The currency is fixed at
// construction and must match the linked registration's currency.
//
// Lifecycle: see OrderStatus.
type Order struct {
	ID             string
	Kind           OrderKind
	RegistrationID string
	CertificateCtx *CertificateContext

	// Certificate batch order (organizer-pays)
	BatchCtx           *BatchContext
	BatchAttendeeCount int

	// Identity: populated with UserID (shadow or real) and GuestEmail
	// if the attendee is a guest.
	UserID     string
	GuestEmail string

	Currency string
	Items    []OrderItem

	Subtotal      int64
	DiscountTotal int64
	TotalAmount   int64

	Status      OrderStatus
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PaidAt      *time.Time
	CancelledAt *time.Time

	// ------------------------------------------------------------
	// Billing — frozen at order creation.
	// ------------------------------------------------------------
	BilledAccountID string
	PlatformFeeRate float64

	// ProcessingFeeRate snapshots the payment provider's cut at the
	// moment the payment method is known (M-Pesa 1.5%, local card 2.9%,
	// intl card 3.8%). Zero for certificate-batch orders (100% to
	// Nuruvent) and for orders whose payment method has not yet been
	// resolved at hydration time.
	//
	// Set once via SetProcessingFeeRate; never changes afterwards for
	// the same order.
	ProcessingFeeRate float64

	SettledAt *time.Time
	PayoutRef string

	// ------------------------------------------------------------
	// Payment — populated on read.
	// ------------------------------------------------------------
	//
	// This is not persisted on the order row. It is hydrated by the
	// read-side repository from the latest payment attached to the
	// order. Guests use this to poll GET /orders/:id instead of
	// needing access to the protected GET /payments/:id endpoint.
	//
	// Empty string when no payment has been initiated yet.
	PaymentStatus PaymentStatus
}

// NewOrder constructs an order from a set of items. Status is always
// OrderStatusPending; the caller sets ExpiresAt via ttl.
//
// billedAccountID, platformFeeRate, and processingFeeRate are
// snapshotted onto the order here and never change afterwards.
//
// processingFeeRate may be 0 if the payment method is not yet known;
// the caller (initiate / webhook) will set it later via
// SetProcessingFeeRate.
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
	processingFeeRate float64,
) (*Order, error) {
	if id == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if registrationID == "" {
		return nil, fmt.Errorf("registration id is required")
	}

	// Identity: must have at least one identity source.
	// Both may be set for shadow users (guests who get a lightweight
	// user record for the attendance/Zoom pipeline, but whose contact
	// email is still tracked on the order for notifications).
	if userID == "" && guestEmail == "" {
		return nil, ErrIdentityRequired
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
	if processingFeeRate < 0 || processingFeeRate > 1 {
		return nil, fmt.Errorf(
			"processing fee rate must be in [0, 1], got %v",
			processingFeeRate,
		)
	}

	// Validate items and compute totals.
	var subtotal, discount int64
	for i, it := range items {
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
		ID:                id,
		Kind:              OrderKindRegistration,
		RegistrationID:    registrationID,
		UserID:            userID,
		GuestEmail:        guestEmail,
		Currency:          currency,
		Items:             items,
		Subtotal:          subtotal,
		DiscountTotal:     discount,
		TotalAmount:       total,
		Status:            OrderStatusPending,
		ExpiresAt:         now.Add(ttl),
		CreatedAt:         now,
		UpdatedAt:         now,
		BilledAccountID:   billedAccountID,
		PlatformFeeRate:   platformFeeRate,
		ProcessingFeeRate: processingFeeRate,
	}, nil
}

// NewCertificateOrder constructs a pending order for a certificate
// purchase. Certificates are single-amount purchases — no items.
func NewCertificateOrder(
	id string,
	certCtx CertificateContext,
	userID string,
	guestEmail string,
	currency string,
	amountMinor int64,
	ttl time.Duration,
	now time.Time,
	billedAccountID string,
	platformFeeRate float64,
	processingFeeRate float64,
) (*Order, error) {
	if id == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if certCtx.ContextType == "" {
		return nil, fmt.Errorf("certificate context_type is required")
	}
	if certCtx.ContextID == "" {
		return nil, fmt.Errorf("certificate context_id is required")
	}
	if certCtx.RecipientName == "" {
		return nil, fmt.Errorf("certificate recipient name is required")
	}
	if certCtx.RecipientEmail == "" {
		return nil, fmt.Errorf("certificate recipient email is required")
	}

	// Identity: must have at least one identity source.
	if userID == "" && guestEmail == "" {
		return nil, ErrIdentityRequired
	}

	if currency == "" {
		return nil, ErrInvalidCurrency
	}
	if amountMinor <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("order ttl must be positive")
	}
	if billedAccountID == "" {
		return nil, fmt.Errorf("billed account id is required")
	}
	if platformFeeRate < 0 || platformFeeRate > 1 {
		return nil, fmt.Errorf(
			"platform fee rate must be in [0, 1], got %v",
			platformFeeRate,
		)
	}
	if processingFeeRate < 0 || processingFeeRate > 1 {
		return nil, fmt.Errorf(
			"processing fee rate must be in [0, 1], got %v",
			processingFeeRate,
		)
	}

	return &Order{
		ID:                id,
		Kind:              OrderKindCertificate,
		RegistrationID:    "",
		CertificateCtx:    &certCtx,
		UserID:            userID,
		GuestEmail:        guestEmail,
		Currency:          currency,
		Items:             nil, // certificate orders have no items
		Subtotal:          amountMinor,
		DiscountTotal:     0,
		TotalAmount:       amountMinor,
		Status:            OrderStatusPending,
		ExpiresAt:         now.Add(ttl),
		CreatedAt:         now,
		UpdatedAt:         now,
		BilledAccountID:   billedAccountID,
		PlatformFeeRate:   platformFeeRate,
		ProcessingFeeRate: processingFeeRate,
	}, nil
}

// NewCertificateBatchOrder constructs a pending order for a
// certificate batch. The organizer pays; attendees receive
// certificates free once the order is settled.
//
// Batch orders are 100% Nuruvent — no processing fee is applied
// because the batch is purchased via the organizer's own billing
// arrangement, not routed through Paystack per attendee.
func NewCertificateBatchOrder(
	id string,
	batchCtx BatchContext,
	currency string,
	totalAmountMinor int64,
	ttl time.Duration,
	now time.Time,
	billedAccountID string,
) (*Order, error) {
	if id == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if batchCtx.EventID == "" {
		return nil, fmt.Errorf("batch context event_id is required")
	}
	if batchCtx.AccountID == "" {
		return nil, fmt.Errorf("batch context account_id is required")
	}
	if batchCtx.TemplateID == "" {
		return nil, fmt.Errorf("batch context template_id is required")
	}
	if batchCtx.AttendeeCount <= 0 {
		return nil, fmt.Errorf("%w: attendee count must be positive", ErrInvalidAmount)
	}
	if len(batchCtx.Recipients) != batchCtx.AttendeeCount {
		return nil, fmt.Errorf("batch context recipient count (%d) does not match attendee_count (%d)",
			len(batchCtx.Recipients), batchCtx.AttendeeCount)
	}
	if totalAmountMinor <= 0 {
		return nil, fmt.Errorf("%w: total amount must be positive", ErrInvalidAmount)
	}
	if currency == "" {
		return nil, ErrInvalidCurrency
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("order ttl must be positive")
	}
	if billedAccountID == "" {
		return nil, fmt.Errorf("billed account id is required")
	}

	return &Order{
		ID:                 id,
		Kind:               OrderKindCertificateBatch,
		RegistrationID:     "",
		CertificateCtx:     nil,
		BatchCtx:           &batchCtx,
		BatchAttendeeCount: batchCtx.AttendeeCount,
		UserID:             "",
		GuestEmail:         "",
		Currency:           currency,
		Items:              nil,
		Subtotal:           totalAmountMinor,
		DiscountTotal:      0,
		TotalAmount:        totalAmountMinor,
		Status:             OrderStatusPending,
		ExpiresAt:          now.Add(ttl),
		CreatedAt:          now,
		UpdatedAt:          now,
		BilledAccountID:    billedAccountID,
		PlatformFeeRate:    0, // 100% to Nuruvent
		ProcessingFeeRate:  0, // batch orders skip processing fee entirely
	}, nil
}

func HydrateOrder(
	id string,
	kind OrderKind,
	certCtx *CertificateContext,
	batchCtx *BatchContext,
	batchAttendeeCount int,
	registrationID, userID, guestEmail, currency string,
	items []OrderItem,
	subtotal, discountTotal, totalAmount int64,
	status OrderStatus,
	expiresAt, createdAt, updatedAt time.Time,
	paidAt, cancelledAt *time.Time,
	billedAccountID string,
	platformFeeRate float64,
	processingFeeRate float64,
	settledAt *time.Time,
	payoutRef string,
) *Order {
	return &Order{
		ID:                 id,
		Kind:               kind,
		RegistrationID:     registrationID,
		CertificateCtx:     certCtx,
		BatchCtx:           batchCtx,
		BatchAttendeeCount: batchAttendeeCount,
		UserID:             userID,
		GuestEmail:         guestEmail,
		Currency:           currency,
		Items:              items,
		Subtotal:           subtotal,
		DiscountTotal:      discountTotal,
		TotalAmount:        totalAmount,
		Status:             status,
		ExpiresAt:          expiresAt,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		PaidAt:             paidAt,
		CancelledAt:        cancelledAt,
		BilledAccountID:    billedAccountID,
		PlatformFeeRate:    platformFeeRate,
		ProcessingFeeRate:  processingFeeRate,
		SettledAt:          settledAt,
		PayoutRef:          payoutRef,
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

// IsGuest reports whether the order belongs to a guest (i.e., the
// attendee had no real Nuruvent account at checkout time, even if a
// shadow user record was created for them).
func (o *Order) IsGuest() bool {
	return o.GuestEmail != ""
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
// units.
func (o *Order) PlatformFee() int64 {
	if o.Kind == OrderKindCertificateBatch {
		return o.TotalAmount
	}
	return int64(float64(o.TotalAmount) * o.PlatformFeeRate)
}

// ProcessingFeeAmount returns the payment provider's cut (M-Pesa,
// card, etc.) of the order total, in minor units.
//
// Truncates toward zero, matching PlatformFee. Certificate-batch
// orders are 100% to Nuruvent and carry no processing fee.
//
// A ProcessingFeeRate of 0 is valid and common — it means the method
// has not yet been resolved (order created but payment not initiated),
// or that the order kind does not carry a processing fee (batch).
func (o *Order) ProcessingFeeAmount() int64 {
	if o.Kind == OrderKindCertificateBatch {
		return 0
	}
	if o.ProcessingFeeRate <= 0 {
		return 0
	}
	return int64(float64(o.TotalAmount) * o.ProcessingFeeRate)
}

// SetProcessingFeeRate snapshots the processing fee rate onto the
// order. Called once the payment method is resolved (at initiate time
// or in the webhook handler). The rate never changes after the first
// successful call.
//
// Pass rate = 0 to explicitly record "no processing fee" (e.g. batch
// orders or a free order); passing a negative or >1 value is a
// programming error.
func (o *Order) SetProcessingFeeRate(rate float64) error {
	if o.Kind == OrderKindCertificateBatch {
		// Batch orders are 100% to Nuruvent. Ignore attempts to set a
		// processing rate on them to keep the invariant simple.
		return nil
	}
	if rate < 0 || rate > 1 {
		return fmt.Errorf(
			"processing fee rate must be in [0, 1], got %v",
			rate,
		)
	}
	o.ProcessingFeeRate = rate
	o.UpdatedAt = time.Now()
	return nil
}

// IsCertificateBatch reports whether this is an organizer-pays
// certificate batch order.
func (o *Order) IsCertificateBatch() bool {
	return o.Kind == OrderKindCertificateBatch
}

// NetToOrganizer returns the amount owed to the billed account after
// both the platform fee and the payment-processing fee are deducted,
// in minor units. Never negative.
//
// On a KES 1,000 M-Pesa ticket with 4.5% platform fee and 1.5%
// processing fee, this returns 940 (minor units: 94000).
func (o *Order) NetToOrganizer() int64 {
	net := o.TotalAmount - o.PlatformFee() - o.ProcessingFeeAmount()
	if net < 0 {
		return 0
	}
	return net
}

// IsSettled reports whether the organizer's net earnings for this
// order have been disbursed.
func (o *Order) IsSettled() bool {
	return o.SettledAt != nil
}

// ============================================================
// INTERNAL
// ============================================================

// validateItem re-checks an OrderItem's invariants.
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

// IsCertificate reports whether this is a certificate order.
func (o *Order) IsCertificate() bool {
	return o.Kind == OrderKindCertificate
}

// IsRegistration reports whether this is a registration order.
func (o *Order) IsRegistration() bool {
	return o.Kind == OrderKindRegistration
}