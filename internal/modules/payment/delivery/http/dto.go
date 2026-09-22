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
	Amount            int64      `json:"amount"`   // minor units
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	ProviderReference string     `json:"provider_reference,omitempty"`
	FailureReason     string     `json:"failure_reason,omitempty"`
	InitiatedAt       time.Time  `json:"initiated_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	FailedAt          *time.Time `json:"failed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}