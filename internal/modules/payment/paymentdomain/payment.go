// internal/modules/payment/paymentdomain/payment.go

package paymentdomain

import (
	"fmt"
	"time"
)

// Payment is a single attempt to pay for an Order via a specific
// provider.
//
// An order may have multiple payments over its lifetime — initial
// attempt, retry after failure, retry after a timeout. At most one
// payment per order can end up in `succeeded` state; the rest are
// `failed` or `expired`.
//
// All monetary values are in minor units. Currency must match the
// parent Order's currency.
type Payment struct {
	ID      string
	OrderID string

	// Provider is a stable identifier for the gateway ("mpesa",
	// "card", "stub").
	Provider string

	// Method is how the user pays. Must agree with the provider.
	Method PaymentMethod

	Amount   int64
	Currency string

	Status PaymentStatus

	// IdempotencyKey is client-supplied. Together with OrderID it
	// forms a unique constraint — re-submitting the same key returns
	// the original payment.
	IdempotencyKey string

	// ProviderReference is the provider's own identifier for the
	// transaction. Populated when the payment succeeds (or, for some
	// providers, as soon as the request is accepted).
	ProviderReference string

	RedirectURL       string

	// FailureReason is set when Status == failed.
	FailureReason string

	InitiatedAt time.Time
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
	FailedAt    *time.Time
}

// NewPayment constructs a new Payment in pending state.
func NewPayment(
	id, orderID, provider string,
	method PaymentMethod,
	amount int64,
	currency string,
	idempotencyKey string,
	ttl time.Duration,
	now time.Time,
) (*Payment, error) {
	if id == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	if orderID == "" {
		return nil, fmt.Errorf("order id is required")
	}
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if !method.IsValid() {
		return nil, fmt.Errorf("invalid payment method: %q", method)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}
	if currency == "" {
		return nil, ErrInvalidCurrency
	}
	if idempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key is required")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("payment ttl must be positive")
	}

	return &Payment{
		ID:             id,
		OrderID:        orderID,
		Provider:       provider,
		Method:         method,
		Amount:         amount,
		Currency:       currency,
		Status:         PaymentStatusPending,
		IdempotencyKey: idempotencyKey,
		InitiatedAt:    now,
		ExpiresAt:      now.Add(ttl),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// HydratePayment reconstructs a payment from persistence without
// re-validating.
func HydratePayment(
	id, orderID, provider string,
	method PaymentMethod,
	amount int64,
	currency string,
	status PaymentStatus,
	idempotencyKey, providerReference, failureReason string, redirectURL string,
	initiatedAt time.Time,
	completedAt, failedAt *time.Time,
	expiresAt, createdAt, updatedAt time.Time,
) *Payment {
	return &Payment{
		ID:                id,
		OrderID:           orderID,
		Provider:          provider,
		Method:            method,
		Amount:            amount,
		Currency:          currency,
		Status:            status,
		IdempotencyKey:    idempotencyKey,
		ProviderReference: providerReference,
		RedirectURL:       redirectURL,
		FailureReason:     failureReason,
		InitiatedAt:       initiatedAt,
		CompletedAt:       completedAt,
		FailedAt:          failedAt,
		ExpiresAt:         expiresAt,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

// ============================================================
// LIFECYCLE
// ============================================================

// MarkSucceeded transitions the payment to succeeded. Idempotent.
// Requires a non-empty provider reference — this is what reconciliation
// and refunds key on.
func (p *Payment) MarkSucceeded(providerReference string, now time.Time) error {
	if p.Status == PaymentStatusSucceeded {
		return nil
	}
	if !p.Status.CanTransitionTo(PaymentStatusSucceeded) {
		return fmt.Errorf(
			"%w: payment %s → %s",
			ErrInvalidStatusTransition, p.Status, PaymentStatusSucceeded,
		)
	}
	if providerReference == "" {
		return fmt.Errorf("provider reference is required to mark succeeded")
	}

	p.Status = PaymentStatusSucceeded
	p.ProviderReference = providerReference
	p.CompletedAt = &now
	p.UpdatedAt = now
	return nil
}

// MarkFailed transitions the payment to failed. Idempotent.
func (p *Payment) MarkFailed(reason string, now time.Time) error {
	if p.Status == PaymentStatusFailed {
		return nil
	}
	if !p.Status.CanTransitionTo(PaymentStatusFailed) {
		return fmt.Errorf(
			"%w: payment %s → %s",
			ErrInvalidStatusTransition, p.Status, PaymentStatusFailed,
		)
	}
	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	p.FailedAt = &now
	p.UpdatedAt = now
	return nil
}

// MarkExpired transitions the payment to expired. Idempotent.
func (p *Payment) MarkExpired(now time.Time) error {
	if p.Status == PaymentStatusExpired {
		return nil
	}
	if !p.Status.CanTransitionTo(PaymentStatusExpired) {
		return fmt.Errorf(
			"%w: payment %s → %s",
			ErrInvalidStatusTransition, p.Status, PaymentStatusExpired,
		)
	}
	p.Status = PaymentStatusExpired
	p.UpdatedAt = now
	return nil
}

// MarkRefunded transitions the payment to refunded. Idempotent.
//
// Note: this is a terminal state on the Payment. Partial refunds do
// NOT transition the payment here — the payment stays `succeeded` and
// individual Refund records track the amounts.
func (p *Payment) MarkRefunded(now time.Time) error {
	if p.Status == PaymentStatusRefunded {
		return nil
	}
	if !p.Status.CanTransitionTo(PaymentStatusRefunded) {
		return fmt.Errorf(
			"%w: payment %s → %s",
			ErrInvalidStatusTransition, p.Status, PaymentStatusRefunded,
		)
	}
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = now
	return nil
}

// ============================================================
// QUERIES
// ============================================================

// IsPending reports whether the payment is in pending state.
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsSucceeded reports whether the payment has succeeded.
func (p *Payment) IsSucceeded() bool {
	return p.Status == PaymentStatusSucceeded
}

// IsFinal reports whether the payment's status is terminal.
func (p *Payment) IsFinal() bool {
	return p.Status.IsFinal()
}

// IsExpired reports whether the deadline has passed (by wall clock,
// not by status). A payment can be past its deadline but still pending
// until a background job transitions it.
func (p *Payment) IsExpired(now time.Time) bool {
	return !p.ExpiresAt.IsZero() && now.After(p.ExpiresAt)
}

// IsRetryable reports whether the payment can be attempted again with
// the same idempotency key. Failed and expired payments are retryable
// in that sense (a new payment will be created with the same key),
// but succeeded payments are not — retrying would double-charge.
func (p *Payment) IsRetryable() bool {
	switch p.Status {
	case PaymentStatusFailed, PaymentStatusExpired:
		return true
	}
	return false
}