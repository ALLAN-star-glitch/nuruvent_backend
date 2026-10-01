// internal/modules/video/delivery/http/fetch_attendance_handler.go

package http

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// FetchAttendance handles POST /meetings/:id/fetch-attendance.
//
// Polls Google Meet for conference records and participants, and
// records attendance for matched attendees. Only supported for
// google_meet meetings.
func (h *MeetingSDKHandler) FetchAttendance(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	meetingID := c.Params("id")
	if meetingID == "" {
		return response.BadRequest(c, "Meeting ID is required", nil)
	}

	result, err := h.svc.FetchGoogleMeetAttendance(
		c.Context(),
		service.FetchGoogleMeetAttendanceCommand{
			UserID:    userID,
			MeetingID: meetingID,
		},
	)
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Attendance fetched", result)
}

// ============================================================
// LINK PARTICIPANT
// ============================================================

// LinkParticipantRequest is the body for
// POST /meetings/:id/link-participant.
type LinkParticipantRequest struct {
	AttendeeID       string `json:"attendee_id" binding:"required,uuid"`
	GoogleMeetUserID string `json:"google_meet_user_id" binding:"required"`
}

// LinkParticipantResponse summarizes the re-poll that follows a
// successful link.
type LinkParticipantResponse struct {
	MeetingNumber     string `json:"meeting_number"`
	ConferenceRecords int    `json:"conference_records"`
	Participants      int    `json:"participants"`
	EventsDispatched  int    `json:"events_dispatched"`
}

// LinkParticipant handles POST /meetings/:id/link-participant.
//
// Binds a Meet participant's Google user id to a registered attendee,
// then re-polls the meeting so the join is recorded. The response
// summarizes the re-poll — the roster UI refetches via its own tag
// invalidation, so we don't return the full result here.
func (h *MeetingSDKHandler) LinkParticipant(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	meetingID := c.Params("id")
	if meetingID == "" {
		return response.BadRequest(c, "Meeting ID is required", nil)
	}

	var body LinkParticipantRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := h.svc.LinkParticipant(c.Context(), service.LinkParticipantCommand{
		UserID:           userID,
		MeetingID:        meetingID,
		AttendeeID:       body.AttendeeID,
		GoogleMeetUserID: body.GoogleMeetUserID,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Participant linked", LinkParticipantResponse{
		MeetingNumber:     result.MeetingNumber,
		ConferenceRecords: result.ConferenceRecords,
		Participants:      result.Participants,
		EventsDispatched:  result.EventsDispatched,
	})
}

// ============================================================
// UNMATCHED PARTICIPANTS
// ============================================================

// UnmatchedParticipantResponse is the wire format for one unmatched
// Meet participant.
type UnmatchedParticipantResponse struct {
	GoogleMeetUserID string  `json:"google_meet_user_id"`
	DisplayName      string  `json:"display_name"`
	JoinedAt         string  `json:"joined_at"`
	LeftAt           *string `json:"left_at,omitempty"`
}

// GetUnmatchedParticipants handles
// GET /meetings/:id/unmatched-participants.
//
// Polls the meeting and returns any participants that couldn't be
// resolved to a registered attendee. Used by the roster's "link
// participant" UI.
func (h *MeetingSDKHandler) GetUnmatchedParticipants(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	meetingID := c.Params("id")
	if meetingID == "" {
		return response.BadRequest(c, "Meeting ID is required", nil)
	}

	result, err := h.svc.GetUnmatchedParticipants(
		c.Context(),
		service.GetUnmatchedParticipantsCommand{
			UserID:    userID,
			MeetingID: meetingID,
		},
	)
	if err != nil {
		return mapDomainError(c, err)
	}

	out := make([]UnmatchedParticipantResponse, 0, len(result))
	for _, p := range result {
		var left *string
		if p.LeftAt != nil {
			s := p.LeftAt.Format(time.RFC3339)
			left = &s
		}
		out = append(out, UnmatchedParticipantResponse{
			GoogleMeetUserID: p.GoogleMeetUserID,
			DisplayName:      p.DisplayName,
			JoinedAt:         p.JoinedAt.Format(time.RFC3339),
			LeftAt:           left,
		})
	}

	return response.Success(c, "Unmatched participants retrieved", out)
}