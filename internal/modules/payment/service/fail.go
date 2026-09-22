// internal/modules/payment/service/fail.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// FailPayment transitions a payment to failed with a reason. Idempotent.
//
// Called when:
//   - The webhook reports a failed charge.
//   - The provider rejects the initiation.
//   - Reconciliation determines the payment will not succeed.
func (s *service) FailPayment(ctx context.Context, paymentID, reason string) error {
	now := s.deps.Clock.Now()

	payment, err := s.deps.Payments.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Idempotent: already failed → no-op.
	if payment.Status == paymentdomain.PaymentStatusFailed {
		return nil
	}

	if err := payment.MarkFailed(reason, now); err != nil {
		return err
	}

	if err := s.deps.Payments.Update(ctx, payment); err != nil {
		return fmt.Errorf("update payment: %w", err)
	}

	// Load the order for the notification (best-effort).
	order, err := s.deps.Orders.FindByID(ctx, payment.OrderID)
	if err == nil {
		_ = s.deps.Notifier.PaymentFailed(ctx, payment, order)
	}

	return nil
}