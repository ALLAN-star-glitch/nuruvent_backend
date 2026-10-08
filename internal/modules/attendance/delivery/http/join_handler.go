// internal/modules/attendance/delivery/http/join_handler.go

package http

import (

	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
)

// JoinHandler handles the public join-link redemption endpoint.
//
// No authentication — the token is the credential. But when the token
// resolves to a user, we issue a short-lived session cookie so guests
// land on the meeting page authenticated (required for the Zoom SDK).
type JoinHandler struct {
	svc  service.Service
	auth attendance.AuthSessionIssuer
}

func NewJoinHandler(
	svc service.Service,
	auth attendance.AuthSessionIssuer,
) *JoinHandler {
	return &JoinHandler{svc: svc, auth: auth}
}

// RedeemJoinToken handles GET /join/:token.
//
// Flow:
//  1. Extract the raw token from the URL.
//  2. Redeem it — the service validates, records the join, and
//     returns the Nuruvent-hosted meeting URL to redirect to.
//  3. Trigger a status recompute for the affected session.
//  4. If the token carries a user and the caller has no session yet,
//     issue a short-lived auth cookie so the meeting page knows who
//     they are.
//  5. Redirect the browser to the meeting page.
//
// If anything goes wrong, render a small HTML page explaining the
// error, rather than a JSON error — the caller is a browser, not an
// API client.
func (h *JoinHandler) RedeemJoinToken(c fiber.Ctx) error {
	rawToken := c.Params("token")
	if rawToken == "" {
		return renderJoinError(c, "missing token", "This link is incomplete.")
	}

	result, err := h.svc.RedeemJoinToken(c.Context(), rawToken)
	if err != nil {
		return renderJoinError(c, friendlyJoinMessage(err), friendlyJoinDetail(err))
	}

	// Best-effort recompute. If this fails, a background job will
	// recompute; the join itself was recorded.
	_ = h.svc.RecomputeSessionStatuses(c.Context(), result.SessionID)

	// Issue a session cookie for the user behind the token — but only
	// when the caller doesn't already have one. Regular users who are
	// already logged in keep their existing session; guests get a new
	// short-lived one so the meeting page treats them as authenticated.
	if result.UserID != "" && c.Cookies("access_token") == "" && h.auth != nil {
		const guestTTL = 4 * time.Hour

		token, expiresAt, err := h.auth.IssueSessionForUser(
			c.Context(),
			result.UserID,
			guestTTL,
		)
		if err != nil {
			// Non-fatal — the meeting page still loads, but the SDK
			// may prompt for sign-in. Log and continue.
			log.Printf("[join] issue guest session failed user=%s err=%v",
				result.UserID, err)
		} else {
			setSessionCookie(c, token, expiresAt)
		}
	}

	// Redirect the browser to the meeting page. The page handles
	// platform-specific handoff.
	if result.RedirectTo == "" {
		return renderJoinError(c, "no redirect url", "The meeting URL is not available yet. Please contact the host.")
	}
	return c.Redirect().To(result.RedirectTo)
}

// ============================================================
// COOKIE HELPER
// ============================================================

// setSessionCookie writes the auth cookie in the same shape the auth
// handler uses, so the frontend and the middleware accept it without
// special-casing.
//
// Domain is left empty — the cookie is scoped to the exact host that
// served the join link. This matches what the auth handler does.
func setSessionCookie(c fiber.Ctx, token string, expiresAt time.Time) {
	isSecure := c.Protocol() == "https" ||
		c.Get("X-Forwarded-Proto") == "https"

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		Path:     "/",
		Domain:   "",
	})
}

// ============================================================
// ERROR RENDERING
// ============================================================

// renderJoinError renders a minimal HTML page for join failures.
// Browsers don't render JSON usefully.
func renderJoinError(c fiber.Ctx, title, detail string) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>` + title + `</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { font-family: system-ui, -apple-system, sans-serif; max-width: 520px; margin: 4rem auto; padding: 0 1rem; color: #202124; }
    h1 { font-size: 1.5rem; }
    p  { color: #5F6368; line-height: 1.6; }
    a  { color: #1A73E8; }
  </style>
</head>
<body>
  <h1>` + title + `</h1>
  <p>` + detail + `</p>
  <p><a href="/">Back to Nuruvent</a></p>
</body>
</html>`
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Status(fiber.StatusGone).SendString(html)
}

// friendlyJoinMessage maps a domain error to a short user-facing
// title.
func friendlyJoinMessage(err error) string {
	if err == nil {
		return "Something went wrong"
	}
	return "We couldn't process this link"
}

// friendlyJoinDetail maps a domain error to a longer user-facing
// explanation. The domain errors are already descriptive; we just
// avoid leaking internals.
func friendlyJoinDetail(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "expired"):
		return "This link has expired. Please contact the host for a new one."
	case strings.Contains(msg, "revoked"):
		return "This link is no longer valid. Please contact the host for a new one."
	case strings.Contains(msg, "not found"):
		return "This link doesn't match any registration. Please check that you copied it correctly."
	case strings.Contains(msg, "cancelled"):
		return "This session has been cancelled."
	default:
		return "The link may be invalid, or the session is no longer available."
	}
}