// internal/modules/payment/delivery/http/dto.go

package http

import "time"

// ============================================================
// REQUESTS
// ============================================================

// InitiatePaymentRequest is the body for POST /payments/initiate.
type InitiatePaymentRequest struct {
	OrderID        string           `json:"order_id"`
	Method         string           `json:"method"` // "mpesa" | "card"
	PayerPhone     string           `json:"payer_phone,omitempty"`
	PayerEmail     string           `json:"payer_email,omitempty"`
	Card           *CardDetailsJSON `json:"card,omitempty"`
	IdempotencyKey string           `json:"idempotency_key"`
	ReturnURL      string           `json:"return_url,omitempty"`
	Description    string           `json:"description,omitempty"`
}

// CardDetailsJSON is the wire format for card details.
//
// SECURITY: these values are PCI-sensitive. The HTTP layer must
// never log request bodies containing them. They are forwarded to
// the service and immediately encrypted by the provider.
type CardDetailsJSON struct {
	Number string `json:"number"`
	CVV    string `json:"cvv"`
	Expiry string `json:"expiry"`
}

// RefundRequest is the body for POST /payments/:id/refund.
type RefundRequest struct {
	Amount         int64  `json:"amount"` // minor units
	Reason         string `json:"reason,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ============================================================
// RESPONSES
// ============================================================

// PaymentResponse is the canonical output for a payment.
type PaymentResponse struct {
	ID                string     `json:"id"`
	OrderID           string     `json:"order_id"`
	Provider          string     `json:"provider"`
	Method            string     `json:"method"`
	Amount            int64      `json:"amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	ProviderReference string     `json:"provider_reference,omitempty"`
	RedirectURL       string     `json:"redirect_url,omitempty"`  
	FailureReason     string     `json:"failure_reason,omitempty"`
	InitiatedAt       time.Time  `json:"initiated_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	AccessCode  string `json:"access_code,omitempty"`
	FailedAt          *time.Time `json:"failed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ============================================================
// ORDER REQUESTS
// ============================================================

// CreateOrderRequest is the body for POST /orders.
//
// Only `registration_id` is accepted. The payment module resolves the
// pricing, items, and totals from the registration — the client cannot
// influence the amount.
type CreateOrderRequest struct {
    RegistrationID string `json:"registration_id"`
    GuestEmail     string `json:"guest_email,omitempty"` // for guests
}

// ============================================================
// ORDER RESPONSES
// ============================================================

// OrderResponse is the canonical output for an order.
type OrderResponse struct {
	ID             string              `json:"id"`
	RegistrationID string              `json:"registration_id"`
	UserID         string              `json:"user_id,omitempty"`
	GuestEmail     string              `json:"guest_email,omitempty"`
	Currency       string              `json:"currency"`
	Subtotal       int64               `json:"subtotal"`       // minor units
	DiscountTotal  int64               `json:"discount_total"` // minor units
	TotalAmount    int64               `json:"total_amount"`   // minor units
	Status         string              `json:"status"`
	ExpiresAt      time.Time           `json:"expires_at"`
	PaidAt         *time.Time          `json:"paid_at,omitempty"`
	CancelledAt    *time.Time          `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Items          []OrderItemResponse `json:"items"`
}

// OrderItemResponse is one line of an order.
type OrderItemResponse struct {
	TicketTypeID string `json:"ticket_type_id"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"` // minor units
	Discount     int64  `json:"discount"`   // minor units
	LineTotal    int64  `json:"line_total"` // minor units
}