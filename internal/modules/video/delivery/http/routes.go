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
	MeetingSDK *MeetingSDKHandler
}

// NewHandlers constructs all handlers.
func NewHandlers(svc service.Service) *Handlers {
	return &Handlers{
		OAuth:      NewOAuthHandler(svc),
		Connection: NewConnectionHandler(svc),
		MeetingSDK: NewMeetingSDKHandler(svc),
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

	// ---- Meeting SDK (embedded frontend) ----
	// Signature issues a short-lived JWT that authorizes the browser
	// to join a specific meeting. ZAK returns the host's Zoom Access
	// Key token for embedded host controls. Join-info wraps both for
	// the meeting page: it loads the meeting, decides the caller's
	// role, and returns everything the SDK needs in one response.
	r.Post("/meetings/signature", authMiddleware, h.MeetingSDK.Signature)
	r.Get("/meetings/zak", authMiddleware, h.MeetingSDK.ZAK)
	r.Get("/meetings/:id/join-info", authMiddleware, h.MeetingSDK.JoinInfo)
}

