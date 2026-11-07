// internal/modules/payment/service/initiate.go

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// InitiatePayment creates a Payment for the given order and calls the
// provider to begin the flow.
//
// Steps:
//  1. Load the order and validate it's pending and not expired.
//  2. Check idempotency: if a payment already exists for
//     (order_id, idempotency_key), return it.
//  3. Resolve the provider. Fails before any state change if the
//     provider is not available.
//  3b. Resolve the processing fee rate for the chosen payment method
//      and snapshot it onto the order. Fails before any state change
//      if the rate is out of range.
//  4. Build the Payment in pending state with the provider's name.
//  5. Persist the payment.
//  6. Call provider.Initiate.
//  7. Update the payment with the provider's reference.
//  8. Notify.
//
// If the provider call fails, the payment is marked failed and the
// error is returned.
//
// IMPORTANT: a successful provider.Initiate does NOT mean the payment
// has succeeded. It only means the provider created a transaction. The
// actual charge completes later and is reported via webhook. Never
// transition the payment to succeeded from here — that breaks the
// webhook-driven confirmation path (see ConfirmPayment).
func (s *service) InitiatePayment(
	ctx context.Context,
	cmd InitiateCommand,
) (*paymentdomain.Payment, error) {
	now := s.deps.Clock.Now()

	// ------------------------------------------------------------
	// 1. Load the order
	// ------------------------------------------------------------
	order, err := s.deps.Orders.FindByID(ctx, cmd.OrderID)
	if err != nil {
		return nil, err
	}
	if !order.IsPending() {
		return nil, fmt.Errorf("%w: order %s is %s",
			paymentdomain.ErrOrderNotPending, order.ID, order.Status)
	}
	if order.IsExpired(now) {
		return nil, paymentdomain.ErrOrderExpired
	}

	// ------------------------------------------------------------
	// 2. Idempotency check
	// ------------------------------------------------------------
	if cmd.IdempotencyKey != "" {
		existing, err := s.deps.Payments.FindByIdempotencyKey(ctx, order.ID, cmd.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
		if err != nil && !errors.Is(err, paymentdomain.ErrPaymentNotFound) {
			return nil, fmt.Errorf("idempotency check: %w", err)
		}
	}

	// ------------------------------------------------------------
	// 2b. Pending payment check — only one live payment per order
	// ------------------------------------------------------------
	existingPending, err := s.deps.Payments.FindPendingByOrder(ctx, order.ID)
	if err == nil && existingPending != nil {
		if existingPending.IsExpired(now) {
			if failErr := existingPending.MarkFailed("payment expired", now); failErr == nil {
				_ = s.deps.Payments.Update(ctx, existingPending)
			}
		} else {
			return existingPending, nil
		}
	} else if err != nil && !errors.Is(err, paymentdomain.ErrPaymentNotFound) {
		return nil, fmt.Errorf("pending payment check: %w", err)
	}

	// ------------------------------------------------------------
	// 3. Resolve the provider before any state change
	// ------------------------------------------------------------
	provider, err := s.deps.Providers.ByMethod(cmd.Method)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 3b. Snapshot the processing fee rate for this method onto the
	//     order, then persist. This is the one point where the
	//     processing fee becomes concrete.
	//
	//     NOTE: PaymentMethod only distinguishes mpesa vs card. Card
	//     orders are charged the international rate (3.8%) as a
	//     conservative default until the webhook reports the card
	//     origin; a refund-difference step can refine this later.
	// ------------------------------------------------------------
	processingFeeRate, err := processingFeeRateForMethod(s.deps.Config.PaymentBilling, cmd.Method)
	if err != nil {
		return nil, fmt.Errorf("resolve processing fee rate: %w", err)
	}
	if err := order.SetProcessingFeeRate(processingFeeRate); err != nil {
		return nil, fmt.Errorf("set processing fee rate: %w", err)
	}
	if err := s.deps.Orders.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("persist order processing fee rate: %w", err)
	}

	// ------------------------------------------------------------
	// 4. Build the payment entity
	// ------------------------------------------------------------
	paymentID := s.deps.IDGenerator.NewID()
	providerTTL := 30 * time.Minute

	payment, err := paymentdomain.NewPayment(
		paymentID,
		order.ID,
		provider.Name(),
		cmd.Method,
		order.TotalAmount,
		order.Currency,
		cmd.IdempotencyKey,
		providerTTL,
		now,
	)
	if err != nil {
		return nil, err
	}

	// ------------------------------------------------------------
	// 5. Persist the payment
	// ------------------------------------------------------------
	if err := s.deps.Payments.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("persist payment: %w", err)
	}

	// ------------------------------------------------------------
	// 6. Call the provider
	// ------------------------------------------------------------
	initReq := paymentdomain.InitiateRequest{
		PaymentID:      payment.ID,
		OrderID:        order.ID,
		Amount:         order.TotalAmount,
		Currency:       order.Currency,
		Method:         cmd.Method,
		PayerPhone:     cmd.PayerPhone,
		PayerEmail:     cmd.PayerEmail,
		Card:           cmd.Card,
		IdempotencyKey: cmd.IdempotencyKey,
		Description:    cmd.Description,
		ReturnURL:      cmd.ReturnURL,
	}

	result, err := provider.Initiate(ctx, initReq)
	if err != nil {
		_ = payment.MarkFailed(err.Error(), now)
		_ = s.deps.Payments.Update(ctx, payment)
		return nil, fmt.Errorf("provider initiate: %w", err)
	}

	// ------------------------------------------------------------
	// 7. Update the payment with the provider's metadata
	// ------------------------------------------------------------
	if result.ProviderReference != "" {
		payment.ProviderReference = result.ProviderReference
	}
	if result.RedirectURL != "" {
		payment.RedirectURL = result.RedirectURL
	}
	if result.AccessCode != "" {
		payment.AccessCode = result.AccessCode
	}

	if result.Status == paymentdomain.PaymentStatusSucceeded {
		fmt.Printf("[InitiatePayment] provider %s reported immediate success for payment %s; awaiting webhook confirmation\n",
			provider.Name(), payment.ID)
	}

	payment.UpdatedAt = now

	if err := s.deps.Payments.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("update payment: %w", err)
	}

	// ------------------------------------------------------------
	// 8. Notify (best-effort)
	// ------------------------------------------------------------
	_ = s.deps.Notifier.PaymentInitiated(ctx, payment, order)

	return payment, nil
}

// ============================================================
// PROCESSING FEE RATE RESOLUTION
// ============================================================

// processingFeeRateForMethod maps a payment method to the
// corresponding processing fee rate from config.
//
// Returns an error for unknown methods rather than defaulting silently,
// so a misconfigured provider surfaces as a startup failure instead of
// a wrong payout later.
//
// NOTE ON CARD RATE
// -----------------
// The domain only distinguishes PaymentMethodMpesa vs
// PaymentMethodCard. The config carries two card rates (local 2.9%,
// international 3.8%), but the local/international distinction is not
// available at this layer — PaymentMethod doesn't carry it, and
// InitiateCommand doesn't either.
//
// We charge the international rate (3.8%) on all cards as a
// conservative default. This ensures Nuruvent never under-recovers
// processing cost. When the webhook reports card origin, a follow-up
// step can refund the 0.9% difference to the host for local cards.
// See the "card origin refinement" TODO in webhook.go / ConfirmPayment.
//
// Batch certificate orders are unaffected — SetProcessingFeeRate
// no-ops on that kind.
func processingFeeRateForMethod(
	billing config.PaymentBillingConfig,
	method paymentdomain.PaymentMethod,
) (float64, error) {
	switch method {
	case paymentdomain.PaymentMethodMpesa:
		// M-Pesa rate. Airtel is not a separate method in the current
		// enum — if/when it is added, route it through the same rate
		// (or add a config field for it) at that point.
		return billing.ProcessingFeeRateMpesa, nil

	case paymentdomain.PaymentMethodCard:
		// Conservative: charge the international rate on all cards
		// until origin is resolved. See NOTE above.
		return billing.ProcessingFeeRateCardIntl, nil

	default:
		return 0, fmt.Errorf("unknown payment method %q", method)
	}
}