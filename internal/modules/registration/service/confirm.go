package service

import (
    "context"
    "fmt"

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

    if err := reg.Confirm(now); err != nil {
        return err
    }

    // Load the event registration to know how many to add to the counter.
    eventReg, err := s.deps.EventRegistrations.FindByID(ctx, registrationID)
    if err != nil {
        return fmt.Errorf("load event registration: %w", err)
    }

    registrable, err := s.deps.Registrables.Resolve("event", eventReg.EventID)
    if err != nil {
        return fmt.Errorf("resolve event: %w", err)
    }

    return s.deps.Registrations.WithTx(ctx, func(tx registrationdomain.RegistrationRepository) error {
        if err := tx.Update(ctx, reg); err != nil {
            return fmt.Errorf("update registration: %w", err)
        }
        if err := registrable.AdjustCount(+eventReg.TotalQuantity()); err != nil {
            return fmt.Errorf("adjust count: %w", err)
        }
        return nil
    })
}