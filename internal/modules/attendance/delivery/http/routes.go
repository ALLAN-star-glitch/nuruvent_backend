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

	// svc is retained so the delivery layer can construct middleware
	// that needs a specific use case (e.g. SessionContextMiddleware,
	// which resolves a session's owning account before authz runs).
	svc service.Service // ← NEW
}

// NewHandlers constructs all handlers.
func NewHandlers(
	svc service.Service,
	providers map[attendance.SessionProvider]service.ProviderAdapter,
	authIssuer attendance.AuthSessionIssuer,
) *Handlers {
	return &Handlers{
		Sessions:   NewSessionHandler(svc),
		Attendance: NewAttendanceHandler(svc),
		Webhooks:   NewWebhookHandler(svc, providers),
		Join:       NewJoinHandler(svc, authIssuer),
		svc:        svc, // ← NEW
	}
}

// Service returns the attendance service backing these handlers.
//
// Used by server/routes.go to build middleware that depends on a
// specific use case — currently SessionContextMiddleware, which needs
// ResolveSessionAccount to inject the Casbin domain before
// authzMiddleware runs on /sessions/:id/* routes.
func (h *Handlers) Service() service.Service { // ← NEW
	return h.svc
}

// RegisterRoutes wires the attendance module's routes into the given
// router.
//
// Auth policy:
//
//   - Public routes: join-token redemption and provider webhooks.
//     Neither has a session; the token or the provider signature is
//     the authentication.
//   - Consumer-facing POSTs (register attendee, upsert session) use
//     authMiddleware. They operate on data the caller already owns,
//     so a domain-scoped authz check is deferred to the service.
//   - Host-facing reads and mutations use authMiddleware AND
//     authzMiddleware. These endpoints expose attendance for a
//     session, an event, or the whole account, and must be
//     authorized against the caller's role.
//   - Session-scoped routes (/sessions/:id/*) additionally insert
//     sessionCtx between authMiddleware and authzMiddleware. sessionCtx
//     resolves the session's owning account and writes account:<uuid>
//     into ContextKeyDomain, so authzMiddleware can run a domain-scoped
//     Casbin check. Without it, authzMiddleware has no domain to check
//     against on these routes.
func RegisterRoutes(
	r fiber.Router,
	h *Handlers,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
	sessionCtx fiber.Handler, // ← NEW
) {
	// ---- Public ----
	r.Get("/join/:token", h.Join.RedeemJoinToken)
	r.Post("/webhooks/zoom", h.Webhooks.ZoomWebhook)
	// r.Post("/webhooks/google-meet", h.Webhooks.GoogleMeetWebhook) // deferred

	// ---- Consumer-facing (other modules) ----
	// Authn only. The service scopes writes to the caller.
	r.Post("/attendees", authMiddleware, h.Sessions.RegisterAttendee)
	r.Post("/attendees/:id/register-for-external", authMiddleware, h.Sessions.RegisterAttendeeForExternal)
	r.Post("/sessions", authMiddleware, h.Sessions.UpsertSession)

	// ---- Host-facing: session join tokens ----
	// NOTE: these still lack a domain resolver and will 500 under
	// authzMiddleware. Fix tracked separately; sessionCtx will be
	// inserted here in a follow-up once the attendance routes are
	// verified.
	r.Post("/sessions/:id/join-tokens", authMiddleware, authzMiddleware, h.Sessions.IssueJoinToken)
	r.Delete("/sessions/:id/join-tokens", authMiddleware, authzMiddleware, h.Sessions.RevokeJoinTokens)

	// ---- Host-facing: attendance reads and mutations ----
	// Previously optionalAuth — anyone with a session ID or attendee
	// ID could read attendance. Now authenticated and authorized.
	//
	// sessionCtx runs between authn and authz on each of these so the
	// Casbin check gets the correct account domain. It calls
	// ResolveSessionAccount on the attendance service; on not-found it
	// defers to the handler, which returns a proper 404.
	r.Get("/sessions/:id/attendance", authMiddleware, sessionCtx, authzMiddleware, h.Attendance.ListSessionAttendance)
	r.Get("/sessions/:id/attendance/export", authMiddleware, sessionCtx, authzMiddleware, h.Attendance.ExportSessionAttendance)

	r.Post("/sessions/:id/attendance/:attendeeID/confirm", authMiddleware, sessionCtx, authzMiddleware, h.Attendance.ConfirmAttendance)
	r.Post("/sessions/:id/attendance/:attendeeID/override", authMiddleware, sessionCtx, authzMiddleware, h.Attendance.OverrideAttendance)
	r.Post("/sessions/:id/attendance/bulk-confirm", authMiddleware, sessionCtx, authzMiddleware, h.Attendance.BulkConfirm)

	r.Get("/attendees/:id/summary", authMiddleware, authzMiddleware, h.Attendance.GetAttendeeSummary)

	// ---- Event-facing ----
	r.Get("/events/:eventId/summary", authMiddleware, authzMiddleware, h.Attendance.EventSummary)
	r.Get("/events/:eventId/attendees", authMiddleware, authzMiddleware, h.Attendance.ListEventAttendees)
	r.Get("/events/:eventId/attendees/:attendeeId", authMiddleware, authzMiddleware, h.Attendance.GetEventAttendeeDetail)
	r.Get("/attendees", authMiddleware, authzMiddleware, h.Attendance.ListAttendees)
}