// internal/modules/payment/delivery/http/order_handler_test.go

package http

import (
	"context"
	"testing"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
)

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	fake := &fakeService{
		createOrderFunc: func(ctx context.Context, cmd service.CreateOrderCommand) (*paymentdomain.Order, error) {
			if cmd.RegistrationID != "reg-1" {
				t.Errorf("registration_id: got %q", cmd.RegistrationID)
			}
			return &paymentdomain.Order{
				ID:             "order-1",
				RegistrationID: "reg-1",
				UserID:         "user-1",
				Currency:       "KES",
				Subtotal:       1500_00,
				DiscountTotal:  0,
				TotalAmount:    1500_00,
				Status:         paymentdomain.OrderStatusPending,
				Items: []paymentdomain.OrderItem{
					{
						TicketTypeID: "tkt-1",
						Quantity:     1,
						UnitPrice:    1500_00,
						LineTotal:    1500_00,
					},
				},
			}, nil
		},
	}

	app := newTestApp(t, fake)

	status, body := doJSON(t, app, "POST", "/api/v1/orders", map[string]any{
		"registration_id": "reg-1",
	})

	if status != 201 {
		t.Fatalf("status: got %d, want 201. body: %v", status, body)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data missing or wrong type: %T", body["data"])
	}
	if data["id"] != "order-1" {
		t.Errorf("id: got %v", data["id"])
	}
	if data["total_amount"] != float64(1500_00) {
		t.Errorf("total_amount: got %v", data["total_amount"])
	}
}

func TestOrderHandler_CreateOrder_MissingRegistrationID(t *testing.T) {
	app := newTestApp(t, &fakeService{})

	status, _ := doJSON(t, app, "POST", "/api/v1/orders", map[string]any{})

	if status != 400 {
		t.Fatalf("status: got %d, want 400", status)
	}
}

func TestOrderHandler_CreateOrder_DomainError(t *testing.T) {
	fake := &fakeService{
		createOrderFunc: func(ctx context.Context, cmd service.CreateOrderCommand) (*paymentdomain.Order, error) {
			return nil, paymentdomain.ErrOrderNotFound
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "POST", "/api/v1/orders", map[string]any{
		"registration_id": "reg-1",
	})

	if status != 404 {
		t.Fatalf("status: got %d, want 404", status)
	}
}

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	fake := &fakeService{
		getOrderFunc: func(ctx context.Context, orderID string) (*paymentdomain.Order, error) {
			return &paymentdomain.Order{
				ID:             "order-1",
				RegistrationID: "reg-1",
				Currency:       "KES",
				TotalAmount:    1500_00,
				Status:         paymentdomain.OrderStatusPending,
			}, nil
		},
	}
	app := newTestApp(t, fake)

	status, body := doJSON(t, app, "GET", "/api/v1/orders/order-1", nil)

	if status != 200 {
		t.Fatalf("status: got %d, want 200. body: %v", status, body)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data missing or wrong type: %T", body["data"])
	}
	if data["id"] != "order-1" {
		t.Errorf("id: got %v", data["id"])
	}
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	fake := &fakeService{
		getOrderFunc: func(ctx context.Context, orderID string) (*paymentdomain.Order, error) {
			return nil, paymentdomain.ErrOrderNotFound
		},
	}
	app := newTestApp(t, fake)

	status, _ := doJSON(t, app, "GET", "/api/v1/orders/missing", nil)

	if status != 404 {
		t.Fatalf("status: got %d, want 404", status)
	}
}