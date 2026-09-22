
// internal/modules/payment/infrastructure/notifier/noop.go

package notifier

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// Noop is a Notifier that discards all events.
//
// It exists so the payment service can be wired before the real
// notification adapter is built. When the notification module
// integration lands, swap this out — the service interface doesn't
// change.
type Noop struct{}

// NewNoop constructs a Noop notifier.
func NewNoop() *Noop { return &Noop{} }

func (*Noop) PaymentInitiated(_ context.Context, _ *paymentdomain.Payment, _ *paymentdomain.Order) error {
	return nil
}
func (*Noop) PaymentSucceeded(_ context.Context, _ *paymentdomain.Payment, _ *paymentdomain.Order) error {
	return nil
}
func (*Noop) PaymentFailed(_ context.Context, _ *paymentdomain.Payment, _ *paymentdomain.Order) error {
	return nil
}
func (*Noop) PaymentExpired(_ context.Context, _ *paymentdomain.Payment, _ *paymentdomain.Order) error {
	return nil
}
func (*Noop) RefundIssued(_ context.Context, _ *paymentdomain.Refund, _ *paymentdomain.Payment) error {
	return nil
}

// Compile-time assertion.
var _ paymentdomain.Notifier = (*Noop)(nil)