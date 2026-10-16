// internal/modules/payment/service/confirm.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ConfirmPayment transitions a payment to succeeded and confirms the
// associated order's post-payment side effects (registration confirm,
// certificate batch issuance).
//
// Idempotent for the payment state transition: calling it twice for the
// same payment will not double-transition.
//
// NOT idempotent-with-early-return for side effects. If the payment is
// already succeeded (e.g. an earlier replay ran but the batch issuance
// failed), we still attempt the batch reconciliation. IssueBatch is
// expected to be idempotent on its own side (checks
// certificate_batches.status == 'issued').
//
// Called by the webhook handler after verifying the transaction with
// the provider.
func (s *service) ConfirmPayment(ctx context.Context, paymentID string) error {
	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 1. Load the payment
	// ------------------------------------------------------------
	payment, err := s.deps.Payments.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 1a. Already succeeded — skip the state transition but still
	// reconcile post-payment side effects. This covers two cases:
	//
	//   (i)  A prior ConfirmPayment ran and completed everything.
	//        IssueBatch should no-op.
	//   (ii) A prior ConfirmPayment ran, transitioned the payment,
	//        but the batch issuance failed or never ran. This replay
	//        gives it another chance.
	//
	// IssueBatch is required to be idempotent; if it isn't, this will
	// double-issue. See certificate/service/issue_batch.go.
	// ------------------------------------------------------------
	if payment.Status == paymentdomain.PaymentStatusSucceeded {
		order, err := s.deps.Orders.FindByID(ctx, payment.OrderID)
		if err == nil && order.Kind == paymentdomain.OrderKindCertificateBatch &&
			s.deps.CertificateBatchIssuer != nil {
			if err := s.deps.CertificateBatchIssuer.IssueBatch(ctx, order, payment); err != nil {
				// Reconciliation catches this. Not fatal — the webhook
				// will be marked processed and the operator can retry.
				fmt.Printf("[ConfirmPayment] reconcile batch for already-succeeded payment %s: %v\n",
					payment.ID, err)
			}
		}
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
	// 5. Post-payment side effects — dispatched by order kind.
	// ------------------------------------------------------------
	switch order.Kind {
	case paymentdomain.OrderKindRegistration:
		if err := s.deps.Registrations.ConfirmRegistration(ctx, order.RegistrationID); err != nil {
			// Reconciliation catches this. Not fatal.
		}

	case paymentdomain.OrderKindCertificate:
		// Legacy attendee-pays path — kept for backward compat with
		// any pending certificate orders created before the batch
		// model shipped.
		if s.deps.CertificateIssuer != nil {
			if err := s.deps.CertificateIssuer.IssueFromPaidOrder(ctx, order, payment); err != nil {
				// Reconciliation catches this. Not fatal.
			}
		}

	case paymentdomain.OrderKindCertificateBatch:
    if s.deps.CertificateBatchIssuer != nil {
        if err := s.deps.CertificateBatchIssuer.IssueBatch(ctx, order, payment); err != nil {
            // Reconciliation catches this. Not fatal — log so silent
            // failures don't go unnoticed.
            log.Printf("[ConfirmPayment] batch issuance for order %s failed: %v", order.ID, err)
        }
    }
	}



	

	// ------------------------------------------------------------
	// 6. Notify (best-effort)
	// ------------------------------------------------------------
	_ = s.deps.Notifier.PaymentSucceeded(ctx, payment, order)

	return nil
}