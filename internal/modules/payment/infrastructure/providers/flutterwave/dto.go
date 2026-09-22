package flutterwave

import "encoding/json"

// ============================================================
// ENVELOPE
// ============================================================

// apiEnvelope is the standard Flutterwave v3 response envelope.
type apiEnvelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// ============================================================
// CHARGE REQUEST — M-PESA
// ============================================================

type mpesaChargeRequest struct {
	Amount            int64  `json:"amount"`
	Currency          string `json:"currency"`
	PhoneNumber       string `json:"phone_number"`
	Email             string `json:"email"`
	TxRef             string `json:"tx_ref"`
	FullName          string `json:"fullname"`
	ClientIP          string `json:"client_ip,omitempty"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
}

// ============================================================
// CHARGE REQUEST — CARD
// ============================================================

type cardChargeRequest struct {
	CardNumber        string                 `json:"card_number"`
	CVV               string                 `json:"cvv"`
	Expiry            string                 `json:"expiry"`
	Currency          string                 `json:"currency"`
	Amount            int64                  `json:"amount"`
	FullName          string                 `json:"fullname"`
	Email             string                 `json:"email"`
	TxRef             string                 `json:"tx_ref"`
	PhoneNumber       string                 `json:"phone_number,omitempty"`
	RedirectURL       string                 `json:"redirect_url,omitempty"`
	Authorization     cardAuthorization      `json:"authorization"`
	ClientIP          string                 `json:"client_ip,omitempty"`
	DeviceFingerprint string                 `json:"device_fingerprint,omitempty"`
	Meta              map[string]interface{} `json:"meta,omitempty"`
}

type cardAuthorization struct {
	Mode     string `json:"mode"`               // "pin", "redirect", "avs_noauth"
	Pin      string `json:"pin,omitempty"`      // for "pin" mode
	Address  string `json:"address,omitempty"`  // for "avs_noauth"
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	ZipCode  string `json:"zipcode,omitempty"`
	Country  string `json:"country,omitempty"`
}

// ============================================================
// CHARGE RESPONSE
// ============================================================

// chargeResponse is what Flutterwave returns from /v3/charges.
type chargeResponse struct {
	ID        int64  `json:"id"`
	TxRef     string `json:"tx_ref"`
	FlwRef    string `json:"flw_ref"`
	Status    string `json:"status"`     // "pending" | "successful" | "failed"
	AuthModel string `json:"auth_model"` // "LIPA_MPESA", "PIN", "VBVSECURECODE", etc.
	Amount    float64 `json:"amount"`
	Currency  string `json:"currency"`
	RedirectURL string `json:"redirect_url,omitempty"`
	ProcessorResponse string `json:"processor_response,omitempty"`
}

// ============================================================
// VERIFY RESPONSE
// ============================================================

type verifyResponse struct {
	ID                int64   `json:"id"`
	TxRef             string  `json:"tx_ref"`
	FlwRef            string  `json:"flw_ref"`
	Status            string  `json:"status"`     // "successful" | "failed" | "pending"
	Amount            float64 `json:"amount"`
	ChargedAmount     float64 `json:"charged_amount"`
	Currency          string  `json:"currency"`
	PaymentType       string  `json:"payment_type"` // "card" | "mpesa" | ...
	ProcessorResponse string  `json:"processor_response"`
}

// ============================================================
// REFUND REQUEST / RESPONSE
// ============================================================

type refundRequest struct {
	Amount int64 `json:"amount,omitempty"` // omit for full refund
}

type refundResponse struct {
	ID     int64   `json:"id"`
	TxRef  string  `json:"tx_ref"`
	FlwRef string  `json:"flw_ref"`
	Status string  `json:"status"` // "completed" | "pending" | "failed"
	Amount float64 `json:"amount"`
}

// ============================================================
// WEBHOOK PAYLOAD
// ============================================================

// webhookPayload is the body Flutterwave POSTs to your webhook endpoint.
type webhookPayload struct {
	Event string          `json:"event"` // "charge.completed", "transfer.completed", etc.
	Data  webhookData     `json:"data"`
}

type webhookData struct {
	ID                int64   `json:"id"`
	TxRef             string  `json:"tx_ref"`
	FlwRef            string  `json:"flw_ref"`
	Status            string  `json:"status"` // "successful" | "failed" | "pending"
	Amount            float64 `json:"amount"`
	ChargedAmount     float64 `json:"charged_amount"`
	Currency          string  `json:"currency"`
	PaymentType       string  `json:"payment_type"`
	ProcessorResponse string  `json:"processor_response"`
	Customer          struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		PhoneNumber string `json:"phone_number"`
	} `json:"customer"`
}