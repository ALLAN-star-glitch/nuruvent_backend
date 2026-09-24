// internal/modules/payment/paymentdomain/order_test.go

package paymentdomain

import (
	"errors"
	"testing"
	"time"
)

// ============================================================
// CONSTRUCTION
// ============================================================

func TestNewOrder(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)
	ttl := 30 * time.Minute

	items := []OrderItem{
		{TicketTypeID: "tkt-general", Quantity: 2, UnitPrice: 1500_00, Discount: 0, LineTotal: 3000_00},
		{TicketTypeID: "tkt-vip", Quantity: 1, UnitPrice: 5000_00, Discount: 500_00, LineTotal: 4500_00},
	}

	t.Run("valid authenticated", func(t *testing.T) {
		o, err := NewOrder(
			"order-1", "reg-1",
			"user-1", "", "KES",
			items, ttl, now,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Status != OrderStatusPending {
			t.Errorf("status: got %s, want %s", o.Status, OrderStatusPending)
		}
		if o.Subtotal != 8000_00 {
			t.Errorf("subtotal: got %d, want %d", o.Subtotal, 8000_00)
		}
		if o.DiscountTotal != 500_00 {
			t.Errorf("discount_total: got %d, want %d", o.DiscountTotal, 500_00)
		}
		if o.TotalAmount != 7500_00 {
			t.Errorf("total_amount: got %d, want %d", o.TotalAmount, 7500_00)
		}
		if !o.ExpiresAt.Equal(now.Add(ttl)) {
			t.Errorf("expires_at: got %v, want %v", o.ExpiresAt, now.Add(ttl))
		}
	})

	t.Run("valid guest", func(t *testing.T) {
		o, err := NewOrder(
			"order-1", "reg-1",
			"", "guest@example.com", "KES",
			items, ttl, now,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !o.IsGuest() {
			t.Error("expected IsGuest true")
		}
	})

	t.Run("rejects empty id", func(t *testing.T) {
		_, err := NewOrder("", "reg-1", "user-1", "", "KES", items, ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects empty registration id", func(t *testing.T) {
		_, err := NewOrder("order-1", "", "user-1", "", "KES", items, ttl, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects no identity", func(t *testing.T) {
		_, err := NewOrder("order-1", "reg-1", "", "", "KES", items, ttl, now)
		if !errors.Is(err, ErrIdentityRequired) {
			t.Fatalf("expected ErrIdentityRequired, got %v", err)
		}
	})

	t.Run("rejects both identities", func(t *testing.T) {
		_, err := NewOrder("order-1", "reg-1", "user-1", "guest@example.com", "KES", items, ttl, now)
		if !errors.Is(err, ErrIdentityRequired) {
			t.Fatalf("expected ErrIdentityRequired, got %v", err)
		}
	})

	t.Run("rejects empty currency", func(t *testing.T) {
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "", items, ttl, now)
		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("expected ErrInvalidCurrency, got %v", err)
		}
	})

	t.Run("rejects empty items", func(t *testing.T) {
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "KES", nil, ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects non-positive ttl", func(t *testing.T) {
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "KES", items, 0, now)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects item with zero quantity", func(t *testing.T) {
		bad := []OrderItem{
			{TicketTypeID: "tkt-1", Quantity: 0, UnitPrice: 100, LineTotal: 0},
		}
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "KES", bad, ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects item with empty ticket type id", func(t *testing.T) {
		bad := []OrderItem{
			{TicketTypeID: "", Quantity: 1, UnitPrice: 100, LineTotal: 100},
		}
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "KES", bad, ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects item with discount exceeding subtotal", func(t *testing.T) {
		bad := []OrderItem{
			{TicketTypeID: "tkt-1", Quantity: 1, UnitPrice: 100, Discount: 200, LineTotal: 0},
		}
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "KES", bad, ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("rejects item with inconsistent line total", func(t *testing.T) {
		// unit_price * quantity - discount != line_total
		bad := []OrderItem{
			{TicketTypeID: "tkt-1", Quantity: 2, UnitPrice: 100, Discount: 0, LineTotal: 999},
		}
		_, err := NewOrder("order-1", "reg-1", "user-1", "", "KES", bad, ttl, now)
		if !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("expected ErrInvalidAmount, got %v", err)
		}
	})
}

// ============================================================
// TRANSITIONS: MarkPaid
// ============================================================

func TestOrder_MarkPaid(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		o := newTestOrder(t, now)
		if err := o.MarkPaid(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Status != OrderStatusPaid {
			t.Errorf("status: got %s, want %s", o.Status, OrderStatusPaid)
		}
		if o.PaidAt == nil {
			t.Error("paid_at should be set")
		}
	})

	t.Run("idempotent when already paid", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.MarkPaid(now)
		firstPaid := *o.PaidAt

		later := now.Add(time.Hour)
		if err := o.MarkPaid(later); err != nil {
			t.Fatalf("second call should be no-op, got: %v", err)
		}
		if !o.PaidAt.Equal(firstPaid) {
			t.Errorf("paid_at changed: got %v, want %v", *o.PaidAt, firstPaid)
		}
	})

	t.Run("rejects when already expired", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.Expire(now)

		err := o.MarkPaid(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})

	t.Run("rejects when already cancelled", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.Cancel(now)

		err := o.MarkPaid(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

// ============================================================
// TRANSITIONS: Expire, Cancel
// ============================================================

func TestOrder_Expire(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		o := newTestOrder(t, now)
		if err := o.Expire(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Status != OrderStatusExpired {
			t.Errorf("status: got %s, want %s", o.Status, OrderStatusExpired)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.Expire(now)
		if err := o.Expire(now); err != nil {
			t.Fatalf("second call should be no-op: %v", err)
		}
	})

	t.Run("rejects when already paid", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.MarkPaid(now)

		err := o.Expire(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

func TestOrder_Cancel(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("transitions from pending", func(t *testing.T) {
		o := newTestOrder(t, now)
		if err := o.Cancel(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Status != OrderStatusCancelled {
			t.Errorf("status: got %s, want %s", o.Status, OrderStatusCancelled)
		}
		if o.CancelledAt == nil {
			t.Error("cancelled_at should be set")
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.Cancel(now)
		if err := o.Cancel(now); err != nil {
			t.Fatalf("second call should be no-op: %v", err)
		}
	})

	t.Run("rejects when already paid", func(t *testing.T) {
		o := newTestOrder(t, now)
		_ = o.MarkPaid(now)

		err := o.Cancel(now)
		if !errors.Is(err, ErrInvalidStatusTransition) {
			t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
		}
	})
}

// ============================================================
// QUERIES
// ============================================================

func TestOrder_Queries(t *testing.T) {
	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)

	t.Run("IsPending, IsPaid", func(t *testing.T) {
		o := newTestOrder(t, now)
		if !o.IsPending() {
			t.Error("expected pending")
		}
		if o.IsPaid() {
			t.Error("should not be paid")
		}

		_ = o.MarkPaid(now)
		if o.IsPending() {
			t.Error("should not be pending")
		}
		if !o.IsPaid() {
			t.Error("expected paid")
		}
	})

	t.Run("IsExpired uses wall clock", func(t *testing.T) {
		o := newTestOrder(t, now)

		before := now.Add(10 * time.Minute)
		if o.IsExpired(before) {
			t.Error("should not be expired before deadline")
		}

		after := now.Add(31 * time.Minute)
		if !o.IsExpired(after) {
			t.Error("should be expired after deadline")
		}
	})

	t.Run("IsFinal", func(t *testing.T) {
		o := newTestOrder(t, now)
		if o.IsFinal() {
			t.Error("pending should not be final")
		}
		_ = o.MarkPaid(now)
		if !o.IsFinal() {
			t.Error("paid should be final")
		}
	})

	t.Run("IsGuest", func(t *testing.T) {
		o, _ := NewOrder("order-1", "reg-1", "", "g@example.com", "KES",
			[]OrderItem{{TicketTypeID: "tkt-1", Quantity: 1, UnitPrice: 100, LineTotal: 100}},
			30*time.Minute, now)
		if !o.IsGuest() {
			t.Error("expected IsGuest true")
		}
	})

	t.Run("ItemCount and TotalQuantity", func(t *testing.T) {
		o := newTestOrder(t, now)
		// newTestOrder has 2 items, quantities 2 and 1
		if o.ItemCount() != 2 {
			t.Errorf("ItemCount: got %d, want 2", o.ItemCount())
		}
		if o.TotalQuantity() != 3 {
			t.Errorf("TotalQuantity: got %d, want 3", o.TotalQuantity())
		}
	})
}

// ============================================================
// HELPERS
// ============================================================

func newTestOrder(t *testing.T, now time.Time) *Order {
	t.Helper()
	items := []OrderItem{
		{TicketTypeID: "tkt-general", Quantity: 2, UnitPrice: 1500_00, LineTotal: 3000_00},
		{TicketTypeID: "tkt-vip", Quantity: 1, UnitPrice: 5000_00, Discount: 500_00, LineTotal: 4500_00},
	}
	o, err := NewOrder(
		"order-1", "reg-1",
		"user-1", "", "KES",
		items, 30*time.Minute, now,
	)
	if err != nil {
		t.Fatalf("newTestOrder: %v", err)
	}
	return o
}