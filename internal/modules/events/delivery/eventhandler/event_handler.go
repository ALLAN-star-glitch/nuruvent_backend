// internal/modules/events/handler/eventhandler.go

package eventhandler

import (
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gorilla/schema"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// EVENT HANDLER
// ============================================================

type EventHandler struct {
	svc     service.Service
	decoder *schema.Decoder
}

func NewEventHandler(svc service.Service) *EventHandler {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("form")
	decoder.IgnoreUnknownKeys(true)

	return &EventHandler{
		svc:     svc,
		decoder: decoder,
	}
}

// ============================================================
// INTERNAL SCOPE RESOLUTION (token-first, body ignored)
// ============================================================

// resolveTeamType returns the team type in priority order:
//   1. Authenticated token context (authoritative — from JWT/session)
//   2. Explicit query param (admin override)
//   3. "institution" default
//
// NOTE: request body is intentionally NOT consulted — body is untrusted.
func resolveTeamType(c fiber.Ctx) string {
	if tt := handlerhelper.GetTeamType(c); tt != "" {
		return tt
	}
	if tt := handlerhelper.GetQueryString(c, "team_type", ""); tt != "" {
		return tt
	}
	return "institution"
}

// resolveAccountID returns the account ID in priority order:
//   1. Authenticated token context (authoritative)
//   2. Explicit query param (admin override)
//
// NOTE: request body is intentionally NOT consulted.
func resolveAccountID(c fiber.Ctx) string {
	if acc := handlerhelper.GetAccountID(c); acc != "" {
		return acc
	}
	return handlerhelper.GetQueryString(c, "account_id", "")
}

// resolveTeamIDForUnified returns the target team ID for unified create/list ops:
//   1. Query param (explicit admin target)
//   2. Authenticated token context
//   3. Falls back to the user's personal scope (userID)
func resolveTeamIDForUnified(c fiber.Ctx, userID string) string {
	if tid := handlerhelper.GetQueryString(c, "team_id", ""); tid != "" {
		return tid
	}
	if tid := handlerhelper.GetTeamID(c); tid != "" {
		return tid
	}
	return userID
}

// buildEventResponses builds EventResponse slices with appropriate creator info
func (h *EventHandler) buildEventResponses(c fiber.Ctx, events []*domain.Event) []EventResponse {
	if len(events) == 0 {
		return []EventResponse{}
	}

	responses := make([]EventResponse, len(events))
	for i, event := range events {
		if event.Creator != nil {
			responses[i] = NewEventResponseFromEventWithCreator(event)
		} else {
			responses[i] = NewEventResponseFromEvent(event)
		}
	}
	return responses
}

// ============================================================
// PUBLIC HANDLERS (No Auth Required)
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

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.GetEventByID(ctx, id)
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

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.GetEventBySlug(ctx, slug)
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
	limit := handlerhelper.GetQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, err := h.svc.GetUpcomingEvents(ctx, "", limit)
	if err != nil {
		return response.InternalError(c, "Failed to get upcoming events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Upcoming events retrieved successfully", responses)
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
	limit := handlerhelper.GetQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, err := h.svc.GetPastEvents(ctx, "", limit)
	if err != nil {
		return response.InternalError(c, "Failed to get past events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Past events retrieved successfully", responses)
}

// ListEvents godoc
// @Summary List events (public feed)
// @Description List events with filters
// @Tags Events
// @Produce json
// @Param team_id query string false "Team ID"
// @Param team_type query string false "Team Type (personal or institution)"
// @Param user_id query string false "User ID (creator)"
// @Param event_type_id query string false "Event Type ID"
// @Param event_status_id query string false "Event Status ID"
// @Param category_id query string false "Category ID"
// @Param include_deleted query bool false "Include soft-deleted events"
// @Param only_deleted query bool false "Show ONLY soft-deleted events"
// @Param include_creator query bool false "Include creator details"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field (created_at, start_date, name)" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Param visibility query string false "Visibility (public, private, unlisted)"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events [get]
func (h *EventHandler) ListEvents(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	var req ListEventsRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.BadRequest(c, "Invalid query parameters", nil)
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	team := domain.TeamFilter{}
	if req.TeamID != "" && req.TeamType != "" {
		team = domain.TeamFilter{
			ID:   req.TeamID,
			Type: req.TeamType,
		}
	}

	filters := service.ListEventsFilters{
		Team:           team,
		TeamID:         req.TeamID,
		UserID:         req.UserID,
		EventTypeID:    req.EventTypeID,
		EventStatusID:  req.EventStatusID,
		CategoryID:     req.CategoryID,
		IncludeDeleted: req.IncludeDeleted,
		OnlyDeleted:    req.OnlyDeleted,
		IncludeCreator: req.IncludeCreator,
		Limit:          req.Limit,
		Offset:         req.Offset,
		SortBy:         req.SortBy,
		SortOrder:      req.SortOrder,
		Visibility:     req.Visibility,
	}

	events, total, err := h.svc.ListEvents(ctx, filters)
	if err != nil {
		return response.InternalError(c, "Failed to list events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":       responses,
		"total":      total,
		"limit":      req.Limit,
		"offset":     req.Offset,
		"sort_by":    req.SortBy,
		"sort_order": req.SortOrder,
	})
}

// GetEventsByType godoc
// @Summary Get events by type
// @Description Get all events of a specific type
// @Tags Events
// @Produce json
// @Param type path string true "Event Type Slug"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/type/{type} [get]
func (h *EventHandler) GetEventsByType(c fiber.Ctx) error {
	eventTypeSlug := c.Params("type")
	if eventTypeSlug == "" {
		return response.BadRequest(c, "Event type is required", nil)
	}

	normalizedSlug := eventTypeSlug
	if !strings.HasPrefix(eventTypeSlug, "event-type-") {
		normalizedSlug = "event-type-" + eventTypeSlug
	}

	page := handlerhelper.GetQueryInt(c, "page", 1)
	pageSize := handlerhelper.GetQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, total, err := h.svc.GetEventsByType(ctx, normalizedSlug, page, pageSize)
	if err != nil {
		if normalizedSlug != eventTypeSlug {
			events, total, err = h.svc.GetEventsByType(ctx, eventTypeSlug, page, pageSize)
		}
		if err != nil {
			return response.InternalError(c, "Failed to get events", fiber.Map{
				"error": err.Error(),
			})
		}
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
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
	ctx := handlerhelper.EnrichUserContext(c)

	types, err := h.svc.GetEventTypes(ctx)
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
	ctx := handlerhelper.EnrichUserContext(c)

	statuses, err := h.svc.GetEventStatuses(ctx)
	if err != nil {
		return response.InternalError(c, "Failed to get event statuses", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Event statuses retrieved successfully", statuses)
}

// GetCategories godoc
// @Summary Get all event categories
// @Description Get list of all event categories (public)
// @Tags Events
// @Produce json
// @Success 200 {object} response.BaseResponse{data=[]CategoryDTO}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/categories [get]
func (h *EventHandler) GetCategories(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	categories, err := h.svc.GetCategories(ctx)
	if err != nil {
		return response.InternalError(c, "Failed to get categories", fiber.Map{
			"error": err.Error(),
		})
	}

	categoryDTOs := make([]CategoryDTO, len(categories))
	for i, cat := range categories {
		categoryDTOs[i] = CategoryDTO{
			ID:          cat.ID,
			Slug:        cat.Slug,
			Name:        cat.Name,
			DisplayName: cat.DisplayName,
			Description: cat.Description,
			Icon:        cat.Icon,
			Color:       cat.Color,
		}
	}

	return response.Success(c, "Categories retrieved successfully", categoryDTOs)
}

// SearchEvents godoc
// @Summary Search events
// @Description Search events. Public route (/events/search) returns only public
//              events with no creator info. Authenticated route (/events/me/search)
//              scopes results based on the caller's permissions (read_all vs read_own).
// @Tags Events
// @Produce json
// @Param q query string true "Search query"
// @Param team_id query string false "Team ID override (authenticated only)"
// @Param team_type query string false "Team type override: personal | institution"
// @Param event_type_id query string false "Event Type ID"
// @Param category_id query string false "Category ID"
// @Param include_deleted query bool false "Include soft-deleted events (authenticated only)"
// @Param only_deleted query bool false "Show ONLY soft-deleted events (authenticated only)"
// @Param include_creator query bool false "Include creator details (authenticated only)"
// @Param visibility query string false "Visibility filter (authenticated only)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/search [get]
// @Router /api/v1/events/me/search [get]
func (h *EventHandler) SearchEvents(c fiber.Ctx) error {
	query := handlerhelper.GetQueryString(c, "q", "")
	if query == "" {
		return response.BadRequest(c, "Search query is required", nil)
	}

	page := handlerhelper.GetQueryInt(c, "page", 1)
	pageSize := handlerhelper.GetQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	// ─── Auth context ────────────────────────────────────────────
	userID := handlerhelper.GetUserIDOptional(c)
	tokenTeamID := handlerhelper.GetTeamID(c)
	tokenTeamType := handlerhelper.GetTeamType(c)
	isAuthenticated := userID != ""

	// ─── Team scope resolution ───────────────────────────────────
	teamID := handlerhelper.GetQueryString(c, "team_id", "")
	teamType := handlerhelper.GetQueryString(c, "team_type", "")

	if teamID == "" && isAuthenticated {
		teamID = tokenTeamID
		teamType = tokenTeamType
	}
	if teamID != "" && teamID == tokenTeamID && teamType == "" {
		teamType = tokenTeamType
	}

	team := domain.TeamFilter{}
	if teamID != "" {
		team = domain.TeamFilter{ID: teamID, Type: teamType}
	}

	// ─── Auth-only filters ───────────────────────────────────────
	includeDeleted := false
	onlyDeleted := false
	includeCreator := false
	visibility := handlerhelper.GetQueryString(c, "visibility", "")

	if isAuthenticated {
		includeDeleted = handlerhelper.GetQueryBool(c, "include_deleted", false)
		onlyDeleted = handlerhelper.GetQueryBool(c, "only_deleted", false)
		includeCreator = handlerhelper.GetQueryBool(c, "include_creator", false)
	} else {
		visibility = "public"
	}

	// ─── Build filters ───────────────────────────────────────────
	filters := service.SearchFilters{
		Team:           team,
		TeamID:         teamID,
		EventTypeID:    handlerhelper.GetQueryString(c, "event_type_id", ""),
		CategoryID:     handlerhelper.GetQueryString(c, "category_id", ""),
		IncludeDeleted: includeDeleted,
		OnlyDeleted:    onlyDeleted,
		IncludeCreator: includeCreator,
		Limit:          pageSize,
		Offset:         (page - 1) * pageSize,
		Visibility:     visibility,
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, total, err := h.svc.SearchEvents(ctx, query, filters)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to search events", nil)
		}
		return response.InternalError(c, "Failed to search events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ============================================================
// PROTECTED HANDLERS (Auth Required)
// ============================================================

// ListUserEvents godoc
// @Summary List my events (unified)
// @Description List events scoped to the authenticated user. Uses token scope; ?scope=team or ?team_id= overrides.
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param scope query string false "Scope: personal or team"
// @Param team_id query string false "Team ID (optional; explicit override)"
// @Param event_type_id query string false "Event Type ID"
// @Param event_status_id query string false "Event Status ID"
// @Param category_id query string false "Category ID"
// @Param include_deleted query bool false "Include soft-deleted events"
// @Param only_deleted query bool false "Show ONLY soft-deleted events"
// @Param include_creator query bool false "Include creator details"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events [get]
func (h *EventHandler) ListUserEvents(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	page := handlerhelper.GetQueryInt(c, "page", 1)
	pageSize := handlerhelper.GetQueryInt(c, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}
	limit, offset := pageSize, (page-1)*pageSize

	scope := handlerhelper.GetQueryString(c, "scope", "")
	overrideTeamID := handlerhelper.GetQueryString(c, "team_id", "")

	includeDeleted := handlerhelper.GetQueryBool(c, "include_deleted", false)
	onlyDeleted := handlerhelper.GetQueryBool(c, "only_deleted", false)
	includeCreator := handlerhelper.GetQueryBool(c, "include_creator", false)

	tokenTeamID := handlerhelper.GetTeamID(c)
	tokenTeamType := resolveTeamType(c)

	filters := service.ListEventsFilters{
		Limit:          limit,
		Offset:         offset,
		IncludeDeleted: includeDeleted,
		OnlyDeleted:    onlyDeleted,
		IncludeCreator: includeCreator,
	}

	isTeamScope := scope == "team" || overrideTeamID != "" || (scope == "" && tokenTeamID != "")

	if isTeamScope {
		teamID := overrideTeamID
		if teamID == "" {
			teamID = tokenTeamID
		}
		if teamID == "" {
			return response.BadRequest(c, "team_id is required for team scope", nil)
		}

		teamType := tokenTeamType
		if teamType == "" {
			teamType = "institution"
		}

		filters.TeamID = teamID
		filters.Team = domain.TeamFilter{ID: teamID, Type: teamType}
	} else {
		personalTeamID := tokenTeamID
		if personalTeamID == "" {
			personalTeamID = userID
		}

		filters.Team = domain.TeamFilter{ID: personalTeamID, Type: "personal"}
		filters.UserID = userID
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, total, err := h.svc.ListEvents(ctx, filters)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to view these events", nil)
		}
		return response.InternalError(c, "Failed to list events", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Events retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ============================================================
// CREATE - Draft (Unified)
// ============================================================

// CreateEventDraft godoc
// @Summary Create a draft event (unified)
// @Description Create a draft event. Team scope resolved from token, then query, then personal fallback.
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
// @Router /api/v1/events/draft [post]
func (h *EventHandler) CreateEventDraft(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreateDraftRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	teamID := resolveTeamIDForUnified(c, userID)
	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	if teamID == userID && handlerhelper.GetTeamID(c) == "" && handlerhelper.GetQueryString(c, "team_type", "") == "" {
		teamType = "personal"
	}

	cmd := ConvertCreateDraftRequestToCommand(req, userID, teamID)
	cmd.TeamType = teamType
	cmd.AccountID = accountID

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CreateDraft(ctx, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to create an event for this team", nil)
		}
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
// CREATE - Published (Unified)
// ============================================================

// CreateEvent godoc
// @Summary Create a published event (unified)
// @Description Create a published event. Team scope resolved from token, then query, then personal fallback.
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
// @Router /api/v1/events [post]
func (h *EventHandler) CreateEvent(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

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

	teamID := resolveTeamIDForUnified(c, userID)
	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	if teamID == userID && handlerhelper.GetTeamID(c) == "" && handlerhelper.GetQueryString(c, "team_type", "") == "" {
		teamType = "personal"
	}

	cmd := ConvertCreateEventRequestToCommand(req, userID, teamID)
	cmd.TeamType = teamType
	cmd.AccountID = accountID

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CreateEvent(ctx, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to create an event for this team", nil)
		}
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	cmd := ConvertUpdateEventRequestToCommand(req, id, userID)
	cmd.TeamType = resolveTeamType(c)
	cmd.AccountID = resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.UpdateEvent(ctx, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to update this event", nil)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteEvent(ctx, id, userID, accountID, teamType); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to delete this event", nil)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.PermanentlyDeleteEvent(ctx, id, userID, accountID, teamType); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to permanently delete this event", nil)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.RestoreEvent(ctx, id, userID, accountID, teamType)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to restore this event", nil)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.DeleteEvents(ctx, req.IDs, userID, accountID, teamType)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.PermanentlyDeleteEvents(ctx, req.IDs, userID, accountID, teamType)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamType := resolveTeamType(c)
	accountID := resolveAccountID(c)

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.RestoreEvents(ctx, req.IDs, userID, accountID, teamType)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.PublishEvent(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to publish this event", nil)
		}
		if errors.Is(err, domain.ErrEventScheduleRequired) {
			return response.BadRequest(c, "Cannot publish: event must have at least one schedule", nil)
		}
		if errors.Is(err, domain.ErrEventTicketRequired) {
			return response.BadRequest(c, "Cannot publish: event must have at least one ticket", nil)
		}
		if errors.Is(err, domain.ErrInvalidEventName) {
			return response.BadRequest(c, "Cannot publish: event name is required", nil)
		}
		if errors.Is(err, domain.ErrInvalidEventDescription) {
			return response.BadRequest(c, "Cannot publish: event description is required", nil)
		}
		if errors.Is(err, domain.ErrInvalidEventType) {
			return response.BadRequest(c, "Cannot publish: event type is required", nil)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CancelEvent(ctx, id, userID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to cancel this event", nil)
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

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CompleteEvent(ctx, id)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.BulkPublishEvents(ctx, req.IDs, userID)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.BulkCancelEvents(ctx, req.IDs, userID)
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

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.BulkCompleteEvents(ctx, req.IDs)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	var req DuplicateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	targetTeamID := handlerhelper.GetQueryString(c, "team_id", "")
	if targetTeamID == "" {
		targetTeamID = handlerhelper.GetTeamID(c)
	}

	cmd := service.DuplicateEventCommand{
		Name:      req.Name,
		Date:      req.Date,
		IsDraft:   req.IsDraft,
		CreatedBy: userID,
		TeamID:    targetTeamID,
		TeamType:  resolveTeamType(c),
		AccountID: resolveAccountID(c),
	}

	event, err := h.svc.DuplicateEvent(ctx, id, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to duplicate this event", nil)
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	cmd := service.BulkDuplicateCommand{
		NamePrefix:     req.NamePrefix,
		DateOffsetDays: req.DateOffsetDays,
		IsDraft:        req.IsDraft,
		CreatedBy:      userID,
		AccountID:      resolveAccountID(c),
	}

	result, err := h.svc.BulkDuplicateEvents(ctx, req.IDs, cmd)
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
// @Param id path string true "Event ID"
// @Param image formData file true "Event image"
// @Success 200 {object} response.BaseResponse{data=service.MediaInfo}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/image [post]
func (h *EventHandler) UploadEventImage(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

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

	contentType := imageFile.Header.Get("Content-Type")
	if contentType == "application/octet-stream" || contentType == "" {
		contentType = detectMimeType(imageFile.Filename, imageData)
	}

	cmd := service.UploadEventImageCommand{
		EventID:     eventID,
		ImageData:   imageData,
		ImageName:   imageFile.Filename,
		ContentType: contentType,
		UploadedBy:  userID,
	}

	media, err := h.svc.UploadEventImage(ctx, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to upload an image for this event", nil)
		}
		return response.InternalError(c, "Failed to upload image", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Image uploaded successfully", media)
}

// detectMimeType detects MIME type from filename and file data
func detectMimeType(filename string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	}

	if len(data) >= 4 {
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return "image/png"
		}
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
		if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
			return "image/gif"
		}
		if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
			data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return "image/webp"
		}
	}

	return "application/octet-stream"
}

// UploadCertificateTemplate godoc
// @Summary Upload certificate template
// @Description Upload a certificate template for an event
// @Tags Events
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param certificate formData file true "Certificate template (PDF or image)"
// @Success 200 {object} response.BaseResponse{data=service.MediaInfo}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/certificate [post]
func (h *EventHandler) UploadCertificateTemplate(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

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

	media, err := h.svc.UploadCertificateTemplate(ctx, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to upload a certificate for this event", nil)
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
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/image [delete]
func (h *EventHandler) DeleteEventImage(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteEventImage(ctx, eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to delete this event's image", nil)
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
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/certificate [delete]
func (h *EventHandler) DeleteEventCertificate(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteEventCertificate(ctx, eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to delete this event's certificate", nil)
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
// @Param id path string true "Event ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/{id}/media [delete]
func (h *EventHandler) DeleteAllEventMedia(c fiber.Ctx) error {
	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteAllEventMedia(ctx, eventID, userID); err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			return response.NotFound(c, "Event not found", nil)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to delete this event's media", nil)
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
// @Param request body BulkIDsRequest true "Event IDs to delete media for"
// @Success 200 {object} response.BaseResponse{data=service.BulkDeleteResult}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/bulk/media [delete]
func (h *EventHandler) BulkDeleteEventMedia(c fiber.Ctx) error {
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

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.BulkDeleteEventMedia(ctx, req.IDs, userID)
	if err != nil {
		return response.InternalError(c, "Failed to delete media", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Media deleted successfully", result)
}

// GetTicketTypes godoc
// @Summary Get all ticket types
// @Description Get list of all active ticket types (public)
// @Tags Events
// @Produce json
// @Success 200 {object} response.BaseResponse{data=[]TicketTypeDTO}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/events/ticket-types [get]
func (h *EventHandler) GetTicketTypes(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	ticketTypes, err := h.svc.GetTicketTypes(ctx)
	if err != nil {
		return response.InternalError(c, "Failed to get ticket types", fiber.Map{
			"error": err.Error(),
		})
	}

	dtos := make([]TicketTypeDTO, len(ticketTypes))
	for i, tt := range ticketTypes {
		dtos[i] = TicketTypeDTO{
			ID:          tt.ID,
			Slug:        tt.Slug,
			Name:        tt.Name,
			DisplayName: tt.DisplayName,
			Description: tt.Description,
			SortOrder:   tt.SortOrder,
			IsActive:    tt.IsActive,
		}
	}

	return response.Success(c, "Ticket types retrieved successfully", dtos)
}