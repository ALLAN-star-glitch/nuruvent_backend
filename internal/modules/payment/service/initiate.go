// internal/modules/payment/service/initiate.go

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
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
//  4. Build the Payment in pending state with the provider's name.
//  5. Persist the payment.
//  6. Call provider.Initiate.
//  7. Update the payment with the provider's reference.
//  8. Notify.
//
// If the provider call fails, the payment is marked failed and the
// error is returned.
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
	// 3. Resolve the provider before any state change
	// ------------------------------------------------------------
	provider, err := s.deps.Providers.ByMethod(cmd.Method)
	if err != nil {
		return nil, err
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
	// 7. Update the payment with the provider's reference
	// ------------------------------------------------------------
	if result.ProviderReference != "" {
		payment.ProviderReference = result.ProviderReference
	}

	// If the provider immediately succeeded (rare, but possible for
	// pre-authorized charges), transition state accordingly.
	if result.Status == paymentdomain.PaymentStatusSucceeded {
		if err := payment.MarkSucceeded(result.ProviderReference, now); err != nil {
			return nil, err
		}
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