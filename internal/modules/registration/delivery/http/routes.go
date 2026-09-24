// internal/modules/registration/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(
	r fiber.Router,
	h *Handler,
	authMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// ------------------------------------------------------------
	// Event-scoped routes (registration flow)
	// ------------------------------------------------------------
	r.Post("/events/:id/register", optionalAuth, h.RegisterForEvent)
	r.Get("/events/:id/registrations", authMiddleware, h.ListByEvent)
	r.Post("/events/:id/waitlist", optionalAuth, h.JoinWaitlist)
	r.Post("/events/:id/waitlist/promote", authMiddleware, h.PromoteFromWaitlist)

	// ------------------------------------------------------------
	// User-scoped routes (authenticated)
	// ------------------------------------------------------------
	me := r.Group("/me", authMiddleware)
	me.Get("/registrations", h.ListMine)

	// ------------------------------------------------------------
	// Registration-scoped routes
	// ------------------------------------------------------------
	// GET allows guests to fetch their own registration by
	// supplying `?email=<guest_email>`.
	// DELETE requires full authentication.
	reg := r.Group("/registrations")
	reg.Get("/:id", optionalAuth, h.GetByID)
	reg.Delete("/:id", authMiddleware, h.Cancel)
}