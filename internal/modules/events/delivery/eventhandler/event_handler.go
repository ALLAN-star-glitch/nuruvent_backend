// internal/modules/events/handler/eventhandler.go

package eventhandler

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// EVENT HANDLER
// ============================================================

type EventHandler struct {
	svc service.Service
}

func NewEventHandler(svc service.Service) *EventHandler {
	return &EventHandler{
		svc: svc,
	}
}

// ============================================================
// HELPERS
// ============================================================

func getQueryInt(c fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getQueryString(c fiber.Ctx, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func getUserID(c fiber.Ctx) (string, error) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", errors.New("user not authenticated")
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user ID")
	}
	return userIDStr, nil
}

// ============================================================
// PUBLIC HANDLERS
// ============================================================

// GetEvent godoc
// @Summary Get event by ID
// @Description Get event details by ID (public)
// @Tags Events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	event, err := h.svc.GetEventByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to get event", fiber.Map{
			"error": err.Error(),
		})
	}

	if event == nil {
		return response.NotFound(c, "Event not found", nil)
	}

	return response.Success(c, "Event retrieved successfully", event)
}

// GetEventBySlug godoc
// @Summary Get event by slug
// @Description Get event details by slug (public)
// @Tags Events
// @Produce json
// @Param slug path string true "Event Slug"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/slug/{slug} [get]
func (h *EventHandler) GetEventBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Event slug is required", nil)
	}

	event, err := h.svc.GetEventBySlug(c.Context(), slug)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to get event", fiber.Map{
			"error": err.Error(),
		})
	}

	if event == nil {
		return response.NotFound(c, "Event not found", nil)
	}

	return response.Success(c, "Event retrieved successfully", event)
}

// ListEvents godoc
// @Summary List events
// @Description List events with filters
// @Tags Events
// @Produce json
// @Param account_id query string false "Account ID"
// @Param event_type_id query string false "Event Type ID"
// @Param event_status_id query string false "Event Status ID"
// @Param include_deleted query bool false "Include soft-deleted events"
//@Param only_deleted query bool false "Show ONLY soft-deleted events"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events [get]
func (h *EventHandler) ListEvents(c fiber.Ctx) error {
	accountID := getQueryString(c, "account_id", "")
	eventTypeID := getQueryString(c, "event_type_id", "")
	eventStatusID := getQueryString(c, "event_status_id", "")
	includeDeleted := c.Query("include_deleted") == "true"
	onlyDeleted := c.Query("only_deleted") == "true"
	limit := getQueryInt(c, "limit", 20)
	offset := getQueryInt(c, "offset", 0)

	// ✅ Debug: Log the incoming parameters
    fmt.Printf("DEBUG Handler: onlyDeleted=%v, accountID=%s, limit=%d, offset=%d\n", onlyDeleted, accountID, limit, offset)

	if limit > 100 {
		limit = 100
	}

	filters := service.ListEventsFilters{
		AccountID:      accountID,
		EventTypeID:    eventTypeID,
		EventStatusID:  eventStatusID,
		IncludeDeleted: includeDeleted,
		OnlyDeleted:    onlyDeleted,
		Limit:          limit,
		Offset:         offset,
	}

	events, total, err := h.svc.ListEvents(c.Context(), filters)
	if err != nil {
		return response.InternalError(c, "Failed to list events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":   events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEventsByType godoc
// @Summary Get events by type
// @Description Get all events of a specific type
// @Tags Events
// @Produce json
// @Param type path string true "Event Type" Enums(workshop, webinar, meetup, bootcamp)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/type/{type} [get]
func (h *EventHandler) GetEventsByType(c fiber.Ctx) error {
	eventType := c.Params("type")
	if eventType == "" {
		return response.BadRequest(c, "Event type is required", nil)
	}

	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := h.svc.GetEventsByType(c.Context(), eventType, page, pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to get events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        events,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetUpcomingEvents godoc
// @Summary Get upcoming events
// @Description Get all upcoming published events
// @Tags Events
// @Produce json
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} response.BaseResponse{data=[]domain.Event}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/upcoming [get]
func (h *EventHandler) GetUpcomingEvents(c fiber.Ctx) error {
	limit := getQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	events, err := h.svc.GetUpcomingEvents(c.Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get upcoming events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Upcoming events retrieved successfully", events)
}

// GetPastEvents godoc
// @Summary Get past events
// @Description Get all past published events
// @Tags Events
// @Produce json
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} response.BaseResponse{data=[]domain.Event}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/past [get]
func (h *EventHandler) GetPastEvents(c fiber.Ctx) error {
	limit := getQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	events, err := h.svc.GetPastEvents(c.Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get past events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Past events retrieved successfully", events)
}

// GetEventTypes godoc
// @Summary Get all event types
// @Description Get list of all event types
// @Tags Events
// @Produce json
// @Success 200 {object} response.BaseResponse{data=[]domain.EventType}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/types [get]
func (h *EventHandler) GetEventTypes(c fiber.Ctx) error {
	types, err := h.svc.GetEventTypes(c.Context())
	if err != nil {
		return response.InternalError(c, "Failed to get event types", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event types retrieved successfully", types)
}

// GetEventStatuses godoc
// @Summary Get all event statuses
// @Description Get list of all event statuses
// @Tags Events
// @Produce json
// @Success 200 {object} response.BaseResponse{data=[]domain.EventStatus}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/statuses [get]
func (h *EventHandler) GetEventStatuses(c fiber.Ctx) error {
	statuses, err := h.svc.GetEventStatuses(c.Context())
	if err != nil {
		return response.InternalError(c, "Failed to get event statuses", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event statuses retrieved successfully", statuses)
}

// SearchEvents godoc
// @Summary Search events
// @Description Search events by title or description
// @Tags Events
// @Produce json
// @Param q query string false "Search query"
// @Param account_id query string false "Account ID"
// @Param event_type_id query string false "Event Type ID"
// @Param include_deleted query bool false "Include soft-deleted events"
// @Param only_deleted query bool false "Show ONLY soft-deleted events"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/search [get]
func (h *EventHandler) SearchEvents(c fiber.Ctx) error {
	query := getQueryString(c, "q", "")
	accountID := getQueryString(c, "account_id", "")
	eventTypeID := getQueryString(c, "event_type_id", "")
	includeDeleted := c.Query("include_deleted") == "true"
	onlyDeleted := c.Query("only_deleted") == "true"
	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)

	if pageSize > 100 {
		pageSize = 100
	}

	filters := service.SearchFilters{
		AccountID:      accountID,
		EventTypeID:    eventTypeID,
		IncludeDeleted: includeDeleted,
		OnlyDeleted:    onlyDeleted,
		Limit:          pageSize,
		Offset:         (page - 1) * pageSize,
	}

	events, total, err := h.svc.SearchEvents(c.Context(), query, filters)
	if err != nil {
		return response.InternalError(c, "Failed to search events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        events,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ============================================================
// PROTECTED HANDLERS (Auth Required)
// ============================================================

// GetEventsByAccount godoc
// @Summary Get events by account
// @Description Get all events for an account
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events [get]
func (h *EventHandler) GetEventsByAccount(c fiber.Ctx) error {
	accountID := c.Params("accountId")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := h.svc.GetEventsByAccount(c.Context(), accountID, page, pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to get events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        events,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ============================================================
// CREATE DRAFT
// ============================================================

// CreateDraft godoc
// @Summary Create a draft event
// @Description Create a draft event with minimal validation (all fields optional)
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param name formData string false "Event name" default(Untitled Event)
// @Param description formData string false "Event description"
// @Param event_type_id formData string false "Event Type ID"
// @Param date formData string false "Event date (YYYY-MM-DD)"
// @Param time formData string false "Event time (HH:MM)"
// @Param duration formData int false "Event duration in minutes" default(60)
// @Param price formData number false "Event price"
// @Param certificate_price formData number false "Certificate price"
// @Param location formData string false "Event location"
// @Param is_virtual formData bool false "Is virtual event" default(true)
// @Param zoom_link formData string false "Zoom link"
// @Param meet_link formData string false "Meet link"
// @Param max_attendees formData int false "Maximum attendees"
// @Param image formData file false "Event image"
// @Success 201 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/draft [post]
func (h *EventHandler) CreateDraft(c fiber.Ctx) error {
	accountID := c.Params("accountId")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreateDraftRequest
	if err := c.Bind().Form(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// Get image file if provided (optional)
	var imageData []byte
	var imageName string
	var contentType string

	file, err := c.FormFile("image")
	if err == nil && file != nil {
		fileContent, err := file.Open()
		if err != nil {
			return response.InternalError(c, "Failed to open image file", nil)
		}
		defer fileContent.Close()

		imageData, err = io.ReadAll(fileContent)
		if err != nil {
			return response.InternalError(c, "Failed to read image file", nil)
		}
		imageName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	// ✅ Build command - user only types Name
	// Service will generate display_name and slug from this
	cmd := service.CreateDraftCommand{
		Name:             req.Name,        // ✅ User input - what they typed
		Description:      req.Description,
		EventTypeID:      req.EventTypeID,
		AccountID:        accountID,
		CreatedBy:        userID,
		Date:             req.Date,
		Time:             req.Time,
		Duration:         req.Duration,
		Price:            req.Price,
		CertificatePrice: req.CertificatePrice,
		Location:         req.Location,
		IsVirtual:        req.IsVirtual,
		ZoomLink:         req.ZoomLink,
		MeetLink:         req.MeetLink,
		MaxAttendees:     req.MaxAttendees,
		IsFeatured:       req.IsFeatured,
		IsPrivate:        req.IsPrivate,
		ImageData:        imageData,
		ImageName:        imageName,
		ContentType:      contentType,
	}

	event, err := h.svc.CreateDraft(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventStatusNotFound) {
			return response.BadRequest(c, "Invalid event status", nil)
		}
		if errors.Is(err, domain.ErrEventTypeNotFound) {
			return response.BadRequest(c, "Invalid event type", nil)
		}
		return response.InternalError(c, "Failed to create draft", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Draft created successfully", event)
}

// ============================================================
// CREATE PUBLISHED EVENT
// ============================================================

// CreateEvent godoc
// @Summary Create a published event
// @Description Create a published event with full validation (all fields required)
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param name formData string true "Event name"
// @Param description formData string false "Event description"
// @Param event_type_id formData string true "Event Type ID"
// @Param date formData string true "Event date (YYYY-MM-DD)"
// @Param time formData string true "Event time (HH:MM)"
// @Param duration formData int true "Event duration in minutes"
// @Param price formData number false "Event price"
// @Param certificate_price formData number false "Certificate price"
// @Param location formData string false "Event location (required for in-person)"
// @Param is_virtual formData bool false "Is virtual event" default(true)
// @Param is_featured formData bool false "Is featured event"
// @Param is_private formData bool false "Is private event"
// @Param zoom_link formData string false "Zoom link (required for virtual)"
// @Param meet_link formData string false "Meet link (required for virtual)"
// @Param max_attendees formData int false "Maximum attendees"
// @Param image formData file false "Event image"
// @Success 201 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events [post]
func (h *EventHandler) CreateEvent(c fiber.Ctx) error {
	accountID := c.Params("accountId")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreateEventWithImageRequest
	if err := c.Bind().Form(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// Strict validation for published events
	if req.Name == "" {
		return response.BadRequest(c, "Event name is required", nil)
	}
	if req.EventTypeID == "" {
		return response.BadRequest(c, "Event type is required", nil)
	}
	if req.Date == "" {
		return response.BadRequest(c, "Event date is required", nil)
	}
	if req.Time == "" {
		return response.BadRequest(c, "Event time is required", nil)
	}
	if req.Duration <= 0 {
		return response.BadRequest(c, "Event duration must be greater than 0", nil)
	}
	if req.Duration < 15 {
		return response.BadRequest(c, "Duration must be at least 15 minutes", nil)
	}
	if !req.IsVirtual && req.Location == "" {
		return response.BadRequest(c, "Location is required for in-person events", nil)
	}
	if req.IsVirtual && req.ZoomLink == "" && req.MeetLink == "" {
		return response.BadRequest(c, "At least one meeting link is required for virtual events", nil)
	}

	// Get image file if provided (optional)
	var imageData []byte
	var imageName string
	var contentType string

	file, err := c.FormFile("image")
	if err == nil && file != nil {
		fileContent, err := file.Open()
		if err != nil {
			return response.InternalError(c, "Failed to open image file", nil)
		}
		defer fileContent.Close()

		imageData, err = io.ReadAll(fileContent)
		if err != nil {
			return response.InternalError(c, "Failed to read image file", nil)
		}
		imageName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	// ✅ Build command - user only types Name
	cmd := service.CreateEventCommand{
		Name:             req.Name,        // ✅ User input - what they typed
		Description:      req.Description,
		EventTypeID:      req.EventTypeID,
		AccountID:        accountID,
		CreatedBy:        userID,
		Date:             req.Date,
		Time:             req.Time,
		Duration:         req.Duration,
		Price:            req.Price,
		CertificatePrice: req.CertificatePrice,
		Location:         req.Location,
		IsVirtual:        req.IsVirtual,
		ZoomLink:         req.ZoomLink,
		MeetLink:         req.MeetLink,
		MaxAttendees:     req.MaxAttendees,
		IsFeatured:       req.IsFeatured,
		IsPrivate:        req.IsPrivate,
		ImageData:        imageData,
		ImageName:        imageName,
		ContentType:      contentType,
	}

	event, err := h.svc.CreateEvent(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventStatusNotFound) {
			return response.BadRequest(c, "Invalid event status", nil)
		}
		if errors.Is(err, domain.ErrEventTypeNotFound) {
			return response.BadRequest(c, "Invalid event type", nil)
		}
		return response.InternalError(c, "Failed to create event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Event created successfully", event)
}

// ============================================================
// UPDATE EVENT
// ============================================================

// UpdateEvent godoc
// @Summary Update an event
// @Description Update event details
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body UpdateEventRequest true "Event update details"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id} [put]
func (h *EventHandler) UpdateEvent(c fiber.Ctx) error {
    id := c.Params("id")
    if id == "" {
        return response.BadRequest(c, "Event ID is required", nil)
    }

    userID, err := getUserID(c)
    if err != nil {
        return response.Unauthorized(c, "User not authenticated", nil)
    }

    var req UpdateEventRequest
    if err := c.Bind().Body(&req); err != nil {
        return response.BadRequest(c, "Invalid request", fiber.Map{
            "error": err.Error(),
        })
    }

    cmd := service.UpdateEventCommand{
        ID:               id,
        UpdatedBy:        userID,
        Name:             req.Name,
        Description:      req.Description,
        EventTypeID:      req.EventTypeID,
        EventStatusID:    req.EventStatusID,
        Date:             req.Date,
        Time:             req.Time,
        Duration:         req.Duration,
        Price:            req.Price,
        CertificatePrice: req.CertificatePrice,
        Location:         req.Location,
        IsVirtual:        req.IsVirtual,
        ZoomLink:         req.ZoomLink,
        MeetLink:         req.MeetLink,
        MaxAttendees:     req.MaxAttendees,
        IsFeatured:       req.IsFeatured,  // ✅ Pass pointer
        IsPrivate:        req.IsPrivate,   // ✅ Pass pointer
    }

    event, err := h.svc.UpdateEvent(c.Context(), cmd)
    if err != nil {
        if errors.Is(err, domain.ErrEventNotFound) {
            return response.NotFound(c, "Event not found", nil)
        }
        return response.InternalError(c, "Failed to update event", fiber.Map{
            "error": err.Error(),
        })
    }

    return response.Success(c, "Event updated successfully", event)
}

// ============================================================
// DELETE - Single Event
// ============================================================

// DeleteEvent godoc
// @Summary Soft delete an event
// @Description Soft delete an event (sets deleted_at timestamp)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id} [delete]
func (h *EventHandler) DeleteEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	if err := h.svc.DeleteEvent(c.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to delete event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event deleted successfully", nil)
}

// PermanentlyDeleteEvent godoc
// @Summary Permanently delete an event
// @Description Permanently delete an event (hard delete - removes from database)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/permanent [delete]
func (h *EventHandler) PermanentlyDeleteEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	if err := h.svc.PermanentlyDeleteEvent(c.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to permanently delete event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event permanently deleted successfully", nil)
}

// RestoreEvent godoc
// @Summary Restore a soft-deleted event
// @Description Restore an event that was soft-deleted
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/restore [post]
func (h *EventHandler) RestoreEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	event, err := h.svc.RestoreEvent(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to restore event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event restored successfully", event)
}

// ============================================================
// DELETE - Bulk Events
// ============================================================

// BulkDeleteEvents godoc
// @Summary Soft delete multiple events
// @Description Soft delete multiple events by IDs
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkIDsRequest true "Event IDs to delete"
// @Success 200 {object} response.BaseResponse{data=service.BulkDeleteResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk [delete]
func (h *EventHandler) BulkDeleteEvents(c fiber.Ctx) error {
	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be deleted at once", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	result, err := h.svc.DeleteEvents(c.Context(), req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to delete events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events deleted successfully", result)
}

// BulkPermanentlyDeleteEvents godoc
// @Summary Permanently delete multiple events
// @Description Hard delete multiple events by IDs
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkIDsRequest true "Event IDs to permanently delete"
// @Success 200 {object} response.BaseResponse{data=service.BulkDeleteResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/permanent [delete]
func (h *EventHandler) BulkPermanentlyDeleteEvents(c fiber.Ctx) error {
	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be deleted at once", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	result, err := h.svc.PermanentlyDeleteEvents(c.Context(), req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to permanently delete events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events permanently deleted successfully", result)
}

// BulkRestoreEvents godoc
// @Summary Restore multiple soft-deleted events
// @Description Restore multiple events by IDs
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkIDsRequest true "Event IDs to restore"
// @Success 200 {object} response.BaseResponse{data=service.BulkRestoreResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/restore [post]
func (h *EventHandler) BulkRestoreEvents(c fiber.Ctx) error {
	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be restored at once", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	result, err := h.svc.RestoreEvents(c.Context(), req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to restore events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events restored successfully", result)
}

// ============================================================
// STATUS - Single
// ============================================================

// PublishEvent godoc
// @Summary Publish an event
// @Description Publish an event to make it public
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/publish [post]
func (h *EventHandler) PublishEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	event, err := h.svc.PublishEvent(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to publish event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event published successfully", event)
}

// CancelEvent godoc
// @Summary Cancel an event
// @Description Cancel an event
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/cancel [post]
func (h *EventHandler) CancelEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	event, err := h.svc.CancelEvent(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to cancel event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event cancelled successfully", event)
}

// CompleteEvent godoc
// @Summary Complete an event
// @Description Mark an event as completed
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/complete [post]
func (h *EventHandler) CompleteEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	event, err := h.svc.CompleteEvent(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to complete event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event completed successfully", event)
}

// ============================================================
// STATUS - Bulk
// ============================================================

// BulkPublishEvents godoc
// @Summary Publish multiple events
// @Description Publish multiple events by IDs
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkIDsRequest true "Event IDs to publish"
// @Success 200 {object} response.BaseResponse{data=service.BulkStatusResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/publish [post]
func (h *EventHandler) BulkPublishEvents(c fiber.Ctx) error {
	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be published at once", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	result, err := h.svc.BulkPublishEvents(c.Context(), req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to publish events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events published successfully", result)
}

// BulkCancelEvents godoc
// @Summary Cancel multiple events
// @Description Cancel multiple events by IDs
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkIDsRequest true "Event IDs to cancel"
// @Success 200 {object} response.BaseResponse{data=service.BulkStatusResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/cancel [post]
func (h *EventHandler) BulkCancelEvents(c fiber.Ctx) error {
	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be cancelled at once", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	result, err := h.svc.BulkCancelEvents(c.Context(), req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to cancel events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events cancelled successfully", result)
}

// BulkCompleteEvents godoc
// @Summary Complete multiple events
// @Description Complete multiple events by IDs
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkIDsRequest true "Event IDs to complete"
// @Success 200 {object} response.BaseResponse{data=service.BulkStatusResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/complete [post]
func (h *EventHandler) BulkCompleteEvents(c fiber.Ctx) error {
	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be completed at once", nil)
	}

	result, err := h.svc.BulkCompleteEvents(c.Context(), req.IDs)
	if err != nil {
		return response.InternalError(c, "Failed to complete events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events completed successfully", result)
}

// ============================================================
// DUPLICATE - Single Event
// ============================================================

// DuplicateEvent godoc
// @Summary Duplicate an event
// @Description Create a copy of an existing event (always creates as draft)
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param request body DuplicateEventRequest false "Duplicate options"
// @Success 200 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/duplicate [post]
func (h *EventHandler) DuplicateEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	var req DuplicateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	cmd := service.DuplicateEventCommand{
		Name:    req.Name,
		Date:    req.Date,
		IsDraft: req.IsDraft,
	}

	event, err := h.svc.DuplicateEvent(c.Context(), id, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to duplicate event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event duplicated successfully", event)
}

// ============================================================
// DUPLICATE - Bulk Events
// ============================================================

// BulkDuplicateEvents godoc
// @Summary Duplicate multiple events
// @Description Create copies of multiple existing events
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkDuplicateRequest true "Events to duplicate"
// @Success 200 {object} response.BaseResponse{data=service.BulkDuplicateResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/duplicate [post]
func (h *EventHandler) BulkDuplicateEvents(c fiber.Ctx) error {
	var req BulkDuplicateRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be duplicated at once", nil)
	}

	cmd := service.BulkDuplicateCommand{
		NamePrefix:     req.NamePrefix,
		DateOffsetDays: req.DateOffsetDays,
		IsDraft:        req.IsDraft,
	}

	result, err := h.svc.BulkDuplicateEvents(c.Context(), req.IDs, cmd)
	if err != nil {
		return response.InternalError(c, "Failed to duplicate events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events duplicated successfully", result)
}

// ============================================================
// MEDIA - UPLOAD
// ============================================================

// UploadEventImage godoc
// @Summary Upload event image
// @Description Upload an image for an event
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param eventId path string true "Event ID"
// @Param image formData file true "Event image"
// @Success 200 {object} response.BaseResponse{data=domain.MediaInfo}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/{eventId}/image [post]
func (h *EventHandler) UploadEventImage(c fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Get file
	imageFile, err := c.FormFile("image")
	if err != nil {
		return response.BadRequest(c, "Image file is required", nil)
	}

	fileContent, err := imageFile.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open image file", nil)
	}
	defer fileContent.Close()

	imageData, err := io.ReadAll(fileContent)
	if err != nil {
		return response.InternalError(c, "Failed to read image file", nil)
	}

	cmd := service.UploadEventImageCommand{
		EventID:     eventID,
		ImageData:   imageData,
		ImageName:   imageFile.Filename,
		ContentType: imageFile.Header.Get("Content-Type"),
		UploadedBy:  userID,
	}

	media, err := h.svc.UploadEventImage(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to upload image", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Image uploaded successfully", media)
}

// UploadCertificateTemplate godoc
// @Summary Upload certificate template
// @Description Upload a certificate template for an event
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param eventId path string true "Event ID"
// @Param certificate formData file true "Certificate template (PDF or image)"
// @Success 200 {object} response.BaseResponse{data=domain.MediaInfo}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/{eventId}/certificate [post]
func (h *EventHandler) UploadCertificateTemplate(c fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Get file
	certFile, err := c.FormFile("certificate")
	if err != nil {
		return response.BadRequest(c, "Certificate file is required", nil)
	}

	fileContent, err := certFile.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open certificate file", nil)
	}
	defer fileContent.Close()

	certData, err := io.ReadAll(fileContent)
	if err != nil {
		return response.InternalError(c, "Failed to read certificate file", nil)
	}

	cmd := service.UploadCertificateCommand{
		EventID:         eventID,
		CertificateData: certData,
		CertificateName: certFile.Filename,
		ContentType:     certFile.Header.Get("Content-Type"),
		UploadedBy:      userID,
	}

	media, err := h.svc.UploadCertificateTemplate(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to upload certificate template", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Certificate template uploaded successfully", media)
}

// ============================================================
// MEDIA - DELETE Single
// ============================================================

// DeleteEventImage godoc
// @Summary Delete event image
// @Description Delete the image for an event
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/{eventId}/image [delete]
func (h *EventHandler) DeleteEventImage(c fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	if err := h.svc.DeleteEventImage(c.Context(), eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to delete event image", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event image deleted successfully", nil)
}

// DeleteEventCertificate godoc
// @Summary Delete certificate template
// @Description Delete the certificate template for an event
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/{eventId}/certificate [delete]
func (h *EventHandler) DeleteEventCertificate(c fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	if err := h.svc.DeleteEventCertificate(c.Context(), eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to delete certificate template", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Certificate template deleted successfully", nil)
}

// DeleteAllEventMedia godoc
// @Summary Delete all media for an event
// @Description Delete all media (images, certificates) for an event
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/{eventId}/media [delete]
func (h *EventHandler) DeleteAllEventMedia(c fiber.Ctx) error {
	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	if err := h.svc.DeleteAllEventMedia(c.Context(), eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to delete all media", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "All media deleted successfully", nil)
}

// ============================================================
// MEDIA - DELETE Bulk
// ============================================================

// BulkDeleteEventMedia godoc
// @Summary Delete media for multiple events
// @Description Delete all media for multiple events
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body BulkIDsRequest true "Event IDs to delete media for"
// @Success 200 {object} response.BaseResponse{data=service.BulkDeleteResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/bulk/media [delete]
func (h *EventHandler) BulkDeleteEventMedia(c fiber.Ctx) error {
	accountID := c.Params("accountId")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	var req BulkIDsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(req.IDs) == 0 {
		return response.BadRequest(c, "At least one event ID is required", nil)
	}

	if len(req.IDs) > 100 {
		return response.BadRequest(c, "Maximum 100 events can be processed at once", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	result, err := h.svc.BulkDeleteEventMedia(c.Context(), req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to delete media", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Media deleted successfully", result)
}

// internal/modules/events/handler/eventhandler.go

// ============================================================
// PUBLIC HANDLERS WITH CREATOR INFO
// ============================================================

// GetUpcomingEventsWithCreator godoc
// @Summary Get upcoming events with creator details
// @Description Get all upcoming published events with full creator information
// @Tags Events
// @Produce json
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} response.BaseResponse{data=[]EventResponse}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/upcoming [get]
func (h *EventHandler) GetUpcomingEventsWithCreator(c fiber.Ctx) error {
	limit := getQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	events, err := h.svc.GetUpcomingEventsWithCreator(c.Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get upcoming events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]EventResponse, len(events))
	for i, event := range events {
		responses[i] = NewEventResponseFromEvent(event)
	}

	return response.Success(c, "Upcoming events retrieved successfully", responses)
}

// GetEventBySlugWithCreator godoc
// @Summary Get event by slug with creator details
// @Description Get event details by slug with full creator information
// @Tags Events
// @Produce json
// @Param slug path string true "Event Slug"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/slug/{slug} [get]
func (h *EventHandler) GetEventBySlugWithCreator(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Event slug is required", nil)
	}

	event, err := h.svc.GetEventBySlugWithCreator(c.Context(), slug)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to get event", fiber.Map{
			"error": err.Error(),
		})
	}

	if event == nil {
		return response.NotFound(c, "Event not found", nil)
	}

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

// GetEventByIDWithCreator godoc
// @Summary Get event by ID with creator details
// @Description Get event details by ID with full creator information
// @Tags Events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEventByIDWithCreator(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	event, err := h.svc.GetEventByIDWithCreator(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		return response.InternalError(c, "Failed to get event", fiber.Map{
			"error": err.Error(),
		})
	}

	if event == nil {
		return response.NotFound(c, "Event not found", nil)
	}

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}


// internal/modules/events/delivery/eventhandler/event_handler.go

// GetEventsByAccountWithCreator godoc
// @Summary Get events by account with creator details
// @Description Get all events for an account with full creator information
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/with-creator [get]
func (h *EventHandler) GetEventsByAccountWithCreator(c fiber.Ctx) error {
	accountID := c.Params("accountId")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := h.svc.GetEventsByAccountWithCreator(c.Context(), accountID, page, pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to get events", fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to response DTOs with creator info
	responses := make([]EventResponse, len(events))
	for i, event := range events {
		responses[i] = NewEventResponseFromEvent(event)
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}
