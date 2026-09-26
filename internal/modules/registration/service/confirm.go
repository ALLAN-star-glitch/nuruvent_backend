// internal/modules/registration/service/confirm.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// ConfirmRegistration transitions a pending registration to confirmed.
// Called by the payment module when a webhook reports success.
// Idempotent.
func (s *service) ConfirmRegistration(ctx context.Context, registrationID string) error {
	now := s.deps.Clock.Now()

	reg, err := s.deps.Registrations.FindByID(ctx, registrationID)
	if err != nil {
		return err
	}

	// Idempotent: already confirmed → no-op.
	if reg.Status == registrationdomain.StatusConfirmed {
		return nil
	}

	// Load the event registration BEFORE the status change so we have the
	// event ID and total quantity for the counter adjustment.
	eventReg, err := s.deps.EventRegistrations.FindByID(ctx, registrationID)
	if err != nil {
		return fmt.Errorf("load event registration: %w", err)
	}

	// Transition the entity.
	if err := reg.Confirm(now); err != nil {
		return err
	}

	// Persist the status change. This is the only transactional write.
	if err := s.deps.Registrations.Update(ctx, reg); err != nil {
		return fmt.Errorf("update registration: %w", err)
	}

	// Adjust the event counter. Best-effort — a failure here does NOT
	// roll back the confirmation, because the registration itself is
	// the source of truth. Drift is reconciled by a background job.
	registrable, err := s.deps.Registrables.Resolve("event", eventReg.EventID)
	if err != nil {
		log.Printf("[confirm] resolve event %s: %v", eventReg.EventID, err)
		return nil
	}
	if err := registrable.AdjustCount(+eventReg.TotalQuantity()); err != nil {
		log.Printf("[confirm] adjust count for event %s: %v", eventReg.EventID, err)
		return nil
	}

	// Sync to attendance (best-effort). Populates reg.JoinLinks.
	s.syncRegistrationToAttendance(ctx, eventReg)

	// Notify (best-effort).
	_ = s.deps.Notifier.RegistrationConfirmed(ctx, reg)

	return nil
}