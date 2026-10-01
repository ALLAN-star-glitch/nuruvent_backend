package registrationdomain

import "context"

// Notifier is the port for outbound notifications. Implemented by an
// adapter that enqueues jobs onto the shared queue; consumed by the
// notification module.
type Notifier interface {
	RegistrationCreated(ctx context.Context, r *Registration) error
	RegistrationConfirmed(ctx context.Context, p RegistrationConfirmedPayload) error
	RegistrationCancelled(ctx context.Context, r *Registration) error
	WaitlistPromoted(ctx context.Context, w *WaitlistEntry) error
}

// RegistrationConfirmedPayload carries everything the notification
// layer needs to render a "your registration is confirmed" message.
//
// JoinLinks is populated when the registration has been synced to
// attendance. It is empty for events with no sessions or when the
// sync was skipped (no email, provider error).
type RegistrationConfirmedPayload struct {
	Registration *Registration
	EventID      string
	EventName    string
	JoinLinks    []JoinLink
}