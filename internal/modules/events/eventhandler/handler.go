// internal/modules/events/eventhandler/handler.go

package eventhandler

import (
	"mime/multipart"
	"strconv"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events/eventservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type EventHandler struct {
	service *eventservice.EventService
}

func NewEventHandler(service *eventservice.EventService) *EventHandler {
	return &EventHandler{service: service}
}

// ================================================
// REQUEST MODELS
// ================================================

type CreateEventRequest struct {
	Name            string    `json:"name" binding:"required" example:"Annual Tech Conference 2024"`
	Description      string    `json:"description" example:"A conference for tech enthusiasts"`
	EventTypeID      string    `json:"event_type_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Date             time.Time `json:"date" binding:"required" example:"2024-12-01"`
	Time             string    `json:"time" binding:"required" example:"09:00"`
	Duration         int       `json:"duration" binding:"required" example:"120"`
	Price            float64   `json:"price" example:"2500"`
	CertificatePrice float64   `json:"certificate_price" example:"500"`
	Location         string    `json:"location" example:"Nairobi, Kenya"`
	ZoomLink         string    `json:"zoom_link" example:"https://zoom.us/j/123456789"`
	MeetLink         string    `json:"meet_link" example:"https://meet.google.com/abc-defg-hij"`
	MaxAttendees     int       `json:"max_attendees" example:"100"`
	IsVirtual        bool      `json:"is_virtual" example:"true"`
}

type CreateEventWithImageRequest struct {
	Name            string    `form:"name" binding:"required" example:"Annual Tech Conference 2024"`
	Description      string    `form:"description" example:"A conference for tech enthusiasts"`
	EventTypeID      string    `form:"event_type_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Date             time.Time `form:"date" binding:"required" example:"2024-12-01"`
	Time             string    `form:"time" binding:"required" example:"09:00"`
	Duration         int       `form:"duration" binding:"required" example:"120"`
	Price            float64   `form:"price" example:"2500"`
	CertificatePrice float64   `form:"certificate_price" example:"500"`
	Location         string    `form:"location" example:"Nairobi, Kenya"`
	ZoomLink         string    `form:"zoom_link" example:"https://zoom.us/j/123456789"`
	MeetLink         string    `form:"meet_link" example:"https://meet.google.com/abc-defg-hij"`
	MaxAttendees     int       `form:"max_attendees" example:"100"`
	IsVirtual        bool      `form:"is_virtual" example:"true"`
}

type UpdateEventRequest struct {
	Title            string    `json:"title" example:"Updated Conference Title"`
	Description      string    `json:"description" example:"Updated description"`
	EventTypeID      string    `json:"event_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventStatusID    string    `json:"event_status_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Date             time.Time `json:"date" example:"2024-12-01"`
	Time             string    `json:"time" example:"09:00"`
	Duration         int       `json:"duration" example:"120"`
	Price            float64   `json:"price" example:"2500"`
	CertificatePrice float64   `json:"certificate_price" example:"500"`
	Location         string    `json:"location" example:"Nairobi, Kenya"`
	ZoomLink         string    `json:"zoom_link" example:"https://zoom.us/j/123456789"`
	MeetLink         string    `json:"meet_link" example:"https://meet.google.com/abc-defg-hij"`
	MaxAttendees     int       `json:"max_attendees" example:"100"`
	IsVirtual        bool      `json:"is_virtual" example:"true"`
}

type GetEventsRequest struct {
	EventTypeID   string `query:"event_type_id"`
	EventStatusID string `query:"event_status_id"`
	Page          int    `query:"page"`
	PageSize      int    `query:"page_size"`
}

type SearchEventsRequest struct {
	Query       string `query:"q"`
	BusinessID  string `query:"business_id"`
	EventTypeID string `query:"event_type_id"`
	Page        int    `query:"page"`
	PageSize    int    `query:"page_size"`
}

// ================================================
// HELPER FUNCTIONS FOR QUERY PARAMETERS
// ================================================

// getQueryInt gets an integer query parameter with a default value
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

// getQueryString gets a string query parameter with a default value
func getQueryString(c fiber.Ctx, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// ================================================
// PUBLIC HANDLERS
// ================================================

// GetEvent godoc
// @Summary Get event by ID
// @Description Get event details by ID (public)
// @Tags Events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=models.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequest(c, "Invalid event ID", nil)
	}

	event, err := h.service.GetEventByIDPublic(c.Context(), uid)
	if err != nil {
		if err.Error() == "event not found" {
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
// @Success 200 {object} response.BaseResponse{data=models.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/slug/{slug} [get]
func (h *EventHandler) GetEventBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Event slug is required", nil)
	}

	event, err := h.service.GetEventBySlug(c.Context(), slug)
	if err != nil {
		if err.Error() == "event not found" {
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

// GetEventsByType godoc
// @Summary Get events by type
// @Description Get all events of a specific type (workshop, webinar, meetup, bootcamp)
// @Tags Events
// @Produce json
// @Param type path string true "Event Type" Enums(workshop, webinar, meetup, bootcamp)
// @Param page query int false "Page number" default(1) example:"1"
// @Param page_size query int false "Page size" default(20) example:"20"
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

	events, total, err := h.service.GetEventsByType(c.Context(), eventType, page, pageSize)
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
// @Param limit query int false "Number of events to return" default(10) example:"10"
// @Success 200 {object} response.BaseResponse{data=[]models.Event}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/upcoming [get]
func (h *EventHandler) GetUpcomingEvents(c fiber.Ctx) error {
	limit := getQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	events, err := h.service.GetUpcomingEvents(c.Context(), limit)
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
// @Param limit query int false "Number of events to return" default(10) example:"10"
// @Success 200 {object} response.BaseResponse{data=[]models.Event}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/past [get]
func (h *EventHandler) GetPastEvents(c fiber.Ctx) error {
	limit := getQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	events, err := h.service.GetPastEvents(c.Context(), limit)
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
// @Success 200 {object} response.BaseResponse{data=[]models.EventType}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/types [get]
func (h *EventHandler) GetEventTypes(c fiber.Ctx) error {
	types, err := h.service.GetEventTypes(c.Context())
	if err != nil {
		return response.InternalError(c, "Failed to get event types", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event types retrieved successfully", types)
}

// SearchEvents godoc
// @Summary Search events
// @Description Search events by title or description
// @Tags Events
// @Produce json
// @Param q query string false "Search query" example:"conference"
// @Param business_id query string false "Business ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param event_type_id query string false "Event Type ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param page query int false "Page number" default(1) example:"1"
// @Param page_size query int false "Page size" default(20) example:"20"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/search [get]
func (h *EventHandler) SearchEvents(c fiber.Ctx) error {
	query := getQueryString(c, "q", "")
	businessIDStr := getQueryString(c, "business_id", "")
	eventTypeIDStr := getQueryString(c, "event_type_id", "")
	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)

	if pageSize > 100 {
		pageSize = 100
	}

	var businessID *uuid.UUID
	if businessIDStr != "" {
		id, err := uuid.Parse(businessIDStr)
		if err != nil {
			return response.BadRequest(c, "Invalid business ID", nil)
		}
		businessID = &id
	}

	var eventTypeID *uuid.UUID
	if eventTypeIDStr != "" {
		id, err := uuid.Parse(eventTypeIDStr)
		if err != nil {
			return response.BadRequest(c, "Invalid event type ID", nil)
		}
		eventTypeID = &id
	}

	events, total, err := h.service.SearchEvents(c.Context(), query, businessID, eventTypeID, page, pageSize)
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

// ================================================
// PROTECTED HANDLERS
// ================================================

// CreateEvent godoc
// @Summary Create a new event
// @Description Create a new event for a business (requires event_manager or host role)
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param businessId path string true "Business ID"
// @Param request body CreateEventRequest true "Event details"
// @Success 201 {object} response.BaseResponse{data=models.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{businessId}/events [post]
func (h *EventHandler) CreateEvent(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("businessId")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	var req CreateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate required fields
	if req.Name == "" {
		return response.BadRequest(c, "Event title is required", nil)
	}
	if req.EventTypeID == "" {
		return response.BadRequest(c, "Event type is required", nil)
	}
	if req.Date.IsZero() {
		return response.BadRequest(c, "Start date is required", nil)
	}
	if req.Time == "" {
		return response.BadRequest(c, "Time is required", nil)
	}
	if req.Duration <= 0 {
		return response.BadRequest(c, "Duration is required and must be greater than 0", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventTypeID, err := uuid.Parse(req.EventTypeID)
	if err != nil {
		return response.BadRequest(c, "Invalid event type ID", nil)
	}

	event := &models.Event{
		Name:            req.Name,
		Description:      req.Description,
		EventTypeID:      eventTypeID,
		Date:             req.Date,
		Time:             req.Time,
		Duration:         req.Duration,
		Price:            req.Price,
		CertificatePrice: req.CertificatePrice,
		Location:         req.Location,
		ZoomLink:         req.ZoomLink,
		MeetLink:         req.MeetLink,
		MaxAttendees:     req.MaxAttendees,
		IsVirtual:        req.IsVirtual,
		IsActive:         true,
	}

	created, err := h.service.CreateEvent(c.Context(), uid, bizUID, event)
	if err != nil {
		if err.Error() == "insufficient permissions to create events for this business" {
			return response.Forbidden(c, "You don't have permission to create events for this business", nil)
		}
		return response.InternalError(c, "Failed to create event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Event created successfully", created)
}

// CreateEventWithImage godoc
// @Summary Create a new event with image
// @Description Create a new event for a business with an image upload
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param businessId path string true "Business ID"
// @Param title formData string true "Event title"
// @Param description formData string false "Event description"
// @Param event_type_id formData string true "Event Type ID"
// @Param date formData string true "Event date (YYYY-MM-DD)"
// @Param time formData string true "Event time (HH:MM)"
// @Param duration formData int true "Event duration in minutes"
// @Param price formData number false "Event price"
// @Param certificate_price formData number false "Certificate price"
// @Param location formData string false "Event location"
// @Param zoom_link formData string false "Zoom link"
// @Param meet_link formData string false "Meet link"
// @Param max_attendees formData int false "Maximum attendees"
// @Param is_virtual formData bool false "Is virtual event"
// @Param image formData file false "Event image"
// @Success 201 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{businessId}/events/with-image [post]
func (h *EventHandler) CreateEventWithImage(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("businessId")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	// Parse form data
	var req CreateEventWithImageRequest
	if err := c.Bind().Form(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate required fields
	if req.Name == "" {
		return response.BadRequest(c, "Event title is required", nil)
	}
	if req.EventTypeID == "" {
		return response.BadRequest(c, "Event type is required", nil)
	}
	if req.Date.IsZero() {
		return response.BadRequest(c, "Start date is required", nil)
	}
	if req.Time == "" {
		return response.BadRequest(c, "Time is required", nil)
	}
	if req.Duration <= 0 {
		return response.BadRequest(c, "Duration is required and must be greater than 0", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventTypeID, err := uuid.Parse(req.EventTypeID)
	if err != nil {
		return response.BadRequest(c, "Invalid event type ID", nil)
	}

	// Get file if provided
	var file multipart.File
	var fileHeader *multipart.FileHeader

	imageFile, err := c.FormFile("image")
	if err == nil && imageFile != nil {
		// Validate file size (max 5MB for event images)
		if imageFile.Size > 5*1024*1024 {
			return response.BadRequest(c, "Image file size exceeds 5MB limit", nil)
		}

		fileContent, err := imageFile.Open()
		if err != nil {
			return response.InternalError(c, "Failed to open image file", nil)
		}
		defer fileContent.Close()

		var ok bool
		file, ok = fileContent.(multipart.File)
		if !ok {
			return response.InternalError(c, "Failed to convert image file", nil)
		}
		fileHeader = imageFile
	}

	event := &models.Event{
		Name:            req.Name,
		Description:      req.Description,
		EventTypeID:      eventTypeID,
		Date:             req.Date,
		Time:             req.Time,
		Duration:         req.Duration,
		Price:            req.Price,
		CertificatePrice: req.CertificatePrice,
		Location:         req.Location,
		ZoomLink:         req.ZoomLink,
		MeetLink:         req.MeetLink,
		MaxAttendees:     req.MaxAttendees,
		IsVirtual:        req.IsVirtual,
		IsActive:         true,
	}

	createdEvent, media, err := h.service.CreateEventWithImage(
		c.Context(),
		uid,
		bizUID,
		event,
		file,
		fileHeader,
	)
	if err != nil {
		return response.InternalError(c, "Failed to create event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Event created successfully", fiber.Map{
		"event": createdEvent,
		"media": media,
	})
}

// GetBusinessEvents godoc
// @Summary Get business events
// @Description Get all events for a business
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param businessId path string true "Business ID"
// @Param event_type_id query string false "Filter by event type ID"
// @Param event_status_id query string false "Filter by event status ID"
// @Param page query int false "Page number" default(1) example:"1"
// @Param page_size query int false "Page size" default(20) example:"20"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{businessId}/events [get]
func (h *EventHandler) GetBusinessEvents(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("businessId")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventTypeIDStr := getQueryString(c, "event_type_id", "")
	eventStatusIDStr := getQueryString(c, "event_status_id", "")
	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)

	if pageSize > 100 {
		pageSize = 100
	}

	var eventTypeID *uuid.UUID
	if eventTypeIDStr != "" {
		id, err := uuid.Parse(eventTypeIDStr)
		if err != nil {
			return response.BadRequest(c, "Invalid event type ID", nil)
		}
		eventTypeID = &id
	}

	var eventStatusID *uuid.UUID
	if eventStatusIDStr != "" {
		id, err := uuid.Parse(eventStatusIDStr)
		if err != nil {
			return response.BadRequest(c, "Invalid event status ID", nil)
		}
		eventStatusID = &id
	}

	events, total, err := h.service.GetBusinessEvents(
		c.Context(),
		uid,
		bizUID,
		eventTypeID,
		eventStatusID,
		page,
		pageSize,
	)
	if err != nil {
		if err.Error() == "insufficient permissions to view events for this business" {
			return response.Forbidden(c, "You don't have permission to view events for this business", nil)
		}
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

// UploadEventImage godoc
// @Summary Upload event image
// @Description Upload an image for an event
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param businessId path string true "Business ID"
// @Param eventId path string true "Event ID"
// @Param image formData file true "Event image"
// @Success 200 {object} response.BaseResponse{data=models.Media}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{businessId}/events/{eventId}/image [post]
func (h *EventHandler) UploadEventImage(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	eventUID, err := uuid.Parse(eventID)
	if err != nil {
		return response.BadRequest(c, "Invalid event ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	// Get file
	imageFile, err := c.FormFile("image")
	if err != nil {
		return response.BadRequest(c, "Image file is required", nil)
	}

	// Validate file size (max 5MB for event images)
	if imageFile.Size > 5*1024*1024 {
		return response.BadRequest(c, "Image file size exceeds 5MB limit", nil)
	}

	fileContent, err := imageFile.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open image file", nil)
	}
	defer fileContent.Close()

	file, ok := fileContent.(multipart.File)
	if !ok {
		return response.InternalError(c, "Failed to convert image file", nil)
	}

	media, err := h.service.UploadEventImage(c.Context(), uid, eventUID, file, imageFile)
	if err != nil {
		if err.Error() == "event not found" {
			return response.NotFound(c, "Event not found", nil)
		}
		if err.Error() == "insufficient permissions to update this event" {
			return response.Forbidden(c, "You don't have permission to update this event", nil)
		}
		return response.InternalError(c, "Failed to upload event image", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event image uploaded successfully", media)
}

// UploadCertificateTemplate godoc
// @Summary Upload certificate template
// @Description Upload a certificate template for an event
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param businessId path string true "Business ID"
// @Param eventId path string true "Event ID"
// @Param certificate formData file true "Certificate template (PDF or image)"
// @Success 200 {object} response.BaseResponse{data=models.Media}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{businessId}/events/{eventId}/certificate [post]
func (h *EventHandler) UploadCertificateTemplate(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventID := c.Params("eventId")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	eventUID, err := uuid.Parse(eventID)
	if err != nil {
		return response.BadRequest(c, "Invalid event ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	// Get file
	certFile, err := c.FormFile("certificate")
	if err != nil {
		return response.BadRequest(c, "Certificate file is required", nil)
	}

	// Validate file size (max 10MB for certificates)
	if certFile.Size > 10*1024*1024 {
		return response.BadRequest(c, "Certificate file size exceeds 10MB limit", nil)
	}

	fileContent, err := certFile.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open certificate file", nil)
	}
	defer fileContent.Close()

	file, ok := fileContent.(multipart.File)
	if !ok {
		return response.InternalError(c, "Failed to convert certificate file", nil)
	}

	media, err := h.service.UploadCertificateTemplate(c.Context(), uid, eventUID, file, certFile)
	if err != nil {
		if err.Error() == "event not found" {
			return response.NotFound(c, "Event not found", nil)
		}
		if err.Error() == "insufficient permissions to upload certificate template" {
			return response.Forbidden(c, "You don't have permission to upload certificate template", nil)
		}
		return response.InternalError(c, "Failed to upload certificate template", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Certificate template uploaded successfully", media)
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
// @Success 200 {object} response.BaseResponse{data=models.Event}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id} [put]
func (h *EventHandler) UpdateEvent(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	eventUID, err := uuid.Parse(eventID)
	if err != nil {
		return response.BadRequest(c, "Invalid event ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	var req UpdateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.EventTypeID != "" {
		updates["event_type_id"] = req.EventTypeID
	}
	if req.EventStatusID != "" {
		updates["event_status_id"] = req.EventStatusID
	}
	if !req.Date.IsZero() {
		updates["date"] = req.Date
	}
	if req.Time != "" {
		updates["time"] = req.Time
	}
	if req.Duration > 0 {
		updates["duration"] = req.Duration
	}
	if req.Price >= 0 {
		updates["price"] = req.Price
	}
	if req.CertificatePrice >= 0 {
		updates["certificate_price"] = req.CertificatePrice
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.ZoomLink != "" {
		updates["zoom_link"] = req.ZoomLink
	}
	if req.MeetLink != "" {
		updates["meet_link"] = req.MeetLink
	}
	if req.MaxAttendees > 0 {
		updates["max_attendees"] = req.MaxAttendees
	}
	updates["is_virtual"] = req.IsVirtual

	if len(updates) == 0 {
		return response.BadRequest(c, "No fields to update", nil)
	}

	updated, err := h.service.UpdateEvent(c.Context(), uid, eventUID, updates)
	if err != nil {
		if err.Error() == "event not found" {
			return response.NotFound(c, "Event not found", nil)
		}
		if err.Error() == "insufficient permissions to update this event" {
			return response.Forbidden(c, "You don't have permission to update this event", nil)
		}
		if err.Error() == "event cannot be edited in its current status" {
			return response.BadRequest(c, "Event cannot be edited in its current status", nil)
		}
		return response.InternalError(c, "Failed to update event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event updated successfully", updated)
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
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	eventUID, err := uuid.Parse(eventID)
	if err != nil {
		return response.BadRequest(c, "Invalid event ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	err = h.service.DeleteEvent(c.Context(), uid, eventUID)
	if err != nil {
		if err.Error() == "event not found" {
			return response.NotFound(c, "Event not found", nil)
		}
		if err.Error() == "insufficient permissions to delete this event" {
			return response.Forbidden(c, "You don't have permission to delete this event", nil)
		}
		return response.InternalError(c, "Failed to delete event", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event deleted successfully", nil)
}