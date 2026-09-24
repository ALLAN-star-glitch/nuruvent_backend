// internal/modules/payment/infrastructure/providers/intasend/provider.go

package intasend

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

const (
	ProviderName = "intasend"
)

// Provider implements paymentdomain.PaymentProvider for IntaSend.
type Provider struct {
	client *client
	cfg    config.IntaSendConfig
}

// NewProvider constructs an IntaSend provider.
func NewProvider(cfg config.IntaSendConfig) *Provider {
	return &Provider{
		client: newClient(cfg),
		cfg:    cfg,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string { return ProviderName }

// Method reports the primary payment method this provider serves.
func (p *Provider) Method() paymentdomain.PaymentMethod {
	return paymentdomain.PaymentMethodMpesa
}

// Methods returns every payment method this provider serves.
func (p *Provider) Methods() []paymentdomain.PaymentMethod {
	return []paymentdomain.PaymentMethod{
		paymentdomain.PaymentMethodMpesa,
		paymentdomain.PaymentMethodCard,
	}
}

// ============================================================
// INITIATE
// ============================================================

func (p *Provider) Initiate(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
) (*paymentdomain.InitiateResult, error) {
	if req.PayerEmail == "" {
		return nil, fmt.Errorf("%w: payer email is required",
			paymentdomain.ErrProviderRejected)
	}

	amountStr := strconv.FormatInt(req.Amount/100, 10)

	switch req.Method {
	case paymentdomain.PaymentMethodMpesa:
		return p.initiateMpesa(ctx, req, amountStr)
	case paymentdomain.PaymentMethodCard:
		return p.initiateCard(ctx, req, amountStr)
	default:
		return nil, fmt.Errorf("%w: unsupported payment method %q",
			paymentdomain.ErrProviderRejected, req.Method)
	}
}

// initiateMpesa triggers an M-Pesa STK push via /payment/collection/.
func (p *Provider) initiateMpesa(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
	amountStr string,
) (*paymentdomain.InitiateResult, error) {
	if req.PayerPhone == "" {
		return nil, fmt.Errorf("%w: payer phone is required for M-Pesa",
			paymentdomain.ErrProviderRejected)
	}

	body := checkoutRequest{
		Currency:    req.Currency,
		Amount:      amountStr,
		Email:       req.PayerEmail,
		PhoneNumber: normalizeKenyanPhone(req.PayerPhone),
		APIRef:      req.PaymentID,
		RedirectURL: req.ReturnURL,
		Comment:     req.Description,
		Method:      "M-PESA",
		PublicKey:   p.cfg.PublishableKey,
	}

	raw, err := p.client.do(ctx, "POST", "/payment/collection/", body)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG MPESA: raw response: %s", string(raw))

	var data checkoutResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode checkout response: %w", err)
	}

	if data.ID == "" {
		return nil, fmt.Errorf("%w: IntaSend returned no invoice ID",
			paymentdomain.ErrProviderRejected)
	}

	return &paymentdomain.InitiateResult{
		ProviderReference: data.ID,
		Status:            paymentdomain.PaymentStatusPending,
		CustomerMessage:   "Check your phone and enter your M-Pesa PIN to complete payment.",
		Raw: map[string]any{
			"id":      data.ID,
			"api_ref": data.APIRef,
			"state":   data.State,
		},
	}, nil
}

// initiateCard creates a hosted checkout session via /checkout/.
func (p *Provider) initiateCard(
	ctx context.Context,
	req paymentdomain.InitiateRequest,
	amountStr string,
) (*paymentdomain.InitiateResult, error) {
	// IntaSend's sandbox 3DS flow drops card submissions with an empty
	// phone_number. Use the caller's phone if provided, otherwise a
	// sandbox default. This is only a hint — the actual cardholder
	// phone doesn't affect the auth outcome in sandbox.
	phone := normalizeKenyanPhone(req.PayerPhone)
	if phone == "" {
		phone = "254708374149" // IntaSend sandbox test number
	}

	body := checkoutLinkRequest{
		PublicKey:   p.cfg.PublishableKey,
		Currency:    req.Currency,
		Amount:      amountStr,
		Email:       req.PayerEmail,
		PhoneNumber: phone,                // ← ADD
		APIRef:      req.PaymentID,
		Comment:     req.Description,
		RedirectURL: req.ReturnURL,
		Method:      "CARD-PAYMENT",
	}

	raw, err := p.client.doPublic(ctx, "POST", "/checkout/", body)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG CARD: raw response: %s", string(raw))

	var data checkoutLinkResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode checkout link response: %w", err)
	}

	if data.URL == "" {
		return nil, fmt.Errorf("%w: IntaSend returned no checkout URL",
			paymentdomain.ErrProviderRejected)
	}

	return &paymentdomain.InitiateResult{
		ProviderReference: data.ID,
		RedirectURL:       data.URL,
		Status:            paymentdomain.PaymentStatusPending,
		CustomerMessage:   "Redirecting you to a secure card checkout.",
		Raw: map[string]any{
			"id":  data.ID,
			"url": data.URL,
		},
	}, nil
}


// ============================================================
// VERIFY
// ============================================================

func (p *Provider) Verify(
	ctx context.Context,
	reference string,
) (*paymentdomain.ProviderStatus, error) {
	body := map[string]string{"invoice_id": reference}
	raw, err := p.client.do(ctx, "POST", "/payment/status/", body)
	if err != nil {
		return nil, err
	}

	var data statusResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode status response: %w", err)
	}

	if data.Detail != "" && data.Invoice.InvoiceID == "" {
		return nil, fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, data.Detail)
	}

	inv := data.Invoice
	amount := parseAmount(inv.Value)

	return &paymentdomain.ProviderStatus{
		Reference: inv.InvoiceID,
		Status:    mapInvoiceState(inv.State),
		Amount:    amount,
		Currency:  inv.Currency,
		Message:   inv.FailedReason,
		UpdatedAt: parseTime(inv.UpdatedAt),
	}, nil
}

// ============================================================
// REFUND
// ============================================================

func (p *Provider) Refund(
	ctx context.Context,
	req paymentdomain.RefundRequest,
) (*paymentdomain.RefundResult, error) {
	body := refundRequest{
		APIRef: req.ProviderReference,
		Amount: strconv.FormatInt(req.Amount/100, 10),
	}

	raw, err := p.client.do(ctx, "POST", "/payment/refund/", body)
	if err != nil {
		return nil, err
	}

	var data refundResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("decode refund response: %w", err)
	}

	return &paymentdomain.RefundResult{
		ProviderReference: data.ID,
		Status:            mapRefundState(data.State),
		Message:           data.Reason,
	}, nil
}

// ============================================================
// HELPERS
// ============================================================

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

func mapInvoiceState(state string) paymentdomain.PaymentStatus {
	switch strings.ToUpper(state) {
	case "COMPLETE", "COMPLETED", "SUCCESS":
		return paymentdomain.PaymentStatusSucceeded
	case "FAILED", "CANCELLED":
		return paymentdomain.PaymentStatusFailed
	case "PENDING", "PROCESSING":
		return paymentdomain.PaymentStatusPending
	default:
		return paymentdomain.PaymentStatusPending
	}
}

func mapRefundState(state string) paymentdomain.PaymentStatus {
	switch strings.ToUpper(state) {
	case "COMPLETE", "COMPLETED":
		return paymentdomain.PaymentStatusSucceeded
	case "FAILED":
		return paymentdomain.PaymentStatusFailed
	default:
		return paymentdomain.PaymentStatusPending
	}
}

func parseAmount(v any) int64 {
	switch x := v.(type) {
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0
		}
		return int64(f * 100)
	case float64:
		return int64(x * 100)
	case int64:
		return x * 100
	case nil:
		return 0
	default:
		return 0
	}
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Now().UTC()
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

func customerMessageFor(method paymentdomain.PaymentMethod) string {
	switch method {
	case paymentdomain.PaymentMethodMpesa:
		return "Check your phone and enter your M-Pesa PIN to complete payment."
	case paymentdomain.PaymentMethodCard:
		return "Redirecting you to a secure card checkout."
	default:
		return "Redirecting you to complete payment."
	}
}

var _ paymentdomain.PaymentProvider = (*Provider)(nil)