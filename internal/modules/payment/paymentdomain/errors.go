// internal/modules/payment/paymentdomain/errors.go

package paymentdomain

import "errors"

// Sentinel errors for the payment module. Callers use errors.Is to
// classify errors across layers (service → delivery) without coupling
// to concrete error types.
//
// Every error returned by the domain layer is either one of these or
// wraps one with fmt.Errorf("%w: ...").
var (
	// ============================================================
	// ORDER
	// ============================================================

	// ErrOrderNotFound is returned when an order does not exist.
	ErrOrderNotFound = errors.New("order not found")

	// ErrOrderNotPending is returned when an operation requires the
	// order to be in pending state but it isn't.
	ErrOrderNotPending = errors.New("order is not in a pending state")

	// ErrOrderExpired is returned when the order's deadline has passed.
	ErrOrderExpired = errors.New("order has expired")

	// ErrDuplicateOrder is returned when attempting to create a second
	// active order for a registration that already has one.
	ErrDuplicateOrder = errors.New("an active order already exists for this registration")

	// ErrOrderAlreadyPaid is returned when attempting to pay an order
	// that has already been paid.
	ErrOrderAlreadyPaid = errors.New("order is already paid")

	// ============================================================
	// PAYMENT
	// ============================================================

	// ErrPaymentNotFound is returned when a payment does not exist.
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrPaymentAlreadyInitiated is returned when a second payment
	// attempt is made against an order that already has an active one.
	ErrPaymentAlreadyInitiated = errors.New("payment has already been initiated for this order")

	// ErrPaymentAlreadySucceeded is returned when an operation assumes
	// the payment has not succeeded but it has.
	ErrPaymentAlreadySucceeded = errors.New("payment has already succeeded")

	// ErrPaymentNotPending is returned when an operation requires the
	// payment to be in pending state but it isn't.
	ErrPaymentNotPending = errors.New("payment is not pending")

	// ErrPaymentNotSucceeded is returned when an operation requires the
	// payment to have succeeded (e.g. refund).
	ErrPaymentNotSucceeded = errors.New("payment has not succeeded")

	// ErrDuplicatePayment is returned when an idempotency key collides.
	// The original payment is returned to the caller.
	ErrDuplicatePayment = errors.New("duplicate payment request")

	// ============================================================
	// REFUND
	// ============================================================

	// ErrRefundNotFound is returned when a refund does not exist.
	ErrRefundNotFound = errors.New("refund not found")

	// ErrRefundExceedsPayment is returned when a refund would exceed
	// the total paid amount (including prior refunds).
	ErrRefundExceedsPayment = errors.New("refund amount exceeds the paid amount")

	// ErrRefundAlreadyExists is returned when a refund request
	// duplicates an existing one (same idempotency key).
	ErrRefundAlreadyExists = errors.New("a refund with this idempotency key already exists")

	// ErrRefundNotPending is returned when a refund operation requires
	// the refund to be pending but it isn't.
	ErrRefundNotPending = errors.New("refund is not pending")

	// ============================================================
	// PROVIDER
	// ============================================================

	// ErrProviderNotFound is returned when a requested provider name
	// is not registered.
	ErrProviderNotFound = errors.New("payment provider not found")

	// ErrProviderUnavailable is returned when the provider is reachable
	// but reports it cannot serve the request (5xx, maintenance).
	ErrProviderUnavailable = errors.New("payment provider is unavailable")

	// ErrProviderRejected is returned when the provider rejects the
	// request due to client error (4xx). Not retryable.
	ErrProviderRejected = errors.New("payment provider rejected the request")

	// ErrProviderTimeout is returned when a provider call times out.
	// Retryable with the same idempotency key.
	ErrProviderTimeout = errors.New("payment provider request timed out")

	// ============================================================
	// WEBHOOK
	// ============================================================

	// ErrInvalidWebhookSignature is returned when a webhook's signature
	// verification fails. The webhook is not processed.
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

	// ErrDuplicateWebhook is returned when a webhook event has already
	// been received and recorded.
	ErrDuplicateWebhook = errors.New("duplicate webhook event")

	// ErrMalformedWebhook is returned when a webhook payload cannot be
	// parsed.
	ErrMalformedWebhook = errors.New("malformed webhook payload")

	ErrWebhookEventNotFound    = errors.New("webhook event not found") 

	// ============================================================
	// DOMAIN-LEVEL
	// ============================================================

	// ErrInvalidStatusTransition is returned by any entity method that
	// rejects an illegal state change.
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrInvalidAmount is returned when an amount is zero, negative, or
	// fails an invariant (e.g. discount > subtotal).
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrInvalidCurrency is returned when a currency is empty or
	// mismatched.
	ErrInvalidCurrency = errors.New("invalid currency")

	// ErrIdentityRequired is returned when neither a user nor a guest
	// identity is provided, or both are provided.
	ErrIdentityRequired = errors.New("user identity is required")

	// ErrNotOwner is returned when an actor does not own the resource
	// they are trying to act on.
	ErrNotOwner = errors.New("actor does not own this resource")

	// ErrNotAuthorized is returned when an actor lacks permission for
	// an action regardless of ownership.
	ErrNotAuthorized = errors.New("actor is not authorized for this action")



	ErrPricingMismatch = errors.New("pricing mismatch")
)