package payment

import (
    "context"
    "fmt"

    paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
    registrationService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
)

// BillingResolver implements paymentdomain.RegistrationBillingResolver
// by delegating to the registration module's service.
//
// Same shape as RegistrationConfirmer — the adapter is the only bridge
// between the payment and registration modules.
type BillingResolver struct {
    regSvc registrationService.Service
}

// NewBillingResolver constructs the adapter.
func NewBillingResolver(regSvc registrationService.Service) *BillingResolver {
    return &BillingResolver{regSvc: regSvc}
}

// ResolveBilledAccount returns the account that owns the event behind
// the given registration.
func (r *BillingResolver) ResolveBilledAccount(
    ctx context.Context,
    registrationID string,
) (string, error) {
    if registrationID == "" {
        return "", fmt.Errorf("registration id is required")
    }

    accountID, err := r.regSvc.ResolveBilledAccount(ctx, registrationID)
    if err != nil {
        return "", fmt.Errorf("resolve billed account for %s: %w", registrationID, err)
    }
    if accountID == "" {
        return "", fmt.Errorf("no account found for registration %s", registrationID)
    }
    return accountID, nil
}

// Compile-time assertion.
var _ paymentdomain.RegistrationBillingResolver = (*BillingResolver)(nil)