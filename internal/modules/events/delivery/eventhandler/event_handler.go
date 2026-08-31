// internal/modules/events/handler/eventhandler.go

package eventhandler

import (
	"errors"
	"io"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"

	"github.com/gorilla/schema"
)

// ============================================================
// EVENT HANDLER
// ============================================================


type EventHandler struct {
	svc      service.Service
	enforcer *authorization.Enforcer
	decoder  *schema.Decoder
}

func NewEventHandler(svc service.Service, enforcer *authorization.Enforcer) *EventHandler {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("form") // Use the form tag
	decoder.IgnoreUnknownKeys(true)
	
	return &EventHandler{
		svc:      svc,
		enforcer: enforcer,
		decoder:  decoder,
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
	userID := c.Locals(authdomain.ContextKeyUserID)
	if userID == nil {
		return "", errors.New("user not authenticated")
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user ID")
	}
	return userIDStr, nil
}

// canViewCreatorInfo checks if a user has permission to view creator info
func (h *EventHandler) canViewCreatorInfo(c fiber.Ctx, userID string, event *domain.Event) bool {
	if event.CreatedBy == userID {
		return true
	}

	var teamDomain string
	if event.InstitutionID != nil && *event.InstitutionID != "" {
		teamDomain = authdomain.InstitutionTeamDomain(*event.InstitutionID)
	} else {
		teamDomain = authdomain.PersonalTeamDomain(event.CreatedBy)
	}

	roles := h.enforcer.GetRolesForUserInDomain(userID, teamDomain)
	if len(roles) == 0 {
		return false
	}

	for _, role := range roles {
		switch role {
		case authdomain.RoleAccountAdmin.String(),
			authdomain.RoleEventManager.String(),
			authdomain.RoleTeamMember.String():
			return true
		}
	}

	return false
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
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

// GetEventBySlug godoc
// @Summary Get event by slug
// @Description Get event details by slug (public)
// @Tags Events
// @Produce json
// @Param slug path string true "Event Slug"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

// GetUpcomingEvents godoc
// @Summary Get upcoming events
// @Description Get all upcoming published events
// @Tags Events
// @Produce json
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} response.BaseResponse{data=[]EventResponse}
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

	return response.Success(c, "Upcoming events retrieved successfully", NewEventResponseFromEvents(events))
}

// ListEvents godoc
// @Summary List events
// @Description List events with filters
// @Tags Events
// @Produce json
// @Param institution_id query string false "Institution ID"
// @Param user_id query string false "User ID (creator)"
// @Param event_type_id query string false "Event Type ID"
// @Param event_status_id query string false "Event Status ID"
// @Param category_id query string false "Category ID"
// @Param include_deleted query bool false "Include soft-deleted events"
// @Param only_deleted query bool false "Show ONLY soft-deleted events"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field (created_at, start_date, name)" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events [get]
func (h *EventHandler) ListEvents(c fiber.Ctx) error {
	institutionID := getQueryString(c, "institution_id", "")
	userID := getQueryString(c, "user_id", "")
	eventTypeID := getQueryString(c, "event_type_id", "")
	eventStatusID := getQueryString(c, "event_status_id", "")
	categoryID := getQueryString(c, "category_id", "")
	includeDeleted := c.Query("include_deleted") == "true"
	onlyDeleted := c.Query("only_deleted") == "true"
	limit := getQueryInt(c, "limit", 20)
	offset := getQueryInt(c, "offset", 0)
	sortBy := getQueryString(c, "sort_by", "created_at")
	sortOrder := getQueryString(c, "sort_order", "desc")

	if limit > 100 {
		limit = 100
	}

	filters := service.ListEventsFilters{
		InstitutionID:  institutionID,
		UserID:         userID,
		EventTypeID:    eventTypeID,
		EventStatusID:  eventStatusID,
		CategoryID:     categoryID,
		IncludeDeleted: includeDeleted,
		OnlyDeleted:    onlyDeleted,
		Limit:          limit,
		Offset:         offset,
		SortBy:         sortBy,
		SortOrder:      sortOrder,
	}

	events, total, err := h.svc.ListEvents(c.Context(), filters)
	if err != nil {
		return response.InternalError(c, "Failed to list events", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        NewEventResponseFromEvents(events),
		"total":       total,
		"limit":       limit,
		"offset":      offset,
		"sort_by":     sortBy,
		"sort_order":  sortOrder,
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
		"data":        NewEventResponseFromEvents(events),
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetPastEvents godoc
// @Summary Get past events
// @Description Get all past published events
// @Tags Events
// @Produce json
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} response.BaseResponse{data=[]EventResponse}
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

	return response.Success(c, "Past events retrieved successfully", NewEventResponseFromEvents(events))
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
// @Param institution_id query string false "Institution ID"
// @Param user_id query string false "User ID (creator)"
// @Param event_type_id query string false "Event Type ID"
// @Param category_id query string false "Category ID"
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
	institutionID := getQueryString(c, "institution_id", "")
	userID := getQueryString(c, "user_id", "")
	eventTypeID := getQueryString(c, "event_type_id", "")
	categoryID := getQueryString(c, "category_id", "")
	includeDeleted := c.Query("include_deleted") == "true"
	onlyDeleted := c.Query("only_deleted") == "true"
	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)

	if pageSize > 100 {
		pageSize = 100
	}

	filters := service.SearchFilters{
		InstitutionID:  institutionID,
		UserID:         userID,
		EventTypeID:    eventTypeID,
		CategoryID:     categoryID,
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
		"data":        NewEventResponseFromEvents(events),
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ============================================================
// PUBLIC HANDLERS WITH CREATOR INFO (Auth Required)
// ============================================================

// GetUpcomingEventsWithCreator godoc
// @Summary Get upcoming events with creator details
// @Description Get all upcoming published events with full creator information (requires auth)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of events to return" default(10)
// @Success 200 {object} response.BaseResponse{data=[]EventResponse}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/upcoming/with-creator [get]
func (h *EventHandler) GetUpcomingEventsWithCreator(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

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
		if h.canViewCreatorInfo(c, userID, event) {
			responses[i] = NewEventResponseFromEventWithCreator(event)
		} else {
			responses[i] = NewEventResponseFromEvent(event)
		}
	}

	return response.Success(c, "Upcoming events retrieved successfully", responses)
}

// GetEventBySlugWithCreator godoc
// @Summary Get event by slug with creator details
// @Description Get event details by slug with full creator information (requires auth)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Event Slug"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/slug/{slug}/with-creator [get]
func (h *EventHandler) GetEventBySlugWithCreator(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

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

	if h.canViewCreatorInfo(c, userID, event) {
		return response.Success(c, "Event retrieved successfully", NewEventResponseFromEventWithCreator(event))
	}

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

// GetEventByIDWithCreator godoc
// @Summary Get event by ID with creator details
// @Description Get event details by ID with full creator information (requires auth)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/with-creator [get]
func (h *EventHandler) GetEventByIDWithCreator(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

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

	if h.canViewCreatorInfo(c, userID, event) {
		return response.Success(c, "Event retrieved successfully", NewEventResponseFromEventWithCreator(event))
	}

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

// ============================================================
// PROTECTED HANDLERS (Auth Required)
// ============================================================

// GetEventsByInstitution godoc
// @Summary Get events by institution
// @Description Get all events for an institution (public - filters private events)
// @Tags Events
// @Produce json
// @Param institutionId path string true "Institution ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events [get]
func (h *EventHandler) GetEventsByInstitution(c fiber.Ctx) error {
    institutionID := c.Params("institutionId")
    if institutionID == "" {
        return response.BadRequest(c, "Institution ID is required", nil)
    }

    // ✅ Auth is OPTIONAL - user may or may not be logged in
    var userID string
    if user := c.Locals(authdomain.ContextKeyUserID); user != nil {
        if id, ok := user.(string); ok {
            userID = id
        }
    }

    page := getQueryInt(c, "page", 1)
    pageSize := getQueryInt(c, "page_size", 20)
    if pageSize > 100 {
        pageSize = 100
    }

    events, total, err := h.svc.GetEventsByInstitution(c.Context(), institutionID, page, pageSize)
    if err != nil {
        return response.InternalError(c, "Failed to get events", fiber.Map{
            "error": err.Error(),
        })
    }

    // If user is authenticated, check permissions for each event
    var responses []EventResponse
    if userID != "" {
        responses = make([]EventResponse, len(events))
        for i, event := range events {
            if h.canViewCreatorInfo(c, userID, event) {
                responses[i] = NewEventResponseFromEventWithCreator(event)
            } else {
                responses[i] = NewEventResponseFromEvent(event)
            }
        }
    } else {
        // Not authenticated - return events without creator info
        responses = NewEventResponseFromEvents(events)
    }

    return response.Success(c, "Events retrieved successfully", fiber.Map{
        "data":        responses,
        "total":       total,
        "page":        page,
        "page_size":   pageSize,
        "total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
    })
}

// GetEventsByInstitutionWithCreator godoc
// @Summary Get events by institution with creator details
// @Description Get all events for an institution with full creator information (requires team admin)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param institutionId path string true "Institution ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/with-creator [get]
func (h *EventHandler) GetEventsByInstitutionWithCreator(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	institutionID := c.Params("institutionId")
	if institutionID == "" {
		return response.BadRequest(c, "Institution ID is required", nil)
	}

	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := h.svc.GetEventsByInstitutionWithCreator(c.Context(), institutionID, page, pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to get events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]EventResponse, len(events))
	for i, event := range events {
		if h.canViewCreatorInfo(c, userID, event) {
			responses[i] = NewEventResponseFromEventWithCreator(event)
		} else {
			responses[i] = NewEventResponseFromEvent(event)
		}
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetEventsByUserWithCreator godoc
// @Summary Get events by user with creator details
// @Description Get all events created by a specific user with full creator information (requires auth)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/{userId}/events/with-creator [get]
func (h *EventHandler) GetEventsByUserWithCreator(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	targetUserID := c.Params("userId")
	if targetUserID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := h.svc.GetEventsByUserWithCreator(c.Context(), targetUserID, page, pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to get events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]EventResponse, len(events))
	for i, event := range events {
		if h.canViewCreatorInfo(c, userID, event) {
			responses[i] = NewEventResponseFromEventWithCreator(event)
		} else {
			responses[i] = NewEventResponseFromEvent(event)
		}
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetMyEvents godoc
// @Summary Get my events
// @Description Get all events created by the authenticated user (personal events)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/events [get]
func (h *EventHandler) GetMyEvents(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	page := getQueryInt(c, "page", 1)
	pageSize := getQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	events, total, err := h.svc.GetEventsByUser(c.Context(), userID, page, pageSize)
	if err != nil {
		return response.InternalError(c, "Failed to get events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]EventResponse, len(events))
	for i, event := range events {
		responses[i] = NewEventResponseFromEventWithCreator(event)
	}

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetMyEventsWithCreator godoc
// @Summary Get my events with creator details
// @Description Get all events created by the authenticated user with creator info (same as GetMyEvents)
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/events/with-creator [get]
func (h *EventHandler) GetMyEventsWithCreator(c fiber.Ctx) error {
	return h.GetMyEvents(c)
}

// ============================================================
// CREATE DRAFT - Personal
// ============================================================

// ============================================================
// CREATE DRAFT - Personal
// ============================================================

// CreatePersonalDraft godoc
// @Summary Create a personal draft event
// @Description Create a draft event for personal account (no institution required)
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateDraftRequest true "Draft event details"
// @Success 201 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/events/draft [post]
func (h *EventHandler) CreatePersonalDraft(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Bind JSON request
	var req CreateDraftRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// ✅ Convert to command - set owner_type from URL context
	// Personal endpoint = "personal", no institution ID
	cmd := ConvertCreateDraftRequestToCommand(req, userID, "personal", "")

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

	return response.Created(c, "Draft created successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// CREATE DRAFT - Institution
// ============================================================

// internal/modules/events/handler/eventhandler.go

// ============================================================
// CREATE DRAFT - Institution
// ============================================================

// CreateDraft godoc
// @Summary Create a draft event for institution
// @Description Create a draft event for institution account (institution ID required)
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param institutionId path string true "Institution ID"
// @Param request body CreateDraftRequest true "Draft event details"
// @Success 201 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/draft [post]
func (h *EventHandler) CreateDraft(c fiber.Ctx) error {
	institutionID := c.Params("institutionId")
	if institutionID == "" {
		return response.BadRequest(c, "Institution ID is required", nil)
	}

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Bind JSON request
	var req CreateDraftRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// ✅ Convert to command - set owner_type from URL context
	// Institution endpoint = "institution", institution ID from URL
	cmd := ConvertCreateDraftRequestToCommand(req, userID, "institution", institutionID)

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

	return response.Created(c, "Draft created successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// CREATE PUBLISHED EVENT - Personal
// ============================================================

// ============================================================
// CREATE PUBLISHED EVENT - Personal
// ============================================================

// CreatePersonalEvent godoc
// @Summary Create a published personal event
// @Description Create a published event for personal account (no institution required)
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateEventRequest true "Event details"
// @Success 201 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/events [post]
func (h *EventHandler) CreatePersonalEvent(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Bind JSON request
	var req CreateEventRequest
	if err := c.Bind().Body(&req); err != nil {
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
	if req.Description == "" {
		return response.BadRequest(c, "Description is required for published events", nil)
	}
	if len(req.Schedules) == 0 {
		return response.BadRequest(c, "At least one schedule is required", nil)
	}
	if len(req.Tickets) == 0 {
		return response.BadRequest(c, "At least one ticket is required", nil)
	}
	if req.Visibility == "" {
		return response.BadRequest(c, "Visibility is required (public, private, unlisted)", nil)
	}

	// ✅ Convert to command - set owner_type from URL context
	// Personal endpoint = "personal", no institution ID
	cmd := ConvertCreateEventRequestToCommand(req, userID, "personal", "")

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

	return response.Created(c, "Event created successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// CREATE PUBLISHED EVENT - Institution
// ============================================================

// ============================================================
// CREATE PUBLISHED EVENT - Institution
// ============================================================

// CreateEvent godoc
// @Summary Create a published event for institution
// @Description Create a published event for institution account (institution ID required)
// @Tags Events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param institutionId path string true "Institution ID"
// @Param request body CreateEventRequest true "Event details"
// @Success 201 {object} response.BaseResponse{data=EventResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events [post]
func (h *EventHandler) CreateEvent(c fiber.Ctx) error {
    institutionID := c.Params("institutionId")
    if institutionID == "" {
        return response.BadRequest(c, "Institution ID is required", nil)
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

    // ✅ ADD DEBUG LOGGING
    log.Printf("🔍 DEBUG: Request - Name='%s', EventTypeID='%s', Tickets count=%d", 
        req.Name, req.EventTypeID, len(req.Tickets))
    for i, ticket := range req.Tickets {
        log.Printf("🔍 DEBUG: Ticket %d - TypeID='%s', Name='%s', Price=%f, Quantity=%d", 
            i, ticket.TicketTypeID, ticket.Name, ticket.Price, ticket.Quantity)
    }

    // ✅ Convert to command - set owner_type from URL context
    cmd := ConvertCreateEventRequestToCommand(req, userID, "institution", institutionID)

    // ✅ ADD DEBUG LOGGING FOR COMMAND
    log.Printf("🔍 DEBUG: Command - OwnerType='%s', InstitutionID='%s', Tickets count=%d", 
        cmd.OwnerType, cmd.InstitutionID, len(cmd.Tickets))
    for i, ticket := range cmd.Tickets {
        log.Printf("🔍 DEBUG: Command Ticket %d - TypeID='%s', Name='%s', Price=%f, Quantity=%d", 
            i, ticket.TicketTypeID, ticket.Name, ticket.Price, ticket.Quantity)
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

    return response.Created(c, "Event created successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// UPDATE EVENT
// ============================================================

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
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

    // ✅ Convert to command - no owner_type needed
    // The service will determine owner_type from the existing event
    cmd := ConvertUpdateEventRequestToCommand(req, id, userID)

    event, err := h.svc.UpdateEvent(c.Context(), cmd)
    if err != nil {
        if errors.Is(err, domain.ErrEventNotFound) {
            return response.NotFound(c, "Event not found", nil)
        }
        return response.InternalError(c, "Failed to update event", fiber.Map{
            "error": err.Error(),
        })
    }

    return response.Success(c, "Event updated successfully", NewEventResponseFromEventWithCreator(event))
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
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event restored successfully", NewEventResponseFromEventWithCreator(event))
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
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event published successfully", NewEventResponseFromEventWithCreator(event))
}

// CancelEvent godoc
// @Summary Cancel an event
// @Description Cancel an event
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event cancelled successfully", NewEventResponseFromEventWithCreator(event))
}

// CompleteEvent godoc
// @Summary Complete an event
// @Description Mark an event as completed
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event completed successfully", NewEventResponseFromEventWithCreator(event))
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
// @Success 200 {object} response.BaseResponse{data=EventResponse}
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

	return response.Success(c, "Event duplicated successfully", NewEventResponseFromEventWithCreator(event))
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
// @Param institutionId path string true "Institution ID"
// @Param eventId path string true "Event ID"
// @Param image formData file true "Event image"
// @Success 200 {object} response.BaseResponse{data=MediaInfoResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/{eventId}/image [post]
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
// @Param institutionId path string true "Institution ID"
// @Param eventId path string true "Event ID"
// @Param certificate formData file true "Certificate template (PDF or image)"
// @Success 200 {object} response.BaseResponse{data=MediaInfoResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/{eventId}/certificate [post]
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
// @Param institutionId path string true "Institution ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/{eventId}/image [delete]
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
// @Param institutionId path string true "Institution ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/{eventId}/certificate [delete]
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
// @Param institutionId path string true "Institution ID"
// @Param eventId path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/{eventId}/media [delete]
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
// @Param institutionId path string true "Institution ID"
// @Param request body BulkIDsRequest true "Event IDs to delete media for"
// @Success 200 {object} response.BaseResponse{data=service.BulkDeleteResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/events/bulk/media [delete]
func (h *EventHandler) BulkDeleteEventMedia(c fiber.Ctx) error {
	institutionID := c.Params("institutionId")
	if institutionID == "" {
		return response.BadRequest(c, "Institution ID is required", nil)
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