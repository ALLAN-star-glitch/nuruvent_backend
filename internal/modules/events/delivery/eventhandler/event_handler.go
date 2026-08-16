package eventhandler

import (
	"errors"
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

func validateCreateEventRequest(req *CreateEventRequest) error {
	if req.Name == "" {
		return errors.New("event name is required")
	}
	if req.EventTypeID == "" {
		return errors.New("event type is required")
	}
	if req.Date == "" {
		return errors.New("event date is required")
	}
	if req.Time == "" {
		return errors.New("event time is required")
	}
	if req.Duration <= 0 {
		return errors.New("event duration must be greater than 0")
	}
	return nil
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
	limit := getQueryInt(c, "limit", 20)
	offset := getQueryInt(c, "offset", 0)

	if limit > 100 {
		limit = 100
	}

	filters := service.ListEventsFilters{
		AccountID:     accountID,
		EventTypeID:   eventTypeID,
		EventStatusID: eventStatusID,
		Limit:         limit,
		Offset:        offset,
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
	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)

	if pageSize > 100 {
		pageSize = 100
	}

	filters := service.SearchFilters{
		AccountID:   accountID,
		EventTypeID: eventTypeID,
		Limit:       pageSize,
		Offset:      (page - 1) * pageSize,
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

// CreateEvent godoc
// @Summary Create a new event
// @Description Create a new event for an account
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body CreateEventRequest true "Event details"
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

	var req CreateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	if err := validateCreateEventRequest(&req); err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	cmd := service.CreateEventCommand{
		Name:             req.Name,
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

// CreateEventWithImage godoc
// @Summary Create a new event with image
// @Description Create a new event for an account with an image upload
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
// @Param location formData string false "Event location"
// @Param is_virtual formData bool false "Is virtual event"
// @Param zoom_link formData string false "Zoom link"
// @Param meet_link formData string false "Meet link"
// @Param max_attendees formData int false "Maximum attendees"
// @Param image formData file false "Event image"
// @Success 201 {object} response.BaseResponse{data=domain.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/events/with-image [post]
func (h *EventHandler) CreateEventWithImage(c fiber.Ctx) error {
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

	// Validate
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

	// Get image file if provided
	var imageFile interface{}
	var imageHeader interface{}

	file, err := c.FormFile("image")
	if err == nil && file != nil {
		fileContent, err := file.Open()
		if err != nil {
			
			return response.InternalError(c, "Failed to open image file", nil)
		}
		defer fileContent.Close()
		imageFile = fileContent
		imageHeader = file
	}

	cmd := service.CreateEventWithImageCommand{
		CreateEventCommand: service.CreateEventCommand{
			Name:             req.Name,
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
		},
		ImageFile:   imageFile,
		ImageHeader: imageHeader,
	}

	event, err := h.svc.CreateEventWithImage(c.Context(), cmd)
	if err != nil {
		return response.InternalError(c, "Failed to create event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Event created successfully", event)
}

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
		UpdatedBy:        userID,
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

// DeleteEvent godoc
// @Summary Delete an event
// @Description Delete an event
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

	cmd := service.UploadEventImageCommand{
		EventID:     eventID,
		ImageFile:   fileContent,
		ImageHeader: imageFile,
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

	cmd := service.UploadCertificateCommand{
		EventID:         eventID,
		CertificateFile:   fileContent,
		CertificateHeader: certFile,
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