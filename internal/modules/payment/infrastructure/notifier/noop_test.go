
// internal/modules/payment/infrastructure/notifier/noop_test.go

package notifier

import (
	"context"
	"testing"
)

func TestNoop_AllMethodsReturnNil(t *testing.T) {
	n := NewNoop()
	ctx := context.Background()

	if err := n.PaymentInitiated(ctx, nil, nil); err != nil {
		t.Errorf("PaymentInitiated: %v", err)
	}
	if err := n.PaymentSucceeded(ctx, nil, nil); err != nil {
		t.Errorf("PaymentSucceeded: %v", err)
	}
	if err := n.PaymentFailed(ctx, nil, nil); err != nil {
		t.Errorf("PaymentFailed: %v", err)
	}
	if err := n.PaymentExpired(ctx, nil, nil); err != nil {
		t.Errorf("PaymentExpired: %v", err)
	}
	if err := n.RefundIssued(ctx, nil, nil); err != nil {
		t.Errorf("RefundIssued: %v", err)
	}
}