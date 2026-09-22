// internal/modules/payment/paymentdomain/repository.go

package paymentdomain

import (
	"context"
	"time"
)

// ============================================================
// ORDER REPOSITORY
// ============================================================

// OrderRepository persists Order entities.
type OrderRepository interface {
	Create(ctx context.Context, o *Order) error
	Update(ctx context.Context, o *Order) error
	FindByID(ctx context.Context, id string) (*Order, error)

	// FindActiveByRegistration returns the pending order (if any) for a
	// registration. "Active" means status = pending. Used to enforce
	// the one-active-order-per-registration rule.
	//
	// Returns ErrOrderNotFound when none exists.
	FindActiveByRegistration(ctx context.Context, registrationID string) (*Order, error)

	// FindExpired returns orders that are still pending but past their
	// deadline. Used by the expiry sweeper job.
	FindExpired(ctx context.Context, before time.Time, limit int) ([]*Order, error)
}

// ============================================================
// PAYMENT REPOSITORY
// ============================================================

// PaymentRepository persists Payment entities.
type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) error
	Update(ctx context.Context, p *Payment) error
	FindByID(ctx context.Context, id string) (*Payment, error)

	// FindByProviderReference looks up a payment by the provider's
	// transaction ID. This is the primary lookup for webhooks and
	// reconciliation.
	//
	// Returns ErrPaymentNotFound when none exists.
	FindByProviderReference(
		ctx context.Context,
		provider, reference string,
	) (*Payment, error)

	// FindByIdempotencyKey enforces the (order_id, idempotency_key)
	// uniqueness contract. Returns the existing payment if one was
	// already created with the same key.
	//
	// Returns ErrPaymentNotFound when none exists.
	FindByIdempotencyKey(
		ctx context.Context,
		orderID, key string,
	) (*Payment, error)

	// FindPendingExpired returns pending payments past their deadline,
	// for the expiry sweeper.
	FindPendingExpired(ctx context.Context, before time.Time, limit int) ([]*Payment, error)

	// FindByOrderID returns all payments (any state) for an order.
	// Used for retry history and reporting.
	FindByOrderID(ctx context.Context, orderID string) ([]*Payment, error)
}

// ============================================================
// REFUND REPOSITORY
// ============================================================

// RefundRepository persists Refund entities.
type RefundRepository interface {
	Create(ctx context.Context, r *Refund) error
	Update(ctx context.Context, r *Refund) error
	FindByID(ctx context.Context, id string) (*Refund, error)

	// ListByPayment returns all refunds for a payment, newest first.
	// Used to compute the remaining refundable balance and for
	// auditing.
	ListByPayment(ctx context.Context, paymentID string) ([]*Refund, error)

	// FindPendingBefore returns pending refunds older than the given
	// time. Used by a reconciliation job to time out refunds that
	// never completed.
	FindPendingBefore(ctx context.Context, before time.Time, limit int) ([]*Refund, error)
}

// ============================================================
// WEBHOOK EVENT REPOSITORY
// ============================================================

// WebhookEventRepository persists raw webhook deliveries.
type WebhookEventRepository interface {
	// RecordIfNew inserts the event. Returns ErrDuplicateWebhook if
	// the (provider, provider_event_id) pair already exists — the
	// caller should treat that as a no-op, not an error.
	//
	// Implementations must enforce this uniqueness at the database
	// level, not via application checks — a race between two
	// concurrent deliveries must resolve to one insert and one
	// duplicate.
	RecordIfNew(ctx context.Context, e *WebhookEvent) error

	// Update persists changes to a webhook event (processing status,
	// errors).
	Update(ctx context.Context, e *WebhookEvent) error

	// FindByID returns a webhook event by its internal ID.
	FindByID(ctx context.Context, id string) (*WebhookEvent, error)

	// FindUnprocessed returns webhook events that have not been
	// successfully processed. Includes events that failed (retryable)
	// and events that were never attempted.
	//
	// The limit caps batch size; implementations must order by
	// ReceivedAt ascending so oldest events are processed first.
	FindUnprocessed(ctx context.Context, limit int) ([]*WebhookEvent, error)
}

// ============================================================
// UNIT OF WORK
// ============================================================

// UnitOfWork executes multiple repository operations inside a single
// database transaction.
//
// Use it when an operation writes to more than one aggregate and the
// writes must succeed or fail together. Examples:
//
//   - Create order + payment
//   - Confirm payment + mark order paid
//   - Refund payment + update payment status
//
// Do NOT include in the transaction:
//
//   - Provider network calls
//   - Cross-module calls (registration.ConfirmRegistration)
//   - Notification dispatch
//   - Metric emission
//
// Those are side effects that run after commit. If they fail, the
// source-of-truth write must not roll back.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(Repositories) error) error
}

// Repositories is the set of repositories bound to a single
// transaction. All operations through these repositories share the
// same database connection and commit or roll back together.
type Repositories struct {
	Orders   OrderRepository
	Payments PaymentRepository
	Refunds  RefundRepository
	Webhooks WebhookEventRepository
}