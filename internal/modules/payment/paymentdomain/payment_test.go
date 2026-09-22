// internal/modules/payment/paymentdomain/payment_test.go

package paymentdomain

import (
	"errors"
	"testing"
	"time"
)

// ============================================================
// CONSTRUCTION
// ============================================================

func TestNewPayment(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)
	ttl := 30 * time.Minute

	t.Run("valid", func(t *testing.T) {
		p, err := NewPayment(
			"pay-1", "order-1", "mpesa",
			PaymentMethodMpesa,
			10000, "KES", "idem-1",
			ttl, now,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Status != PaymentStatusPending {
			t.Errorf("status: got %s, want %s", p.Status, PaymentStatusPending)
		}
		if !p.ExpiresAt.Equal(now.Add(ttl)) {
			t.Errorf("expires_at: got %v, want %v", p.ExpiresAt, now.Add(ttl))
		}
		if p.InitiatedAt != now {
			t.Errorf("initiated_at: got %v, want %v", p.InitiatedAt, now)
		}
		if p.CompletedAt != nil {
			t.Errorf("completed_at: got %v, want nil", p.CompletedAt)
		}
	})

	t.Run("rejects empty id", func(t *testing.T) {
		_, err := NewPayment("", "order-1", "mpesa",
			PaymentMethodMpesa, 10000, "KES", "idem-1", ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects empty order id", func(t *testing.T) {
		_, err := NewPayment("pay-1", "", "mpesa",
			PaymentMethodMpesa, 10000, "KES", "idem-1", ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects empty provider", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "",
			PaymentMethodMpesa, 10000, "KES", "idem-1", ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects invalid method", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethod("crypto"), 10000, "KES", "idem-1", ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects zero amount", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 0, "KES", "idem-1", ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects negative amount", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, -100, "KES", "idem-1", ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects empty currency", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000, "", "idem-1", ttl, now)
		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("expected ErrInvalidCurrency, got %v", err)
		}
	})

	t.Run("rejects empty idempotency key", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000, "KES", "", ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects non-positive ttl", func(t *testing.T) {
		_, err := NewPayment("pay-1", "order-1", "mpesa",
			PaymentMethodMpesa, 10000, "KES", "idem-1", 0, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ============================================================
// TRANSITIONS: MarkSucceeded
// ============================================================

func TestPayment_MarkSucceeded(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		p := newTestPayment(t, now)
		if err := p.MarkSucceeded("ref-1", now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Status != PaymentStatusSucceeded {
			t.Errorf("status: got %s, want %s", p.Status, PaymentStatusSucceeded)
		}
		if p.ProviderReference != "ref-1" {
			t.Errorf("provider_reference: got %q, want %q", p.ProviderReference, "ref-1")
		}
		if p.CompletedAt == nil {
			t.Error("completed_at should be set")
		}
	})

	t.Run("idempotent when already succeeded", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkSucceeded("ref-1", now)
		firstCompleted := *p.CompletedAt

		later := now.Add(time.Hour)
		if err := p.MarkSucceeded("ref-1", later); err != nil {
			t.Fatalf("second call should be no-op, got: %v", err)
		}
		if !p.CompletedAt.Equal(firstCompleted) {
			t.Errorf("completed_at changed: got %v, want %v", *p.CompletedAt, firstCompleted)
		}
	})

	t.Run("rejects when already failed", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkFailed("insufficient funds", now)

		err := p.MarkSucceeded("ref-1", now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})

	t.Run("rejects when already expired", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkExpired(now)

		err := p.MarkSucceeded("ref-1", now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})

	t.Run("requires non-empty provider reference", func(t *testing.T) {
		p := newTestPayment(t, now)
		err := p.MarkSucceeded("", now)
		if err == nil {
			t.Fatal("expected error for empty reference")
		}
		if p.Status != PaymentStatusPending {
			t.Errorf("status must not change on failure: got %s", p.Status)
		}
	})
}

// ============================================================
// TRANSITIONS: MarkFailed
// ============================================================

func TestPayment_MarkFailed(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		p := newTestPayment(t, now)
		if err := p.MarkFailed("insufficient funds", now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Status != PaymentStatusFailed {
			t.Errorf("status: got %s, want %s", p.Status, PaymentStatusFailed)
		}
		if p.FailureReason != "insufficient funds" {
			t.Errorf("failure_reason: got %q", p.FailureReason)
		}
		if p.FailedAt == nil {
			t.Error("failed_at should be set")
		}
	})

	t.Run("idempotent when already failed", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkFailed("first reason", now)
		if err := p.MarkFailed("second reason", now); err != nil {
			t.Fatalf("second call should be no-op, got: %v", err)
		}
		// Failure reason should not be overwritten once final.
		if p.FailureReason != "first reason" {
			t.Errorf("failure_reason changed: got %q", p.FailureReason)
		}
	})

	t.Run("rejects when already succeeded", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkSucceeded("ref-1", now)

		err := p.MarkFailed("late failure", now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

// ============================================================
// TRANSITIONS: MarkExpired
// ============================================================

func TestPayment_MarkExpired(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		p := newTestPayment(t, now)
		if err := p.MarkExpired(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Status != PaymentStatusExpired {
			t.Errorf("status: got %s, want %s", p.Status, PaymentStatusExpired)
		}
	})

	t.Run("idempotent when already expired", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkExpired(now)
		if err := p.MarkExpired(now); err != nil {
			t.Fatalf("second call should be no-op, got: %v", err)
		}
	})

	t.Run("rejects when already succeeded", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkSucceeded("ref-1", now)

		err := p.MarkExpired(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

// ============================================================
// TRANSITIONS: MarkRefunded
// ============================================================

func TestPayment_MarkRefunded(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from succeeded", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkSucceeded("ref-1", now)

		if err := p.MarkRefunded(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Status != PaymentStatusRefunded {
			t.Errorf("status: got %s, want %s", p.Status, PaymentStatusRefunded)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkSucceeded("ref-1", now)
		_ = p.MarkRefunded(now)

		if err := p.MarkRefunded(now); err != nil {
			t.Fatalf("second call should be no-op, got: %v", err)
		}
	})

	t.Run("rejects when pending", func(t *testing.T) {
		p := newTestPayment(t, now)

		err := p.MarkRefunded(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})

	t.Run("rejects when failed", func(t *testing.T) {
		p := newTestPayment(t, now)
		_ = p.MarkFailed("insufficient funds", now)

		err := p.MarkRefunded(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

// ============================================================
// QUERIES
// ============================================================

func TestPayment_Queries(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("IsPending", func(t *testing.T) {
		p := newTestPayment(t, now)
		if !p.IsPending() {
			t.Error("expected IsPending")
		}
		_ = p.MarkSucceeded("ref-1", now)
		if p.IsPending() {
			t.Error("should not be pending after MarkSucceeded")
		}
	})

	t.Run("IsSucceeded", func(t *testing.T) {
		p := newTestPayment(t, now)
		if p.IsSucceeded() {
			t.Error("should not be succeeded initially")
		}
		_ = p.MarkSucceeded("ref-1", now)
		if !p.IsSucceeded() {
			t.Error("expected IsSucceeded")
		}
	})

	t.Run("IsFinal", func(t *testing.T) {
		p := newTestPayment(t, now)
		if p.IsFinal() {
			t.Error("pending should not be final")
		}
		_ = p.MarkSucceeded("ref-1", now)
		if !p.IsFinal() {
			t.Error("succeeded should be final")
		}
	})

	t.Run("IsExpired takes wall clock", func(t *testing.T) {
		p := newTestPayment(t, now)

		// Before deadline
		before := now.Add(10 * time.Minute)
		if p.IsExpired(before) {
			t.Error("should not be expired before deadline")
		}

		// After deadline
		after := now.Add(31 * time.Minute)
		if !p.IsExpired(after) {
			t.Error("should be expired after deadline")
		}
	})

	t.Run("IsRetryable", func(t *testing.T) {
		cases := []struct {
			name   string
			setup  func(p *Payment)
			expect bool
		}{
			{"pending is not retryable", func(p *Payment) {}, false},
			{"failed is retryable", func(p *Payment) { _ = p.MarkFailed("x", now) }, true},
			{"expired is retryable", func(p *Payment) { _ = p.MarkExpired(now) }, true},
			{"succeeded is not retryable", func(p *Payment) { _ = p.MarkSucceeded("ref", now) }, false},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				p := newTestPayment(t, now)
				tc.setup(p)
				if got := p.IsRetryable(); got != tc.expect {
					t.Errorf("got %v, want %v", got, tc.expect)
				}
			})
		}
	})
}

// ============================================================
// HELPERS
// ============================================================

func newTestPayment(t *testing.T, now time.Time) *Payment {
	t.Helper()
	p, err := NewPayment(
		"pay-1", "order-1", "mpesa",
		PaymentMethodMpesa,
		10000, "KES", "idem-1",
		30*time.Minute, now,
	)
	if err != nil {
		t.Fatalf("newTestPayment: %v", err)
	}
	return p
}