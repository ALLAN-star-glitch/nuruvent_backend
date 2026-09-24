// internal/modules/payment/service/commands.go

package service

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"

// ============================================================
// INITIATE
// ============================================================

// InitiateCommand is the input for InitiatePayment.
type InitiateCommand struct {
	// OrderID is the order to pay for. Must be pending.
	OrderID string

	// Method is the payment method (mpesa or card).
	Method paymentdomain.PaymentMethod

	// PayerPhone is required for M-Pesa. The provider normalizes it.
	PayerPhone string

	// PayerEmail is required for card processors.
	//
	// For guest checkouts, this field doubles as proof of ownership:
	// it must match the email stored on the order.
	PayerEmail string

	// Card carries card details when Method == card. Nil for M-Pesa.
	//
	// SECURITY: this struct holds PCI-sensitive data. It must never
	// be logged or persisted. After InitiatePayment returns, callers
	// should zero the fields.
	Card *paymentdomain.CardDetails

	// IdempotencyKey prevents duplicate payment creation. Supplied by
	// the client, stable across retries.
	IdempotencyKey string

	// ReturnURL is where the card provider redirects after payment.
	ReturnURL string

	// Description is shown to the user in the provider's checkout.
	Description string

	// ------------------------------------------------------------
	// Ownership proof
	// ------------------------------------------------------------
	// Exactly one of these is required, mirroring CreateOrder:
	//
	//   - ActorID set    → authenticated user; must own the order.
	//   - PayerEmail set → guest; must match the order's GuestEmail.
	//
	// If neither is set, the request is rejected as unauthorized.
	ActorID string
}

// ============================================================
// REFUND
// ============================================================

// RefundCommand is the input for Refund.
type RefundCommand struct {
	// PaymentID is the payment to refund.
	PaymentID string

	// Amount is in minor units. For a full refund, set this to the
	// payment's amount.
	Amount int64

	// Reason is a human-readable explanation for audit.
	Reason string

	// ActorID is the user initiating the refund.
	ActorID string

	// IdempotencyKey prevents duplicate refund creation.
	IdempotencyKey string
}

// ============================================================
// CREATE ORDER
// ============================================================

// CreateOrderCommand is the input for CreateOrder.
//
// The caller supplies only the registration ID. The payment module
// resolves the pricing, builds the order items, and computes the
// totals from the registration's snapshot.
//
// This means the client cannot influence the order amount — the price
// always reflects what the registration agreed to.
type CreateOrderCommand struct {
	RegistrationID string

	// ------------------------------------------------------------
	// Ownership proof
	// ------------------------------------------------------------
	// Exactly one of these is required:
	//
	//   - ActorID set    → authenticated user; must own the
	//                      registration (registration.user_id == ActorID).
	//   - GuestEmail set → guest; must match registration.guest_email.
	//
	// If neither is set, the request is rejected as unauthorized.
	ActorID    string
	GuestEmail string
}