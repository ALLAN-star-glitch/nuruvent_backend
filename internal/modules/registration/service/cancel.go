// internal/modules/registration/service/cancel.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// CancelRegistration cancels a registration and adjusts the counter.
func (s *service) CancelRegistration(ctx context.Context, cmd CancelCommand) error {
	now := s.deps.Clock.Now()

	reg, err := s.deps.Registrations.FindByID(ctx, cmd.RegistrationID)
	if err != nil {
		return err
	}

	// Idempotent: already cancelled → no-op.
	if reg.Status == registrationdomain.StatusCancelled {
		return nil
	}

	// Authorization: actor must own it or be the event organizer.
	// (Organizer check is delegated to the events module via a port —
	//  added in a later step. For now, owner-only.)
	if reg.UserID != "" && reg.UserID != cmd.ActorID {
		return registrationdomain.ErrNotOwner
	}

	eventReg, err := s.deps.EventRegistrations.FindByID(ctx, cmd.RegistrationID)
	if err != nil {
		return fmt.Errorf("load event registration: %w", err)
	}

	wasConfirmed := reg.Status == registrationdomain.StatusConfirmed

	if err := reg.Cancel(cmd.ActorID, cmd.Reason, now); err != nil {
		return err
	}

	// Persist the status change.
	if err := s.deps.Registrations.Update(ctx, reg); err != nil {
		return fmt.Errorf("update registration: %w", err)
	}

	// Mark the event_registrations row inactive so the partial unique
	// index no longer blocks the user from re-registering for the same
	// event. Best-effort — if this fails, the cancellation still stands;
	// a re-register attempt will retry the deactivation path.
	if err := s.deps.EventRegistrations.Deactivate(ctx, cmd.RegistrationID); err != nil {
		log.Printf("[cancel] deactivate event registration for %s: %v", cmd.RegistrationID, err)
	}

	// Adjust the event counter — best-effort. Failure must NOT roll back
	// the cancellation; the counter is a derived value.
	if wasConfirmed {
		registrable, err := s.deps.Registrables.Resolve("event", eventReg.EventID)
		if err != nil {
			log.Printf("[cancel] resolve event %s: %v", eventReg.EventID, err)
		} else if err := registrable.AdjustCount(-eventReg.TotalQuantity()); err != nil {
			log.Printf("[cancel] adjust count for event %s: %v", eventReg.EventID, err)
		}
	}

	// Notify — best-effort.
	_ = s.deps.Notifier.RegistrationCancelled(ctx, reg)

	// If a spot opened up, try to promote from the waitlist (best-effort).
	if wasConfirmed {
		go s.promoteFromWaitlistAsync(context.Background(), eventReg.EventID)
	}

	return nil
}

// promoteFromWaitlistAsync runs promotion in the background so the cancel
// response isn't blocked on waitlist processing.
func (s *service) promoteFromWaitlistAsync(ctx context.Context, eventID string) {
	_, _ = s.PromoteFromWaitlist(ctx, eventID)
}