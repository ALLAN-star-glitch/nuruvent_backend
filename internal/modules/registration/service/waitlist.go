package service

import (
    "context"
    "fmt"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// JoinWaitlist adds a user (or guest) to the event's waitlist.
func (s *service) JoinWaitlist(
    ctx context.Context,
    cmd JoinWaitlistCommand,
) (*registrationdomain.WaitlistEntry, error) {
    now := s.deps.Clock.Now()

    registrable, err := s.deps.Registrables.Resolve("event", cmd.EventID)
    if err != nil {
        return nil, fmt.Errorf("resolve event: %w", err)
    }
    if !registrable.IsRegistrationOpen() {
        return nil, registrationdomain.ErrEventNotOpen
    }

    position, err := s.deps.Waitlist.NextPosition(ctx, cmd.EventID)
    if err != nil {
        return nil, fmt.Errorf("next position: %w", err)
    }

	entry, err := registrationdomain.NewWaitlistEntry(
		s.deps.IDGenerator.NewID(),
		cmd.EventID,
		cmd.UserID,
		guestEmail(cmd.Guest),
		guestName(cmd.Guest),      // ← add
		cmd.TicketTypeID,          // ← add
		position,
		now,
	)
    if err != nil {
        return nil, err
    }

    if err := s.deps.Waitlist.Create(ctx, entry); err != nil {
        return nil, fmt.Errorf("create waitlist entry: %w", err)
    }

    return entry, nil
}

// PromoteFromWaitlist attempts to register the next waitlisted user.
// Called after a cancellation frees a spot.
func (s *service) PromoteFromWaitlist(
    ctx context.Context,
    eventID string,
) (*registrationdomain.EventRegistration, error) {
    entry, err := s.deps.Waitlist.PeekNext(ctx, eventID)
    if err != nil {
        return nil, err
    }
    if entry == nil {
        return nil, nil // nothing to promote
    }

    // Reconstruct a RegisterCommand from the waitlist entry.
    cmd := RegisterCommand{
        EventID:    eventID,
        UserID:     entry.UserID,
        Guest:      nil, // for now, waitlist only supports authenticated users
        Selections: nil, // ??? — need to store intent on the waitlist entry
    }

    // NOTE: The waitlist entry needs to remember the ticket selections
    // and quantities the user wanted. Step 2's WaitlistEntry doesn't carry
    // them yet — we should add `Selections []TicketSelection` to it.
    // Flagging this now so we fix it before Step 4.

    eventReg, err := s.RegisterForEvent(ctx, cmd)
    if err != nil {
        return nil, err
    }

    if err := s.deps.Waitlist.Delete(ctx, entry.ID); err != nil {
        return nil, fmt.Errorf("delete waitlist entry: %w", err)
    }

    _ = s.deps.Notifier.WaitlistPromoted(ctx, entry)

    return eventReg, nil
}