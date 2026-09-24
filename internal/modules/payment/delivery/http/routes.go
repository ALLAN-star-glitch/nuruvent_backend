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
    optionalAuth fiber.Handler,   // add this param
) {
    // Orders — allow guests, verify ownership in the handler/service
    r.Post("/orders", optionalAuth, orderHandler.CreateOrder)
    r.Get("/orders/:id", optionalAuth, orderHandler.GetOrder)

    // Payments — allow guests, verify ownership in the handler/service
    r.Post("/payments/initiate", optionalAuth, paymentHandler.InitiatePayment)
    r.Get("/payments/:id", optionalAuth, paymentHandler.GetPayment)

    // Refund — keep auth-only (admin/organizer action)
    r.Post("/payments/:id/refund", authMiddleware, paymentHandler.Refund)

    // Webhooks unchanged
    r.Post("/webhooks/intasend", paymentHandler.IntaSendWebhook)
    r.Post("/webhooks/paystack", paymentHandler.PaystackWebhook)
}