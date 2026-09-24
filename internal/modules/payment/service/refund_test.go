// internal/modules/payment/service/refund_test.go

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// newSucceededPayment builds a succeeded payment for refund tests.
func newSucceededPayment(t *testing.T) *paymentdomain.Payment {
	t.Helper()
	p := newPendingPayment(t, "order-1")
	p.ProviderReference = "FLW-REF-1"
	_ = p.MarkSucceeded("FLW-REF-1", p.CreatedAt)
	return p
}

func TestRefund_FullSuccess(t *testing.T) {
	payment := newSucceededPayment(t)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
		d.Providers.(*fakeProviderRegistry).byNameFunc = func(name string) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{
				name:   name,
				method: paymentdomain.PaymentMethodMpesa,
				refundFunc: func(ctx context.Context, req paymentdomain.RefundRequest) (*paymentdomain.RefundResult, error) {
					return &paymentdomain.RefundResult{
						ProviderReference: "REFUND-REF-1",
						Status:            paymentdomain.PaymentStatusSucceeded,
					}, nil
				},
			}, nil
		}
	})

	err := svc.Refund(context.Background(), RefundCommand{
		PaymentID: "pay-1",
		Amount:    payment.Amount, // full refund
		Reason:    "customer request",
		ActorID:   "admin-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Payment should be marked refunded.
	updatedPayments := deps.Payments.(*fakePaymentRepository).updated
	if len(updatedPayments) != 1 {
		t.Fatalf("expected 1 payment update, got %d", len(updatedPayments))
	}
	if updatedPayments[0].Status != paymentdomain.PaymentStatusRefunded {
		t.Errorf("payment status: got %s, want refunded", updatedPayments[0].Status)
	}

	// Refund should be persisted and marked succeeded.
	updatedRefunds := deps.Refunds.(*fakeRefundRepository).updated
	if len(updatedRefunds) != 1 {
		t.Fatalf("expected 1 refund update, got %d", len(updatedRefunds))
	}
	if updatedRefunds[0].Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("refund status: got %s, want succeeded", updatedRefunds[0].Status)
	}

	// Notification should have fired.
	if len(deps.Notifier.(*fakeNotifier).refunded) != 1 {
		t.Error("expected 1 refund notification")
	}
}

func TestRefund_PartialDoesNotMarkPaymentRefunded(t *testing.T) {
	payment := newSucceededPayment(t)

	svc, deps := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
		d.Providers.(*fakeProviderRegistry).byNameFunc = func(name string) (paymentdomain.PaymentProvider, error) {
			return &fakeProvider{name: name, method: paymentdomain.PaymentMethodMpesa}, nil
		}
	})

	err := svc.Refund(context.Background(), RefundCommand{
		PaymentID: "pay-1",
		Amount:    payment.Amount / 2, // partial
		Reason:    "partial refund",
		ActorID:   "admin-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Payment should stay succeeded.
	updated := deps.Payments.(*fakePaymentRepository).updated
	if len(updated) != 0 {
		t.Error("payment should not be updated for a partial refund")
	}
}

func TestRefund_ExceedsPayment(t *testing.T) {
	payment := newSucceededPayment(t)

	svc, _ := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
	})

	err := svc.Refund(context.Background(), RefundCommand{
		PaymentID: "pay-1",
		Amount:    payment.Amount + 1,
		Reason:    "too much",
		ActorID:   "admin-1",
	})
	if !errors.Is(err, paymentdomain.ErrRefundExceedsPayment) {
		t.Fatalf("expected ErrRefundExceedsPayment, got %v", err)
	}
}

func TestRefund_PaymentNotSucceeded(t *testing.T) {
	payment := newPendingPayment(t, "order-1") // still pending

	svc, _ := newTestService(t, func(d *Dependencies) {
		d.Payments.(*fakePaymentRepository).findByIDFunc = func(ctx context.Context, id string) (*paymentdomain.Payment, error) {
			return payment, nil
		}
	})

	err := svc.Refund(context.Background(), RefundCommand{
		PaymentID: "pay-1",
		Amount:    1000,
		ActorID:   "admin-1",
	})
	if !errors.Is(err, paymentdomain.ErrPaymentNotSucceeded) {
		t.Fatalf("expected ErrPaymentNotSucceeded, got %v", err)
	}
}