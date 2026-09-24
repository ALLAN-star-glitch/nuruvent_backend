// internal/app/adapters/payment/registration_confirmer.go

package payment

import (
	"context"
	"fmt"

	paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	registrationService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
)

// RegistrationConfirmer implements paymentdomain.RegistrationConfirmer
// by delegating to the registration module's service.
//
// The payment module never imports the registration module directly.
// This adapter is the only bridge between them, and it lives in the
// app's composition layer.
//
// Both methods are best-effort from the payment service's perspective:
// if they fail, the payment still succeeded (or expired). The failure
// is logged but not rolled back. Reconciliation picks up the drift.
type RegistrationConfirmer struct {
	regSvc registrationService.Service
}

// NewRegistrationConfirmer constructs the adapter.
func NewRegistrationConfirmer(regSvc registrationService.Service) *RegistrationConfirmer {
	return &RegistrationConfirmer{regSvc: regSvc}
}

// ConfirmRegistration transitions a pending registration to confirmed.
// Called by the payment service when a payment succeeds.
func (c *RegistrationConfirmer) ConfirmRegistration(
	ctx context.Context,
	registrationID string,
) error {
	if err := c.regSvc.ConfirmRegistration(ctx, registrationID); err != nil {
		return fmt.Errorf("confirm registration %s: %w", registrationID, err)
	}
	return nil
}

// ExpirePending transitions a pending registration to expired.
// Called by the payment service when a payment window elapses.
//
// NOTE: the registration module's method is called ExpirePending,
// which is why this adapter's method has the same name.
func (c *RegistrationConfirmer) ExpirePending(
	ctx context.Context,
	registrationID string,
) error {
	if err := c.regSvc.ExpirePending(ctx, registrationID); err != nil {
		return fmt.Errorf("expire registration %s: %w", registrationID, err)
	}
	return nil
}

// Compile-time assertion.
var _ paymentdomain.RegistrationConfirmer = (*RegistrationConfirmer)(nil)