package postgres

import (
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// MODEL → DOMAIN
// ============================================================

func toRegistrationDomain(m *RegistrationModel) (*registrationdomain.Registration, error) {
	status, ok := types.ParseRegistrationStatus(m.Status.Slug)
	if !ok {
		return nil, fmt.Errorf("unknown registration status slug: %s", m.Status.Slug)
	}

	return registrationdomain.HydrateRegistration(
		m.ID,
		m.RegistrationNumber,
		derefString(m.UserID),
		derefString(m.GuestEmail),
		derefString(m.GuestName),
		derefString(m.GuestPhone),
		status,
		m.Currency,
		m.TotalAmount,
		m.CreatedAt,
		m.UpdatedAt,
		m.ConfirmedAt,
		m.CancelledAt,
		derefString(m.CancelledBy),
		derefString(m.CancellationReason),
	), nil
}

func toEventRegistrationDomain(m *EventRegistrationModel, selections []registrationdomain.TicketSelection) (*registrationdomain.EventRegistration, error) {
	if m.Registration == nil {
		return nil, fmt.Errorf("event registration model missing preloaded Registration")
	}
	reg, err := toRegistrationDomain(m.Registration)
	if err != nil {
		return nil, err
	}
	pricing := registrationdomain.PricingSnapshot{
		Currency:      reg.Currency,
		Subtotal:      m.Registration.Subtotal,
		DiscountTotal: m.Registration.DiscountTotal,
		Total:         m.Registration.TotalAmount,
		SnapshotAt:    m.Registration.CreatedAt,
	}
	return registrationdomain.NewEventRegistration(reg, m.EventID, selections, pricing)
}

func toWaitlistDomain(m *WaitlistModel) *registrationdomain.WaitlistEntry {
	return &registrationdomain.WaitlistEntry{
		ID:           m.ID,
		EventID:      m.EventID,
		UserID:       derefString(m.UserID),
		GuestEmail:   derefString(m.GuestEmail),
		GuestName:    derefString(m.GuestName),
		TicketTypeID: derefString(m.TicketTypeID),
		Position:     m.Position,
		CreatedAt:    m.CreatedAt,
		PromotedAt:   m.ConvertedAt,
	}
}

// ============================================================
// DOMAIN → MODEL
// ============================================================

func toRegistrationModel(r *registrationdomain.Registration) (*RegistrationModel, error) {
	// Resolve status ID from slug — requires a lookup or cache.
	// For simplicity, assume the DB assigns it via a subquery; here we
	// only carry the slug and let the repository resolve it.
	//
	// NOTE: this requires the repository to look up the status_id
	// from the registration_statuses table by slug before inserting.
	// We'll add that in a follow-up helper.
	panic("not implemented — resolve status_id before calling")
}

// ============================================================
// SMALL HELPERS
// ============================================================

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}