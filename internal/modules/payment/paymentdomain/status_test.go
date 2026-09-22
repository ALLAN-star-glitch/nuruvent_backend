// internal/modules/payment/paymentdomain/status_test.go

package paymentdomain

import "testing"

// ============================================================
// ORDER STATUS
// ============================================================

func TestOrderStatus_AllTransitions(t *testing.T) {
	// The complete transition table. Every (from, to) pair not listed
	// here must be rejected.
	// map[OrderStatus]bool is the inner map
	allowed := map[OrderStatus]map[OrderStatus]bool{ // 2D Matrix in map form.. Outer key (from) is the starting status and the inner key (to) is the target status
		OrderStatusPending: {
			OrderStatusPaid:      true,
			OrderStatusExpired:   true,
			OrderStatusCancelled: true,
		},
		OrderStatusPaid:      {},
		OrderStatusExpired:   {},
		OrderStatusCancelled: {},
	}

	for _, from := range AllOrderStatuses {
		for _, to := range AllOrderStatuses {
			expected := allowed[from][to]
			got := from.CanTransitionTo(to)

			if got != expected {
				t.Errorf(
					"order transition %s → %s: got %v, want %v",
					from, to, got, expected,
				)
			}
		}
	}
}

func TestOrderStatus_IsFinal(t *testing.T) {
	cases := map[OrderStatus]bool{
		OrderStatusPending:   false,
		OrderStatusPaid:      true,
		OrderStatusExpired:   true,
		OrderStatusCancelled: true,
	}

	for status, want := range cases {
		if got := status.IsFinal(); got != want {
			t.Errorf("%s.IsFinal(): got %v, want %v", status, got, want)
		}
	}
}

func TestOrderStatus_IsValid(t *testing.T) {
	for _, s := range AllOrderStatuses {
		if !s.IsValid() {
			t.Errorf("%s should be valid", s)
		}
	}

	invalid := OrderStatus("unknown")
	if invalid.IsValid() {
		t.Errorf("%s should not be valid", invalid)
	}
}

func TestParseOrderStatus(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, s := range AllOrderStatuses {
			got, err := ParseOrderStatus(string(s))
			if err != nil {
				t.Errorf("ParseOrderStatus(%q): unexpected error: %v", s, err)
			}
			if got != s {
				t.Errorf("ParseOrderStatus(%q): got %q, want %q", s, got, s)
			}
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParseOrderStatus("bogus")
		if err == nil {
			t.Error("expected error for invalid status, got nil")
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := ParseOrderStatus("")
		if err == nil {
			t.Error("expected error for empty status, got nil")
		}
	})
}

// ============================================================
// PAYMENT STATUS
// ============================================================

func TestPaymentStatus_AllTransitions(t *testing.T) {
	allowed := map[PaymentStatus]map[PaymentStatus]bool{
		PaymentStatusPending: {
			PaymentStatusSucceeded: true,
			PaymentStatusFailed:    true,
			PaymentStatusExpired:   true,
		},
		PaymentStatusSucceeded: {
			PaymentStatusRefunded: true,
		},
		PaymentStatusFailed:   {},
		PaymentStatusExpired:  {},
		PaymentStatusRefunded: {},
	}

	for _, from := range AllPaymentStatuses {
		for _, to := range AllPaymentStatuses {
			expected := allowed[from][to]
			got := from.CanTransitionTo(to)

			if got != expected {
				t.Errorf(
					"payment transition %s → %s: got %v, want %v",
					from, to, got, expected,
				)
			}
		}
	}
}

func TestPaymentStatus_IsFinal(t *testing.T) {
	cases := map[PaymentStatus]bool{
		PaymentStatusPending:   false,
		PaymentStatusSucceeded: true,
		PaymentStatusFailed:    true,
		PaymentStatusExpired:   true,
		PaymentStatusRefunded:  true,
	}

	for status, want := range cases {
		if got := status.IsFinal(); got != want {
			t.Errorf("%s.IsFinal(): got %v, want %v", status, got, want)
		}
	}
}

func TestPaymentStatus_IsValid(t *testing.T) {
	for _, s := range AllPaymentStatuses {
		if !s.IsValid() {
			t.Errorf("%s should be valid", s)
		}
	}

	invalid := PaymentStatus("unknown")
	if invalid.IsValid() {
		t.Errorf("%s should not be valid", invalid)
	}
}

func TestParsePaymentStatus(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, s := range AllPaymentStatuses {
			got, err := ParsePaymentStatus(string(s))
			if err != nil {
				t.Errorf("ParsePaymentStatus(%q): unexpected error: %v", s, err)
			}
			if got != s {
				t.Errorf("ParsePaymentStatus(%q): got %q, want %q", s, got, s)
			}
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParsePaymentStatus("bogus")
		if err == nil {
			t.Error("expected error for invalid status, got nil")
		}
	})
}

// ============================================================
// PAYMENT METHOD
// ============================================================

func TestPaymentMethod_IsValid(t *testing.T) {
	for _, m := range AllPaymentMethods {
		if !m.IsValid() {
			t.Errorf("%s should be valid", m)
		}
	}

	invalid := PaymentMethod("crypto")
	if invalid.IsValid() {
		t.Errorf("%s should not be valid", invalid)
	}
}

func TestParsePaymentMethod(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, m := range AllPaymentMethods {
			got, err := ParsePaymentMethod(string(m))
			if err != nil {
				t.Errorf("ParsePaymentMethod(%q): unexpected error: %v", m, err)
			}
			if got != m {
				t.Errorf("ParsePaymentMethod(%q): got %q, want %q", m, got, m)
			}
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := ParsePaymentMethod("crypto")
		if err == nil {
			t.Error("expected error for invalid method, got nil")
		}
	})
}