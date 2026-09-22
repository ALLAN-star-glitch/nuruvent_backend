// internal/modules/payment/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires the handler methods into the Fiber router.
//
// The webhook route is intentionally NOT behind auth middleware —
// Flutterwave identifies itself via the verif-hash signature header,
// which the service verifies inside HandleWebhook.
func RegisterRoutes(
	r fiber.Router,
	h *Handler,
	authMiddleware fiber.Handler,
) {
	// Payment initiation — authenticated
	r.Post("/payments/initiate", authMiddleware, h.InitiatePayment)

	// Payment read — authenticated
	r.Get("/payments/:id", authMiddleware, h.GetPayment)

	// Refund — authenticated (organizer-only enforced in service)
	r.Post("/payments/:id/refund", authMiddleware, h.Refund)

	// Webhook — no auth, signature verification inside the service
	r.Post("/webhooks/flutterwave", h.FlutterwaveWebhook)
}