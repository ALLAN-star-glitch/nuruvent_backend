// internal/modules/payment/service/confirm_test.go

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

func TestConfirmPayment_Success(t *testing.T) {
	order := newPendingOrder(t)
	payment := newPendingPayment(t, order.ID)
	payment.ProviderReference = "FLW-REF-1"

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
	})

	err := svc.ConfirmPayment(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("payment status: got %s, want succeeded", payment.Status)
	}
	if order.Status != paymentdomain.OrderStatusPaid {
		t.Errorf("order status: got %s, want paid", order.Status)
	}

	// Verify registration confirmation was called.
	confirmer := deps.Registrations.(*fakeRegistrationConfirmer)
	if len(confirmer.confirmed) != 1 {
		t.Fatalf("expected 1 registration confirm, got %d", len(confirmer.confirmed))
	}
	if confirmer.confirmed[0] != order.RegistrationID {
		t.Errorf("registration id: got %q, want %q", confirmer.confirmed[0], order.RegistrationID)
	}

	// Verify notification was fired.
	if len(deps.Notifier.(*fakeNotifier).succeeded) != 1 {
		t.Error("expected 1 succeeded notification")
	}
}

func TestConfirmPayment_IdempotentWhenAlreadySucceeded(t *testing.T) {
	order := newPendingOrder(t)
	payment := newPendingPayment(t, order.ID)
	payment.ProviderReference = "FLW-REF-1"
	_ = payment.MarkSucceeded("FLW-REF-1", payment.CreatedAt)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
	})

	err := svc.ConfirmPayment(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No writes should have happened.
	if len(deps.Payments.(*fakePaymentRepository).updated) != 0 {
		t.Error("should not update payment on idempotent call")
	}
	if len(deps.Registrations.(*fakeRegistrationConfirmer).confirmed) != 0 {
		t.Error("should not confirm registration again")
	}
}

func TestConfirmPayment_PaymentNotFound(t *testing.T) {
	svc, _ := newTestService(t, nil)

	err := svc.ConfirmPayment(context.Background(), "missing")
	if !errors.Is(err, paymentdomain.ErrPaymentNotFound) {
		t.Fatalf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestConfirmPayment_RegistrationFailureIsBestEffort(t *testing.T) {
	order := newPendingOrder(t)
	payment := newPendingPayment(t, order.ID)
	payment.ProviderReference = "FLW-REF-1"

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
		d.Registrations.(*fakeRegistrationConfirmer).confirmFunc = func(ctx context.Context, regID string) error {
			return errors.New("registration module down")
		}
	})

	// The payment still succeeds — registration failure is best-effort.
	err := svc.ConfirmPayment(context.Background(), "pay-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("payment status: got %s, want succeeded", payment.Status)
	}

	// Verify the registration confirm was attempted.
	if len(deps.Registrations.(*fakeRegistrationConfirmer).confirmed) != 1 {
		t.Error("registration confirm should have been attempted")
	}
}