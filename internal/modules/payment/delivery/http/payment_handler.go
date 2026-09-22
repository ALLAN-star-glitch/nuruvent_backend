// internal/modules/payment/delivery/http/payment_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// Handler holds the service and exposes HTTP methods.
type Handler struct {
	svc service.Service
}

// NewHandler constructs a Handler.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// ============================================================
// INITIATE PAYMENT
// ============================================================

// InitiatePayment handles POST /payments/initiate.
func (h *Handler) InitiatePayment(c fiber.Ctx) error {
	var body InitiatePaymentRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	// Basic shape validation
	if body.OrderID == "" {
		return response.BadRequest(c, "order_id is required", nil)
	}
	if body.IdempotencyKey == "" {
		return response.BadRequest(c, "idempotency_key is required", nil)
	}

	method := paymentdomain.PaymentMethod(body.Method)
	if !method.IsValid() {
		return response.BadRequest(c, "invalid payment method", nil)
	}

	// Convert card DTO to domain type.
	var card *paymentdomain.CardDetails
	if body.Card != nil {
		card = &paymentdomain.CardDetails{
			Number: body.Card.Number,
			CVV:    body.Card.CVV,
			Expiry: body.Card.Expiry,
		}
	}

	cmd := service.InitiateCommand{
		OrderID:        body.OrderID,
		Method:         method,
		PayerPhone:     body.PayerPhone,
		PayerEmail:     body.PayerEmail,
		Card:           card,
		IdempotencyKey: body.IdempotencyKey,
		ReturnURL:      body.ReturnURL,
		Description:    body.Description,
	}

	payment, err := h.svc.InitiatePayment(c.Context(), cmd)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Payment initiated successfully", toPaymentResponse(payment))
}

// ============================================================
// GET PAYMENT
// ============================================================

// GetPayment handles GET /payments/:id.


// GetPayment handles GET /payments/:id.
//
// For MVP, any authenticated user can read a payment by ID. A stricter
// version would check ownership (payment → order → user) and reject
// requests from users who don't own the payment.
func (h *Handler) GetPayment(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Payment ID is required", nil)
	}

	payment, err := h.svc.GetPayment(c.Context(), id)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Payment retrieved successfully", toPaymentResponse(payment))
}

// ============================================================
// REFUND
// ============================================================

// Refund handles POST /payments/:id/refund.
func (h *Handler) Refund(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Payment ID is required", nil)
	}

	actorID := handlerhelper.GetUserIDOptional(c)
	if actorID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body RefundRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}
	if body.Amount <= 0 {
		return response.BadRequest(c, "amount must be positive", nil)
	}

	cmd := service.RefundCommand{
		PaymentID:      id,
		Amount:         body.Amount,
		Reason:         body.Reason,
		ActorID:        actorID,
		IdempotencyKey: body.IdempotencyKey,
	}

	if err := h.svc.Refund(c.Context(), cmd); err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Refund processed successfully", nil)
}

// ============================================================
// WEBHOOK
// ============================================================

// FlutterwaveWebhook handles POST /webhooks/flutterwave.
//
// This route is NOT behind auth middleware. The Flutterwave signature
// header is the sole authentication mechanism. The service verifies it
// inside HandleWebhook.
func (h *Handler) FlutterwaveWebhook(c fiber.Ctx) error {
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

	if err := h.svc.HandleWebhook(c.Context(), "flutterwave", payload, headers); err != nil {
		return mapDomainError(c, err)
	}

	// Flutterwave expects a 200 with no body or a small acknowledgment.
	return response.Success(c, "Webhook received", nil)
}