// internal/modules/payment/infrastructure/providers/paystack/provider.go

package paystack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

const ProviderName = "paystack"

// Provider implements paymentdomain.PaymentProvider for Paystack.
//
// Two flows, two endpoints:
//
//   - M-Pesa: POST /charge — direct STK push. The user stays on your
//     site. Only M-Pesa is offered (no Airtel, no Till).
//
//   - Card: POST /transaction/initialize — returns a hosted checkout
//     URL. Paystack collects the card on their PCI-compliant page.
type Provider struct {
	client *client
	cfg    config.PaystackConfig
}

// NewProvider constructs a Paystack provider.
func NewProvider(cfg config.PaystackConfig) *Provider {
	return &Provider{
		client: newClient(cfg),
		cfg:    cfg,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string { return ProviderName }

// Method reports the primary method this provider serves.
func (p *Provider) Method() paymentdomain.PaymentMethod {
	return paymentdomain.PaymentMethodMpesa
}

// Methods returns every method this provider serves.
func (p *Provider) Methods() []paymentdomain.PaymentMethod {
	return []paymentdomain.PaymentMethod{
		paymentdomain.PaymentMethodMpesa,
		paymentdomain.PaymentMethodCard,
	}
}

// ============================================================
// INITIATE
// ============================================================

// Initiate routes to the appropriate Paystack flow based on method.
//
//   - M-Pesa → /charge (direct STK push, no redirect)
//   - Card   → /transaction/initialize (hosted checkout)
func (p *Provider) Initiate(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
	if req.PayerEmail == "" {
		return nil, fmt.Errorf("%w: payer email is required",
			paymentdomain.ErrProviderRejected)
	}

	switch req.Method {
	case paymentdomain.PaymentMethodMpesa:
		return p.initiateMpesa(ctx, req)
	case paymentdomain.PaymentMethodCard:
		return p.initiateCard(ctx, req)
	default:
		return nil, fmt.Errorf("%w: unsupported payment method %q",
			paymentdomain.ErrProviderRejected, req.Method)
	}
}

// ============================================================
// INITIATE — M-PESA (STK push via /charge)
// ============================================================

// initiateMpesa triggers a direct M-Pesa STK push.
//
// The user stays on your site. Paystack sends the PIN prompt to their
// phone. No redirect URL is returned — the outcome arrives via webhook.
func (p *Provider) initiateMpesa(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
	if req.PayerPhone == "" {
		return nil, fmt.Errorf("%w: payer phone is required for M-Pesa",
			paymentdomain.ErrProviderRejected)
	}

	phone := normalizePhone(req.PayerPhone)
	log.Printf("DEBUG MPESA: raw=%q normalized=%q", req.PayerPhone, phone)
	if phone == "" {
		return nil, fmt.Errorf("%w: invalid phone number format",
			paymentdomain.ErrProviderRejected)
	}

	body := chargeRequest{
		Email:     req.PayerEmail,
		Amount:    strconv.FormatInt(req.Amount, 10),
		Currency:  req.Currency,
		Reference: req.PaymentID,
		MobileMoney: mobileMoneyDetails{
			Phone:    phone,
			Provider: "mpesa",
		},
		Metadata: map[string]interface{}{
			"payment_id": req.PaymentID,
			"order_id":   req.OrderID,
		},
	}

	raw, err := p.client.do(ctx, "POST", "/charge", body)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG MPESA raw response: %s", string(raw))

	var data chargeResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode charge response: %w", err)
	}

	// Paystack's /charge returns status=false for charges that
	// require further customer action (pay_offline, send_otp, etc).
	// This is not a rejection — the charge was accepted and is in
	// flight. The final result arrives via webhook.
	//
	// We only treat it as a rejection if the inner status indicates
	// an actual failure.
	innerStatus := strings.ToLower(data.Data.Status)

	switch innerStatus {
	case "pay_offline", "send_otp", "send_pin", "send_phone", "open_url", "pending", "success":
		// Valid in-flight status — return pending.
	default:
		msg := data.Message
		if msg == "" {
			msg = "Paystack rejected the charge"
		}
		return nil, fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, msg)
	}

	message := data.Data.DisplayText
	if message == "" {
		message = "Check your phone and enter your M-Pesa PIN to complete payment."
	}

	return &paymentdomain.InitiateResult{
		ProviderReference: data.Data.Reference,
		Status:            paymentdomain.PaymentStatusPending,
		CustomerMessage:   message,
		Raw: map[string]any{
			"reference":    data.Data.Reference,
			"status":       data.Data.Status,
			"display_text": data.Data.DisplayText,
		},
	}, nil
}

// ============================================================
// INITIATE — CARD (hosted checkout via /transaction/initialize)
// ============================================================

// initiateCard creates a Paystack checkout session for card payments.
//
// Returns a hosted checkout URL. The user enters card details on
// Paystack's PCI-compliant page.
func (p *Provider) initiateCard(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
	body := initializeRequest{
		Email:       req.PayerEmail,
		Amount:      strconv.FormatInt(req.Amount, 10),
		Currency:    req.Currency,
		Reference:   req.PaymentID,
		CallbackURL: req.ReturnURL,
		Channels:    []string{"card"},
		Metadata: map[string]interface{}{
			"payment_id": req.PaymentID,
			"order_id":   req.OrderID,
			"method":     "card",
		},
	}

	raw, err := p.client.do(ctx, "POST", "/transaction/initialize", body)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG CARD raw response: %s", string(raw))

	var data initializeResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode initialize response: %w", err)
	}

	if !data.Status || data.Data.AuthorizationURL == "" {
		return nil, fmt.Errorf("%w: Paystack returned no authorization URL",
			paymentdomain.ErrProviderRejected)
	}

	return &paymentdomain.InitiateResult{
		ProviderReference: data.Data.Reference,
		RedirectURL:       data.Data.AuthorizationURL,
		Status:            paymentdomain.PaymentStatusPending,
		CustomerMessage:   "Redirecting you to a secure card checkout.",
		Raw: map[string]any{
			"reference":         data.Data.Reference,
			"authorization_url": data.Data.AuthorizationURL,
			"access_code":       data.Data.AccessCode,
		},
	}, nil
}

// ============================================================
// VERIFY
// ============================================================

// Verify fetches the current state of a transaction from Paystack.
//
// Works for both M-Pesa and card transactions — Paystack serves both
// from the same verify endpoint.
func (p *Provider) Verify(
	ctx context.Context,
	reference string,
) (*paymentdomain.ProviderStatus, error) {
	path := "/transaction/verify/" + reference
	raw, err := p.client.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var data verifyResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode verify response: %w", err)
	}

	if !data.Status {
		return nil, fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, data.Message)
	}

	return &paymentdomain.ProviderStatus{
		Reference: data.Data.Reference,
		Status:    mapPaystackStatus(data.Data.Status),
		Amount:    data.Data.Amount,
		Currency:  data.Data.Currency,
		Message:   data.Data.GatewayResponse,
		UpdatedAt: parseTime(data.Data.PaidAt),
	}, nil
}

// ============================================================
// REFUND
// ============================================================

// Refund reverses a Paystack transaction.
func (p *Provider) Refund(
	ctx context.Context,
	req paymentdomain.RefundRequest,
) (*paymentdomain.RefundResult, error) {
	body := refundRequest{
		Transaction: req.ProviderReference,
		Amount:      strconv.FormatInt(req.Amount, 10),
		Currency:    req.Currency,
	}

	raw, err := p.client.do(ctx, "POST", "/refund", body)
	if err != nil {
		return nil, err
	}

	var data refundResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode refund response: %w", err)
	}

	if !data.Status {
		return nil, fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, data.Message)
	}

	return &paymentdomain.RefundResult{
		ProviderReference: data.Data.Reference,
		Status:            mapRefundStatus(data.Data.Status),
		Message:           data.Message,
	}, nil
}

// ============================================================
// HELPERS
// ============================================================

// normalizePhone converts a phone number into +254XXXXXXXXX format
// as required by Paystack's /charge endpoint.
//
// Accepts +254, 254, 0, and 9-digit (7XXXXXXXX) inputs.
func normalizePhone(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "+", "") // strip any existing +

	switch {
	case strings.HasPrefix(s, "254"):
		return "+" + s
	case strings.HasPrefix(s, "0"):
		return "+254" + s[1:]
	case len(s) == 9 && s[0] == '7':
		return "+254" + s
	default:
		return ""
	}
}

// mapPaystackStatus translates Paystack's transaction statuses into
// our domain statuses.
func mapPaystackStatus(status string) paymentdomain.PaymentStatus {
	switch strings.ToLower(status) {
	case "success":
		return paymentdomain.PaymentStatusSucceeded
	case "failed", "abandoned", "reversed":
		return paymentdomain.PaymentStatusFailed
	case "pending", "processing", "ongoing", "queued", "pay_offline", "send_otp":
		return paymentdomain.PaymentStatusPending
	default:
		return paymentdomain.PaymentStatusPending
	}
}

// mapRefundStatus translates Paystack's refund statuses into our
// domain statuses.
func mapRefundStatus(status string) paymentdomain.PaymentStatus {
	switch strings.ToLower(status) {
	case "success", "processed":
		return paymentdomain.PaymentStatusSucceeded
	case "failed":
		return paymentdomain.PaymentStatusFailed
	default:
		return paymentdomain.PaymentStatusPending
	}
}

// parseTime parses Paystack timestamps. Returns now if it can't parse.
func parseTime(s string) time.Time {
	if s == "" {
		return time.Now().UTC()
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}


// Compile-time assertion.
var _ paymentdomain.PaymentProvider = (*Provider)(nil)