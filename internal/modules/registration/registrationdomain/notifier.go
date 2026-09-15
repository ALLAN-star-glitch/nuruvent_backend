package registrationdomain

import "context"

// Notifier is the port for outbound notifications. Implemented by an adapter
// that enqueues jobs onto the shared queue; consumed by the notification module.
type Notifier interface {
    RegistrationCreated(ctx context.Context, r *Registration) error
    RegistrationConfirmed(ctx context.Context, r *Registration) error
    RegistrationCancelled(ctx context.Context, r *Registration) error
    WaitlistPromoted(ctx context.Context, w *WaitlistEntry) error
}