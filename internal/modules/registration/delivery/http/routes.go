// internal/modules/registration/delivery/http/routes.go

package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires the registration module's routes into the given
// router.
//
// Auth policy:
//
//   - Public/guest routes use optionalAuth. Guest registration and
//     lookup enforce ownership by matching the guest email in the
//     handler/service. They do NOT need authzMiddleware — the caller
//     may not even be a Nuruvent user.
//   - Organizer-facing reads and mutations use authMiddleware AND
//     authzMiddleware. These expose registration data for an event
//     or for the whole account and must be authorized against the
//     caller's role.
//   - Self-service /me routes use authMiddleware only. They operate
//     strictly on the caller's own data, so a domain-scoped authz
//     check is deferred to the service (or not needed at all).
func RegisterRoutes(
	r fiber.Router,
	h *Handler,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// ---- Event-scoped routes ----

	// Guest-eligible: caller may register without a session.
	// Ownership enforced by guest-email match in the service.
	r.Post("/events/:id/register", optionalAuth, h.RegisterForEvent)

	// Organizer-facing: reads the event's registration list.
	r.Get("/events/:id/registrations", authMiddleware, authzMiddleware, h.ListByEvent)

	// Guest-eligible: join a waitlist without a session.
	r.Post("/events/:id/waitlist", optionalAuth, h.JoinWaitlist)

	// Organizer-facing: promote from waitlist (mutates the event's
	// registration set).
	r.Post("/events/:id/waitlist/promote", authMiddleware, authzMiddleware, h.PromoteFromWaitlist)

	// ---- Registration-scoped routes ----
	// Register BEFORE the /registrations group's parameterised routes
	// so the static path is not captured by /registrations/:id.

	// Organizer-facing: cross-event registration view for the account.
	r.Get("/registrations", authMiddleware, authzMiddleware, h.ListAllRegistrations)

	// Guest-eligible: caller may read a registration by ID without a
	// session. Ownership enforced in the service.
	r.Get("/registrations/:id", optionalAuth, h.GetByID)

	// Organizer-facing: cancel a registration. Mutates account-scoped
	// data.
	r.Delete("/registrations/:id", authMiddleware, authzMiddleware, h.Cancel)

	// ---- User-scoped routes ----
	// Self-service. Operates on the caller's own data only. No authz
	// check needed — the service scopes by the authenticated user.
	me := r.Group("/me", authMiddleware)
	me.Get("/registrations", h.ListMine)
	me.Get("/session-links", h.GetMySessionLinks)
}