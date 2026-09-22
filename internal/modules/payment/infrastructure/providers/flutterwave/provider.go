// internal/modules/payment/infrastructure/providers/flutterwave/provider.go

package flutterwave

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

const (
	ProviderName = "flutterwave"
)

// Provider implements paymentdomain.PaymentProvider for Flutterwave v3.
type Provider struct {
	client *client
	cfg    config.FlutterwaveConfig
}

// NewProvider constructs a Flutterwave provider.
func NewProvider(cfg config.FlutterwaveConfig) *Provider {
	return &Provider{
		client: newClient(cfg),
		cfg:    cfg,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string { return ProviderName }

// Method reports the primary payment method this provider serves.
// Flutterwave serves both M-Pesa and card; the registry uses this as
// the default when ByMethod is called with an ambiguous request.
func (p *Provider) Method() paymentdomain.PaymentMethod {
	return paymentdomain.PaymentMethodMpesa
}

// ============================================================
// INITIATE
// ============================================================

// Initiate starts a charge via Flutterwave.
//
// For M-Pesa: sends an STK push to the customer's phone.
// For card: encrypts card data and submits it for processing.
//
// The method is chosen from req.Method.
func (p *Provider) Initiate(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
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
// M-PESA
// ============================================================

// initiateMpesa sends an STK push via Flutterwave.
func (p *Provider) initiateMpesa(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
	phone := normalizeKenyanPhone(req.PayerPhone)
	if phone == "" {
		return nil, fmt.Errorf("%w: invalid phone number",
			paymentdomain.ErrProviderRejected)
	}

	body := mpesaChargeRequest{
		Amount:      req.Amount / 100, // Flutterwave expects major units (KES)
		Currency:    req.Currency,
		PhoneNumber: phone,
		Email:       req.PayerEmail,
		TxRef:       req.PaymentID, // our internal ID as the reference
		FullName:    req.Description,
	}

	raw, err := p.client.do(ctx, "POST", "/charges?type=mpesa", body)
	if err != nil {
		return nil, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var data chargeResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decode charge data: %w", err)
	}

	return &paymentdomain.InitiateResult{
		ProviderReference: data.FlwRef,
		Status:            mapChargeStatus(data.Status),
		CustomerMessage:   "Enter your M-Pesa PIN on your phone to complete payment.",
		Raw: map[string]any{
			"id":         data.ID,
			"tx_ref":     data.TxRef,
			"flw_ref":    data.FlwRef,
			"status":     data.Status,
			"auth_model": data.AuthModel,
		},
	}, nil
}

// ============================================================
// CARD
// ============================================================

// initiateCard submits an encrypted card charge via Flutterwave.
//
// Flutterwave requires card details to be encrypted with AES-256-GCM
// using the configured encryption key before submission. The response
// may indicate:
//
//   - "successful" — charge completed immediately (no 3DS)
//   - "pending"    — customer must complete an action (PIN, OTP, 3DS)
//
// For "pending" responses, the caller must handle the next step
// (usually a redirect to the auth URL).
func (p *Provider) initiateCard(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
	// Card details come from the request; the DTO carries them in a
	// nested object to keep them clearly separated from the M-Pesa
	// fields.
	if req.Card == nil {
		return nil, fmt.Errorf("%w: card details are required",
			paymentdomain.ErrProviderRejected)
	}
	if p.cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("%w: encryption key is not configured",
			paymentdomain.ErrProviderRejected)
	}

	// Encrypt each sensitive field individually. Flutterwave expects
	// the encrypted values in the same fields as the plaintext would
	// occupy.
	encNumber, err := encryptCardPayload(req.Card.Number, p.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt card number: %w", err)
	}
	encCVV, err := encryptCardPayload(req.Card.CVV, p.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt cvv: %w", err)
	}
	encExpiry, err := encryptCardPayload(req.Card.Expiry, p.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt expiry: %w", err)
	}

	body := cardChargeRequest{
		CardNumber:  encNumber,
		CVV:         encCVV,
		Expiry:      encExpiry,
		Currency:    req.Currency,
		Amount:      req.Amount / 100,
		FullName:    req.Description,
		Email:       req.PayerEmail,
		PhoneNumber: normalizeKenyanPhone(req.PayerPhone),
		TxRef:       req.PaymentID,
		RedirectURL: req.ReturnURL,
		Authorization: cardAuthorization{
			Mode: "redirect", // default to 3DS; Flutterwave will tell us
		},
		Meta: map[string]interface{}{
			"payment_id": req.PaymentID,
			"order_id":   req.OrderID,
		},
	}

	raw, err := p.client.do(ctx, "POST", "/charges?type=card", body)
	if err != nil {
		return nil, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var data chargeResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decode charge data: %w", err)
	}

	result := &paymentdomain.InitiateResult{
		ProviderReference: data.FlwRef,
		Status:            mapChargeStatus(data.Status),
		Raw: map[string]any{
			"id":           data.ID,
			"tx_ref":       data.TxRef,
			"flw_ref":      data.FlwRef,
			"status":       data.Status,
			"auth_model":   data.AuthModel,
			"redirect_url": data.RedirectURL,
		},
	}

	// Populate the redirect URL if the auth model requires a browser
	// step (3DS, VBV, or secure code).
	if data.RedirectURL != "" {
		result.RedirectURL = data.RedirectURL
		result.CustomerMessage = "Redirecting you to your bank for verification…"
	} else {
		result.CustomerMessage = "Processing your card…"
	}

	return result, nil
}

// ============================================================
// VERIFY
// ============================================================

// Verify fetches a transaction's current state from Flutterwave using
// our internal tx_ref.
func (p *Provider) Verify(
	ctx context.Context,
	reference string,
) (*paymentdomain.ProviderStatus, error) {
	path := "/transactions/verify_by_reference?tx_ref=" + reference
	raw, err := p.client.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var data verifyResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decode verify data: %w", err)
	}

	return &paymentdomain.ProviderStatus{
		Reference: data.TxRef,
		Status:    mapChargeStatus(data.Status),
		Amount:    int64(data.Amount * 100),
		Currency:  data.Currency,
		Message:   data.ProcessorResponse,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// ============================================================
// REFUND
// ============================================================

// Refund reverses a transaction. Requires the Flutterwave transaction
// ID, which is stored as part of the payment's provider reference.
func (p *Provider) Refund(
	ctx context.Context,
	req paymentdomain.RefundRequest,
) (*paymentdomain.RefundResult, error) {
	body := refundRequest{
		Amount: req.Amount / 100,
	}

	path := fmt.Sprintf("/transactions/%s/refund", req.ProviderReference)
	raw, err := p.client.do(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var data refundResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decode refund data: %w", err)
	}

	return &paymentdomain.RefundResult{
		ProviderReference: data.FlwRef,
		Status:            mapRefundStatus(data.Status),
		Message:           env.Message,
	}, nil
}

// ============================================================
// HELPERS
// ============================================================

// normalizeKenyanPhone converts a phone number into the format
// Flutterwave expects: 254XXXXXXXXX (no leading + or 0).
func normalizeKenyanPhone(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")

	switch {
	case strings.HasPrefix(s, "+254"):
		return s[1:]
	case strings.HasPrefix(s, "254"):
		return s
	case strings.HasPrefix(s, "0"):
		return "254" + s[1:]
	case len(s) == 9 && s[0] == '7':
		return "254" + s
	default:
		return ""
	}
}

// mapChargeStatus translates Flutterwave's charge statuses into our
// domain statuses.
func mapChargeStatus(flwStatus string) paymentdomain.PaymentStatus {
	switch strings.ToLower(flwStatus) {
	case "successful", "success":
		return paymentdomain.PaymentStatusSucceeded
	case "failed":
		return paymentdomain.PaymentStatusFailed
	case "pending":
		return paymentdomain.PaymentStatusPending
	default:
		return paymentdomain.PaymentStatusPending
	}
}

// mapRefundStatus translates Flutterwave's refund statuses into ours.
func mapRefundStatus(flwStatus string) paymentdomain.PaymentStatus {
	switch strings.ToLower(flwStatus) {
	case "completed", "successful":
		return paymentdomain.PaymentStatusSucceeded
	case "failed":
		return paymentdomain.PaymentStatusFailed
	default:
		return paymentdomain.PaymentStatusPending
	}
}

// Compile-time assertion.
var _ paymentdomain.PaymentProvider = (*Provider)(nil)