// internal/app/adapters/payment/pricing_resolver.go

package payment

import (
	"context"
	"errors"
	"fmt"

	paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	registrationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	registrationService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
)

// PricingResolver implements paymentdomain.RegistrationPricingResolver
// by delegating to the registration module.
//
// The payment module uses this to fetch a registration's pricing at
// the moment an order is created. This ensures the order total always
// matches what the user was quoted at registration time.
type PricingResolver struct {
	regSvc registrationService.Service
}

// NewPricingResolver constructs the adapter.
func NewPricingResolver(regSvc registrationService.Service) *PricingResolver {
	return &PricingResolver{regSvc: regSvc}
}

// ResolvePricing loads a registration and returns a flat snapshot of
// its pricing for the payment module.
//
// The registration must be in pending status. If it's already
// confirmed, cancelled, or expired, an error is returned — the caller
// (payment service) surfaces this as a 409 Conflict.
//
// Reads pricing from the persisted Registration fields (Currency,
// Subtotal, DiscountTotal, TotalAmount). The domain's PricingSnapshot
// is only populated at creation time and is empty after hydration.
func (r *PricingResolver) ResolvePricing(
	ctx context.Context,
	registrationID string,
) (*paymentdomain.RegistrationPricing, error) {
	if registrationID == "" {
		return nil, errors.New("registration_id is required")
	}

	// The registration service's GetByID takes an actorID for
	// authorization. The payment module isn't an "actor" — it's an
	// internal caller. Pass an empty string; the service treats empty
	// actorID as an internal/system call.
	reg, err := r.regSvc.GetByID(ctx, registrationID, "")
	if err != nil {
		return nil, fmt.Errorf("load registration: %w", err)
	}
	if reg == nil {
		return nil, fmt.Errorf("%w: %s", paymentdomain.ErrRegistrationNotFound, registrationID)
	}

	// Only pending registrations can become orders. A confirmed
	// registration is already paid for; a cancelled one shouldn't be
	// charged.
	if reg.Registration.Status != registrationdomain.StatusPending {
		return nil, fmt.Errorf(
			"registration %s is not pending (status: %s)",
			registrationID, reg.Registration.Status,
		)
	}

	// Flatten the ticket selections into the payment module's shape.
	items := make([]paymentdomain.RegistrationPricingItem, 0, len(reg.Selections))
	for _, s := range reg.Selections {
		items = append(items, paymentdomain.RegistrationPricingItem{
			TicketTypeID: s.TicketTypeID,
			Quantity:     s.Quantity,
			UnitPrice:    s.UnitPrice,
			Discount:     s.Discount,
		})
	}

	// Use the persisted Registration fields, not reg.Pricing.*. The
	// pricing snapshot on the domain is only set at creation time and
	// is empty when the registration is hydrated from the database.
	return &paymentdomain.RegistrationPricing{
		RegistrationID: reg.Registration.ID,
		UserID:         reg.Registration.UserID,
		GuestEmail:     reg.Registration.GuestEmail,
		Currency:       reg.Registration.Currency,
		Subtotal:       reg.Registration.Subtotal,
		DiscountTotal:  reg.Registration.DiscountTotal,
		TotalAmount:    reg.Registration.TotalAmount,
		Items:          items,
	}, nil
}


// Compile-time assertion.
var _ paymentdomain.RegistrationPricingResolver = (*PricingResolver)(nil)