package service

import "context"

// ResolveBilledAccount returns the account that owns the event behind
// the given registration.
//
// Thin pass-through to the repository — no business logic lives here.
// The ownership chain is enforced by the repo's SQL; the service just
// forwards.
//
// Used by the payment module's billing adapter to snapshot the billed
// account onto orders at creation time.
func (s *service) ResolveBilledAccount(
    ctx context.Context,
    registrationID string,
) (string, error) {
    return s.deps.EventRegistrations.ResolveBilledAccount(ctx, registrationID)
}