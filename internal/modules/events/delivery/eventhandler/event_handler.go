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

func resolveTeamType(c fiber.Ctx) string {
	if tt := handlerhelper.GetTeamType(c); tt != "" {
		return tt
	}
	if tt := handlerhelper.GetQueryString(c, "team_type", ""); tt != "" {
		return tt
	}
	return "institution"
}

func resolveAccountID(c fiber.Ctx) string {
	if acc := handlerhelper.GetAccountID(c); acc != "" {
		return acc
	}
	return handlerhelper.GetQueryString(c, "account_id", "")
}

func resolveTeamIDForUnified(c fiber.Ctx, userID string) string {
	if tid := handlerhelper.GetQueryString(c, "team_id", ""); tid != "" {
		return tid
	}
	if tid := handlerhelper.GetTeamID(c); tid != "" {
		return tid
	}
	return userID
}

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
// PUBLIC HANDLERS
// ============================================================

func (h *EventHandler) GetEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.GetEventByID(ctx, id)
	if err != nil {
		return respondClassifiedError(c, "get event", err)
	}
	if event == nil {
		return response.NotFound(c, "Event not found", nil)
	}

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

func (h *EventHandler) GetEventBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Event slug is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.GetEventBySlug(ctx, slug)
	if err != nil {
		return respondClassifiedError(c, "get event", err)
	}
	if event == nil {
		return response.NotFound(c, "Event not found", nil)
	}

	return response.Success(c, "Event retrieved successfully", NewEventResponseFromEvent(event))
}

func (h *EventHandler) GetUpcomingEvents(c fiber.Ctx) error {
	limit := handlerhelper.GetQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, err := h.svc.GetUpcomingEvents(ctx, "", limit)
	if err != nil {
		return respondClassifiedError(c, "get upcoming events", err)
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Upcoming events retrieved successfully", responses)
}

func (h *EventHandler) GetPastEvents(c fiber.Ctx) error {
	limit := handlerhelper.GetQueryInt(c, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	ctx := handlerhelper.EnrichUserContext(c)

	events, err := h.svc.GetPastEvents(ctx, "", limit)
	if err != nil {
		return respondClassifiedError(c, "get past events", err)
	}

	responses := h.buildEventResponses(c, events)

	return response.Success(c, "Past events retrieved successfully", responses)
}

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
		team = domain.TeamFilter{ID: req.TeamID, Type: req.TeamType}
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
		return respondClassifiedError(c, "list events", err)
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
			return respondClassifiedError(c, "get events by type", err)
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

func (h *EventHandler) GetEventTypes(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	types, err := h.svc.GetEventTypes(ctx)
	if err != nil {
		return respondClassifiedError(c, "get event types", err)
	}

	dtos := make([]EventTypeDTO, 0, len(types))
	for _, t := range types {
		if built := buildEventTypeDTO(t); built != nil {
			dtos = append(dtos, *built)
		}
	}

	return response.Success(c, "Event types retrieved successfully", dtos)
}

func (h *EventHandler) GetEventStatuses(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	statuses, err := h.svc.GetEventStatuses(ctx)
	if err != nil {
		return respondClassifiedError(c, "get event statuses", err)
	}

	dtos := make([]EventStatusDTO, 0, len(statuses))
	for _, s := range statuses {
		if built := buildEventStatusDTO(s); built != nil {
			dtos = append(dtos, *built)
		}
	}

	return response.Success(c, "Event statuses retrieved successfully", dtos)
}

func (h *EventHandler) GetCategories(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	categories, err := h.svc.GetCategories(ctx)
	if err != nil {
		return respondClassifiedError(c, "get categories", err)
	}

	dtos := make([]CategoryDTO, 0, len(categories))
	for _, cat := range categories {
		if built := buildCategoryDTO(cat); built != nil {
			dtos = append(dtos, *built)
		}
	}

	return response.Success(c, "Categories retrieved successfully", dtos)
}

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

	userID := handlerhelper.GetUserIDOptional(c)
	tokenTeamID := handlerhelper.GetTeamID(c)
	tokenTeamType := handlerhelper.GetTeamType(c)
	isAuthenticated := userID != ""

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
		return respondClassifiedError(c, "search events", err)
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
// PROTECTED HANDLERS
// ============================================================

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
		return respondClassifiedError(c, "view these events", err)
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
// CREATE - Draft
// ============================================================

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
		return respondClassifiedError(c, "create this draft", err)
	}

	return response.Created(c, "Draft created successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// CREATE - Published
// ============================================================

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

	// Handler-level shape validation (fast path).
	if req.Name == "" {
		return response.BadRequest(c, "Event name is required", fiber.Map{
			"field": "name",
		})
	}
	if req.EventTypeID == "" {
		return response.BadRequest(c, "Event type is required", fiber.Map{
			"field": "event_type_id",
		})
	}
	if req.Description == "" {
		return response.BadRequest(c, "Description is required for published events", fiber.Map{
			"field": "description",
		})
	}
	if len(req.Schedules) == 0 {
		return response.BadRequest(c, "At least one schedule is required", fiber.Map{
			"field": "schedules",
		})
	}
	if len(req.Tickets) == 0 {
		return response.BadRequest(c, "At least one ticket is required", fiber.Map{
			"field": "tickets",
		})
	}
	if req.Visibility == "" {
		return response.BadRequest(c, "Visibility is required (public, private, unlisted)", fiber.Map{
			"field": "visibility",
		})
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
		return respondClassifiedError(c, "create this event", err)
	}

	return response.Created(c, "Event created successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// UPDATE
// ============================================================

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
		return respondClassifiedError(c, "update this event", err)
	}

	return response.Success(c, "Event updated successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// DELETE - Single
// ============================================================

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
		return respondClassifiedError(c, "delete this event", err)
	}

	return response.Success(c, "Event deleted successfully", nil)
}

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
		return respondClassifiedError(c, "permanently delete this event", err)
	}

	return response.Success(c, "Event permanently deleted successfully", nil)
}

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
		return respondClassifiedError(c, "restore this event", err)
	}

	return response.Success(c, "Event restored successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// DELETE - Bulk
// ============================================================

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
		return respondClassifiedError(c, "delete these events", err)
	}

	return response.Success(c, "Events deleted successfully", result)
}

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
		return respondClassifiedError(c, "permanently delete these events", err)
	}

	return response.Success(c, "Events permanently deleted successfully", result)
}

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
		return respondClassifiedError(c, "restore these events", err)
	}

	return response.Success(c, "Events restored successfully", result)
}

// ============================================================
// STATUS - Single
// ============================================================

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
		return respondClassifiedError(c, "publish this event", err)
	}

	return response.Success(c, "Event published successfully", NewEventResponseFromEventWithCreator(event))
}

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
		return respondClassifiedError(c, "cancel this event", err)
	}

	return response.Success(c, "Event cancelled successfully", NewEventResponseFromEventWithCreator(event))
}

func (h *EventHandler) CompleteEvent(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	event, err := h.svc.CompleteEvent(ctx, id)
	if err != nil {
		return respondClassifiedError(c, "complete this event", err)
	}

	return response.Success(c, "Event completed successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// STATUS - Bulk
// ============================================================

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
		return respondClassifiedError(c, "publish these events", err)
	}

	return response.Success(c, "Events published successfully", result)
}

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
		return respondClassifiedError(c, "cancel these events", err)
	}

	return response.Success(c, "Events cancelled successfully", result)
}

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
		return respondClassifiedError(c, "complete these events", err)
	}

	return response.Success(c, "Events completed successfully", result)
}

// ============================================================
// DUPLICATE - Single
// ============================================================

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
		return respondClassifiedError(c, "duplicate this event", err)
	}

	return response.Success(c, "Event duplicated successfully", NewEventResponseFromEventWithCreator(event))
}

// ============================================================
// DUPLICATE - Bulk
// ============================================================

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
		return respondClassifiedError(c, "duplicate these events", err)
	}

	return response.Success(c, "Events duplicated successfully", result)
}

// ============================================================
// MEDIA - Upload
// ============================================================

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
		return respondClassifiedError(c, "upload an image for this event", err)
	}

	return response.Success(c, "Image uploaded successfully", media)
}

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
		return respondClassifiedError(c, "upload a certificate for this event", err)
	}

	return response.Success(c, "Certificate template uploaded successfully", media)
}

// detectMimeType detects MIME type from filename and file data.
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

// ============================================================
// MEDIA - Delete Single
// ============================================================

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
		return respondClassifiedError(c, "delete this event's image", err)
	}

	return response.Success(c, "Event image deleted successfully", nil)
}

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
		return respondClassifiedError(c, "delete this event's certificate", err)
	}

	return response.Success(c, "Certificate template deleted successfully", nil)
}

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
		return respondClassifiedError(c, "delete this event's media", err)
	}

	return response.Success(c, "All media deleted successfully", nil)
}

// ============================================================
// MEDIA - Delete Bulk
// ============================================================

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
		return respondClassifiedError(c, "delete media for these events", err)
	}

	return response.Success(c, "Media deleted successfully", result)
}

// ============================================================
// Ticket types (public reference)
// ============================================================

func (h *EventHandler) GetTicketTypes(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	ticketTypes, err := h.svc.GetTicketTypes(ctx)
	if err != nil {
		return respondClassifiedError(c, "get ticket types", err)
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

// ============================================================
// AI-assisted draft generation
// ============================================================

func (h *EventHandler) GenerateEventDraft(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil || userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req service.GenerateEventDraftRequest
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

	req.CreatedBy = userID
	req.TeamID = teamID
	req.TeamType = teamType
	req.AccountID = accountID

	ctx := handlerhelper.EnrichUserContext(c)

	result, err := h.svc.GenerateEventDraft(ctx, req)
	if err != nil {
		return mapGenerateEventDraftError(c, err)
	}

	return response.Success(c, "Draft generated successfully", result)
}

// ============================================================
// AI draft error mapping
// ============================================================
//
// The AI endpoint has its own error surface because it produces
// different failure modes (provider failure, parse failure,
// unpublishable output) that don't map cleanly to the general
// classifier. Kept local to this file.

func mapGenerateEventDraftError(c fiber.Ctx, err error) error {
	if errors.Is(err, service.ErrAIDisabled) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "AI service is not configured",
			"errors": fiber.Map{
				"reason": "openrouter.api_key is empty",
			},
		})
	}
	if errors.Is(err, service.ErrAIProvider) {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"message": "AI service is unavailable",
			"errors": fiber.Map{
				"reason": err.Error(),
			},
		})
	}
	if errors.Is(err, service.ErrAIParseFailure) {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"success": false,
			"message": "AI returned an unparseable response",
			"errors": fiber.Map{
				"reason": err.Error(),
			},
		})
	}

	var unpublishable *service.DraftUnpublishableError
	if errors.As(err, &unpublishable) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "AI-generated draft could not be made publishable",
			"errors": fiber.Map{
				"reason":            unpublishable.Reason,
				"validation_errors": unpublishable.ValidationErrors,
				"raw_draft":         unpublishable.RawDraft,
			},
		})
	}

	// Everything else (invalid lookups, permissions, etc.) goes
	// through the shared classifier.
	return respondClassifiedError(c, "generate a draft", err)
}