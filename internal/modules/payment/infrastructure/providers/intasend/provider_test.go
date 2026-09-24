// internal/modules/payment/infrastructure/providers/intasend/provider_test.go

package intasend

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// newTestProvider spins up an httptest server that simulates IntaSend
// and returns a Provider pointed at it.
func newTestProvider(t *testing.T, handler http.HandlerFunc) (*Provider, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := config.IntaSendConfig{
		BaseURL:        srv.URL,
		SecretKey:      "ISSecretKey_test_fake",
		PublishableKey: "ISPubKey_test_fake",
		Challenge:      "test-challenge",
		Enabled:        true,
	}
	return NewProvider(cfg), srv
}

// ============================================================
// INITIATE — M-PESA
// ============================================================

func TestProvider_Initiate_Mpesa_Success(t *testing.T) {
	var capturedBody map[string]any

	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payment/collection/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer ISSecretKey_test_fake" {
			t.Errorf("wrong auth header: %q", r.Header.Get("Authorization"))
		}

		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "INV-123",
			"url":     "https://sandbox.intasend.com/checkout/INV-123",
			"api_ref": "pay-123",
			"state":   "PENDING",
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:   "pay-123",
		OrderID:     "order-123",
		Amount:      1500, // minor units (KES 15.00)
		Currency:    "KES",
		PayerPhone:  "+254709929220",
		PayerEmail:  "test@example.com",
		Description: "Test payment",
		Method:      paymentdomain.PaymentMethodMpesa,
		ReturnURL:   "https://example.com/return",
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProviderReference != "INV-123" {
		t.Errorf("provider_reference: got %q", result.ProviderReference)
	}
	if result.RedirectURL != "https://sandbox.intasend.com/checkout/INV-123" {
		t.Errorf("redirect_url: got %q", result.RedirectURL)
	}
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}

	// Verify the request body IntaSend received.
	if capturedBody["amount"] != "15" {
		t.Errorf("amount should be '15' (major units as string), got %v", capturedBody["amount"])
	}
	if capturedBody["method"] != "M-PESA" {
		t.Errorf("method should be 'M-PESA', got %v", capturedBody["method"])
	}
	if capturedBody["phone_number"] != "254709929220" {
		t.Errorf("phone should be normalized, got %v", capturedBody["phone_number"])
	}
	if capturedBody["api_ref"] != "pay-123" {
		t.Errorf("api_ref should be our payment ID, got %v", capturedBody["api_ref"])
	}
}

func TestProvider_Initiate_Mpesa_RejectedByProvider(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"detail": "phone_number is invalid",
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-123",
		Amount:     1500,
		Currency:   "KES",
		PayerPhone: "+254709929220",
		PayerEmail: "test@example.com",
		Method:     paymentdomain.PaymentMethodMpesa,
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "phone_number") {
		t.Errorf("expected phone_number error, got %v", err)
	}
	if !strings.Contains(err.Error(), "provider rejected") {
		t.Errorf("expected ErrProviderRejected wrapper, got %v", err)
	}
}

func TestProvider_Initiate_Mpesa_RequiresPhone(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-123",
		Amount:     1500,
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
			"id":      "INV-CARD-1",
			"url":     "https://sandbox.intasend.com/checkout/INV-CARD-1",
			"api_ref": "pay-card-1",
			"state":   "PENDING",
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:   "pay-card-1",
		OrderID:     "order-card-1",
		Amount:      1500,
		Currency:    "KES",
		PayerEmail:  "test@example.com",
		PayerPhone:  "+254709929220",
		Description: "Card payment",
		Method:      paymentdomain.PaymentMethodCard,
		ReturnURL:   "https://example.com/payments/return",
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}
	if result.RedirectURL == "" {
		t.Error("expected redirect_url to be set")
	}

	// Verify the request body uses the correct IntaSend method string.
	if capturedBody["method"] != "CARD-PAYMENT" {
		t.Errorf("method should be 'CARD-PAYMENT', got %v", capturedBody["method"])
	}

	// SECURITY: verify NO card data was sent — hosted checkout means
	// we never touch card details.
	for _, forbidden := range []string{"card_number", "cvv", "expiry"} {
		if _, present := capturedBody[forbidden]; present {
			t.Errorf("SECURITY: %q should not be in the request body (hosted checkout)", forbidden)
		}
	}
}

// ============================================================
// INITIATE — UNSUPPORTED METHOD
// ============================================================

func TestProvider_Initiate_UnsupportedMethod(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID: "pay-123",
		Amount:    1500,
		Currency:  "KES",
		Method:    paymentdomain.PaymentMethod("barter"),
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
	if !strings.Contains(err.Error(), "unsupported payment method") {
		t.Errorf("expected 'unsupported' error, got %v", err)
	}
}

func TestProvider_Initiate_RequiresEmail(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID: "pay-123",
		Amount:    1500,
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

// ============================================================
// VERIFY
// ============================================================

func TestProvider_Verify_Successful(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payment/collection/status/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("api_ref") != "pay-123" {
			t.Errorf("unexpected api_ref: %s", r.URL.Query().Get("api_ref"))
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"invoice": map[string]any{
				"id":           "INV-123",
				"state":        "COMPLETE",
				"provider":     "M-PESA",
				"provider_ref": "MPESA-RECEIPT-123",
				"api_ref":      "pay-123",
				"amount":       "15.00",
				"currency":     "KES",
				"created_at":   "2026-09-23T18:00:00Z",
				"updated_at":   "2026-09-23T18:01:00Z",
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
	if status.Amount != 1500 {
		t.Errorf("amount: got %d, want 1500", status.Amount)
	}
	if status.Currency != "KES" {
		t.Errorf("currency: got %q", status.Currency)
	}
}

func TestProvider_Verify_Pending(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"invoice": map[string]any{
				"id":         "INV-123",
				"state":      "PROCESSING",
				"api_ref":    "pay-123",
				"amount":     "15.00",
				"currency":   "KES",
				"created_at": "2026-09-23T18:00:00Z",
				"updated_at": "2026-09-23T18:00:10Z",
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

func TestProvider_Verify_Failed(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"invoice": map[string]any{
				"id":            "INV-123",
				"state":         "FAILED",
				"api_ref":       "pay-123",
				"amount":        "15.00",
				"currency":      "KES",
				"failed_reason": "Insufficient funds",
				"created_at":    "2026-09-23T18:00:00Z",
				"updated_at":    "2026-09-23T18:00:30Z",
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
	if status.Message != "Insufficient funds" {
		t.Errorf("message: got %q", status.Message)
	}
}

// ============================================================
// REFUND
// ============================================================

func TestProvider_Refund_Success(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payment/refund/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":         "RF-123",
			"invoice_id": "INV-123",
			"api_ref":    "pay-123",
			"amount":     "15.00",
			"state":      "COMPLETE",
			"created_at": "2026-09-23T18:05:00Z",
		})
	})

	result, err := provider.Refund(context.Background(), paymentdomain.RefundRequest{
		PaymentID:         "pay-123",
		ProviderReference: "pay-123",
		Amount:            1500,
		Currency:          "KES",
		Reason:            "customer request",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("status: got %s, want succeeded", result.Status)
	}
}

// ============================================================
// WEBHOOK — CHALLENGE VERIFICATION
// ============================================================

func TestProvider_ParseWebhook_ValidChallenge(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{
		"challenge": "test-challenge",
		"invoice_id": "INV-123",
		"state": "COMPLETE",
		"provider": "M-PESA",
		"api_ref": "pay-123",
		"amount": "15.00",
		"currency": "KES",
		"provider_ref": "MPESA-RECEIPT-123",
		"created_at": "2026-09-23T18:00:00Z",
		"updated_at": "2026-09-23T18:01:00Z"
	}`)

	evt, err := provider.ParseWebhook(context.Background(), body, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.ProviderReference != "pay-123" {
		t.Errorf("reference: got %q", evt.ProviderReference)
	}
	if evt.ProviderEventID != "INV-123" {
		t.Errorf("provider_event_id: got %q", evt.ProviderEventID)
	}
	if evt.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("status: got %s, want succeeded", evt.Status)
	}
	if evt.Amount != 1500 {
		t.Errorf("amount: got %d, want 1500", evt.Amount)
	}
}

func TestProvider_ParseWebhook_InvalidChallenge(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{
		"challenge": "wrong-challenge",
		"invoice_id": "INV-123",
		"state": "COMPLETE",
		"api_ref": "pay-123",
		"amount": "15.00",
		"currency": "KES"
	}`)

	_, err := provider.ParseWebhook(context.Background(), body, map[string]string{})
	if err == nil {
		t.Fatal("expected error for invalid challenge")
	}
	if !strings.Contains(err.Error(), "invalid webhook signature") {
		t.Errorf("expected signature error, got %v", err)
	}
}

func TestProvider_ParseWebhook_MissingChallenge(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"invoice_id":"INV-123","state":"COMPLETE","api_ref":"pay-123"}`)

	_, err := provider.ParseWebhook(context.Background(), body, map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing challenge")
	}
}

func TestProvider_ParseWebhook_NotACollectionEvent(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	// A payload without invoice_id — not a collection event.
	body := []byte(`{"challenge":"test-challenge"}`)

	_, err := provider.ParseWebhook(context.Background(), body, map[string]string{})
	if err == nil {
		t.Fatal("expected error for non-collection event")
	}
	if !strings.Contains(err.Error(), "not a collection event") {
		t.Errorf("expected 'not a collection event' error, got %v", err)
	}
}

// ============================================================
// PHONE NORMALIZATION
// ============================================================

func TestNormalizeKenyanPhone(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"+254709929220", "254709929220"},
		{"254709929220", "254709929220"},
		{"0709929220", "254709929220"},
		{"0709 929 220", "254709929220"},
		{"0709-929-220", "254709929220"},
		{"709929220", "254709929220"},
		{"invalid", ""},
		{"123", ""},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := normalizeKenyanPhone(tc.in)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// ============================================================
// STATE MAPPING
// ============================================================

func TestMapInvoiceState(t *testing.T) {
	cases := []struct {
		in   string
		want paymentdomain.PaymentStatus
	}{
		{"COMPLETE", paymentdomain.PaymentStatusSucceeded},
		{"COMPLETED", paymentdomain.PaymentStatusSucceeded},
		{"complete", paymentdomain.PaymentStatusSucceeded},
		{"FAILED", paymentdomain.PaymentStatusFailed},
		{"PENDING", paymentdomain.PaymentStatusPending},
		{"PROCESSING", paymentdomain.PaymentStatusPending},
		{"", paymentdomain.PaymentStatusPending},
		{"UNKNOWN", paymentdomain.PaymentStatusPending},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := mapInvoiceState(tc.in)
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}