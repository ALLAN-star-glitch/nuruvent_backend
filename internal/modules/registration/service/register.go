package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

func (s *service) RegisterForEvent(
	ctx context.Context,
	cmd RegisterCommand,
) (*registrationdomain.EventRegistration, error) {
	now := s.deps.Clock.Now()

	log.Printf("[registration] ENTER RegisterForEvent: user=%q event=%q selections=%d",
		cmd.UserID, cmd.EventID, len(cmd.Selections))

	// ------------------------------------------------------------
	// 1. Resolve the registrable
	// ------------------------------------------------------------
	registrable, err := s.deps.Registrables.Resolve("event", cmd.EventID)
	if err != nil {
		log.Printf("[registration] STEP 1 FAIL: resolve event: %v", err)
		return nil, fmt.Errorf("resolve event: %w", err)
	}
	log.Printf("[registration] STEP 1 OK: registrable resolved")

		// ------------------------------------------------------------
	// 2. Validate registration is open
	// ------------------------------------------------------------
	if !registrable.IsRegistrationOpen() {
		log.Printf("[registration] STEP 2 FAIL: registration not open for event=%s",
			cmd.EventID)
		return nil, registrationdomain.ErrEventNotOpen
	}
	log.Printf("[registration] STEP 2 OK: registration is open")

	// ------------------------------------------------------------
	// 2b. Enforce authentication for events that require it.
	//
	// Zoom events require a Nuruvent account: attendance is tracked
	// via a JWT signed for a specific user, and guests have no
	// identity to sign.
	// ------------------------------------------------------------
	if cmd.UserID == "" && registrable.RequiresAuth() {
		log.Printf("[registration] STEP 2b FAIL: guest registration not permitted for event=%s",
			cmd.EventID)
		return nil, registrationdomain.ErrAuthRequired
	}
	if registrable.RequiresAuth() {
		log.Printf("[registration] STEP 2b OK: authenticated registration (event requires auth)")
	}
	log.Printf("[registration] STEP 2 OK: registration is open")

	// ------------------------------------------------------------
	// 3. Duplicate check (for authenticated users)
	// ------------------------------------------------------------
	if cmd.UserID != "" {
		existing, err := s.deps.Registrations.FindActiveByUserAndEvent(ctx, cmd.UserID, cmd.EventID)
		if err != nil && err != registrationdomain.ErrRegistrationNotFound {
			log.Printf("[registration] STEP 3 FAIL: duplicate check error: %v", err)
			return nil, fmt.Errorf("duplicate check: %w", err)
		}
		if existing != nil {
			log.Printf("[registration] STEP 3 FAIL: duplicate registration id=%s",
				existing.ID)
			return nil, registrationdomain.ErrDuplicateRegistration
		}
	}
	log.Printf("[registration] STEP 3 OK: no duplicate")

	// ------------------------------------------------------------
	// 4. Validate ticket selections against availability
	// ------------------------------------------------------------
	availability, err := registrable.TicketAvailability()
	if err != nil {
		log.Printf("[registration] STEP 4 FAIL: ticket availability error: %v", err)
		return nil, fmt.Errorf("ticket availability: %w", err)
	}
	log.Printf("[registration] STEP 4 availability=%v", availability)

	if err := validateTicketSelections(cmd.Selections, availability); err != nil {
		log.Printf("[registration] STEP 4 FAIL: ticket selection invalid: %v", err)
		return nil, err
	}
	log.Printf("[registration] STEP 4 OK: selections valid")

	// ------------------------------------------------------------
	// 5. Compute pricing server-side
	// ------------------------------------------------------------
	pricingTable, err := registrable.TicketPricing()
	if err != nil {
		log.Printf("[registration] STEP 5 FAIL: ticket pricing error: %v", err)
		return nil, fmt.Errorf("ticket pricing: %w", err)
	}

	selections, err := buildSelections(cmd.Selections, pricingTable)
	if err != nil {
		log.Printf("[registration] STEP 5 FAIL: build selections error: %v", err)
		return nil, err
	}

	snapshot, err := registrationdomain.NewPricingSnapshot("KES", selections, now)
	if err != nil {
		log.Printf("[registration] STEP 5 FAIL: pricing snapshot error: %v", err)
		return nil, err
	}
	log.Printf("[registration] STEP 5 OK: pricing snapshot built, selections=%d",
		len(selections))

	// ------------------------------------------------------------
	// 6. Capacity check (unless waitlist)
	// ------------------------------------------------------------
	requiresPayment := registrable.RequiresPayment()
	totalQuantity := sumQuantities(selections)

	current := registrable.CurrentRegistrations()
	capacity := registrable.Capacity()

	log.Printf("[registration] STEP 6 capacity check: current=%d capacity=%d requested=%d requiresPayment=%v",
		current, capacity, totalQuantity, requiresPayment)

	// capacity == 0 means unlimited in the events module. The check
	// only applies when a positive cap is set.
	if capacity > 0 && current+totalQuantity > capacity {
		log.Printf("[registration] STEP 6 FAIL: event full (current=%d capacity=%d requested=%d)",
			current, capacity, totalQuantity)
		return nil, registrationdomain.ErrEventFull
	}
	log.Printf("[registration] STEP 6 OK: capacity available (unlimited or within limit)")

	// ------------------------------------------------------------
	// 7. Build entities
	// ------------------------------------------------------------
	id := s.deps.IDGenerator.NewID()

	registrationNumber, err := s.deps.NumberGenerator.Next(ctx)
	if err != nil {
		log.Printf("[registration] STEP 7 FAIL: number generator error: %v", err)
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
		log.Printf("[registration] STEP 7 FAIL: build registration error: %v", err)
		return nil, err
	}

	eventReg, err := registrationdomain.NewEventRegistration(
		reg,
		cmd.EventID,
		selections,
		snapshot,
	)
	if err != nil {
		log.Printf("[registration] STEP 7 FAIL: build event registration error: %v", err)
		return nil, err
	}
	log.Printf("[registration] STEP 7 OK: entities built, regNumber=%s status=%s",
		registrationNumber, reg.Status)

	// ------------------------------------------------------------
	// 8. Persist in a transaction
	// ------------------------------------------------------------
	err = s.deps.Registrations.WithTx(ctx,
		func(
			regRepo registrationdomain.RegistrationRepository,
			eventRegRepo registrationdomain.EventRegistrationRepository,
			waitlistRepo registrationdomain.WaitlistRepository,
		) error {
			if err := regRepo.Create(ctx, reg); err != nil {
				log.Printf("[registration] STEP 8 FAIL: create registration: %v", err)
				return fmt.Errorf("create registration: %w", err)
			}
			if err := eventRegRepo.Create(ctx, eventReg); err != nil {
				log.Printf("[registration] STEP 8 FAIL: create event registration: %v", err)
				return fmt.Errorf("create event registration: %w", err)
			}
			if reg.Status == registrationdomain.StatusConfirmed {
				if err := registrable.AdjustCount(+totalQuantity); err != nil {
					log.Printf("[registration] STEP 8 FAIL: adjust count: %v", err)
					return fmt.Errorf("adjust count: %w", err)
				}
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	log.Printf("[registration] STEP 8 OK: persisted, id=%s", reg.ID)

		// ------------------------------------------------------------
	// 9. Enqueue notifications (best-effort)
	//
	// RegistrationCreated fires unconditionally: the attendee gets a
	// "we've received your registration" message.
	//
	// RegistrationConfirmed fires only when the registration is
	// already confirmed at creation time (free events, zero-total
	// registrations). Paid registrations transition to confirmed
	// later, via ConfirmRegistration, which fires this itself.
	//
	// At this point the attendance sync has not run, so the payload
	// carries no join links. For paid events that's fine — the links
	// go out in the confirmation email sent by ConfirmRegistration.
	// For free events, syncRegistrationToAttendance runs below and
	// then a second notification fires with the links populated.
	// ------------------------------------------------------------
	if err := s.deps.Notifier.RegistrationCreated(ctx, reg); err != nil {
		log.Printf("[registration] STEP 9 WARN: RegistrationCreated notifier failed: %v", err)
	}
	if reg.Status == registrationdomain.StatusConfirmed {
		// Sync attendance so free events also get join links in the
		// confirmation email. Best-effort — failures are logged and
		// swallowed.
		s.syncRegistrationToAttendance(ctx, eventReg)

		if err := s.deps.Notifier.RegistrationConfirmed(ctx, registrationdomain.RegistrationConfirmedPayload{
			Registration: reg,
			EventID:      eventReg.EventID,
			JoinLinks:    eventReg.JoinLinks,
		}); err != nil {
			log.Printf("[registration] STEP 9 WARN: RegistrationConfirmed notifier failed: %v", err)
		}
	}

	log.Printf("[registration] EXIT OK: regID=%s eventID=%s", reg.ID, eventReg.EventID)
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