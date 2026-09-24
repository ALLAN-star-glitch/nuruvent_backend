// internal/modules/payment/infrastructure/providers/paystack/dto.go

package paystack

// ============================================================
// CHARGE — MOBILE MONEY (M-PESA)
// ============================================================

// chargeRequest is the body for POST /charge.
//
// Used to trigger a direct STK push for mobile money. The user stays
// on your site; Paystack sends the PIN prompt to their phone.
//
// Unlike /transaction/initialize, this endpoint accepts a specific
// mobile money provider — so we can force M-Pesa without offering
// Airtel.
type chargeRequest struct {
	Email       string              `json:"email"`
	Amount      string              `json:"amount"`     // minor units, string
	Currency    string              `json:"currency"`
	Reference   string              `json:"reference"`  // our payment ID
	MobileMoney mobileMoneyDetails  `json:"mobile_money"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type mobileMoneyDetails struct {
	Phone    string `json:"phone"`    // 254XXXXXXXXX
	Provider string `json:"provider"` // "mpesa"
}

// chargeResponse is what Paystack returns from /charge.
//
// Status values:
//   - "pay_offline"     → STK push sent, waiting for customer PIN
//   - "send_otp"        → OTP required (rare for M-Pesa)
//   - "success"         → completed immediately (rare)
//   - "failed"          → charge rejected
//
// For M-Pesa, we expect "pay_offline". The final result arrives via
// webhook (charge.success or charge.failed).
type chargeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Reference       string `json:"reference"`
		Status          string `json:"status"`      // "pay_offline", "send_otp", etc.
		DisplayText     string `json:"display_text"` // user-facing instruction
		GatewayResponse string `json:"gateway_response,omitempty"`
	} `json:"data"`
}

// ============================================================
// INITIALIZE TRANSACTION — CARD (HOSTED CHECKOUT)
// ============================================================

// initializeRequest is the body for POST /transaction/initialize.
//
// Used for card payments. Returns a hosted checkout URL.
type initializeRequest struct {
	Email       string                 `json:"email"`
	Amount      string                 `json:"amount"`
	Currency    string                 `json:"currency,omitempty"`
	Reference   string                 `json:"reference,omitempty"`
	CallbackURL string                 `json:"callback_url,omitempty"`
	Channels    []string               `json:"channels,omitempty"` // ["card"]
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type initializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// ============================================================
// VERIFY TRANSACTION — RESPONSE
// ============================================================

type verifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int64  `json:"id"`
		Domain          string `json:"domain"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		CreatedAt       string `json:"created_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		Fees            int64  `json:"fees"`
		Customer        struct {
			ID        int64  `json:"id"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Phone     string `json:"phone"`
		} `json:"customer"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Channel           string `json:"channel"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			CountryCode       string `json:"country_code"`
		} `json:"authorization"`
	} `json:"data"`
}

// ============================================================
// REFUND — REQUEST / RESPONSE
// ============================================================

type refundRequest struct {
	Transaction string `json:"transaction"`
	Amount      string `json:"amount,omitempty"`
	Currency    string `json:"currency,omitempty"`
}

type refundResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID        int64  `json:"id"`
		Status    string `json:"status"`
		Amount    int64  `json:"amount"`
		Currency  string `json:"currency"`
		Reference string `json:"reference"`
	} `json:"data"`
}

// ============================================================
// WEBHOOK PAYLOAD
// ============================================================

type webhookPayload struct {
	Event string      `json:"event"`
	Data  webhookData `json:"data"`
}

type webhookData struct {
	ID              int64  `json:"id"`
	Reference       string `json:"reference"`
	Status          string `json:"status"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	GatewayResponse string `json:"gateway_response"`
	PaidAt          string `json:"paid_at"`
	Channel         string `json:"channel"`
	Customer        struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	} `json:"customer"`
}