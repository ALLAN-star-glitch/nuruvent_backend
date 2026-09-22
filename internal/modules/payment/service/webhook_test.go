// internal/modules/payment/service/webhook_test.go

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

func TestHandleWebhook_SuccessfulCharge(t *testing.T) {
	order := newPendingOrder(t)
	payment := newPendingPayment(t, order.ID)
	payment.ProviderReference = "FLW-REF-1"

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Providers.(*fakeProviderRegistry).byNameFunc = func(name string) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{
				name:   name,
				method: paymentdomain.PaymentMethodMpesa,
				parseWebhookFunc: func(ctx context.Context, payload []byte, headers map[string]string) (*paymentdomain.WebhookEventData, error) {
					return &paymentdomain.WebhookEventData{
						ProviderEventID:   "evt-1",
						ProviderReference: "FLW-REF-1",
						OurPaymentID:      "pay-1",
						Status:            paymentdomain.PaymentStatusSucceeded,
						Amount:            payment.Amount,
						Currency:          payment.Currency,
					}, nil
				},
				verifyFunc: func(ctx context.Context, ref string) (*paymentdomain.ProviderStatus, error) {
					return &paymentdomain.ProviderStatus{
						Reference: ref,
						Status:    paymentdomain.PaymentStatusSucceeded,
						Amount:    payment.Amount,
						Currency:  payment.Currency,
					}, nil
				},
			}, nil
		}
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
	})

	err := svc.HandleWebhook(context.Background(), "flutterwave", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("payment status: got %s, want succeeded", payment.Status)
	}

	// Verify the webhook event was recorded and marked processed.
	created := deps.Webhooks.(*fakeWebhookEventRepository).created
	if len(created) != 1 {
		t.Fatalf("expected 1 webhook event recorded, got %d", len(created))
	}
	if !created[0].IsProcessed() {
		t.Error("webhook event should be marked processed")
	}
}

func TestHandleWebhook_DuplicateIgnored(t *testing.T) {
	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Providers.(*fakeProviderRegistry).byNameFunc = func(name string) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{name: name}, nil
		}
		d.Webhooks.(*fakeWebhookEventRepository).recordIfNewFunc = func(ctx context.Context, e *paymentdomain.WebhookEvent) error {
			return paymentdomain.ErrDuplicateWebhook
		}
	})

	err := svc.HandleWebhook(context.Background(), "flutterwave", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("duplicate webhook should be silent: %v", err)
	}

	// No payment/order updates should have occurred.
	if len(deps.Payments.(*fakePaymentRepository).updated) != 0 {
		t.Error("duplicate webhook should not update payments")
	}
}

func TestHandleWebhook_AmountMismatch(t *testing.T) {
	payment := newPendingPayment(t, "order-1")

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Providers.(*fakeProviderRegistry).byNameFunc = func(name string) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{
				name: name,
				parseWebhookFunc: func(ctx context.Context, payload []byte, headers map[string]string) (*paymentdomain.WebhookEventData, error) {
					return &paymentdomain.WebhookEventData{
						ProviderEventID: "evt-1",
						OurPaymentID:    "pay-1",
						Status:          paymentdomain.PaymentStatusSucceeded,
					}, nil
				},
				verifyFunc: func(ctx context.Context, ref string) (*paymentdomain.ProviderStatus, error) {
					return &paymentdomain.ProviderStatus{
						Reference: ref,
						Status:    paymentdomain.PaymentStatusSucceeded,
						Amount:    payment.Amount + 100, // mismatch
						Currency:  payment.Currency,
					}, nil
				},
			}, nil
		}
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
	})

	err := svc.HandleWebhook(context.Background(), "flutterwave", []byte(`{}`), map[string]string{})
	if !errors.Is(err, paymentdomain.ErrPricingMismatch) {
		t.Fatalf("expected ErrPricingMismatch, got %v", err)
	}

	// Webhook should be marked failed.
	updated := deps.Webhooks.(*fakeWebhookEventRepository).updated
	if len(updated) != 1 {
		t.Fatalf("expected 1 webhook update, got %d", len(updated))
	}
	if !updated[0].HasFailed() {
		t.Error("webhook should be marked failed")
	}
}

func TestHandleWebhook_FailedCharge(t *testing.T) {
	payment := newPendingPayment(t, "order-1")

	svc, _ := newTestService(t, func(d *Dependencies) {
		d.Providers.(*fakeProviderRegistry).byNameFunc = func(name string) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{
				name: name,
				parseWebhookFunc: func(ctx context.Context, payload []byte, headers map[string]string) (*paymentdomain.WebhookEventData, error) {
					return &paymentdomain.WebhookEventData{
						ProviderEventID: "evt-1",
						OurPaymentID:    "pay-1",
						Status:          paymentdomain.PaymentStatusFailed,
						Message:         "insufficient funds",
					}, nil
				},
				verifyFunc: func(ctx context.Context, ref string) (*paymentdomain.ProviderStatus, error) {
					return &paymentdomain.ProviderStatus{
						Reference: ref,
						Status:    paymentdomain.PaymentStatusFailed,
						Message:   "insufficient funds",
					}, nil
				},
			}, nil
		}
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
	})

	err := svc.HandleWebhook(context.Background(), "flutterwave", []byte(`{}`), map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.Status != paymentdomain.PaymentStatusFailed {
		t.Errorf("payment status: got %s, want failed", payment.Status)
	}
	if payment.FailureReason != "insufficient funds" {
		t.Errorf("failure_reason: got %q", payment.FailureReason)
	}
}