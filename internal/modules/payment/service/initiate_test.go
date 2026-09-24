// internal/modules/payment/service/initiate_test.go

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

func TestInitiatePayment_Mpesa_Success(t *testing.T) {
	order := newPendingOrder(t)

	var providerCalled bool

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
		d.Providers.(*fakeProviderRegistry).byMethodFunc = func(method paymentdomain.PaymentMethod) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{
				name:   "flutterwave",
				method: paymentdomain.PaymentMethodMpesa,
				initiateFunc: func(ctx context.Context, req paymentdomain.InitiateRequest) (*paymentdomain.InitiateResult, error) {
					providerCalled = true
					if req.PaymentID != "id-1" {
						t.Errorf("payment id: got %q", req.PaymentID)
					}
					return &paymentdomain.InitiateResult{
						ProviderReference: "FLW-REF-1",
						Status:            paymentdomain.PaymentStatusPending,
					}, nil
				},
			}, nil
		}
	})

	payment, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID:        "order-1",
		Method:         paymentdomain.PaymentMethodMpesa,
		PayerPhone:     "+254709929220",
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !providerCalled {
		t.Error("provider.Initiate was not called")
	}
	if payment.ProviderReference != "FLW-REF-1" {
		t.Errorf("provider_reference: got %q", payment.ProviderReference)
	}
	if payment.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", payment.Status)
	}
	if payment.Provider != "flutterwave" {
		t.Errorf("provider name: got %q", payment.Provider)
	}

	// Verify the payment was persisted.
	created := deps.Payments.(*fakePaymentRepository).created
	if len(created) != 1 {
		t.Fatalf("expected 1 create, got %d", len(created))
	}

	// Verify the payment was updated with the provider reference.
	updated := deps.Payments.(*fakePaymentRepository).updated
	if len(updated) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updated))
	}

	// Verify notification was fired.
	notifier := deps.Notifier.(*fakeNotifier)
	if len(notifier.initiated) != 1 {
		t.Errorf("expected 1 initiated notification, got %d", len(notifier.initiated))
	}
}

func TestInitiatePayment_OrderNotFound(t *testing.T) {
	svc, _ := newTestService(t, nil) // default FindByID returns ErrOrderNotFound

	_, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID: "missing",
		Method:  paymentdomain.PaymentMethodMpesa,
	})
	if !errors.Is(err, paymentdomain.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestInitiatePayment_OrderNotPending(t *testing.T) {
	order := newPendingOrder(t)
	order.MarkPaid(order.CreatedAt) // move out of pending

	svc, _ := newTestService(t, func(d *Dependencies) {
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
	})

	_, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID: "order-1",
		Method:  paymentdomain.PaymentMethodMpesa,
	})
	if !errors.Is(err, paymentdomain.ErrOrderNotPending) {
		t.Fatalf("expected ErrOrderNotPending, got %v", err)
	}
}

func TestInitiatePayment_IdempotencyReturnsExisting(t *testing.T) {
	order := newPendingOrder(t)
	existing := newPendingPayment(t, order.ID)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
		d.Payments.(*fakePaymentRepository).findByIdempotencyKeyFunc = func(ctx context.Context, orderID, key string) (*paymentdomain.Payment, error) {
			return existing, nil
		}
	})

	payment, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID:        "order-1",
		Method:         paymentdomain.PaymentMethodMpesa,
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.ID != existing.ID {
		t.Errorf("expected existing payment ID %q, got %q", existing.ID, payment.ID)
	}

	// No new payment should be created.
	if len(deps.Payments.(*fakePaymentRepository).created) != 0 {
		t.Error("should not create a new payment for a duplicate key")
	}
}

func TestInitiatePayment_ProviderNotFound(t *testing.T) {
	order := newPendingOrder(t)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
		// ByMethod defaults to ErrProviderNotFound
	})

	_, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID: "order-1",
		Method:  paymentdomain.PaymentMethodMpesa,
	})
	if !errors.Is(err, paymentdomain.ErrProviderNotFound) {
		t.Fatalf("expected ErrProviderNotFound, got %v", err)
	}

	// No payment should have been created — the provider was
	// resolved before any state change.
	if len(deps.Payments.(*fakePaymentRepository).created) != 0 {
		t.Errorf("expected 0 payment creates, got %d",
			len(deps.Payments.(*fakePaymentRepository).created))
	}

	// No payment update should have occurred either.
	if len(deps.Payments.(*fakePaymentRepository).updated) != 0 {
		t.Errorf("expected 0 payment updates, got %d",
			len(deps.Payments.(*fakePaymentRepository).updated))
	}
}

func TestInitiatePayment_ProviderInitiateFails(t *testing.T) {
	order := newPendingOrder(t)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
		d.Providers.(*fakeProviderRegistry).byMethodFunc = func(method paymentdomain.PaymentMethod) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{
				name:   "flutterwave",
				method: paymentdomain.PaymentMethodMpesa,
				initiateFunc: func(ctx context.Context, req paymentdomain.InitiateRequest) (*paymentdomain.InitiateResult, error) {
					return nil, errors.New("provider rejected")
				},
			}, nil
		}
	})

	_, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID:        "order-1",
		Method:         paymentdomain.PaymentMethodMpesa,
		IdempotencyKey: "idem-provider-fail",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	// Payment should be marked failed.
	updated := deps.Payments.(*fakePaymentRepository).updated
	if len(updated) != 1 {
		t.Fatalf("expected 1 update (mark failed), got %d", len(updated))
	}
	if updated[0].Status != paymentdomain.PaymentStatusFailed {
		t.Errorf("status: got %s, want failed", updated[0].Status)
	}
}

func TestInitiatePayment_UsesClockForTimestamp(t *testing.T) {
	order := newPendingOrder(t)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
		d.Providers.(*fakeProviderRegistry).byMethodFunc = func(method paymentdomain.PaymentMethod) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{name: "flutterwave", method: paymentdomain.PaymentMethodMpesa}, nil
		}
	})

	_, err := svc.InitiatePayment(context.Background(), InitiateCommand{
		OrderID:        "order-1",
		Method:         paymentdomain.PaymentMethodMpesa,
		IdempotencyKey: "idem-clock-test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	created := deps.Payments.(*fakePaymentRepository).created[0]
	expectedTime := deps.Clock.Now()
	if !created.CreatedAt.Equal(expectedTime) {
		t.Errorf("created_at: got %v, want %v", created.CreatedAt, expectedTime)
	}
}