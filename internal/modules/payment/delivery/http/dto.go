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



// ============================================================
// LIST PAYMENTS
// ============================================================

// ListPaymentsQuery captures the query parameters accepted by
// GET /payments. All fields are optional except the implicit account
// scope, which comes from the authenticated user's JWT.
type ListPaymentsQuery struct {
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Search   string `query:"search"`
	Status   string `query:"status"`
	Method   string `query:"method"`
	EventID  string `query:"event_id"`
	SortBy   string `query:"sort_by"`
	SortOrder string `query:"sort_order"`

	// Date range — ISO 8601 date or RFC3339 timestamp, parsed in the
	// handler. Kept as strings here so binding is trivial and
	// validation errors are explicit.
	DateFrom string `query:"date_from"`
	DateTo   string `query:"date_to"`
}

// PaymentListItemResponse is one row in the payments ledger.
//
// Amounts are in minor units. The frontend divides by 100 for
// display. Currency is included so the frontend can format with the
// right symbol and locale.
type PaymentListItemResponse struct {
	ID                string `json:"id"`
	OrderID           string `json:"order_id"`
	RegistrationID    string `json:"registration_id"`
	RegistrationNumber string `json:"registration_number,omitempty"`

	// Attendee
	AttendeeName  string `json:"attendee_name"`
	AttendeeEmail string `json:"attendee_email"`
	AttendeePhone string `json:"attendee_phone,omitempty"`

	// Event
	EventID        string `json:"event_id"`
	EventTitle     string `json:"event_title"`
	EventStartDate string `json:"event_start_date,omitempty"`
	EventImageURL  string `json:"event_image_url,omitempty"`

	// Payment
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
	PlatformFee       int64  `json:"platform_fee"`
	ProcessingFee     int64  `json:"processing_fee"`
	NetToOrganizer    int64  `json:"net_to_organizer"`
	PlatformFeeRate   float64 `json:"platform_fee_rate"`
	ProcessingFeeRate float64 `json:"processing_fee_rate"`

	Status          string `json:"status"`
	StatusLabel     string `json:"status_label"`
	Provider        string `json:"provider"`
	Method          string `json:"method"`
	MethodLabel     string `json:"method_label"`
	TransactionID   string `json:"transaction_id,omitempty"`

	// Timing
	InitiatedAt string `json:"initiated_at"`
	CompletedAt string `json:"completed_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// ListPaymentsResponse is the envelope returned by GET /payments.
type ListPaymentsResponse struct {
	Payments []PaymentListItemResponse `json:"payments"`
	Total    int                       `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// ============================================================
// PAYMENT STATS
// ============================================================

// PaymentStatsQuery captures the query parameters accepted by
// GET /payments/stats. Same filters as the list endpoint minus
// pagination and search.
type PaymentStatsQuery struct {
	EventID  string `query:"event_id"`
	DateFrom string `query:"date_from"`
	DateTo   string `query:"date_to"`
}

// PaymentStatsResponse is the aggregate summary.
type PaymentStatsResponse struct {
	TotalRevenue        int64            `json:"total_revenue"`
	TotalPlatformFees   int64            `json:"total_platform_fees"`
	TotalProcessingFees int64            `json:"total_processing_fees"`
	TotalNet            int64            `json:"total_net"`
	TransactionCount    int64            `json:"transaction_count"`
	Currency            string           `json:"currency"`
	ByStatus            map[string]int64 `json:"by_status"`
}
