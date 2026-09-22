// internal/modules/payment/service/fail_test.go

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

func TestFailPayment_Success(t *testing.T) {
	order := newPendingOrder(t)
	payment := newPendingPayment(t, order.ID)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
		d.Orders.(*fakeOrderRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Order, error) {
			return order, nil
		}
	})

	err := svc.FailPayment(context.Background(), "pay-1", "insufficient funds")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.Status != paymentdomain.PaymentStatusFailed {
		t.Errorf("status: got %s, want failed", payment.Status)
	}
	if payment.FailureReason != "insufficient funds" {
		t.Errorf("failure_reason: got %q", payment.FailureReason)
	}

	// Verify notification was fired.
	if len(deps.Notifier.(*fakeNotifier).failed) != 1 {
		t.Error("expected 1 failed notification")
	}
}

func TestFailPayment_Idempotent(t *testing.T) {
	payment := newPendingPayment(t, "order-1")
	_ = payment.MarkFailed("already failed", payment.CreatedAt)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
	})

	err := svc.FailPayment(context.Background(), "pay-1", "another reason")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No write should have occurred.
	if len(deps.Payments.(*fakePaymentRepository).updated) != 0 {
		t.Error("should not update an already-failed payment")
	}
}

func TestFailPayment_NotFound(t *testing.T) {
	svc, _ := newTestService(t, nil)

	err := svc.FailPayment(context.Background(), "missing", "reason")
	if !errors.Is(err, paymentdomain.ErrPaymentNotFound) {
		t.Fatalf("expected ErrPaymentNotFound, got %v", err)
	}
}