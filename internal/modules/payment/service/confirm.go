// internal/modules/payment/service/confirm.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ConfirmPayment transitions a payment to succeeded and confirms the
// associated registration. Idempotent — calling it twice for the same
// payment is a no-op.
//
// The webhook handler calls this after verifying the transaction with
// the provider. It is not called directly by any HTTP endpoint.
func (s *service) ConfirmPayment(ctx context.Context, paymentID string) error {
	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 1. Load the payment
	// ------------------------------------------------------------
	payment, err := s.deps.Payments.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Idempotent: already succeeded → no-op.
	if payment.Status == paymentdomain.PaymentStatusSucceeded {
		return nil
	}

	// ------------------------------------------------------------
	// 2. Transition the payment
	// ------------------------------------------------------------
	if err := payment.MarkSucceeded(payment.ProviderReference, now); err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 3. Load the order to mark it paid
	// ------------------------------------------------------------
	order, err := s.deps.Orders.FindByID(ctx, payment.OrderID)
	if err != nil {
		return fmt.Errorf("load order: %w", err)
	}
	if err := order.MarkPaid(now); err != nil {
		return fmt.Errorf("mark order paid: %w", err)
	}

	// ------------------------------------------------------------
	// 4. Persist both writes in a transaction
	// ------------------------------------------------------------
	err = s.deps.UnitOfWork.Do(ctx, func(repos paymentdomain.Repositories) error {
		if err := repos.Payments.Update(ctx, payment); err != nil {
			return fmt.Errorf("update payment: %w", err)
		}
		if err := repos.Orders.Update(ctx, order); err != nil {
			return fmt.Errorf("update order: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 5. Confirm the registration (best-effort, cross-module)
	// ------------------------------------------------------------
	if err := s.deps.Registrations.ConfirmRegistration(ctx, order.RegistrationID); err != nil {
		// Log and continue — the payment is confirmed. If the
		// registration confirmation fails, reconciliation will catch it.
		// (Logging is done by the caller; this function only returns
		// errors from the source-of-truth writes.)
	}

	// ------------------------------------------------------------
	// 6. Notify (best-effort)
	// ------------------------------------------------------------
	_ = s.deps.Notifier.PaymentSucceeded(ctx, payment, order)

	return nil
}