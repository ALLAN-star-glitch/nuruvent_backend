// internal/modules/attendance/delivery/http/routes.go

package http

import (
	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
)

// Handlers bundles the HTTP handlers for the attendance module.
type Handlers struct {
	Sessions   *SessionHandler
	Attendance *AttendanceHandler
	Webhooks   *WebhookHandler
	Join       *JoinHandler
}

// NewHandlers constructs all handlers.
func NewHandlers(
	svc service.Service,
	providers map[attendance.SessionProvider]service.ProviderAdapter,
) *Handlers {
	return &Handlers{
		Sessions:   NewSessionHandler(svc),
		Attendance: NewAttendanceHandler(svc),
		Webhooks:   NewWebhookHandler(svc, providers),
		Join:       NewJoinHandler(svc),
	}
}

// RegisterRoutes wires the attendance module's routes into the given
// router.
func RegisterRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// ---- Public ----
	r.Get("/join/:token", h.Join.RedeemJoinToken)
	r.Post("/webhooks/zoom", h.Webhooks.ZoomWebhook)
	// r.Post("/webhooks/google-meet", h.Webhooks.GoogleMeetWebhook) // deferred

	// ---- Consumer-facing (other modules) ----
	r.Post("/attendees", authMiddleware, h.Sessions.RegisterAttendee)
	r.Post("/attendees/:id/register-for-external", authMiddleware, h.Sessions.RegisterAttendeeForExternal)
	r.Post("/sessions", authMiddleware, h.Sessions.UpsertSession)

	// ---- Host-facing ----
	r.Post("/sessions/:id/join-tokens", authMiddleware, h.Sessions.IssueJoinToken)
	r.Delete("/sessions/:id/join-tokens", authMiddleware, h.Sessions.RevokeJoinTokens)

	r.Get("/sessions/:id/attendance", optionalAuth, h.Attendance.ListSessionAttendance)
	r.Get("/sessions/:id/attendance/export", optionalAuth, h.Attendance.ExportSessionAttendance)
	r.Post("/sessions/:id/attendance/:attendeeID/confirm", authMiddleware, h.Attendance.ConfirmAttendance)
	r.Post("/sessions/:id/attendance/:attendeeID/override", authMiddleware, h.Attendance.OverrideAttendance)
	r.Post("/sessions/:id/attendance/bulk-confirm", authMiddleware, h.Attendance.BulkConfirm)

	r.Get("/attendees/:id/summary", optionalAuth, h.Attendance.GetAttendeeSummary)
}