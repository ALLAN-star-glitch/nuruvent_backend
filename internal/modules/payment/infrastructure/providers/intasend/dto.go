package intasend

// ============================================================
// CHECKOUT INITIATE — REQUEST
// ============================================================

// checkoutRequest is the body sent to POST /payment/collection/
//
// IntaSend wraps everything (M-Pesa + cards) behind one checkout
// endpoint. The response contains a `url` — a hosted checkout page
// the user is redirected to, where they choose their method and
// complete payment.
type checkoutRequest struct {
	PublicKey   string `json:"public_key,omitempty"`
	Currency    string `json:"currency"`              // "KES", "USD"
	Amount      string `json:"amount"`                // IntaSend expects a string
	Email       string `json:"email"`                 // required
	PhoneNumber string `json:"phone_number,omitempty"` // required for M-Pesa
	APIRef      string `json:"api_ref"`               // our internal payment ID
	RedirectURL string `json:"redirect_url"`          // where to send user after payment
	Comment     string `json:"comment,omitempty"`     // appears on the checkout page
	Method      string `json:"method,omitempty"`      // "M-PESA" or "CARD-PAYMENT"; omit to let user choose
	Host        string `json:"host,omitempty"`        // optional; required for some setups
}

// ============================================================
// CHECKOUT INITIATE — RESPONSE
// ============================================================

// checkoutResponse is what IntaSend returns from POST /payment/collection/.
//
// `URL` is the hosted checkout page. Redirect the user there.
// Store `ID` as the provider reference — you'll use it for status checks.
type checkoutResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`         // ← the hosted checkout link
	APIRef    string `json:"api_ref"`
	Host      string `json:"host,omitempty"`
	Amount    any    `json:"amount"`      // IntaSend returns this as string or number depending on endpoint
	Currency  string `json:"currency,omitempty"`
	State     string `json:"state,omitempty"` // "PENDING" | "PROCESSING" | "COMPLETE" | "FAILED"
	CreatedAt string `json:"created_at,omitempty"`
}

// ============================================================
// STATUS CHECK — RESPONSE
// ============================================================

// statusResponse is what IntaSend returns from GET /payment/status/?invoice_id=...
//
// Response shape (from IntaSend docs):
//
//	{
//	  "invoice": {
//	    "id": "XMSLWOS",
//	    "invoice_id": "XMSLWOS",
//	    "state": "PENDING",
//	    "provider": "M-PESA",
//	    "charges": "0.00",
//	    "net_amount": 10.36,
//	    "currency": "KES",
//	    "value": "10.36",
//	    "api_ref": "ISL_...",
//	    "failed_reason": null,
//	    "created_at": "...",
//	    "updated_at": "..."
//	  },
//	  "meta": {
//	    "id": "...",
//	    "customer": { ... }
//	  }
//	}
type statusResponse struct {
	Detail  string      `json:"detail"`
	Invoice invoiceData `json:"invoice"`
	Meta    struct {
		ID string `json:"id"`
	} `json:"meta"`
}

// invoiceData is the inner "invoice" object.
type invoiceData struct {
	ID           string `json:"id"`
	InvoiceID    string `json:"invoice_id"`
	State        string `json:"state"`
	Provider     string `json:"provider"`
	Charges      string `json:"charges"`
	NetAmount    any    `json:"net_amount"` // IntaSend returns number OR string
	Currency     string `json:"currency"`
	Value        string `json:"value"`      // string (see docs)
	APIRef       string `json:"api_ref"`
	FailedReason string `json:"failed_reason"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// ============================================================
// REFUND — REQUEST / RESPONSE
// ============================================================

// refundRequest is the body sent to POST /payment/refund/
type refundRequest struct {
	InvoiceID   string `json:"invoice_id,omitempty"` // preferred: the IntaSend invoice ID
	APIRef      string `json:"api_ref,omitempty"`    // fallback: our payment reference
	Amount      string `json:"amount,omitempty"`     // omit for full refund
	Reason      string `json:"reason,omitempty"`
}

// refundResponse is what IntaSend returns from POST /payment/refund/
type refundResponse struct {
	ID          string `json:"id"`
	InvoiceID   string `json:"invoice_id"`
	APIRef      string `json:"api_ref"`
	Amount      string `json:"amount"`
	State       string `json:"state"`     // "PENDING" | "COMPLETE" | "FAILED"
	Reason      string `json:"reason,omitempty"`
	CreatedAt   string `json:"created_at"`
}



// ============================================================
// WEBHOOK PAYLOAD
// ============================================================

// webhookPayload is the body IntaSend POSTs to your webhook endpoint.
//
// Every request includes a `challenge` field — compare it against your
// configured challenge string to verify the request is genuine.
type webhookPayload struct {
	Challenge  string          `json:"challenge"`
	InvoiceID  string          `json:"invoice_id"`
	State      string          `json:"state"`       // "PENDING" | "PROCESSING" | "COMPLETE" | "FAILED"
	Provider   string          `json:"provider"`    // "M-PESA" | "CARD-PAYMENT" | ...
	APIRef     string          `json:"api_ref"`
	Amount     string          `json:"amount"`
	Currency   string          `json:"currency"`
	NetAmount  string          `json:"net_amount,omitempty"`
	Value      string          `json:"value,omitempty"`
	Charges    string          `json:"charges,omitempty"`
	ProviderRef string         `json:"provider_ref,omitempty"`
	CreatedAt  string          `json:"created_at,omitempty"`
	UpdatedAt  string          `json:"updated_at,omitempty"`
	FailedReason string        `json:"failed_reason,omitempty"`
	Customer   webhookCustomer `json:"customer,omitempty"`
}

type webhookCustomer struct {
	CustomerID string `json:"customer_id,omitempty"`
	Phone      string `json:"phone_number,omitempty"`
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
}


// ============================================================
// CHECKOUT LINK — REQUEST
// ============================================================

// ============================================================
// CHECKOUT LINK (hosted card payment)
// ============================================================

// checkoutLinkRequest is the body for POST /checkout/.
//
// This is IntaSend's Checkout Link API — a public endpoint that
// accepts the publishable key in the body (no Authorization header).
// It returns a URL to a hosted payment page.
type checkoutLinkRequest struct {
	PublicKey   string `json:"public_key"`
	Currency    string `json:"currency"`
	Amount      string `json:"amount"`
	PhoneNumber string `json:"phone_number,omitempty"` 
	Email       string `json:"email,omitempty"`
	APIRef      string `json:"api_ref,omitempty"`
	Comment     string `json:"comment,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
	Method      string `json:"method,omitempty"`   // "CARD-PAYMENT", "M-PESA", or empty for all
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
}

// checkoutLinkResponse is what IntaSend returns from /checkout/.
type checkoutLinkResponse struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Signature   string   `json:"signature,omitempty"`
	APIRef      string   `json:"api_ref,omitempty"`
	Amount      float64  `json:"amount,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	Methods     []string `json:"methods,omitempty"`
	RedirectURL string   `json:"redirect_url,omitempty"`
	Paid        bool     `json:"paid,omitempty"`
}