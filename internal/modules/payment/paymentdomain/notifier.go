// internal/modules/payment/paymentdomain/notifier.go

package paymentdomain

import "context"

// Notifier is the outbound port for payment-related notifications.
//
// Implementations (typically a thin adapter that enqueues jobs onto a
// shared queue) MUST NOT block. Notifications are best-effort side
// effects — a notification failure must never fail the caller's
// operation. The service layer logs failures and moves on.
//
// All methods receive the entities involved so the adapter can render
// rich messages without additional lookups. If the adapter needs data
// not present on the entity (e.g. the event name), it is responsible
// for fetching it — the domain layer does not know about other
// modules' entities.
type Notifier interface {
	// PaymentInitiated fires when a payment attempt is created and the
	// provider has accepted the initiation request. Used for messages
	// like "We've sent an M-Pesa prompt to your phone".
	PaymentInitiated(ctx context.Context, p *Payment, o *Order) error

	// PaymentSucceeded fires when a payment transitions to succeeded.
	// This is the moment money has moved and the registration should
	// confirm.
	PaymentSucceeded(ctx context.Context, p *Payment, o *Order) error

	// PaymentFailed fires when a payment transitions to failed. It
	// carries the failure reason for the message.
	PaymentFailed(ctx context.Context, p *Payment, o *Order) error

	// PaymentExpired fires when a payment expires without succeeding.
	// Semantically distinct from failure — the user didn't complete
	// the flow in time, rather than the provider rejecting it.
	PaymentExpired(ctx context.Context, p *Payment, o *Order) error

	// RefundIssued fires when a refund transitions to succeeded.
	// Partial refunds and full refunds share this method — the
	// adapter can distinguish based on the amounts.
	RefundIssued(ctx context.Context, r *Refund, p *Payment) error
}