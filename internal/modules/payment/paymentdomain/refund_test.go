// internal/modules/payment/paymentdomain/refund_test.go

package paymentdomain

import (
	"errors"
	"testing"
	"time"
)

// ============================================================
// CONSTRUCTION
// ============================================================

func TestNewRefund(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("valid", func(t *testing.T) {
		r, err := NewRefund(
			"refund-1", "pay-1",
			5000_00, "KES",
			"customer request", "actor-1",
			now,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status != PaymentStatusPending {
			t.Errorf("status: got %s, want %s", r.Status, PaymentStatusPending)
		}
		if r.Amount != 5000_00 {
			t.Errorf("amount: got %d, want %d", r.Amount, 5000_00)
		}
		if r.Currency != "KES" {
			t.Errorf("currency: got %q, want %q", r.Currency, "KES")
		}
	})

	t.Run("rejects empty id", func(t *testing.T) {
		_, err := NewRefund("", "pay-1", 100, "KES", "x", "actor", now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects empty payment id", func(t *testing.T) {
		_, err := NewRefund("refund-1", "", 100, "KES", "x", "actor", now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects zero amount", func(t *testing.T) {
		_, err := NewRefund("refund-1", "pay-1", 0, "KES", "x", "actor", now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects negative amount", func(t *testing.T) {
		_, err := NewRefund("refund-1", "pay-1", -100, "KES", "x", "actor", now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects empty currency", func(t *testing.T) {
		_, err := NewRefund("refund-1", "pay-1", 100, "", "x", "actor", now)
		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("expected ErrInvalidCurrency, got %v", err)
		}
	})

	t.Run("rejects empty actor", func(t *testing.T) {
		_, err := NewRefund("refund-1", "pay-1", 100, "KES", "x", "", now)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ============================================================
// LIFECYCLE
// ============================================================

func TestRefund_MarkSucceeded(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		r := newTestRefund(t, now)
		if err := r.MarkSucceeded("provider-ref-1", now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status != PaymentStatusSucceeded {
			t.Errorf("status: got %s, want %s", r.Status, PaymentStatusSucceeded)
		}
		if r.ProviderReference != "provider-ref-1" {
			t.Errorf("provider_reference: got %q", r.ProviderReference)
		}
		if r.CompletedAt == nil {
			t.Error("completed_at should be set")
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		r := newTestRefund(t, now)
		_ = r.MarkSucceeded("ref-1", now)
		first := *r.CompletedAt

		later := now.Add(time.Hour)
		if err := r.MarkSucceeded("ref-1", later); err != nil {
			t.Fatalf("second call should be no-op: %v", err)
		}
		if !r.CompletedAt.Equal(first) {
			t.Errorf("completed_at changed")
		}
	})

	t.Run("requires provider reference", func(t *testing.T) {
		r := newTestRefund(t, now)
		err := r.MarkSucceeded("", now)
		if err == nil {
			t.Fatal("expected error for empty reference")
		}
		if r.Status != PaymentStatusPending {
			t.Errorf("status must not change on failure: got %s", r.Status)
		}
	})

	t.Run("rejects when already failed", func(t *testing.T) {
		r := newTestRefund(t, now)
		_ = r.MarkFailed("provider rejected", now)

		err := r.MarkSucceeded("ref-1", now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

func TestRefund_MarkFailed(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		r := newTestRefund(t, now)
		if err := r.MarkFailed("provider rejected", now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Status != PaymentStatusFailed {
			t.Errorf("status: got %s, want %s", r.Status, PaymentStatusFailed)
		}
		if r.FailureReason != "provider rejected" {
			t.Errorf("failure_reason: got %q", r.FailureReason)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		r := newTestRefund(t, now)
		_ = r.MarkFailed("first", now)
		if err := r.MarkFailed("second", now); err != nil {
			t.Fatalf("second call should be no-op: %v", err)
		}
		if r.FailureReason != "first" {
			t.Errorf("failure_reason changed: got %q", r.FailureReason)
		}
	})

	t.Run("rejects when already succeeded", func(t *testing.T) {
		r := newTestRefund(t, now)
		_ = r.MarkSucceeded("ref-1", now)

		err := r.MarkFailed("late failure", now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

// ============================================================
// VALIDATION AGAINST PAYMENT
// ============================================================

func TestRefund_ValidateAgainstPayment(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	// Helper: create a succeeded payment of 10000 KES.
	newSucceededPayment := func(t *testing.T) *Payment {
		t.Helper()
		p, _ := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000_00, "KES", "idem-1",
			30*time.Minute, now)
		_ = p.MarkSucceeded("provider-ref-1", now)
		return p
	}

	t.Run("valid full refund", func(t *testing.T) {
		payment := newSucceededPayment(t)
		r, _ := NewRefund("refund-1", "pay-1", 10000_00, "KES", "full", "actor", now)

		if err := r.ValidateAgainstPayment(payment, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid partial refund", func(t *testing.T) {
		payment := newSucceededPayment(t)
		r, _ := NewRefund("refund-1", "pay-1", 3000_00, "KES", "partial", "actor", now)

		if err := r.ValidateAgainstPayment(payment, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid refund after prior succeeded refunds", func(t *testing.T) {
		payment := newSucceededPayment(t)

		prior := []*Refund{
			{ID: "r1", PaymentID: "pay-1", Amount: 4000_00, Status: PaymentStatusSucceeded},
			{ID: "r2", PaymentID: "pay-1", Amount: 3000_00, Status: PaymentStatusSucceeded},
		}
		// 4000 + 3000 = 7000 prior; new refund 3000 → total 10000 = exact
		r, _ := NewRefund("refund-3", "pay-1", 3000_00, "KES", "rest", "actor", now)

		if err := r.ValidateAgainstPayment(payment, prior); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects payment not succeeded", func(t *testing.T) {
		pendingPayment, _ := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000_00, "KES", "idem-1",
			30*time.Minute, now)

		r, _ := NewRefund("refund-1", "pay-1", 1000_00, "KES", "x", "actor", now)

		err := r.ValidateAgainstPayment(pendingPayment, nil)
		if !errors.Is(err, ErrPaymentNotSucceeded) {
			t.Fatalf("expected ErrPaymentNotSucceeded, got %v", err)
		}
	})

	t.Run("rejects nil payment", func(t *testing.T) {
		r, _ := NewRefund("refund-1", "pay-1", 1000_00, "KES", "x", "actor", now)

		err := r.ValidateAgainstPayment(nil, nil)
		if !errors.Is(err, ErrPaymentNotFound) {
			t.Fatalf("expected ErrPaymentNotFound, got %v", err)
		}
	})

	t.Run("rejects currency mismatch", func(t *testing.T) {
		payment := newSucceededPayment(t)
		r, _ := NewRefund("refund-1", "pay-1", 1000_00, "USD", "x", "actor", now)

		err := r.ValidateAgainstPayment(payment, nil)
		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("expected ErrInvalidCurrency, got %v", err)
		}
	})

	t.Run("rejects zero amount", func(t *testing.T) {
		payment := newSucceededPayment(t)
		// Build a Refund with zero amount by bypassing constructor
		r := &Refund{
			ID: "refund-1", PaymentID: "pay-1", Amount: 0, Currency: "KES",
		}

		err := r.ValidateAgainstPayment(payment, nil)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects exceeding payment with prior succeeded refunds", func(t *testing.T) {
		payment := newSucceededPayment(t)

		prior := []*Refund{
			{ID: "r1", PaymentID: "pay-1", Amount: 6000_00, Status: PaymentStatusSucceeded},
		}
		// 6000 prior + 5000 new = 11000 > 10000
		r, _ := NewRefund("refund-2", "pay-1", 5000_00, "KES", "x", "actor", now)

		err := r.ValidateAgainstPayment(payment, prior)
		if !errors.Is(err, ErrRefundExceedsPayment) {
			t.Fatalf("expected ErrRefundExceedsPayment, got %v", err)
		}
	})

	t.Run("rejects exceeding payment with prior pending refunds", func(t *testing.T) {
		payment := newSucceededPayment(t)

		// Pending refunds consume the budget too — prevents race where
		// two pending refunds both succeed and over-refund.
		prior := []*Refund{
			{ID: "r1", PaymentID: "pay-1", Amount: 6000_00, Status: PaymentStatusPending},
		}
		r, _ := NewRefund("refund-2", "pay-1", 5000_00, "KES", "x", "actor", now)

		err := r.ValidateAgainstPayment(payment, prior)
		if !errors.Is(err, ErrRefundExceedsPayment) {
			t.Fatalf("expected ErrRefundExceedsPayment, got %v", err)
		}
	})

	t.Run("ignores prior failed refunds", func(t *testing.T) {
		payment := newSucceededPayment(t)

		// Failed refunds don't consume the budget — the money didn't move.
		prior := []*Refund{
			{ID: "r1", PaymentID: "pay-1", Amount: 6000_00, Status: PaymentStatusFailed},
		}
		r, _ := NewRefund("refund-2", "pay-1", 6000_00, "KES", "x", "actor", now)

		if err := r.ValidateAgainstPayment(payment, prior); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ignores nil prior refunds", func(t *testing.T) {
		payment := newSucceededPayment(t)

		prior := []*Refund{nil, nil}
		r, _ := NewRefund("refund-1", "pay-1", 5000_00, "KES", "x", "actor", now)

		if err := r.ValidateAgainstPayment(payment, prior); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// ============================================================
// QUERIES
// ============================================================

func TestRefund_Queries(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("IsPending, IsSucceeded, IsFailed", func(t *testing.T) {
		r := newTestRefund(t, now)
		if !r.IsPending() {
			t.Error("expected pending")
		}

		_ = r.MarkSucceeded("ref-1", now)
		if !r.IsSucceeded() {
			t.Error("expected succeeded")
		}
		if r.IsPending() {
			t.Error("should not be pending")
		}
	})

	t.Run("IsFinal", func(t *testing.T) {
		r := newTestRefund(t, now)
		if r.IsFinal() {
			t.Error("pending should not be final")
		}

		_ = r.MarkSucceeded("ref-1", now)
		if !r.IsFinal() {
			t.Error("succeeded should be final")
		}
	})
}

// ============================================================
// AGGREGATES
// ============================================================

func TestTotalRefunded(t *testing.T) {
	refunds := []*Refund{
		{ID: "r1", Amount: 1000_00, Status: PaymentStatusSucceeded},
		{ID: "r2", Amount: 2000_00, Status: PaymentStatusSucceeded},
		{ID: "r3", Amount: 3000_00, Status: PaymentStatusPending}, // ignored
		{ID: "r4", Amount: 4000_00, Status: PaymentStatusFailed},  // ignored
		nil, // ignored
	}

	got := TotalRefunded(refunds)
	want := int64(3000_00)
	if got != want {
		t.Errorf("TotalRefunded: got %d, want %d", got, want)
	}
}

func TestRefundableRemaining(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	newSucceeded := func(t *testing.T) *Payment {
		t.Helper()
		p, _ := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000_00, "KES", "idem-1",
			30*time.Minute, now)
		_ = p.MarkSucceeded("ref", now)
		return p
	}

	t.Run("no prior refunds", func(t *testing.T) {
		p := newSucceeded(t)
		if got := RefundableRemaining(p, nil); got != 10000_00 {
			t.Errorf("got %d, want %d", got, int64(10000_00))
		}
	})

	t.Run("with succeeded refunds", func(t *testing.T) {
		p := newSucceeded(t)
		prior := []*Refund{
			{Amount: 4000_00, Status: PaymentStatusSucceeded},
		}
		if got := RefundableRemaining(p, prior); got != 6000_00 {
			t.Errorf("got %d, want %d", got, int64(6000_00))
		}
	})

	t.Run("with pending refunds", func(t *testing.T) {
		p := newSucceeded(t)
		prior := []*Refund{
			{Amount: 4000_00, Status: PaymentStatusPending},
		}
		if got := RefundableRemaining(p, prior); got != 6000_00 {
			t.Errorf("got %d, want %d", got, int64(6000_00))
		}
	})

	t.Run("with failed refunds", func(t *testing.T) {
		p := newSucceeded(t)
		prior := []*Refund{
			{Amount: 4000_00, Status: PaymentStatusFailed},
		}
		if got := RefundableRemaining(p, prior); got != 10000_00 {
			t.Errorf("got %d, want %d", got, int64(10000_00))
		}
	})

	t.Run("fully refunded", func(t *testing.T) {
		p := newSucceeded(t)
		prior := []*Refund{
			{Amount: 10000_00, Status: PaymentStatusSucceeded},
		}
		if got := RefundableRemaining(p, prior); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("over-refunded clamps to zero", func(t *testing.T) {
		p := newSucceeded(t)
		// Imagine a bug allowed this to be persisted.
		prior := []*Refund{
			{Amount: 15000_00, Status: PaymentStatusSucceeded},
		}
		if got := RefundableRemaining(p, prior); got != 0 {
			t.Errorf("got %d, want 0 (clamped)", got)
		}
	})

	t.Run("nil payment", func(t *testing.T) {
		if got := RefundableRemaining(nil, nil); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("non-succeeded payment", func(t *testing.T) {
		pending, _ := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000_00, "KES", "idem-1",
			30*time.Minute, now)

		if got := RefundableRemaining(pending, nil); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
}

// ============================================================
// HELPERS
// ============================================================

func newTestRefund(t *testing.T, now time.Time) *Refund {
	t.Helper()
	r, err := NewRefund(
		"refund-1", "pay-1",
		5000_00, "KES",
		"customer request", "actor-1",
		now,
	)
	if err != nil {
		t.Fatalf("newTestRefund: %v", err)
	}
	return r
}