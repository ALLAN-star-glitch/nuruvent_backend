// internal/modules/payment/paymentdomain/registration.go

package paymentdomain

import "context"

// RegistrationConfirmer is the outbound port the service uses to tell
// the registration module about payment outcomes.
//
// Implemented by an adapter in internal/app/adapters/payment/. The
// payment module never imports the registration module directly.
//
// Both methods must be idempotent — calling them multiple times for the
// same registration ID produces the same result. Webhook redelivery and
// reconciliation can both trigger them.
//
// The method names mirror the registration module's service exactly, so
// the adapter in app/adapters/payment/ is a trivial pass-through with no
// vocabulary translation.
type RegistrationConfirmer interface {
	// ConfirmRegistration transitions a pending registration to
	// confirmed. Called when a payment succeeds.
	ConfirmRegistration(ctx context.Context, registrationID string) error

	// ExpirePending transitions a pending registration to expired.
	// Called when a payment window elapses without success.
	ExpirePending(ctx context.Context, registrationID string) error
}

// ============================================================
// PRICING RESOLVER
// ============================================================

// RegistrationPricing is a flattened snapshot of a registration's
// pricing, as seen by the payment module.
//
// The payment module cannot import the registration module's domain
// types directly, so the adapter flattens them into this struct.
type RegistrationPricing struct {
	RegistrationID string

	// Exactly one of UserID or GuestEmail is populated.
	UserID     string
	GuestEmail string

	Currency      string
	Subtotal      int64 // minor units
	DiscountTotal int64 // minor units
	TotalAmount   int64 // minor units

	Items []RegistrationPricingItem
}

// RegistrationPricingItem is one line of a registration, flattened for
// the payment module.
type RegistrationPricingItem struct {
	TicketTypeID string
	Quantity     int
	UnitPrice    int64 // minor units
	Discount     int64 // minor units
}

// RegistrationPricingResolver is the outbound port the payment module
// uses to fetch a registration's pricing when creating an order.
//
// It is implemented by an adapter in internal/app/adapters/payment/.
// The adapter calls the registration module's service and returns a
// flattened RegistrationPricing.
//
// Returns an error if the registration doesn't exist, isn't pending, or
// the registration module is unreachable.
type RegistrationPricingResolver interface {
	ResolvePricing(ctx context.Context, registrationID string) (*RegistrationPricing, error)
}