// internal/modules/registration/infrastructure/notifier/noop.go

package notifier

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// Noop is a Notifier that discards all events. It exists so the
// registration module can be wired before the notification module is
// implemented. Replace with the real adapter when notifications land.
type Noop struct{}

// NewNoop constructs a Noop notifier.
func NewNoop() *Noop { return &Noop{} }

func (*Noop) RegistrationCreated(ctx context.Context, r *registrationdomain.Registration) error {
	return nil
}

func (*Noop) RegistrationConfirmed(ctx context.Context, r *registrationdomain.Registration) error {
	return nil
}

func (*Noop) RegistrationCancelled(ctx context.Context, r *registrationdomain.Registration) error {
	return nil
}

func (*Noop) WaitlistPromoted(ctx context.Context, w *registrationdomain.WaitlistEntry) error {
	return nil
}

// Compile-time assertion.
var _ registrationdomain.Notifier = (*Noop)(nil)