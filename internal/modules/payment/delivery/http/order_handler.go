// internal/modules/payment/delivery/http/order_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// OrderHandler exposes order-related HTTP endpoints.
//
// Orders are created from a registration. The client sends only the
// registration ID — the payment module resolves the pricing, items,
// and totals.
type OrderHandler struct {
	svc service.Service
}

// NewOrderHandler constructs the handler.
func NewOrderHandler(svc service.Service) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// ============================================================
// CREATE ORDER
// ============================================================

// CreateOrder handles POST /orders.
//
// Body: { "registration_id": "<uuid>" }
//
// Returns 201 with the created (or existing) order.
//
// Idempotent: calling this endpoint twice with the same registration
// returns the same order. If the registration already has a pending
// order, that order is returned instead of creating a duplicate.
func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
	var body CreateOrderRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if body.RegistrationID == "" {
		return response.BadRequest(c, "registration_id is required", nil)
	}

	order, err := h.svc.CreateOrder(c.Context(), service.CreateOrderCommand{
		RegistrationID: body.RegistrationID,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Order created successfully", toOrderResponse(order))
}


// ============================================================
// GET ORDER
// ============================================================

// GetOrder handles GET /orders/:id.
//
// For MVP, any authenticated user can read any order by ID. A stricter
// version would check ownership (order → registration → user).
func (h *OrderHandler) GetOrder(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Order ID is required", nil)
	}

	order, err := h.svc.GetOrder(c.Context(), id)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Order retrieved successfully", toOrderResponse(order))
}

// PaystackWebhook handles POST /webhooks/paystack.
//
// This route is NOT behind auth middleware. The Paystack signature
// header (x-paystack-signature) is the sole authentication mechanism.
// The service verifies it inside HandleWebhook.
func (h *Handler) PaystackWebhook(c fiber.Ctx) error {
	// Read raw body — signature verification needs the exact bytes.
	payload := c.Body()
	if len(payload) == 0 {
		return response.BadRequest(c, "Empty webhook payload", nil)
	}

	// Collect headers as a map.
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	if err := h.svc.HandleWebhook(c.Context(), "paystack", payload, headers); err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Webhook received", nil)
}