// internal/modules/attendance/delivery/http/attendance_handler.go

package http

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// AttendanceHandler exposes host-facing operations and reporting.
type AttendanceHandler struct {
	svc service.Service
}

func NewAttendanceHandler(svc service.Service) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

// ListSessionAttendance handles GET /sessions/:id/attendance.
func (h *AttendanceHandler) ListSessionAttendance(c fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID is required", nil)
	}

	statuses, err := h.svc.ListSessionAttendance(c.Context(), sessionID)
	if err != nil {
		return mapDomainError(c, err)
	}

	out := make([]*SessionStatusResponse, 0, len(statuses))
	for _, s := range statuses {
		out = append(out, toSessionStatusResponse(s))
	}

	return response.Success(c, "Attendance retrieved successfully", fiber.Map{
		"session_id": sessionID,
		"count":      len(out),
		"attendees":  out,
	})
}

// ConfirmAttendance handles
// POST /sessions/:id/attendance/:attendeeID/confirm.
func (h *AttendanceHandler) ConfirmAttendance(c fiber.Ctx) error {
	sessionID := c.Params("id")
	attendeeID := c.Params("attendeeID")
	if sessionID == "" || attendeeID == "" {
		return response.BadRequest(c, "Session ID and Attendee ID are required", nil)
	}

	actorID := handlerhelper.GetUserIDOptional(c)
	if actorID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body ConfirmAttendanceRequest
	_ = c.Bind().Body(&body) // body is optional

	err := h.svc.ConfirmAttendance(c.Context(), service.ConfirmAttendanceCommand{
		AttendeeID: attendeeID,
		SessionID:  sessionID,
		ActorID:    actorID,
		Reason:     body.Reason,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Attendance confirmed", nil)
}

// OverrideAttendance handles
// POST /sessions/:id/attendance/:attendeeID/override.
func (h *AttendanceHandler) OverrideAttendance(c fiber.Ctx) error {
	sessionID := c.Params("id")
	attendeeID := c.Params("attendeeID")
	if sessionID == "" || attendeeID == "" {
		return response.BadRequest(c, "Session ID and Attendee ID are required", nil)
	}

	actorID := handlerhelper.GetUserIDOptional(c)
	if actorID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body OverrideAttendanceRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}

	newStatus := attendance.AttendanceStatus(body.NewStatus)
	if !newStatus.IsValid() {
		return response.BadRequest(c, "Invalid new_status", nil)
	}

	err := h.svc.OverrideAttendance(c.Context(), service.OverrideAttendanceCommand{
		AttendeeID: attendeeID,
		SessionID:  sessionID,
		ActorID:    actorID,
		NewStatus:  newStatus,
		Reason:     body.Reason,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Attendance overridden", nil)
}

// BulkConfirm handles POST /sessions/:id/attendance/bulk-confirm.
func (h *AttendanceHandler) BulkConfirm(c fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID is required", nil)
	}

	actorID := handlerhelper.GetUserIDOptional(c)
	if actorID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body BulkConfirmRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{"error": err.Error()})
	}
	if len(body.AttendeeIDs) == 0 {
		return response.BadRequest(c, "attendee_ids must not be empty", nil)
	}

	err := h.svc.BulkConfirm(c.Context(), service.BulkConfirmCommand{
		SessionID:   sessionID,
		AttendeeIDs: body.AttendeeIDs,
		ActorID:     actorID,
		Reason:      body.Reason,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Attendance bulk-confirmed", fiber.Map{
		"confirmed_count": len(body.AttendeeIDs),
	})
}

// ExportSessionAttendance handles
// GET /sessions/:id/attendance/export.
func (h *AttendanceHandler) ExportSessionAttendance(c fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return response.BadRequest(c, "Session ID is required", nil)
	}

	csv, err := h.svc.ExportSessionAttendance(c.Context(), sessionID)
	if err != nil {
		return mapDomainError(c, err)
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=\"attendance-"+sessionID+".csv\"")
	return c.Send(csv)
}

// GetAttendeeSummary handles GET /attendees/:id/summary.
func (h *AttendanceHandler) GetAttendeeSummary(c fiber.Ctx) error {
	attendeeID := c.Params("id")
	if attendeeID == "" {
		return response.BadRequest(c, "Attendee ID is required", nil)
	}

	summary, err := h.svc.GetAttendeeSummary(c.Context(), attendeeID)
	if err != nil {
		return mapDomainError(c, err)
	}

	resp := &AttendeeSummaryResponse{
		Attendee: toAttendeeResponse(summary.Attendee),
		Statuses: make([]*SessionStatusResponse, 0, len(summary.Statuses)),
		Rollups:  make([]*RollupStatusResponse, 0, len(summary.Rollups)),
	}
	for _, s := range summary.Statuses {
		resp.Statuses = append(resp.Statuses, toSessionStatusResponse(s))
	}
	for _, r := range summary.Rollups {
		resp.Rollups = append(resp.Rollups, toRollupStatusResponse(r))
	}

	return response.Success(c, "Attendee summary retrieved", resp)
}

// EventSummary handles GET /events/:eventId/summary.
//
// Returns a per-session and total attendance summary for every
// session under the event. Requires authentication.
//
// TODO: verify the caller has read access to the event. Currently
// any authenticated user can request the summary by event UUID.
func (h *AttendanceHandler) EventSummary(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	summary, err := h.svc.GetEventAttendanceSummary(
		c.Context(),
		service.GetEventAttendanceSummaryCommand{
			UserID:  userID,
			EventID: eventID,
		},
	)
	if err != nil {
		return mapDomainError(c, err)
	}

	out := &EventAttendanceSummaryResponse{
		Sessions: make([]SessionAttendanceSummaryResponse, 0, len(summary.Sessions)),
		Totals: EventAttendanceTotalsResponse{
			TotalSessions:          summary.Totals.TotalSessions,
			SessionsWithAttendance: summary.Totals.SessionsWithAttendance,
			UniqueAttendees:        summary.Totals.UniqueAttendees,
			TotalAttendanceEvents:  summary.Totals.TotalAttendanceEvents,
		},
	}
	for _, s := range summary.Sessions {
		out.Sessions = append(out.Sessions, SessionAttendanceSummaryResponse{
			SessionID:          s.SessionID,
			Title:              s.Title,
			Provider:           s.Provider,
			ScheduledStart:     s.ScheduledStart.Format(time.RFC3339),
			ScheduledEnd:       s.ScheduledEnd.Format(time.RFC3339),
			RegisteredCount:    s.RegisteredCount,
			AttendedCount:      s.AttendedCount,
			AvgDurationSeconds: s.AvgDurationSeconds,
			HasAttendanceData:  s.HasAttendanceData,
			VideoMeetingID:     s.VideoMeetingID,
		})
	}

	return response.Success(c, "Attendance summary retrieved", out)
}

// ============================================================
// EVENT ATTENDEE DIRECTORY
// ============================================================

// ListEventAttendees handles GET /events/:eventId/attendees.
func (h *AttendanceHandler) ListEventAttendees(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var statuses []string
	if raw := c.Query("status"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(s); t != "" {
				statuses = append(statuses, t)
			}
		}
	}

	res, err := h.svc.ListEventAttendees(c.Context(), service.ListEventAttendeesCommand{
		UserID:    userID,
		EventID:   eventID,
		Search:    c.Query("search"),
		Statuses:  statuses,
		SortBy:    c.Query("sort_by", "name"),
		SortOrder: c.Query("sort_order", "asc"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	out := EventAttendeesListResponse{
		Attendees: make([]EventAttendeeResponse, 0, len(res.Attendees)),
		Total:     res.Total,
		Page:      res.Page,
		PageSize:  res.PageSize,
	}
	for _, a := range res.Attendees {
		out.Attendees = append(out.Attendees, toEventAttendeeResponse(a))
	}

	return response.Success(c, "Attendees retrieved successfully", out)
}

// GetEventAttendeeDetail handles GET /events/:eventId/attendees/:attendeeId.
func (h *AttendanceHandler) GetEventAttendeeDetail(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	eventID := c.Params("eventId")
	attendeeID := c.Params("attendeeId")
	if eventID == "" || attendeeID == "" {
		return response.BadRequest(c, "Event ID and Attendee ID are required", nil)
	}

	detail, err := h.svc.GetEventAttendeeDetail(c.Context(), service.GetEventAttendeeDetailCommand{
		UserID:     userID,
		EventID:    eventID,
		AttendeeID: attendeeID,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Attendee retrieved successfully", toEventAttendeeDetailResponse(detail))
}




// ============================================================
// EVENT ATTENDEE DIRECTORY — mappers
// ============================================================

func toEventAttendeeResponse(a *service.EventAttendeeListItem) EventAttendeeResponse {
	return EventAttendeeResponse{
		AttendeeID:           a.AttendeeID,
		DisplayName:          a.DisplayName,
		Email:                a.Email,
		EffectiveStatus:      string(a.EffectiveStatus),
		SessionsTotal:        a.SessionsTotal,
		SessionsAttended:     a.SessionsAttended,
		SessionsConfirmed:    a.SessionsConfirmed,
		TotalDurationSeconds: int64(a.TotalDurationSeconds),
		RegisteredAt:         a.RegisteredAt.Format(time.RFC3339),
		LastActivityAt:       a.LastActivityAt.Format(time.RFC3339),
	}
}

func toEventAttendeeDetailResponse(d *service.EventAttendeeDetail) EventAttendeeDetailResponse {
	out := EventAttendeeDetailResponse{
		EventAttendeeResponse: EventAttendeeResponse{
			AttendeeID:           d.AttendeeID,
			DisplayName:          d.DisplayName,
			Email:                d.Email,
			EffectiveStatus:      string(d.EffectiveStatus),
			SessionsTotal:        d.SessionsTotal,
			SessionsAttended:     d.SessionsAttended,
			SessionsConfirmed:    d.SessionsConfirmed,
			TotalDurationSeconds: int64(d.TotalDurationSeconds),
			RegisteredAt:         d.RegisteredAt.Format(time.RFC3339),
			LastActivityAt:       d.LastActivityAt.Format(time.RFC3339),
		},
		Sessions: make([]EventAttendeeSessionResponse, 0, len(d.Sessions)),
	}
	for _, s := range d.Sessions {
		out.Sessions = append(out.Sessions, EventAttendeeSessionResponse{
			SessionID:        s.SessionID,
			Title:            s.Title,
			Provider:         string(s.Provider),
			ScheduledStart:   s.ScheduledStart.Format(time.RFC3339),
			ScheduledEnd:     s.ScheduledEnd.Format(time.RFC3339),
			DerivedStatus:    string(s.DerivedStatus),
			HostConfirmed:    s.HostConfirmed,
			TotalDurationSec: int64(s.TotalDurationSec),
			LastDerivedAt:    s.LastDerivedAt.Format(time.RFC3339),
		})
	}
	return out
}



// ListAttendees handles GET /attendees.
//
// Cross-event attendee directory. Scoped to the caller's accounts.
func (h *AttendanceHandler) ListAttendees(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	var statuses []string
	if raw := c.Query("status"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(s); t != "" {
				statuses = append(statuses, t)
			}
		}
	}

	res, err := h.svc.ListAttendees(c.Context(), service.ListAttendeesCommand{
		UserID:    userID,
		EventID:   c.Query("event_id"),
		Search:    c.Query("search"),
		Statuses:  statuses,
		SortBy:    c.Query("sort_by", "registered_at"),
		SortOrder: c.Query("sort_order", "desc"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	out := CrossEventAttendeesListResponse{
		Attendees: make([]CrossEventAttendeeResponse, 0, len(res.Attendees)),
		Total:     res.Total,
		Page:      res.Page,
		PageSize:  res.PageSize,
	}
	for _, a := range res.Attendees {
		out.Attendees = append(out.Attendees, toCrossEventAttendeeResponse(a))
	}

	return response.Success(c, "Attendees retrieved successfully", out)
}

// ---- DTOs ----

type CrossEventAttendeesListResponse struct {
	Attendees []CrossEventAttendeeResponse `json:"attendees"`
	Total     int                          `json:"total"`
	Page      int                          `json:"page"`
	PageSize  int                          `json:"page_size"`
}

type CrossEventAttendeeResponse struct {
	AttendeeID           string `json:"attendee_id"`
	DisplayName          string `json:"display_name"`
	Email                string `json:"email"`
	EventID              string `json:"event_id"`
	EventName            string `json:"event_name"`
	EventSlug            string `json:"event_slug"`
	EventStartDate       string `json:"event_start_date"`
	EffectiveStatus      string `json:"effective_status"`
	SessionsTotal        int    `json:"sessions_total"`
	SessionsAttended     int    `json:"sessions_attended"`
	SessionsConfirmed    int    `json:"sessions_confirmed"`
	TotalDurationSeconds int64  `json:"total_duration_seconds"`
	RegisteredAt         string `json:"registered_at"`
	LastActivityAt       string `json:"last_activity_at"`
}

// ---- Mapper ----

func toCrossEventAttendeeResponse(a *service.CrossEventAttendeeItem) CrossEventAttendeeResponse {
	return CrossEventAttendeeResponse{
		AttendeeID:           a.AttendeeID,
		DisplayName:          a.DisplayName,
		Email:                a.Email,
		EventID:              a.EventID,
		EventName:            a.EventName,
		EventSlug:            a.EventSlug,
		EventStartDate:       a.EventStartDate.Format(time.RFC3339),
		EffectiveStatus:      string(a.EffectiveStatus),
		SessionsTotal:        a.SessionsTotal,
		SessionsAttended:     a.SessionsAttended,
		SessionsConfirmed:    a.SessionsConfirmed,
		TotalDurationSeconds: a.TotalDurationSeconds,
		RegisteredAt:         a.RegisteredAt.Format(time.RFC3339),
		LastActivityAt:       a.LastActivityAt.Format(time.RFC3339),
	}
}