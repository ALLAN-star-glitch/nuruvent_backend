// internal/modules/payment/paymentdomain/refund.go

package paymentdomain

import (
	"fmt"
	"time"
)

// Refund records a full or partial reversal of a successful Payment.
//
// Multiple refunds can exist per payment, as long as the sum of their
// amounts does not exceed the payment's amount. Each refund is
// independently tracked through its own lifecycle.
//
// A refund transitions through the same states as a payment attempt:
//
//	pending   → succeeded | failed
//	succeeded → (final)
//	failed    → (final)
//
// A refund that fails does not affect the payment — the payment stays
// in whatever state it was before the attempt. Failed refunds are
// retryable via a new refund record.
type Refund struct {
	ID        string
	PaymentID string

	Amount   int64
	Currency string

	// Reason is a human-readable explanation ("customer request",
	// "event cancelled by organizer", etc.).
	Reason string

	// ActorID is the user who initiated the refund (organizer, admin).
	ActorID string

	// ProviderReference is the provider's reference for the refund
	// transaction, set on success.
	ProviderReference string

	Status        PaymentStatus // reuse: pending | succeeded | failed
	FailureReason string

	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// NewRefund constructs a pending refund.
func NewRefund(
	id, paymentID string,
	amount int64,
	currency string,
	reason, actorID string,
	now time.Time,
) (*Refund, error) {
	if id == "" {
		return nil, fmt.Errorf("refund id is required")
	}
	if paymentID == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: refund amount must be positive", ErrInvalidAmount)
	}
	if currency == "" {
		return nil, ErrInvalidCurrency
	}
	if actorID == "" {
		return nil, fmt.Errorf("actor id is required")
	}

	return &Refund{
		ID:        id,
		PaymentID: paymentID,
		Amount:    amount,
		Currency:  currency,
		Reason:    reason,
		ActorID:   actorID,
		Status:    PaymentStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// HydrateRefund reconstructs a refund from persistence.
func HydrateRefund(
	id, paymentID string,
	amount int64,
	currency, reason, actorID, providerReference string,
	status PaymentStatus,
	failureReason string,
	createdAt, updatedAt time.Time,
	completedAt *time.Time,
) *Refund {
	return &Refund{
		ID:                id,
		PaymentID:         paymentID,
		Amount:            amount,
		Currency:          currency,
		Reason:            reason,
		ActorID:           actorID,
		ProviderReference: providerReference,
		Status:            status,
		FailureReason:     failureReason,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		CompletedAt:       completedAt,
	}
}

// ============================================================
// LIFECYCLE
// ============================================================

// MarkSucceeded transitions the refund to succeeded. Idempotent.
// Requires a non-empty provider reference.
func (r *Refund) MarkSucceeded(providerReference string, now time.Time) error {
	if r.Status == PaymentStatusSucceeded {
		return nil
	}
	if !r.Status.CanTransitionTo(PaymentStatusSucceeded) {
		return fmt.Errorf(
			"%w: refund %s → %s",
			ErrInvalidStatusTransition, r.Status, PaymentStatusSucceeded,
		)
	}
	if providerReference == "" {
		return fmt.Errorf("provider reference is required to mark refund succeeded")
	}

	r.Status = PaymentStatusSucceeded
	r.ProviderReference = providerReference
	r.CompletedAt = &now
	r.UpdatedAt = now
	return nil
}

// MarkFailed transitions the refund to failed. Idempotent.
func (r *Refund) MarkFailed(reason string, now time.Time) error {
	if r.Status == PaymentStatusFailed {
		return nil
	}
	if !r.Status.CanTransitionTo(PaymentStatusFailed) {
		return fmt.Errorf(
			"%w: refund %s → %s",
			ErrInvalidStatusTransition, r.Status, PaymentStatusFailed,
		)
	}
	r.Status = PaymentStatusFailed
	r.FailureReason = reason
	r.UpdatedAt = now
	return nil
}

// ============================================================
// QUERIES
// ============================================================

// IsPending reports whether the refund is in pending state.
func (r *Refund) IsPending() bool {
	return r.Status == PaymentStatusPending
}

// IsSucceeded reports whether the refund has succeeded.
func (r *Refund) IsSucceeded() bool {
	return r.Status == PaymentStatusSucceeded
}

// IsFailed reports whether the refund has failed.
func (r *Refund) IsFailed() bool {
	return r.Status == PaymentStatusFailed
}

// IsFinal reports whether the refund's status is terminal.
func (r *Refund) IsFinal() bool {
	return r.Status == PaymentStatusSucceeded || r.Status == PaymentStatusFailed
}

// ============================================================
// VALIDATION
// ============================================================

// ValidateAgainstPayment ensures a refund is legal against the given
// payment and any prior refunds.
//
// Rules:
//   - The payment must be in succeeded state
//   - Currencies must match
//   - The sum of prior pending/succeeded refunds plus this refund must
//     not exceed the payment amount
//
// Failed refunds do not count against the total (they represent
// attempts that did not move money).
func (r *Refund) ValidateAgainstPayment(
	payment *Payment,
	priorRefunds []*Refund,
) error {
	if payment == nil {
		return ErrPaymentNotFound
	}
	if !payment.IsSucceeded() {
		return ErrPaymentNotSucceeded
	}
	if r.Currency != payment.Currency {
		return fmt.Errorf(
			"%w: refund currency %q does not match payment currency %q",
			ErrInvalidCurrency, r.Currency, payment.Currency,
		)
	}
	if r.Amount <= 0 {
		return fmt.Errorf("%w: refund amount must be positive", ErrInvalidAmount)
	}

	var priorTotal int64
	for _, p := range priorRefunds {
		if p == nil {
			continue
		}
		// Only pending and succeeded refunds consume the refundable budget.
		switch p.Status {
		case PaymentStatusPending, PaymentStatusSucceeded:
			priorTotal += p.Amount
		}
	}

	if priorTotal+r.Amount > payment.Amount {
		return fmt.Errorf(
			"%w: prior refunds %d + new refund %d exceed paid amount %d",
			ErrRefundExceedsPayment, priorTotal, r.Amount, payment.Amount,
		)
	}
	return nil
}

// ============================================================
// AGGREGATES
// ============================================================

// TotalRefunded computes the sum of succeeded refunds in a slice.
func TotalRefunded(refunds []*Refund) int64 {
	var total int64
	for _, r := range refunds {
		if r != nil && r.Status == PaymentStatusSucceeded {
			total += r.Amount
		}
	}
	return total
}

// RefundableRemaining computes how much can still be refunded on a
// payment given prior refunds.
func RefundableRemaining(payment *Payment, priorRefunds []*Refund) int64 {
	if payment == nil || !payment.IsSucceeded() {
		return 0
	}

	var consumed int64
	for _, r := range priorRefunds {
		if r == nil {
			continue
		}
		switch r.Status {
		case PaymentStatusPending, PaymentStatusSucceeded:
			consumed += r.Amount
		}
	}

	remaining := payment.Amount - consumed
	if remaining < 0 {
		return 0
	}
	return remaining
}