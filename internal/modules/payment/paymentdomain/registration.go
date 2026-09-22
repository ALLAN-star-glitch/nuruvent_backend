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