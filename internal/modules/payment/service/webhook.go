// internal/modules/payment/service/webhook.go

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// HandleWebhook receives a raw webhook from a provider, verifies it,
// deduplicates it, verifies it server-side, and applies the outcome to
// the relevant payment.
//
// Steps:
//  1. Resolve the provider by name.
//  2. Parse + verify signature via provider.ParseWebhook.
//  3. Record the raw webhook event (dedupe).
//     If it's a duplicate, return nil (already processed).
//  4. Verify server-side via provider.Verify.
//  5. Route: succeeded → ConfirmPayment; failed → FailPayment.
//  6. Mark the webhook event processed.
//
// The method is idempotent — duplicate deliveries result in exactly one
// state change.
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
	// 2. Parse and verify signature
	// ------------------------------------------------------------
	event, err := provider.ParseWebhook(ctx, payload, headers)
	if err != nil {
		return err
	}

	// ------------------------------------------------------------
	// 3. Record the webhook event (dedupe)
	// ------------------------------------------------------------
	webhookEvent := paymentdomain.NewWebhookEvent(
		s.deps.IDGenerator.NewID(),
		providerName,
		event.ProviderEventID,
		true, // signature already verified by ParseWebhook
		payload,
		now,
	)

	err = s.deps.Webhooks.RecordIfNew(ctx, webhookEvent)
	if errors.Is(err, paymentdomain.ErrDuplicateWebhook) {
		// Already processed. Silent success.
		return nil
	}
	if err != nil {
		return fmt.Errorf("record webhook: %w", err)
	}

	// ------------------------------------------------------------
	// 4. Verify server-side
	// ------------------------------------------------------------
	status, err := provider.Verify(ctx, event.OurPaymentID)
	if err != nil {
		webhookEvent.MarkFailed(fmt.Sprintf("verify: %v", err))
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return fmt.Errorf("provider verify: %w", err)
	}

	// Cross-check: the amount and currency from the provider must
	// match what we recorded.
	payment, err := s.deps.Payments.FindByID(ctx, event.OurPaymentID)
	if err != nil {
		webhookEvent.MarkFailed(fmt.Sprintf("load payment: %v", err))
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return fmt.Errorf("load payment: %w", err)
	}
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
	// 5. Route based on the verified status
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

	if applyErr != nil {
		webhookEvent.MarkFailed(applyErr.Error())
		_ = s.deps.Webhooks.Update(ctx, webhookEvent)
		return applyErr
	}

	// ------------------------------------------------------------
	// 6. Mark the webhook event processed
	// ------------------------------------------------------------
	webhookEvent.MarkProcessed(now)
	if err := s.deps.Webhooks.Update(ctx, webhookEvent); err != nil {
		return fmt.Errorf("mark webhook processed: %w", err)
	}

	return nil
}