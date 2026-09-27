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

// CreateMeeting handles POST /api/v1/events/:id/meeting.
//
// Creates a Zoom meeting for every virtual schedule that doesn't
// already have one. Idempotent — schedules with an existing meeting
// are skipped.
func (h *EventHandler) CreateMeeting(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CreateEventMeeting(ctx, id, userID)
	if err != nil {
		return respondClassifiedError(c, "create a meeting for this event", err)
	}

	return response.Success(c, "Meeting created successfully", NewEventResponseFromEventWithCreator(event))
}

// DeleteMeeting handles DELETE /api/v1/events/:id/meeting.
//
// Removes the Zoom meeting for every virtual schedule. The event
// itself is not affected.
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

	return response.Success(c, "Meeting deleted successfully", NewEventResponseFromEventWithCreator(event))
}

// RegenerateMeeting handles POST /api/v1/events/:id/meeting/regenerate.
//
// Deletes the current Zoom meeting and creates a fresh one with a new
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

	return response.Success(c, "Meeting regenerated successfully", NewEventResponseFromEventWithCreator(event))
}