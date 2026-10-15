// internal/modules/certificate/delivery/http/routes.go

package http

import (
	"github.com/gofiber/fiber/v3"

	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
)

// Handlers bundles the certificate module's HTTP handlers.
type Handlers struct {
	Handler *Handler
}

// NewHandlers constructs all handlers.
func NewHandlers(svc certificateService.Service) *Handlers {
	return &Handlers{Handler: NewHandler(svc)}
}

// ============================================================
// ROUTE REGISTRATION
// ============================================================
//
// NOTE ON AUTHZ
//
// Certificate routes use BOTH authzMiddleware and service-level authz
// (via domain.AuthorizationChecker). They are not redundant:
//
//   - authzMiddleware enforces the coarse "does the caller hold any
//     certificate permission in this account" gate at the edge. It
//     resolves (resource, action) from URL+method.
//
//   - Service-level Authz.CanX enforces the fine-grained decision,
//     including own-vs-all scope (ListMine vs ListByAccount) which
//     the middleware cannot express.
//
// Both must pass. If you remove either, one class of authorization
// breaks.
//
// ROUTE ORDERING
//
// /certificates/verify/:code MUST be registered before any other
// /certificates/* route, otherwise Fiber captures "verify" as the
// certificate ID.

// RegisterRoutes mounts the certificate module's main routes under the
// given router (typically the /api/v1 group).
//
// optionalAuth is used on the /request endpoint so guest buyers can
// request certificates with a matching guest_email.
func RegisterRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// ---- Public (no auth) ----
	r.Get("/certificates/verify/:code", h.Handler.Verify)


	// ---- Certificates (auth + authz) ----
		// Batch issuance (organizer-pays)
	r.Post("/events/:id/certificates/issue", authMiddleware, authzMiddleware, h.Handler.InitiateBatchIssuance)
	r.Post("/certificates/bulk", authMiddleware, authzMiddleware, h.Handler.IssueBulk)
	r.Post("/certificates/:id/revoke", authMiddleware, authzMiddleware, h.Handler.RevokeByUser)
	r.Get("/certificates/:id/download", authMiddleware, authzMiddleware, h.Handler.Download)

	// ---- Templates (auth + authz) ----
	r.Post("/certificate-templates", authMiddleware, authzMiddleware, h.Handler.CreateTemplate)
	r.Get("/certificate-templates", authMiddleware, authzMiddleware, h.Handler.ListTemplates)
	r.Put("/certificate-templates/:id", authMiddleware, authzMiddleware, h.Handler.UpdateTemplate)
	r.Delete("/certificate-templates/:id", authMiddleware, authzMiddleware, h.Handler.DeleteTemplate)
	r.Post("/certificate-templates/:id/default", authMiddleware, authzMiddleware, h.Handler.SetDefaultTemplate)
}

// RegisterUserRoutes mounts user-scoped certificate routes
// (/me/certificates).
func RegisterUserRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	r.Get("/me/certificates", authMiddleware, authzMiddleware, h.Handler.ListMine)
}

// RegisterAccountRoutes mounts account-scoped certificate routes
// (/accounts/:id/certificates).
func RegisterAccountRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	r.Get("/accounts/:id/certificates", authMiddleware, authzMiddleware, h.Handler.ListByAccount)
}