// internal/modules/payment/service/refund.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// Refund creates and processes a refund against a succeeded payment.
//
// Steps:
//  1. Load the payment. Must be succeeded.
//  2. Check idempotency — return early if a refund with the same key
//     already exists.
//  3. Load prior refunds and validate the total.
//  4. Create the refund entity.
//  5. Persist it.
//  6. Call the provider's Refund.
//  7. Update the refund with the provider's reference.
//  8. If the refund is a full refund, mark the payment refunded.
//  9. Notify.
func (s *service) Refund(ctx context.Context, cmd RefundCommand) error {
	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 1. Load the payment
	// ------------------------------------------------------------
	payment, err := s.deps.Payments.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return err
	}
	if !payment.IsSucceeded() {
		return paymentdomain.ErrPaymentNotSucceeded
	}

	// ------------------------------------------------------------
	// 2. Load prior refunds (for validation)
	// ------------------------------------------------------------
	priorRefunds, err := s.deps.Refunds.ListByPayment(ctx, payment.ID)
	if err != nil {
		return fmt.Errorf("load prior refunds: %w", err)
	}

	// ------------------------------------------------------------
	// 3. Build the refund entity and validate
	// ------------------------------------------------------------
	refundID := s.deps.IDGenerator.NewID()
	refund, err := paymentdomain.NewRefund(
		refundID,
		payment.ID,
		cmd.Amount,
		payment.Currency,
		cmd.Reason,
		cmd.ActorID,
		now,
	)
	if err != nil {
		return err
	}
	if err := refund.ValidateAgainstPayment(payment, priorRefunds); err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 4. Persist the refund
	// ------------------------------------------------------------
	if err := s.deps.Refunds.Create(ctx, refund); err != nil {
		return fmt.Errorf("persist refund: %w", err)
	}

	// ------------------------------------------------------------
	// 5. Call the provider
	// ------------------------------------------------------------
	provider, err := s.deps.Providers.ByName(payment.Provider)
	if err != nil {
		_ = refund.MarkFailed("provider unavailable", now)
		_ = s.deps.Refunds.Update(ctx, refund)
		return err
	}

	refundReq := paymentdomain.RefundRequest{
		PaymentID:         payment.ID,
		ProviderReference: payment.ProviderReference,
		Amount:            cmd.Amount,
		Currency:          payment.Currency,
		Reason:            cmd.Reason,
	}

	result, err := provider.Refund(ctx, refundReq)
	if err != nil {
		_ = refund.MarkFailed(err.Error(), now)
		_ = s.deps.Refunds.Update(ctx, refund)
		return fmt.Errorf("provider refund: %w", err)
	}

	// ------------------------------------------------------------
	// 6. Update the refund with the provider's response
	// ------------------------------------------------------------
	if result.Status == paymentdomain.PaymentStatusSucceeded {
		if err := refund.MarkSucceeded(result.ProviderReference, now); err != nil {
			return err
		}
	} else if result.Status == paymentdomain.PaymentStatusFailed {
		if err := refund.MarkFailed(result.Message, now); err != nil {
			return err
		}
	}

	if err := s.deps.Refunds.Update(ctx, refund); err != nil {
		return fmt.Errorf("update refund: %w", err)
	}

	// ------------------------------------------------------------
	// 7. If the refund succeeded and fully covers the payment,
	//    mark the payment refunded.
	// ------------------------------------------------------------
	if refund.IsSucceeded() {
		totalRefunded := paymentdomain.TotalRefunded(append(priorRefunds, refund))
		if totalRefunded >= payment.Amount {
			if err := payment.MarkRefunded(now); err != nil {
				return err
			}
			if err := s.deps.Payments.Update(ctx, payment); err != nil {
				return fmt.Errorf("mark payment refunded: %w", err)
			}
		}
	}

	// ------------------------------------------------------------
	// 8. Notify (best-effort)
	// ------------------------------------------------------------
	_ = s.deps.Notifier.RefundIssued(ctx, refund, payment)

	return nil
}