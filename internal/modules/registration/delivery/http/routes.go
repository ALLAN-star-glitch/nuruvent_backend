// internal/modules/registration/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires the handler methods into the Fiber router.
//
// For the register-for-event feature:
//   - Event-scoped routes (POST /events/:id/register) are registered here
//     to keep the flow self-contained. If the events module later needs to
//     own all /events/... routes, this can be moved.
//   - User-scoped and registration-scoped routes are registered here as well.
func RegisterRoutes(r fiber.Router, h *Handler, authMiddleware fiber.Handler) {
	// ------------------------------------------------------------
	// Event-scoped routes (registration flow)
	// ------------------------------------------------------------
	// POST /events/:id/register — register for an event
	// Auth is optional; the handler accepts either an authenticated user
	// or guest details in the body.
	r.Post("/events/:id/register", h.RegisterForEvent)

	// GET /events/:id/registrations — list registrations for an event
	// Auth required; only the event organizer should see this.
	r.Get("/events/:id/registrations", authMiddleware, h.ListByEvent)

	// POST /events/:id/waitlist — join the event waitlist
	// Auth optional; guest details allowed.
	r.Post("/events/:id/waitlist", h.JoinWaitlist)

	// POST /events/:id/waitlist/promote — promote the next waitlisted user
	// Auth required; typically organizer-only.
	r.Post("/events/:id/waitlist/promote", authMiddleware, h.PromoteFromWaitlist)

	// ------------------------------------------------------------
	// User-scoped routes (authenticated)
	// ------------------------------------------------------------
	me := r.Group("/me", authMiddleware)
	me.Get("/registrations", h.ListMine)

	// ------------------------------------------------------------
	// Registration-scoped routes (authenticated for now)
	// ------------------------------------------------------------
	reg := r.Group("/registrations", authMiddleware)
	reg.Get("/:id", h.GetByID)
	reg.Delete("/:id", h.Cancel)
}