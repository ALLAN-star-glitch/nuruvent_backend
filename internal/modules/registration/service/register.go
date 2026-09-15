package service

import (
    "context"
    "fmt"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

func (s *service) RegisterForEvent(
    ctx context.Context,
    cmd RegisterCommand,
) (*registrationdomain.EventRegistration, error) {
    now := s.deps.Clock.Now()

    // ------------------------------------------------------------
    // 1. Resolve the registrable
    // ------------------------------------------------------------
    registrable, err := s.deps.Registrables.Resolve("event", cmd.EventID)
    if err != nil {
        return nil, fmt.Errorf("resolve event: %w", err)
    }

    // ------------------------------------------------------------
    // 2. Validate registration is open
    // ------------------------------------------------------------
    if !registrable.IsRegistrationOpen() {
        return nil, registrationdomain.ErrEventNotOpen
    }

    // ------------------------------------------------------------
    // 3. Duplicate check (for authenticated users)
    // ------------------------------------------------------------
    if cmd.UserID != "" {
        existing, err := s.deps.Registrations.FindActiveByUserAndEvent(ctx, cmd.UserID, cmd.EventID)
        if err != nil && err != registrationdomain.ErrRegistrationNotFound {
            return nil, fmt.Errorf("duplicate check: %w", err)
        }
        if existing != nil {
            return nil, registrationdomain.ErrDuplicateRegistration
        }
    }

    // ------------------------------------------------------------
    // 4. Validate ticket selections against availability
    // ------------------------------------------------------------
    availability, err := registrable.TicketAvailability()
    if err != nil {
        return nil, fmt.Errorf("ticket availability: %w", err)
    }
    if err := validateTicketSelections(cmd.Selections, availability); err != nil {
        return nil, err
    }

    // ------------------------------------------------------------
    // 5. Compute pricing server-side
    // ------------------------------------------------------------
    pricingTable, err := registrable.TicketPricing()
    if err != nil {
        return nil, fmt.Errorf("ticket pricing: %w", err)
    }
    selections, err := buildSelections(cmd.Selections, pricingTable)
    if err != nil {
        return nil, err
    }
    snapshot, err := registrationdomain.NewPricingSnapshot("KES", selections, now)
    if err != nil {
        return nil, err
    }

    // ------------------------------------------------------------
    // 6. Capacity check (unless waitlist)
    // ------------------------------------------------------------
    requiresPayment := registrable.RequiresPayment()
    totalQuantity := sumQuantities(selections)

    if registrable.CurrentRegistrations()+totalQuantity > registrable.Capacity() {
        return nil, registrationdomain.ErrEventFull
    }

    // ------------------------------------------------------------
    // 7. Build entities
    // ------------------------------------------------------------
    id := s.deps.IDGenerator.NewID()

	registrationNumber, err := s.deps.NumberGenerator.Next(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate registration number: %w", err)
	}

	reg, err := registrationdomain.NewRegistration(
		id,
		registrationNumber,
		cmd.UserID,
		guestEmail(cmd.Guest),
		guestName(cmd.Guest),
		guestPhone(cmd.Guest),
		snapshot,
		requiresPayment,
		now,
	)
if err != nil {
	return nil, err
}
    if err != nil {
        return nil, err
    }

    eventReg, err := registrationdomain.NewEventRegistration(
        reg,
        cmd.EventID,
        selections,
        snapshot,
    )
    if err != nil {
        return nil, err
    }

    // ------------------------------------------------------------
    // 8. Persist in a transaction
    // ------------------------------------------------------------
    err = s.deps.Registrations.WithTx(ctx, func(tx registrationdomain.RegistrationRepository) error {
        if err := tx.Create(ctx, reg); err != nil {
            return fmt.Errorf("create registration: %w", err)
        }
        if err := s.deps.EventRegistrations.Create(ctx, eventReg); err != nil {
            return fmt.Errorf("create event registration: %w", err)
        }
        // Only adjust the counter when the registration is confirmed.
        if reg.Status == registrationdomain.StatusConfirmed {
            if err := registrable.AdjustCount(+totalQuantity); err != nil {
                return fmt.Errorf("adjust count: %w", err)
            }
        }
        return nil
    })
    if err != nil {
        return nil, err
    }

    // ------------------------------------------------------------
    // 9. Enqueue notifications (best-effort)
    // ------------------------------------------------------------
    if err := s.deps.Notifier.RegistrationCreated(ctx, reg); err != nil {
        // Log but don't fail — the registration succeeded.
        // (Actual logging depends on your logger convention.)
    }
    if reg.Status == registrationdomain.StatusConfirmed {
        _ = s.deps.Notifier.RegistrationConfirmed(ctx, reg)
    }

    return eventReg, nil
}

// ------------------------------------------------------------
// Helpers local to register.go
// ------------------------------------------------------------

func validateTicketSelections(
    inputs []TicketSelectionInput,
    availability map[string]int,
) error {
    if len(inputs) == 0 {
        return registrationdomain.ErrTicketUnavailable
    }
    for _, in := range inputs {
        avail, ok := availability[in.TicketTypeID]
        if !ok {
            return fmt.Errorf("%w: %s", registrationdomain.ErrTicketUnavailable, in.TicketTypeID)
        }
        if in.Quantity < 1 {
            return fmt.Errorf("%w: quantity must be >= 1", registrationdomain.ErrTicketUnavailable)
        }
        if in.Quantity > avail {
            return fmt.Errorf("%w: %s requested %d, available %d",
                registrationdomain.ErrTicketUnavailable, in.TicketTypeID, in.Quantity, avail)
        }
    }
    return nil
}

func buildSelections(
    inputs []TicketSelectionInput,
    pricing map[string]registrationdomain.TicketPrice,
) ([]registrationdomain.TicketSelection, error) {
    out := make([]registrationdomain.TicketSelection, 0, len(inputs))
    for _, in := range inputs {
        price, ok := pricing[in.TicketTypeID]
        if !ok {
            return nil, fmt.Errorf("%w: %s", registrationdomain.ErrPricingMismatch, in.TicketTypeID)
        }
        out = append(out, registrationdomain.TicketSelection{
            TicketTypeID: in.TicketTypeID,
            Quantity:     in.Quantity,
            UnitPrice:    price.UnitPrice,
            Discount:     price.Discount,
        })
    }
    return out, nil
}

func sumQuantities(sel []registrationdomain.TicketSelection) int {
    n := 0
    for _, s := range sel {
        n += s.Quantity
    }
    return n
}

func guestEmail(g *GuestIdentity) string {
    if g == nil {
        return ""
    }
    return g.Email
}
func guestName(g *GuestIdentity) string {
    if g == nil {
        return ""
    }
    return g.Name
}
func guestPhone(g *GuestIdentity) string {
    if g == nil {
        return ""
    }
    return g.Phone
}