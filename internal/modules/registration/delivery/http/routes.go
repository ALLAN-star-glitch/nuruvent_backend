// internal/modules/registration/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(
	r fiber.Router,
	h *Handler,
	authMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// Event-scoped routes
	r.Post("/events/:id/register", optionalAuth, h.RegisterForEvent)
	r.Get("/events/:id/registrations", authMiddleware, h.ListByEvent)
	r.Post("/events/:id/waitlist", optionalAuth, h.JoinWaitlist)
	r.Post("/events/:id/waitlist/promote", authMiddleware, h.PromoteFromWaitlist)

	// Registration-scoped routes (register BEFORE /registrations group)
	// GET /registrations — cross-event organizer view
	r.Get("/registrations", authMiddleware, h.ListAllRegistrations)

	// Registration-scoped routes with :id
	r.Get("/registrations/:id", optionalAuth, h.GetByID)
	r.Delete("/registrations/:id", authMiddleware, h.Cancel)

	// User-scoped routes
	me := r.Group("/me", authMiddleware)
	me.Get("/registrations", h.ListMine)
	me.Get("/session-links", h.GetMySessionLinks)
}