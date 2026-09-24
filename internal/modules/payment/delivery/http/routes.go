// internal/modules/payment/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires the payment module's handler methods into the
// Fiber router.
//
// Two handlers are registered:
//
//   - OrderHandler: order creation and retrieval
//   - Handler: payment initiation, retrieval, refund, and webhook
//
// All routes are authenticated except the webhook. The webhook uses
// challenge-based verification (inside the service) as its sole
// authentication mechanism — IntaSend has no user session to present.
func RegisterRoutes(
	r fiber.Router,
	orderHandler *OrderHandler,
	paymentHandler *Handler,
	authMiddleware fiber.Handler,
) {
	// ============================================================
	// ORDER ROUTES — authenticated
	// ============================================================
	r.Post("/orders", authMiddleware, orderHandler.CreateOrder)
	r.Get("/orders/:id", authMiddleware, orderHandler.GetOrder)

	// ============================================================
	// PAYMENT ROUTES — authenticated
	// ============================================================
	r.Post("/payments/initiate", authMiddleware, paymentHandler.InitiatePayment)
	r.Get("/payments/:id", authMiddleware, paymentHandler.GetPayment)
	r.Post("/payments/:id/refund", authMiddleware, paymentHandler.Refund)

	// ============================================================
	// WEBHOOK — no auth, challenge verification inside the service
	// ============================================================
	r.Post("/webhooks/intasend", paymentHandler.IntaSendWebhook)
	r.Post("/webhooks/paystack", paymentHandler.PaystackWebhook)
}