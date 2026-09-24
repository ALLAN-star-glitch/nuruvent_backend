// internal/modules/payment/service/webhook.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// HandleWebhook receives a raw webhook from a provider, verifies it,
// records it for audit, cross-checks it with the provider, and applies
// the outcome to the relevant payment.
//
// Steps:
//  1. Resolve the provider by name.
//  2. Parse + verify the challenge/signature via provider.ParseWebhook.
//  3. Record the raw webhook event (append-only audit log).
//  4. Cross-check server-side via provider.Verify (M-Pesa only — see
//     below). If the cross-check fails, the event data is used as the
//     source of truth — the verified webhook payload is already
//     authoritative.
//  5. Route by state: succeeded → ConfirmPayment; failed → FailPayment.
//  6. Mark the webhook event processed.
//
// The method is idempotent. Duplicate deliveries — the same webhook
// sent twice, or successive webhooks for the same transaction — produce
// exactly one state change. This is enforced by the payment state
// machine, not by event deduplication.
func (s *service) HandleWebhook(
	ctx context.Context,
	providerName string,
	payload []byte,
	headers map[string]string,
) error {
	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 1. Resolve the provider
	// ------------------------------------------------------------
	provider, err := s.deps.Providers.ByName(providerName)
	if err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 2. Parse and verify the challenge/signature
	// ------------------------------------------------------------
	event, err := provider.ParseWebhook(ctx, payload, headers)
	if err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 3. Record the webhook event (append-only audit)
	// ------------------------------------------------------------
	webhookEvent := paymentdomain.NewWebhookEvent(
		s.deps.IDGenerator.NewID(),
		providerName,
		event.ProviderEventID,
		true, // signature already verified by ParseWebhook
		payload,
		now,
	)

	if err := s.deps.Webhooks.RecordIfNew(ctx, webhookEvent); err != nil {
		return fmt.Errorf("record webhook: %w", err)
	}

	// ------------------------------------------------------------
	// 4. Load the payment (needed for the card check + amount cross-check)
	// ------------------------------------------------------------
	payment, err := s.deps.Payments.FindByID(ctx, event.OurPaymentID)
	if err != nil {
		webhookEvent.MarkFailed(fmt.Sprintf("load payment: %v", err))
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return fmt.Errorf("load payment: %w", err)
	}

	// ------------------------------------------------------------
	// 5. Cross-check with the provider (best-effort, M-Pesa only)
	// ------------------------------------------------------------
	// Card checkouts use IntaSend's public /checkout/ endpoint. Their
	// status lookup rejects those invoice IDs with "Invalid api token",
	// so we skip Verify for card and trust the challenge-verified
	// webhook payload.
	//
	// For M-Pesa, Verify is a redundant safety net. If it fails, log
	// and continue with the event data rather than aborting.
	var status *paymentdomain.ProviderStatus
	if payment.Method == paymentdomain.PaymentMethodCard {
		status = &paymentdomain.ProviderStatus{
			Reference: event.ProviderReference,
			Status:    event.Status,
			Amount:    event.Amount,
			Currency:  event.Currency,
			Message:   event.Message,
			UpdatedAt: now,
		}
	} else {
		status, err = provider.Verify(ctx, event.ProviderReference)
		if err != nil {
			log.Printf("webhook verify failed (using event data): %v", err)
			status = &paymentdomain.ProviderStatus{
				Reference: event.ProviderReference,
				Status:    event.Status,
				Amount:    event.Amount,
				Currency:  event.Currency,
				Message:   event.Message,
				UpdatedAt: now,
			}
		}
	}

	// Cross-check: the amount and currency from the provider must
	// match what we recorded.
	if status.Amount != 0 && status.Amount != payment.Amount {
		err := fmt.Errorf("%w: expected %d, got %d",
			paymentdomain.ErrPricingMismatch, payment.Amount, status.Amount)
		webhookEvent.MarkFailed(err.Error())
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return err
	}
	if status.Currency != "" && status.Currency != payment.Currency {
		err := fmt.Errorf("%w: expected %s, got %s",
			paymentdomain.ErrInvalidCurrency, payment.Currency, status.Currency)
		webhookEvent.MarkFailed(err.Error())
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return err
	}

	// ------------------------------------------------------------
	// 6. Route based on the verified status
	// ------------------------------------------------------------
	var applyErr error
	switch status.Status {
	case paymentdomain.PaymentStatusSucceeded:
		applyErr = s.ConfirmPayment(ctx, payment.ID)
	case paymentdomain.PaymentStatusFailed:
		applyErr = s.FailPayment(ctx, payment.ID, status.Message)
	default:
		// Pending or unknown — leave the webhook event unprocessed
		// and return. A future event will close it out.
		webhookEvent.MarkProcessed(now)
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return nil
	}

	// ------------------------------------------------------------
	// 7. Handle apply errors
	// ------------------------------------------------------------
	if applyErr != nil {
		// Idempotent replay: the payment is already in the target
		// state (e.g. already succeeded). Treat as success and return
		// 200 so IntaSend stops retrying.
		if errors.Is(applyErr, paymentdomain.ErrInvalidStatusTransition) {
			webhookEvent.MarkProcessed(now)
			_ = s.deps.Webhooks.Update(ctx, webhookEvent)
			return nil
		}

		webhookEvent.MarkFailed(applyErr.Error())
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return applyErr
	}

	// ------------------------------------------------------------
	// 8. Mark the webhook event processed
	// ------------------------------------------------------------
	webhookEvent.MarkProcessed(now)
	if err := s.deps.Webhooks.Update(ctx, webhookEvent); err != nil {
		return fmt.Errorf("mark webhook processed: %w", err)
	}

	return nil
}