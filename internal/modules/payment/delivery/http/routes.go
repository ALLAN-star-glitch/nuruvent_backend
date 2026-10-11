// internal/modules/payment/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires the payment module's handler methods into the
// Fiber router.
//
// Handlers registered:
//
//   - OrderHandler: order creation and retrieval
//   - Handler:     payment initiation, retrieval, listing, stats,
//                  refund, and webhook
//
// Auth policy:
//
//   - Order and payment *creation* endpoints use optionalAuth — guest
//     checkouts are allowed, and ownership is enforced in the
//     handler/service by matching either the JWT subject or the
//     guest email on the order.
//   - The ledger, stats, single-payment lookup, and refund endpoints
//     require a full session (authMiddleware) AND a permission check
//     (authzMiddleware) — they are organizer actions scoped to the
//     caller's active account.
//   - Webhooks are unauthenticated by design; the provider's signature
//     is the sole authentication mechanism, verified inside the
//     service.
//
// Route ordering matters:
//
//   - Static paths MUST be registered before parameterised paths.
//     `/payments/stats` must come before `/payments/:id`, otherwise
//     Fiber matches the parameterised route with id="stats" and
//     returns 404.
//   - `/payments/initiate` is a POST and `/payments/:id` is a GET,
//     so they don't collide by method. It is still placed early for
//     readability — creation endpoints grouped together.
func RegisterRoutes(
	r fiber.Router,
	orderHandler *OrderHandler,
	paymentHandler *Handler,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// ------------------------------------------------------------
	// Orders
	// ------------------------------------------------------------
	// Guests may create an order for a registration they own
	// (matched by guest email). Ownership is verified in the handler
	// and service.
	r.Post("/orders", optionalAuth, orderHandler.CreateOrder)
	r.Get("/orders/:id", optionalAuth, orderHandler.GetOrder)

	// ------------------------------------------------------------
	// Payment initiation
	// ------------------------------------------------------------
	// Guests may initiate payment for an order they own. Ownership
	// is verified in the handler and service.
	r.Post("/payments/initiate", optionalAuth, paymentHandler.InitiatePayment)

	// ------------------------------------------------------------
	// Payment ledger and stats — organizer view
	// ------------------------------------------------------------
	// Require a full session AND a payment:read permission.
	// Scoped to the caller's active account (JWT's account_id).
	// Placed BEFORE /payments/:id so the static "stats" path is not
	// captured by the parameterised route.
	r.Get("/payments", authMiddleware, authzMiddleware, paymentHandler.ListPayments)
	r.Get("/payments/stats", authMiddleware, authzMiddleware, paymentHandler.GetPaymentStats)

	// ------------------------------------------------------------
	// Single payment lookup
	// ------------------------------------------------------------
	// Requires a full session AND a payment:read permission.
	// Previously optionalAuth — a payment ID is not a capability
	// token, and unauthenticated reads let anyone who guessed or
	// intercepted an ID inspect the payment. Guests who need to
	// check their payment status read the order via GET /orders/:id
	// instead, which enforces guest-email ownership.
	//
	// Registered after the static paths above so /payments/stats and
	// /payments resolve to their specific handlers.
	r.Get("/payments/:id", authMiddleware, authzMiddleware, paymentHandler.GetPayment)

	// ------------------------------------------------------------
	// Refunds — organizer or admin action
	// ------------------------------------------------------------
	// Requires a full session AND a payment:refund permission.
	// The service enforces role checks on top of the middleware.
	r.Post("/payments/:id/refund", authMiddleware, authzMiddleware, paymentHandler.Refund)

	// ------------------------------------------------------------
	// Webhooks — signature-verified, no session
	// ------------------------------------------------------------
	// The provider's signature is the only authentication. The
	// service verifies it before processing. No auth middleware here
	// — providers cannot present a user session.
	r.Post("/webhooks/intasend", paymentHandler.IntaSendWebhook)
	r.Post("/webhooks/paystack", paymentHandler.PaystackWebhook)
}