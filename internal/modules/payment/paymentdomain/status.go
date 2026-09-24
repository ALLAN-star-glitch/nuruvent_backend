// internal/modules/payment/paymentdomain/status.go

package paymentdomain

import "fmt"

// ============================================================
// ORDER STATUS
// ============================================================

// OrderStatus is the lifecycle state of an Order.
//
//	         ┌─────────┐
//	         │ pending │
//	         └────┬────┘
//	              │
//	   ┌──────────┼──────────┐
//	   ▼          ▼          ▼
//	┌──────┐  ┌─────────┐  ┌───────────┐
//	│ paid │  │ expired │  │ cancelled │
//	└──────┘  └─────────┘  └───────────┘
//	 (final)   (final)      (final)
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusExpired   OrderStatus = "expired"
	OrderStatusCancelled OrderStatus = "cancelled"
)

var AllOrderStatuses = []OrderStatus{
	OrderStatusPending,
	OrderStatusPaid,
	OrderStatusExpired,
	OrderStatusCancelled,
}

func (s OrderStatus) String() string { return string(s) }

func (s OrderStatus) IsValid() bool {
	for _, v := range AllOrderStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// IsFinal reports whether the status has no outgoing transitions.
func (s OrderStatus) IsFinal() bool {
	switch s {
	case OrderStatusPaid, OrderStatusExpired, OrderStatusCancelled:
		return true
	}
	return false
}

// CanTransitionTo reports whether a status change is legal.
//
// Legal transitions:
//
//	pending   → paid | expired | cancelled
//	paid      → (final)
//	expired   → (final)
//	cancelled → (final)
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	switch s {
	case OrderStatusPending:
		return next == OrderStatusPaid ||
			next == OrderStatusExpired ||
			next == OrderStatusCancelled
	case OrderStatusPaid, OrderStatusExpired, OrderStatusCancelled:
		return false
	}
	return false
}

// ============================================================
// PAYMENT STATUS
// ============================================================

// PaymentStatus is the lifecycle state of a Payment attempt.
//
//	   ┌─────────┐
//	   │ pending │
//	   └────┬────┘
//	        │
//	   ┌────┼──────────┐
//	   ▼    ▼          ▼
//	┌───────────┐  ┌────────┐  ┌─────────┐
//	│ succeeded │  │ failed │  │ expired │
//	└─────┬─────┘  └────────┘  └─────────┘
//	      │        (final)     (final)
//	      ▼
//	┌──────────┐
//	│ refunded │
//	└──────────┘
//	 (final)
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

var AllPaymentStatuses = []PaymentStatus{
	PaymentStatusPending,
	PaymentStatusSucceeded,
	PaymentStatusFailed,
	PaymentStatusExpired,
	PaymentStatusRefunded,
}

func (s PaymentStatus) String() string { return string(s) }

func (s PaymentStatus) IsValid() bool {
	for _, v := range AllPaymentStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// IsFinal reports whether the status has no outgoing transitions.
func (s PaymentStatus) IsFinal() bool {
	switch s {
	case PaymentStatusSucceeded, PaymentStatusFailed, PaymentStatusExpired, PaymentStatusRefunded:
		return true
	}
	return false
}

// CanTransitionTo reports whether a status change is legal.
//
// Legal transitions:
//
//	pending   → succeeded | failed | expired
//	succeeded → refunded
//	failed    → (final)
//	expired   → (final)
//	refunded  → (final)
func (s PaymentStatus) CanTransitionTo(next PaymentStatus) bool {
	switch s {
	case PaymentStatusPending:
		return next == PaymentStatusSucceeded ||
			next == PaymentStatusFailed ||
			next == PaymentStatusExpired
	case PaymentStatusSucceeded:
		return next == PaymentStatusRefunded
	case PaymentStatusFailed, PaymentStatusExpired, PaymentStatusRefunded:
		return false
	}
	return false
}

// ============================================================
// PAYMENT METHOD
// ============================================================

// PaymentMethod is how the user pays.
type PaymentMethod string

const (
	PaymentMethodMpesa PaymentMethod = "mpesa"
	PaymentMethodCard  PaymentMethod = "card"
)

var AllPaymentMethods = []PaymentMethod{
	PaymentMethodMpesa,
	PaymentMethodCard,
}

func (m PaymentMethod) String() string { return string(m) }

func (m PaymentMethod) IsValid() bool {
	for _, v := range AllPaymentMethods {
		if v == m {
			return true
		}
	}
	return false
}

// ============================================================
// PARSE HELPERS
// ============================================================

// ParseOrderStatus converts a string (e.g. from a DB column) into an
// OrderStatus. Returns an error if the value is not a known status.
func ParseOrderStatus(v string) (OrderStatus, error) {
	s := OrderStatus(v)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid order status: %q", v)
	}
	return s, nil
}

// ParsePaymentStatus converts a string into a PaymentStatus.
func ParsePaymentStatus(v string) (PaymentStatus, error) {
	s := PaymentStatus(v)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid payment status: %q", v)
	}
	return s, nil
}

// ParsePaymentMethod converts a string into a PaymentMethod.
func ParsePaymentMethod(v string) (PaymentMethod, error) {
	m := PaymentMethod(v)
	if !m.IsValid() {
		return "", fmt.Errorf("invalid payment method: %q", v)
	}
	return m, nil
}