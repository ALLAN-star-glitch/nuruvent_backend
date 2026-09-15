// internal/modules/registration/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires the handler methods into the Fiber router.
//
// Event-scoped routes (/events/:id/register, etc.) are expected to be
// registered by the events module, which will call this handler's
// methods directly. Only user-scoped and registration-scoped routes
// are registered here.
func RegisterRoutes(r fiber.Router, h *Handler, authMiddleware fiber.Handler) {
	// User-scoped routes (authenticated)
	me := r.Group("/me", authMiddleware)
	me.Get("/registrations", h.ListMine)

	// Registration-scoped routes (authenticated)
	reg := r.Group("/registrations", authMiddleware)
	reg.Get("/:id", h.GetByID)
	reg.Delete("/:id", h.Cancel)

	// Waitlist routes (public or authenticated)
	// Note: /events/:id/... routes are typically owned by the events module.
	// If events delegates here, expose these methods to be called directly.
}