// internal/app/adapters/payment/pricing_resolver.go

package payment

import (
	"context"
	"errors"
	"fmt"

	paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	registrationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// PricingResolver implements paymentdomain.RegistrationPricingResolver
// by reading registrations directly from the registration repository.
//
// The repository's FindByID is unconditional — it does not enforce
// ownership. That's intentional here: the payment module enforces its
// own ownership rule (actor OR matching guest email) inside
// CreateOrder and InitiatePayment, using the UserID / GuestEmail
// returned in the snapshot. Enforcing ownership twice — once here and
// once in the payment service — would break legitimate guest checkouts
// because the payment module has no actor to present.
type PricingResolver struct {
	regRepo registrationdomain.EventRegistrationRepository
}

// NewPricingResolver constructs the adapter.
func NewPricingResolver(regRepo registrationdomain.EventRegistrationRepository) *PricingResolver {
	return &PricingResolver{regRepo: regRepo}
}


// ResolvePricing loads a registration and returns a flat snapshot of
// its pricing for the payment module.
//
// The registration must be in pending status. If it's already
// confirmed, cancelled, or expired, an error is returned — the caller
// (payment service) surfaces this as a 409 Conflict.
func (r *PricingResolver) ResolvePricing(
	ctx context.Context,
	registrationID string,
) (*paymentdomain.RegistrationPricing, error) {
	if registrationID == "" {
		return nil, errors.New("registration_id is required")
	}

	// Unconditional lookup — no actor or guest email required.
	// Ownership is enforced by the payment service on the returned
	// snapshot, not here.
	reg, err := r.regRepo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("load registration: %w", err)
	}
	if reg == nil {
		return nil, fmt.Errorf("%w: %s", paymentdomain.ErrRegistrationNotFound, registrationID)
	}

	// Only pending registrations can become orders.
	if reg.Registration.Status != registrationdomain.StatusPending {
		return nil, fmt.Errorf(
			"registration %s is not pending (status: %s)",
			registrationID, reg.Registration.Status,
		)
	}

	// Flatten ticket selections.
	items := make([]paymentdomain.RegistrationPricingItem, 0, len(reg.Selections))
	for _, s := range reg.Selections {
		items = append(items, paymentdomain.RegistrationPricingItem{
			TicketTypeID: s.TicketTypeID,
			Quantity:     s.Quantity,
			UnitPrice:    s.UnitPrice,
			Discount:     s.Discount,
		})
	}

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