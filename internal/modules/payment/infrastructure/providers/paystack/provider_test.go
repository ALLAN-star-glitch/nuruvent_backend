// internal/modules/payment/infrastructure/providers/paystack/provider_test.go

package paystack

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// newTestProvider spins up an httptest server that simulates Paystack
// and returns a Provider pointed at it.
func newTestProvider(t *testing.T, handler http.HandlerFunc) (*Provider, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := config.PaystackConfig{
		BaseURL:   srv.URL,
		SecretKey: "sk_test_fake",
		PublicKey: "pk_test_fake",
		Enabled:   true,
	}
	return NewProvider(cfg), srv
}

// ============================================================
// INITIATE — M-PESA
// ============================================================

func TestProvider_Initiate_Mpesa_Success(t *testing.T) {
	var capturedBody map[string]any

	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/initialize" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer sk_test_fake" {
			t.Errorf("wrong auth header: %q", r.Header.Get("Authorization"))
		}

		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Authorization URL created",
			"data": map[string]any{
				"authorization_url": "https://checkout.paystack.com/abc123",
				"access_code":       "abc123",
				"reference":         "pay-123",
			},
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:   "pay-123",
		OrderID:     "order-123",
		Amount:      1000, // minor units = KES 10
		Currency:    "KES",
		PayerPhone:  "254708374149",
		PayerEmail:  "test@example.com",
		Description: "Test payment",
		Method:      paymentdomain.PaymentMethodMpesa,
		ReturnURL:   "https://example.com/return",
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProviderReference != "pay-123" {
		t.Errorf("provider_reference: got %q", result.ProviderReference)
	}
	if result.RedirectURL != "https://checkout.paystack.com/abc123" {
		t.Errorf("redirect_url: got %q", result.RedirectURL)
	}
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}

	// Body checks
	if capturedBody["amount"] != "1000" {
		t.Errorf("amount should be '1000' (minor units as string), got %v", capturedBody["amount"])
	}
	if capturedBody["email"] != "test@example.com" {
		t.Errorf("email mismatch: %v", capturedBody["email"])
	}
	if capturedBody["currency"] != "KES" {
		t.Errorf("currency mismatch: %v", capturedBody["currency"])
	}
	if capturedBody["reference"] != "pay-123" {
		t.Errorf("reference mismatch: %v", capturedBody["reference"])
	}

	// channels should be ["mobile_money"]
	channels, ok := capturedBody["channels"].([]any)
	if !ok || len(channels) != 1 || channels[0] != "mobile_money" {
		t.Errorf("expected channels=[mobile_money], got %v", capturedBody["channels"])
	}
}

func TestProvider_Initiate_Mpesa_RequiresPhone(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-123",
		Amount:     1000,
		Currency:   "KES",
		PayerEmail: "test@example.com",
		Method:     paymentdomain.PaymentMethodMpesa,
		// PayerPhone is empty
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing phone")
	}
	if !strings.Contains(err.Error(), "phone") {
		t.Errorf("expected phone error, got %v", err)
	}
}

// ============================================================
// INITIATE — CARD
// ============================================================

func TestProvider_Initiate_Card_Success(t *testing.T) {
	var capturedBody map[string]any

	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Authorization URL created",
			"data": map[string]any{
				"authorization_url": "https://checkout.paystack.com/card123",
				"access_code":       "card123",
				"reference":         "pay-card-1",
			},
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:   "pay-card-1",
		OrderID:     "order-1",
		Amount:      1000,
		Currency:    "KES",
		PayerEmail:  "test@example.com",
		Description: "Card payment",
		Method:      paymentdomain.PaymentMethodCard,
		ReturnURL:   "https://example.com/return",
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RedirectURL == "" {
		t.Error("expected redirect_url to be set")
	}

	// channels should be ["card"]
	channels, ok := capturedBody["channels"].([]any)
	if !ok || len(channels) != 1 || channels[0] != "card" {
		t.Errorf("expected channels=[card], got %v", capturedBody["channels"])
	}

	// SECURITY: no card details should be in the request — Paystack
	// hosts the card form. We only send an email + amount.
	for _, forbidden := range []string{"card_number", "cvv", "expiry", "card"} {
		if _, present := capturedBody[forbidden]; present {
			t.Errorf("SECURITY: %q should not be in the request body", forbidden)
		}
	}
}

// ============================================================
// INITIATE — VALIDATION
// ============================================================

func TestProvider_Initiate_RequiresEmail(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID: "pay-123",
		Amount:    1000,
		Currency:  "KES",
		Method:    paymentdomain.PaymentMethodMpesa,
		// PayerEmail is empty
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing email")
	}
	if !strings.Contains(err.Error(), "email") {
		t.Errorf("expected email error, got %v", err)
	}
}

func TestProvider_Initiate_UnsupportedMethod(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-123",
		Amount:     1000,
		Currency:   "KES",
		PayerEmail: "test@example.com",
		Method:     paymentdomain.PaymentMethod("barter"),
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' error, got %v", err)
	}
}

func TestProvider_Initiate_RejectedByProvider(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  false,
			"message": "Invalid key",
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-123",
		Amount:     1000,
		Currency:   "KES",
		PayerEmail: "test@example.com",
		PayerPhone: "254708374149",
		Method:     paymentdomain.PaymentMethodMpesa,
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Invalid key") {
		t.Errorf("expected 'Invalid key', got %v", err)
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected authentication failed wrapper, got %v", err)
	}
}

// ============================================================
// VERIFY
// ============================================================

func TestProvider_Verify_Success(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/verify/pay-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Verification successful",
			"data": map[string]any{
				"id":               12345,
				"status":           "success",
				"reference":        "pay-123",
				"amount":           1000,
				"currency":         "KES",
				"gateway_response": "Successful",
				"paid_at":          "2026-09-23T12:00:00Z",
				"channel":          "mobile_money",
			},
		})
	})

	status, err := provider.Verify(context.Background(), "pay-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("status: got %s, want succeeded", status.Status)
	}
	if status.Amount != 1000 {
		t.Errorf("amount: got %d, want 1000", status.Amount)
	}
	if status.Currency != "KES" {
		t.Errorf("currency: got %q", status.Currency)
	}
}

func TestProvider_Verify_Failed(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Verification successful",
			"data": map[string]any{
				"id":               12345,
				"status":           "failed",
				"reference":        "pay-123",
				"amount":           1000,
				"currency":         "KES",
				"gateway_response": "Insufficient funds",
				"paid_at":          "2026-09-23T12:00:00Z",
			},
		})
	})

	status, err := provider.Verify(context.Background(), "pay-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != paymentdomain.PaymentStatusFailed {
		t.Errorf("status: got %s, want failed", status.Status)
	}
}

func TestProvider_Verify_Pending(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Verification successful",
			"data": map[string]any{
				"id":         12345,
				"status":     "pending",
				"reference":  "pay-123",
				"amount":     1000,
				"currency":   "KES",
				"paid_at":    "",
			},
		})
	})

	status, err := provider.Verify(context.Background(), "pay-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", status.Status)
	}
}

// ============================================================
// REFUND
// ============================================================

func TestProvider_Refund_Success(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refund" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Refund initiated",
			"data": map[string]any{
				"id":        999,
				"status":    "processing",
				"amount":    1000,
				"currency":  "KES",
				"reference": "refund-1",
			},
		})
	})

	result, err := provider.Refund(context.Background(), paymentdomain.RefundRequest{
		PaymentID:         "pay-123",
		ProviderReference: "pay-123",
		Amount:            1000,
		Currency:          "KES",
		Reason:            "customer request",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}
}

// ============================================================
// WEBHOOK — SIGNATURE VERIFICATION
// ============================================================

// computeSignature reproduces the HMAC-SHA512 signature Paystack sends
// in the x-paystack-signature header.
func computeSignature(secret string, payload []byte) string {
	hash := hmac.New(sha512.New, []byte(secret))
	hash.Write(payload)
	return hex.EncodeToString(hash.Sum(nil))
}

func TestProvider_ParseWebhook_ValidSignature(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{
		"event": "charge.success",
		"data": {
			"id": 12345,
			"reference": "pay-123",
			"status": "success",
			"amount": 1000,
			"currency": "KES",
			"gateway_response": "Successful",
			"paid_at": "2026-09-23T12:00:00Z",
			"channel": "mobile_money"
		}
	}`)

	sig := computeSignature("sk_test_fake", body)
	headers := map[string]string{"x-paystack-signature": sig}

	evt, err := provider.ParseWebhook(context.Background(), body, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.ProviderReference != "pay-123" {
		t.Errorf("reference: got %q", evt.ProviderReference)
	}
	if evt.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("status: got %s, want succeeded", evt.Status)
	}
	if evt.Amount != 1000 {
		t.Errorf("amount: got %d, want 1000", evt.Amount)
	}
}

func TestProvider_ParseWebhook_InvalidSignature(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"event":"charge.success","data":{"reference":"pay-123","status":"success"}}`)
	headers := map[string]string{"x-paystack-signature": "deadbeef"}

	_, err := provider.ParseWebhook(context.Background(), body, headers)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if !strings.Contains(err.Error(), "invalid webhook signature") {
		t.Errorf("expected signature error, got %v", err)
	}
}

func TestProvider_ParseWebhook_MissingSignature(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"event":"charge.success","data":{"reference":"pay-123"}}`)

	_, err := provider.ParseWebhook(context.Background(), body, map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing signature")
	}
}

func TestProvider_ParseWebhook_UnsupportedEvent(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"event":"transfer.success","data":{"reference":"pay-123"}}`)
	sig := computeSignature("sk_test_fake", body)
	headers := map[string]string{"x-paystack-signature": sig}

	_, err := provider.ParseWebhook(context.Background(), body, headers)
	if err == nil {
		t.Fatal("expected error for unsupported event")
	}
	if !strings.Contains(err.Error(), "unsupported event") {
		t.Errorf("expected 'unsupported event' error, got %v", err)
	}
}

func TestProvider_ParseWebhook_ChargeFailed(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{
		"event": "charge.failed",
		"data": {
			"reference": "pay-123",
			"status": "failed",
			"amount": 1000,
			"currency": "KES",
			"gateway_response": "Insufficient funds"
		}
	}`)
	sig := computeSignature("sk_test_fake", body)
	headers := map[string]string{"x-paystack-signature": sig}

	evt, err := provider.ParseWebhook(context.Background(), body, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.Status != paymentdomain.PaymentStatusFailed {
		t.Errorf("status: got %s, want failed", evt.Status)
	}
	if evt.Message != "Insufficient funds" {
		t.Errorf("message: got %q", evt.Message)
	}
}

// ============================================================
// STATE MAPPING
// ============================================================

func TestMapPaystackStatus(t *testing.T) {
	cases := []struct {
		in   string
		want paymentdomain.PaymentStatus
	}{
		{"success", paymentdomain.PaymentStatusSucceeded},
		{"SUCCESS", paymentdomain.PaymentStatusSucceeded},
		{"failed", paymentdomain.PaymentStatusFailed},
		{"abandoned", paymentdomain.PaymentStatusFailed},
		{"reversed", paymentdomain.PaymentStatusFailed},
		{"pending", paymentdomain.PaymentStatusPending},
		{"processing", paymentdomain.PaymentStatusPending},
		{"ongoing", paymentdomain.PaymentStatusPending},
		{"queued", paymentdomain.PaymentStatusPending},
		{"", paymentdomain.PaymentStatusPending},
		{"unknown", paymentdomain.PaymentStatusPending},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := mapPaystackStatus(tc.in)
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}



// ============================================================
// INITIATE — M-PESA (direct charge, not checkout)
// ============================================================

func TestProvider_Initiate_Mpesa_UsesChargeEndpoint(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]any

	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path

		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Charge attempted",
			"data": map[string]any{
				"reference":    "pay-123",
				"status":       "pay_offline",
				"display_text": "Please enter your M-Pesa PIN on your phone",
			},
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:   "pay-123",
		OrderID:     "order-123",
		Amount:      1000,
		Currency:    "KES",
		PayerPhone:  "254708374149",
		PayerEmail:  "test@example.com",
		Method:      paymentdomain.PaymentMethodMpesa,
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must call /charge, not /transaction/initialize.
	if capturedPath != "/charge" {
		t.Errorf("expected /charge, got %s", capturedPath)
	}

	// No redirect URL for M-Pesa.
	if result.RedirectURL != "" {
		t.Errorf("M-Pesa should not return a redirect URL, got %q", result.RedirectURL)
	}

	// Response should be pending, awaiting webhook.
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}

	// Verify the body includes the mobile_money block with mpesa provider.
	mm, ok := capturedBody["mobile_money"].(map[string]any)
	if !ok {
		t.Fatal("expected mobile_money in request body")
	}
	if mm["phone"] != "254708374149" {
		t.Errorf("phone: got %v", mm["phone"])
	}
	if mm["provider"] != "mpesa" {
		t.Errorf("provider should be 'mpesa' (M-Pesa only), got %v", mm["provider"])
	}

	// Confirm the request did not ask for a channel list.
	// /charge doesn't use `channels` — it uses the provider field.
	if _, present := capturedBody["channels"]; present {
		t.Error("charge request should not include a channels array")
	}
}

// ============================================================
// INITIATE — CARD (hosted checkout, unchanged)
// ============================================================

func TestProvider_Initiate_Card_UsesCheckoutEndpoint(t *testing.T) {
	var capturedPath string
	var capturedBody map[string]any

	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path

		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  true,
			"message": "Authorization URL created",
			"data": map[string]any{
				"authorization_url": "https://checkout.paystack.com/card123",
				"access_code":       "card123",
				"reference":         "pay-card-1",
			},
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-card-1",
		OrderID:    "order-1",
		Amount:     1000,
		Currency:   "KES",
		PayerEmail: "test@example.com",
		Method:     paymentdomain.PaymentMethodCard,
		ReturnURL:  "https://example.com/return",
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/transaction/initialize" {
		t.Errorf("expected /transaction/initialize, got %s", capturedPath)
	}

	if result.RedirectURL == "" {
		t.Error("card should return a redirect URL")
	}

	// Channels should be ["card"].
	channels, ok := capturedBody["channels"].([]any)
	if !ok || len(channels) != 1 || channels[0] != "card" {
		t.Errorf("expected channels=[card], got %v", capturedBody["channels"])
	}
}