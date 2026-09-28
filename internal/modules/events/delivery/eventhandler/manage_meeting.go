// internal/modules/events/handler/manage_meeting.go

package eventhandler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// MEETING MANAGEMENT
// ============================================================

// CreateMeetingRequest is the optional request body for CreateMeeting.
//
// Platform is used as an override when a virtual schedule does not
// already have a platform set. Schedules that already carry a platform
// keep their own. An empty or missing platform falls back to the
// schedule's value, then to link-shape inference.
type CreateMeetingRequest struct {
	Platform string `json:"platform" validate:"omitempty,oneof=zoom google_meet"`
}

// CreateMeeting handles POST /api/v1/events/:id/meeting.
//
// Creates a meeting on the host's connected video platform for every
// virtual schedule that doesn't already have one. Idempotent —
// schedules with an existing meeting are skipped.
//
// The request body is optional:
//
//	{ "platform": "zoom" | "google_meet" }
//
// When provided, the platform acts as a fallback for schedules whose
// platform column is empty. When omitted, each schedule's stored
// platform is used, falling back to link-shape inference for legacy
// rows.
func (h *EventHandler) CreateMeeting(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Parse the optional body. An empty body is valid — it means
	// "use whatever platform each schedule already has."
	var req CreateMeetingRequest
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return response.BadRequest(c, "Invalid request body", nil)
		}
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CreateEventMeeting(ctx, id, userID, req.Platform)
	if err != nil {
		return respondClassifiedError(c, "create a meeting for this event", err)
	}

	return response.Success(
		c,
		"Meeting created successfully",
		NewEventResponseFromEventWithCreator(event),
	)
}

// DeleteMeeting handles DELETE /api/v1/events/:id/meeting.
//
// Removes the meeting for every virtual schedule. The event itself
// is not affected.
func (h *EventHandler) DeleteMeeting(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.DeleteEventMeeting(ctx, id, userID)
	if err != nil {
		return respondClassifiedError(c, "delete the meeting for this event", err)
	}

	return response.Success(
		c,
		"Meeting deleted successfully",
		NewEventResponseFromEventWithCreator(event),
	)
}

// RegenerateMeeting handles POST /api/v1/events/:id/meeting/regenerate.
//
// Deletes the current meeting and creates a fresh one with a new
// join link. Used for recovery when the meeting is broken or the link
// has leaked.
func (h *EventHandler) RegenerateMeeting(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.RegenerateEventMeeting(ctx, id, userID)
	if err != nil {
		return respondClassifiedError(c, "regenerate the meeting for this event", err)
	}

	return response.Success(
		c,
		"Meeting regenerated successfully",
		NewEventResponseFromEventWithCreator(event),
	)
}