// internal/modules/attendance/delivery/http/session_handler.go

package http

import (
	"time"

	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// SessionHandler exposes session and attendee management endpoints.
//
// These are consumer-facing: called by other modules (events,
// courses) via HTTP, or by the frontend on behalf of a host.
type SessionHandler struct {
	svc service.Service
}

func NewSessionHandler(svc service.Service) *SessionHandler {
	return &SessionHandler{svc: svc}
}

// RegisterAttendee handles POST /attendees.
func (h *SessionHandler) RegisterAttendee(c fiber.Ctx) error {
	var body RegisterAttendeeRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}

	attendee, err := h.svc.RegisterAttendee(c.Context(), service.RegisterAttendeeCommand{
		External:    attendance.ExternalRef{Type: body.ExternalType, ID: body.ExternalID},
		DisplayName: body.DisplayName,
		Email:       body.Email,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Attendee registered successfully", toAttendeeResponse(attendee))
}

// RegisterAttendeeForExternal handles
// POST /attendees/:id/register-for-external.
//
// Registers the attendee for every session under the external
// reference, issues a join token per session, and returns the
// resulting links. When the request carries a public_base_url, each
// link includes a full join URL.
func (h *SessionHandler) RegisterAttendeeForExternal(c fiber.Ctx) error {
	attendeeID := c.Params("id")
	if attendeeID == "" {
		return response.BadRequest(c, "Attendee ID is required", nil)
	}

	var body RegisterAttendeeForExternalRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}

	var grace time.Duration
	if body.LinkGrace != "" {
		d, err := time.ParseDuration(body.LinkGrace)
		if err != nil {
			return response.BadRequest(c, "Invalid link_grace duration", nil)
		}
		grace = d
	}

	result, err := h.svc.RegisterAttendeeForExternal(
		c.Context(),
		service.RegisterAttendeeForExternalCommand{
			AttendeeID:    attendeeID,
			External:      attendance.ExternalRef{Type: body.ExternalType, ID: body.ExternalID},
			PublicBaseURL: body.PublicBaseURL,
			LinkGrace:     grace,
		},
	)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(
		c,
		"Attendee registered for external reference",
		&RegisterAttendeeForExternalResponse{
			AttendeeID: result.AttendeeID,
			Links:      toAttendeeSessionLinksResponse(result.Links),
		},
	)
}

// UpsertSession handles POST /sessions.
func (h *SessionHandler) UpsertSession(c fiber.Ctx) error {
	var body UpsertSessionRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}

	provider := attendance.SessionProvider(body.Provider)
	if !provider.IsValid() {
		return response.BadRequest(c, "Invalid provider", nil)
	}

	session, err := h.svc.UpsertSession(c.Context(), service.UpsertSessionCommand{
		External:          attendance.ExternalRef{Type: body.ExternalType, ID: body.ExternalID},
		Title:             body.Title,
		ScheduledStart:    body.ScheduledStart,
		ScheduledEnd:      body.ScheduledEnd,
		Provider:          provider,
		ProviderMeetingID: body.ProviderMeetingID,
		ProviderURL:       body.ProviderURL,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Session upserted successfully", toSessionResponse(session))
}

// IssueJoinToken handles POST /sessions/:id/join-tokens.
//
// Returns the raw token, its expiry, and — when the caller passes a
// public_base_url in the request body — a full join URL.
func (h *SessionHandler) IssueJoinToken(c fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID is required", nil)
	}

	var body IssueJoinTokenRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}

	var grace time.Duration
	if body.Grace != "" {
		d, err := time.ParseDuration(body.Grace)
		if err != nil {
			return response.BadRequest(c, "Invalid grace duration", nil)
		}
		grace = d
	}

	result, err := h.svc.IssueJoinToken(c.Context(), service.IssueJoinTokenCommand{
		AttendeeID:    body.AttendeeID,
		SessionID:     sessionID,
		Grace:         grace,
		PublicBaseURL: body.PublicBaseURL,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Created(c, "Join token issued", &IssueJoinTokenResponse{
		RawToken:  result.RawToken,
		ExpiresAt: result.ExpiresAt,
		JoinURL:   result.JoinURL,
	})
}

// RevokeJoinTokens handles DELETE /sessions/:id/join-tokens.
func (h *SessionHandler) RevokeJoinTokens(c fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID is required", nil)
	}

	var body RevokeJoinTokensRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}

	err := h.svc.RevokeJoinTokens(c.Context(), service.RevokeJoinTokensCommand{
		AttendeeID: body.AttendeeID,
		SessionID:  sessionID,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Join tokens revoked", nil)
}