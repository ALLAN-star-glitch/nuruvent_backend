// internal/modules/payment/infrastructure/providers/flutterwave/provider_test.go

package flutterwave

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

// newTestProvider spins up an httptest server that simulates Flutterwave
// and returns a Provider pointed at it.
func newTestProvider(t *testing.T, handler http.HandlerFunc) (*Provider, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	cfg := config.FlutterwaveConfig{
		BaseURL:       srv.URL,
		SecretKey:     "FLWSECK_TEST-fake",
		PublicKey:     "FLWPUBK_TEST-fake",
		SecretHash:    "test-secret-hash",
		// 32-byte key required by AES-256-GCM
		EncryptionKey: "0123456789abcdef0123456789abcdef",
	}
	return NewProvider(cfg), srv
}

// ============================================================
// INITIATE — M-PESA
// ============================================================

func TestProvider_Initiate_Mpesa_Success(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/charges" || r.URL.Query().Get("type") != "mpesa" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		if r.Header.Get("Authorization") != "Bearer FLWSECK_TEST-fake" {
			t.Errorf("missing or wrong auth header: %q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "success",
			"message": "Charge initiated",
			"data": map[string]any{
				"id":         1191376,
				"tx_ref":     "pay-123",
				"flw_ref":    "MC-15852113s09v5050e8",
				"status":     "pending",
				"auth_model": "LIPA_MPESA",
				"amount":     15.0,
				"currency":   "KES",
			},
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
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProviderReference != "MC-15852113s09v5050e8" {
		t.Errorf("provider_reference: got %q", result.ProviderReference)
	}
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}
}

func TestProvider_Initiate_Mpesa_RejectedByProvider(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "phone_number is invalid",
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
}

func TestProvider_Initiate_RejectsInvalidPhone(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-123",
		Amount:     1500,
		Currency:   "KES",
		PayerPhone: "not-a-phone",
		Method:     paymentdomain.PaymentMethodMpesa,
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}
}

// ============================================================
// INITIATE — CARD
// ============================================================

func TestProvider_Initiate_Card_Success(t *testing.T) {
	var capturedBody map[string]any

	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/charges" || r.URL.Query().Get("type") != "card" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}

		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &capturedBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "success",
			"message": "Charge completed",
			"data": map[string]any{
				"id":         1191376,
				"tx_ref":     "pay-card-1",
				"flw_ref":    "FLW-CARD-REF",
				"status":     "successful",
				"auth_model": "NOAUTH",
				"amount":     15.0,
				"currency":   "KES",
			},
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:   "pay-card-1",
		OrderID:     "order-card-1",
		Amount:      1500,
		Currency:    "KES",
		PayerEmail:  "test@example.com",
		PayerPhone:  "+254709929220",
		Description: "Card payment test",
		Method:      paymentdomain.PaymentMethodCard,
		ReturnURL:   "https://example.com/payments/return",
		Card: &paymentdomain.CardDetails{
			Number: "5531886652142950",
			CVV:    "564",
			Expiry: "09/32",
		},
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("status: got %s, want succeeded", result.Status)
	}
	if result.ProviderReference != "FLW-CARD-REF" {
		t.Errorf("provider_reference: got %q", result.ProviderReference)
	}

	// SECURITY: verify the plaintext card number was NOT sent.
	if got, ok := capturedBody["card_number"].(string); ok {
		if got == "5531886652142950" {
			t.Error("SECURITY: card number was sent in plaintext")
		}
		if got == "" {
			t.Error("card_number was empty in the request body")
		}
	} else {
		t.Error("card_number missing from request body")
	}

	// Same for CVV
	if got, ok := capturedBody["cvv"].(string); ok {
		if got == "564" {
			t.Error("SECURITY: CVV was sent in plaintext")
		}
	}

	// Same for expiry
	if got, ok := capturedBody["expiry"].(string); ok {
		if got == "09/32" {
			t.Error("SECURITY: expiry was sent in plaintext")
		}
	}
}

func TestProvider_Initiate_Card_RequiresCardDetails(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not have been called")
	})

	req := paymentdomain.InitiateRequest{
		PaymentID: "pay-123",
		Amount:    1500,
		Currency:  "KES",
		Method:    paymentdomain.PaymentMethodCard,
		// Card is nil
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing card details")
	}
	if !strings.Contains(err.Error(), "card details") {
		t.Errorf("expected 'card details' error, got %v", err)
	}
}

func TestProvider_Initiate_Card_RequiresEncryptionKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	// Config WITHOUT encryption key
	cfg := config.FlutterwaveConfig{
		BaseURL:   srv.URL,
		SecretKey: "FLWSECK_TEST-fake",
		// EncryptionKey is empty
	}
	provider := NewProvider(cfg)

	req := paymentdomain.InitiateRequest{
		PaymentID: "pay-123",
		Amount:    1500,
		Currency:  "KES",
		Method:    paymentdomain.PaymentMethodCard,
		Card: &paymentdomain.CardDetails{
			Number: "5531886652142950",
			CVV:    "564",
			Expiry: "09/32",
		},
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing encryption key")
	}
	if !strings.Contains(err.Error(), "encryption key") {
		t.Errorf("expected 'encryption key' error, got %v", err)
	}
}

func TestProvider_Initiate_Card_RedirectFor3DS(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"id":           1191376,
				"tx_ref":       "pay-card-1",
				"flw_ref":      "FLW-CARD-REF",
				"status":       "pending",
				"auth_model":   "VBVSECURECODE",
				"amount":       15.0,
				"currency":     "KES",
				"redirect_url": "https://bank.example.com/3ds/verify?token=abc",
			},
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID:  "pay-card-1",
		Amount:     1500,
		Currency:   "KES",
		PayerEmail: "test@example.com",
		Method:     paymentdomain.PaymentMethodCard,
		ReturnURL:  "https://example.com/payments/return",
		Card: &paymentdomain.CardDetails{
			Number: "5531886652142950",
			CVV:    "564",
			Expiry: "09/32",
		},
	}

	result, err := provider.Initiate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != paymentdomain.PaymentStatusPending {
		t.Errorf("status: got %s, want pending", result.Status)
	}
	if result.RedirectURL == "" {
		t.Error("expected redirect_url to be set for 3DS flow")
	}
	if result.RedirectURL != "https://bank.example.com/3ds/verify?token=abc" {
		t.Errorf("redirect_url: got %q", result.RedirectURL)
	}
}

func TestProvider_Initiate_Card_RejectedByProvider(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Invalid card number",
		})
	})

	req := paymentdomain.InitiateRequest{
		PaymentID: "pay-123",
		Amount:    1500,
		Currency:  "KES",
		Method:    paymentdomain.PaymentMethodCard,
		Card: &paymentdomain.CardDetails{
			Number: "1234567890123456",
			CVV:    "564",
			Expiry: "09/32",
		},
	}

	_, err := provider.Initiate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Invalid card number") {
		t.Errorf("expected 'Invalid card number' error, got %v", err)
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

// ============================================================
// VERIFY
// ============================================================

func TestProvider_Verify_Successful(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/verify_by_reference" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("tx_ref") != "pay-123" {
			t.Errorf("unexpected tx_ref: %s", r.URL.Query().Get("tx_ref"))
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data": map[string]any{
				"id":                 1191376,
				"tx_ref":             "pay-123",
				"flw_ref":            "MC-15852113s09v5050e8",
				"status":             "successful",
				"amount":             15.0,
				"charged_amount":     15.0,
				"currency":           "KES",
				"payment_type":       "mpesa",
				"processor_response": "Approved",
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
}

// ============================================================
// REFUND
// ============================================================

func TestProvider_Refund_Success(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/FLW-REF-123/refund" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "success",
			"message": "Refund initiated",
			"data": map[string]any{
				"id":     999,
				"tx_ref": "pay-123",
				"flw_ref": "FLW-REF-123",
				"status": "completed",
				"amount": 15.0,
			},
		})
	})

	result, err := provider.Refund(context.Background(), paymentdomain.RefundRequest{
		PaymentID:         "pay-123",
		ProviderReference: "FLW-REF-123",
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
// WEBHOOK SIGNATURE
// ============================================================

func TestProvider_ParseWebhook_ValidSignature(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{
		"event": "charge.completed",
		"data": {
			"id": 1191376,
			"tx_ref": "pay-123",
			"flw_ref": "MC-15852113s09v5050e8",
			"status": "successful",
			"amount": 15.0,
			"currency": "KES",
			"payment_type": "mpesa",
			"processor_response": "Approved"
		}
	}`)

	headers := map[string]string{"verif-hash": "test-secret-hash"}

	evt, err := provider.ParseWebhook(context.Background(), body, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.ProviderReference != "pay-123" {
		t.Errorf("reference: got %q", evt.ProviderReference)
	}
	if evt.Status != paymentdomain.PaymentStatusSucceeded {
		t.Errorf("status: got %s", evt.Status)
	}
	if evt.ProviderEventID != "MC-15852113s09v5050e8" {
		t.Errorf("provider_event_id: got %q", evt.ProviderEventID)
	}
}

func TestProvider_ParseWebhook_InvalidSignature(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"event":"charge.completed","data":{"tx_ref":"pay-123","status":"successful"}}`)
	headers := map[string]string{"verif-hash": "wrong-hash"}

	_, err := provider.ParseWebhook(context.Background(), body, headers)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid webhook signature") {
		t.Errorf("expected signature error, got %v", err)
	}
}

func TestProvider_ParseWebhook_MissingSignature(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"event":"charge.completed","data":{}}`)
	headers := map[string]string{}

	_, err := provider.ParseWebhook(context.Background(), body, headers)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProvider_ParseWebhook_UnsupportedEvent(t *testing.T) {
	provider, _ := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {})

	body := []byte(`{"event":"transfer.completed","data":{}}`)
	headers := map[string]string{"verif-hash": "test-secret-hash"}

	_, err := provider.ParseWebhook(context.Background(), body, headers)
	if err == nil {
		t.Fatal("expected error for unsupported event")
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