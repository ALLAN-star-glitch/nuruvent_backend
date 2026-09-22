package http

// internal/modules/payment/delivery/http/handler_test.go

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// FAKE SERVICE
// ============================================================

type fakeService struct {
	initiateFunc     func(ctx context.Context, cmd service.InitiateCommand) (*paymentdomain.Payment, error)
	confirmFunc      func(ctx context.Context, paymentID string) error
	failFunc         func(ctx context.Context, paymentID, reason string) error
	refundFunc       func(ctx context.Context, cmd service.RefundCommand) error
	handleWebhookFunc func(ctx context.Context, provider string, payload []byte, headers map[string]string) error
	getPaymentFunc    func(ctx context.Context, paymentID string) (*paymentdomain.Payment, error) 
}

func (f *fakeService) InitiatePayment(ctx context.Context, cmd service.InitiateCommand) (*paymentdomain.Payment, error) {
	if f.initiateFunc != nil {
		return f.initiateFunc(ctx, cmd)
	}
	return nil, nil
}

func (f *fakeService) GetPayment(ctx context.Context, paymentID string) (*paymentdomain.Payment, error) {
	if f.getPaymentFunc != nil {
		return f.getPaymentFunc(ctx, paymentID)
	}
	return nil, paymentdomain.ErrPaymentNotFound
}

func (f *fakeService) ConfirmPayment(ctx context.Context, paymentID string) error {
	if f.confirmFunc != nil {
		return f.confirmFunc(ctx, paymentID)
	}
	return nil
}

func (f *fakeService) FailPayment(ctx context.Context, paymentID, reason string) error {
	if f.failFunc != nil {
		return f.failFunc(ctx, paymentID, reason)
	}
	return nil
}

func (f *fakeService) Refund(ctx context.Context, cmd service.RefundCommand) error {
	if f.refundFunc != nil {
		return f.refundFunc(ctx, cmd)
	}
	return nil
}

func (f *fakeService) HandleWebhook(ctx context.Context, provider string, payload []byte, headers map[string]string) error {
	if f.handleWebhookFunc != nil {
		return f.handleWebhookFunc(ctx, provider, payload, headers)
	}
	return nil
}

// ============================================================
// TEST HARNESS
// ============================================================

// newTestApp returns a Fiber app with the payment routes and a fake
// service. The auth middleware is a no-op that sets a user ID.
func newTestApp(t *testing.T, fake *fakeService) *fiber.App {
    t.Helper()

    app := fiber.New()

    // Correct: uses the same constant production code reads from.
    authMiddleware := func(c fiber.Ctx) error {
        c.Locals(types.ContextKeyUserID, "test-user")
        return c.Next()
    }

    handler := NewHandler(fake)
    api := app.Group("/api/v1")
    RegisterRoutes(api, handler, authMiddleware)

    return app
}

// doJSON executes a JSON request against the app and returns the
// status code and parsed body.
func doJSON(t *testing.T, app *fiber.App, method, path string, body any) (int, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(buf)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5000})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	var parsed map[string]any
	raw, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(raw, &parsed)

	return resp.StatusCode, parsed
}

// ============================================================
// INITIATE PAYMENT
// ============================================================

func TestHandler_InitiatePayment_Mpesa_Success(t *testing.T) {
	fake := &fakeService{
		initiateFunc: func(ctx context.Context, cmd service.InitiateCommand) (*paymentdomain.Payment, error) {
			if cmd.OrderID != "order-1" {
				t.Errorf("order_id: got %q", cmd.OrderID)
			}
			if cmd.Method != paymentdomain.PaymentMethodMpesa {
				t.Errorf("method: got %q", cmd.Method)
			}
			if cmd.PayerPhone != "+254709929220" {
				t.Errorf("payer_phone: got %q", cmd.PayerPhone)
			}
			return &paymentdomain.Payment{
				ID:                "pay-1",
				OrderID:           "order-1",
				Provider:          "flutterwave",
				Method:            paymentdomain.PaymentMethodMpesa,
				Amount:            1500_00,
				Currency:          "KES",
				Status:            paymentdomain.PaymentStatusPending,
				ProviderReference: "FLW-REF-1",
			}, nil
		},
	}

	app := newTestApp(t, fake)

	status, body := doJSON(t, app, "POST", "/api/v1/payments/initiate", map[string]any{
		"order_id":        "order-1",
		"method":          "mpesa",
		"payer_phone":     "+254709929220",
		"idempotency_key": "idem-1",
	})

	if status != 201 {
		t.Fatalf("status: got %d, want 201. body: %v", status, body)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data missing or wrong type: %T", body["data"])
	}
	if data["provider_reference"] != "FLW-REF-1" {
		t.Errorf("provider_reference: got %v", data["provider_reference"])
	}
	if data["status"] != "pending" {
		t.Errorf("status: got %v", data["status"])
	}
}

func TestHandler_InitiatePayment_MissingOrderID(t *testing.T) {
	app := newTestApp(t, &fakeService{})

	status, body := doJSON(t, app, "POST", "/api/v1/payments/initiate", map[string]any{
		"method":          "mpesa",
		"idempotency_key": "idem-1",
	})

	if status != 400 {
		t.Fatalf("status: got %d, want 400. body: %v", status, body)
	}
}

func TestHandler_InitiatePayment_InvalidMethod(t *testing.T) {
	app := newTestApp(t, &fakeService{})

	status, _ := doJSON(t, app, "POST", "/api/v1/payments/initiate", map[string]any{
		"order_id":        "order-1",
		"method":          "barter",
		"idempotency_key": "idem-1",
	})

	if status != 400 {
		t.Fatalf("status: got %d, want 400", status)
	}
}

func TestHandler_InitiatePayment_DomainError(t *testing.T) {
	fake := &fakeService{
		initiateFunc: func(ctx context.Context, cmd service.InitiateCommand) (*paymentdomain.Payment, error) {
			return nil, paymentdomain.ErrOrderNotPending
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "POST", "/api/v1/payments/initiate", map[string]any{
		"order_id":        "order-1",
		"method":          "mpesa",
		"idempotency_key": "idem-1",
	})

	if status != 409 {
		t.Fatalf("status: got %d, want 409", status)
	}
}

func TestHandler_InitiatePayment_Card(t *testing.T) {
	fake := &fakeService{
		initiateFunc: func(ctx context.Context, cmd service.InitiateCommand) (*paymentdomain.Payment, error) {
			if cmd.Card == nil {
				t.Error("card details missing")
			} else if cmd.Card.Number != "5531886652142950" {
				t.Errorf("card number: got %q", cmd.Card.Number)
			}
			return &paymentdomain.Payment{
				ID:     "pay-1",
				Status: paymentdomain.PaymentStatusPending,
			}, nil
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "POST", "/api/v1/payments/initiate", map[string]any{
		"order_id":        "order-1",
		"method":          "card",
		"payer_email":     "test@example.com",
		"idempotency_key": "idem-1",
		"card": map[string]string{
			"number": "5531886652142950",
			"cvv":    "564",
			"expiry": "09/32",
		},
	})

	if status != 201 {
		t.Fatalf("status: got %d, want 201", status)
	}
}

// ============================================================
// REFUND
// ============================================================

func TestHandler_Refund_Success(t *testing.T) {
	fake := &fakeService{
		refundFunc: func(ctx context.Context, cmd service.RefundCommand) error {
			if cmd.PaymentID != "pay-1" {
				t.Errorf("payment_id: got %q", cmd.PaymentID)
			}
			if cmd.Amount != 500_00 {
				t.Errorf("amount: got %d", cmd.Amount)
			}
			if cmd.ActorID != "test-user" {
				t.Errorf("actor_id: got %q", cmd.ActorID)
			}
			return nil
		},
	}
	app := newTestApp(t, fake)

	status, body := doJSON(t, app, "POST", "/api/v1/payments/pay-1/refund", map[string]any{
		"amount": 500_00,
		"reason": "customer request",
	})

	if status != 200 {
		t.Fatalf("status: got %d, want 200. body: %v", status, body)
	}
}

func TestHandler_Refund_InvalidAmount(t *testing.T) {
	app := newTestApp(t, &fakeService{})

	status, _ := doJSON(t, app, "POST", "/api/v1/payments/pay-1/refund", map[string]any{
		"amount": 0,
	})

	if status != 400 {
		t.Fatalf("status: got %d, want 400", status)
	}
}

func TestHandler_Refund_DomainError(t *testing.T) {
	fake := &fakeService{
		refundFunc: func(ctx context.Context, cmd service.RefundCommand) error {
			return paymentdomain.ErrRefundExceedsPayment
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "POST", "/api/v1/payments/pay-1/refund", map[string]any{
		"amount": 100_000_00,
	})

	if status != 422 {
		t.Fatalf("status: got %d, want 422", status)
	}
}

// ============================================================
// WEBHOOK
// ============================================================

func TestHandler_FlutterwaveWebhook_Success(t *testing.T) {
	fake := &fakeService{
		handleWebhookFunc: func(ctx context.Context, provider string, payload []byte, headers map[string]string) error {
			if provider != "flutterwave" {
				t.Errorf("provider: got %q", provider)
			}
			if len(payload) == 0 {
				t.Error("empty payload")
			}
			return nil
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "POST", "/api/v1/webhooks/flutterwave", map[string]any{
		"event": "charge.completed",
	})

	if status != 200 {
		t.Fatalf("status: got %d, want 200", status)
	}
}

func TestHandler_FlutterwaveWebhook_InvalidSignature(t *testing.T) {
	fake := &fakeService{
		handleWebhookFunc: func(ctx context.Context, provider string, payload []byte, headers map[string]string) error {
			return paymentdomain.ErrInvalidWebhookSignature
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "POST", "/api/v1/webhooks/flutterwave", map[string]any{
		"event": "charge.completed",
	})

	if status != 401 {
		t.Fatalf("status: got %d, want 401", status)
	}
}

func TestHandler_FlutterwaveWebhook_EmptyBody(t *testing.T) {
	app := newTestApp(t, &fakeService{})

	req := httptest.NewRequest("POST", "/api/v1/webhooks/flutterwave", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5000})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("status: got %d, want 400", resp.StatusCode)
	}
}

// Silence unused import for errors.
var _ = errors.New