package service

import (
    "context"
    "fmt"

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

    registrable, err := s.deps.Registrables.Resolve("event", eventReg.EventID)
    if err != nil {
        return fmt.Errorf("resolve event: %w", err)
    }

    err = s.deps.Registrations.WithTx(ctx, func(tx registrationdomain.RegistrationRepository) error {
        if err := tx.Update(ctx, reg); err != nil {
            return fmt.Errorf("update registration: %w", err)
        }
        if wasConfirmed {
            if err := registrable.AdjustCount(-eventReg.TotalQuantity()); err != nil {
                return fmt.Errorf("adjust count: %w", err)
            }
        }
        return nil
    })
    if err != nil {
        return err
    }

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