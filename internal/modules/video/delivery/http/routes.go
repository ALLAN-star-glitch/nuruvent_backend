// internal/modules/video/delivery/http/routes.go

package http

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
)

// Handlers bundles the HTTP handlers for the video module.
type Handlers struct {
	OAuth      *OAuthHandler
	Connection *ConnectionHandler
}

// NewHandlers constructs all handlers.
func NewHandlers(svc service.Service) *Handlers {
	return &Handlers{
		OAuth:      NewOAuthHandler(svc),
		Connection: NewConnectionHandler(svc),
	}
}

// RegisterRoutes wires the video module's routes into the given router.
func RegisterRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
) {
	// ---- Public (state-bound) ----
	// The callback is publicly reachable by design — the platform's
	// server redirects the user's browser here. Auth is enforced by
	// the single-use state token, not by session.
	r.Get("/oauth/:platform/callback", h.OAuth.Callback)

	// ---- Host-facing ----
	r.Get("/oauth/:platform/connect", authMiddleware, h.OAuth.Connect)
	r.Get("/connections", authMiddleware, h.Connection.List)
	r.Post("/connections/:platform/disconnect", authMiddleware, h.Connection.Disconnect)
}