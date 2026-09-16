// internal/modules/registration/infrastructure/postgres/mappers.go

package postgres

import (
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// DOMAIN → MODEL
// ============================================================

// toRegistrationModel converts a domain Registration into a GORM model.
// The statusID must be resolved by the caller (via resolveStatusID)
// because the domain entity only knows the status slug, not its UUID.
func toRegistrationModel(
	r *registrationdomain.Registration,
	statusID string,
) *RegistrationModel {
	return &RegistrationModel{
		ID:                 r.ID,
		UserID:             nullableString(r.UserID),
		GuestEmail:         nullableString(r.GuestEmail),
		GuestName:          nullableString(r.GuestName),
		GuestPhone:         nullableString(r.GuestPhone),
		RegistrationNumber: r.RegistrationNumber,
		StatusID:           statusID,
		Currency:           r.Currency,
		Subtotal:           r.Subtotal,
		DiscountTotal:      r.DiscountTotal,
		TotalAmount:        r.TotalAmount,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
		ConfirmedAt:        r.ConfirmedAt,
		CancelledAt:        r.CancelledAt,
		CancelledBy:        nullableString(r.CancelledBy),
		CancellationReason: nullableString(r.CancellationReason),
	}
}

// ============================================================
// MODEL → DOMAIN
// ============================================================

// toRegistrationDomain converts a GORM model into a domain Registration.
// Requires the Status relation to be preloaded.
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
		m.Subtotal,
		m.DiscountTotal,
		m.TotalAmount,
		m.CreatedAt,
		m.UpdatedAt,
		m.ConfirmedAt,
		m.CancelledAt,
		derefString(m.CancelledBy),
		derefString(m.CancellationReason),
	), nil
}

// toEventRegistrationDomain composes a Registration and its selections
// into an EventRegistration domain entity.
// Requires the Registration relation to be preloaded.
func toEventRegistrationDomain(
	m *EventRegistrationModel,
	selections []registrationdomain.TicketSelection,
) (*registrationdomain.EventRegistration, error) {
	if m.Registration == nil {
		return nil, fmt.Errorf("event registration model missing preloaded Registration")
	}

	reg, err := toRegistrationDomain(m.Registration)
	if err != nil {
		return nil, err
	}

	pricing := registrationdomain.PricingSnapshot{
		Currency:      reg.Currency,
		Subtotal:      reg.Subtotal,
		DiscountTotal: reg.DiscountTotal,
		Total:         reg.TotalAmount,
		SnapshotAt:    reg.CreatedAt,
	}

	return registrationdomain.NewEventRegistration(reg, m.EventID, selections, pricing)
}

// toWaitlistDomain converts a GORM waitlist model into a domain entity.
func toWaitlistDomain(m *WaitlistModel) *registrationdomain.WaitlistEntry {
	status, ok := types.ParseWaitlistStatus(m.Status)
	if !ok {
		status = types.WaitlistStatusWaiting
	}

	return registrationdomain.HydrateWaitlistEntry(
		m.ID,
		m.EventID,
		derefString(m.UserID),
		derefString(m.GuestEmail),
		derefString(m.GuestName),
		derefString(m.TicketTypeID),
		m.Position,
		status,
		m.CreatedAt,
		m.NotifiedAt,
		m.OfferedAt,
		m.OfferedExpiresAt,
		m.ConvertedAt,
	)
}

// ============================================================
// HELPERS
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