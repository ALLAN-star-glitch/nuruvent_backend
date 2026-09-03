// internal/modules/events/service/read.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// READ - Basic (No Creator Info)
// ============================================================

// GetEventByID retrieves an event by its ID
func (s *eventService) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	if id == "" {
		return nil, errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// Check if user can view this event using Scope
	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	// ✅ Populate organizer (public-facing - always shown)
	organizer, err := s.getOrganizerInfo(ctx, event)
	if err != nil {
		log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
	} else {
		event.Organizer = organizer
	}

	// ✅ Populate creator info (internal - only for authorized users)
	if userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
		event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
	}

	return event, nil
}

// GetEventBySlug retrieves an event by its slug
func (s *eventService) GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error) {
	if slug == "" {
		return nil, errors.New("event slug is required")
	}

	event, err := s.repo.GetEventBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// Check if user can view this event using Scope
	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	// ✅ Populate organizer (public-facing - always shown)
	organizer, err := s.getOrganizerInfo(ctx, event)
	if err != nil {
		log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
	} else {
		event.Organizer = organizer
	}

	// ✅ Populate creator info (internal - only for authorized users)
	if userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
		event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
	}

	return event, nil
}

// ListEvents lists events with filters using TeamFilter
func (s *eventService) ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error) {
	userID := s.getUserIDFromContext(ctx)

	log.Printf("🔍 SERVICE: IncludeDeleted=%v, OnlyDeleted=%v, IncludeCreator=%v", 
		filters.IncludeDeleted, filters.OnlyDeleted, filters.IncludeCreator)

	// If user is not authenticated, only show public events
	if userID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
		filters.IncludeCreator = false // Unauthenticated users cannot see creator info
	}

	// Check if the team filter is set and user has permission
	if !s.isEmptyTeam(filters.Team) {
		// Create scope from team filter
		scope := s.scopeFromTeamFilter(filters.Team)

		// Check if user can read ALL events in this scope
		canReadAll, err := s.permChecker.CanReadAllEvents(ctx, userID, scope)
		if err != nil {
			return nil, 0, fmt.Errorf("permission check failed: %w", err)
		}

		// If user cannot read all, they can only read their own events
		if !canReadAll {
			// Check if user can read their own events
			canReadOwn, err := s.permChecker.CanReadOwnEvents(ctx, userID, scope)
			if err != nil {
				return nil, 0, fmt.Errorf("permission check failed: %w", err)
			}
			if !canReadOwn {
				return nil, 0, errors.New("insufficient permissions to view events")
			}
			// Set UserID to current user to filter by their own events
			filters.UserID = userID
		}
	}

	domainFilters := domain.ListEventsFilters{
		Team:           filters.Team,
		UserID:         filters.UserID,
		EventTypeID:    filters.EventTypeID,
		EventStatusID:  filters.EventStatusID,
		CategoryID:     filters.CategoryID,
		IncludeDeleted: filters.IncludeDeleted,
		OnlyDeleted:    filters.OnlyDeleted,
		Limit:          filters.Limit,
		Offset:         filters.Offset,
		SortBy:         filters.SortBy,
		SortOrder:      filters.SortOrder,
		Visibility:     domain.Visibility(filters.Visibility),
	}

	log.Printf("🔍 DOMAIN FILTERS: IncludeDeleted=%v, OnlyDeleted=%v", domainFilters.IncludeDeleted, domainFilters.OnlyDeleted)

	events, total, err := s.repo.ListEvents(ctx, domainFilters)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions (handles deleted events properly)
	showDeleted := filters.IncludeDeleted || filters.OnlyDeleted
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events, showDeleted)

	// ✅ Populate organizer for ALL events (public-facing - always shown)
	for _, event := range filteredEvents {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer
	}

	// ✅ Populate creator info ONLY if:
	// 1. IncludeCreator is true
	// 2. User is authenticated
	// 3. User has permission (via canViewCreatorInfo)
	if filters.IncludeCreator && userID != "" {
		for _, event := range filteredEvents {
			if s.canViewCreatorInfo(ctx, userID, event) {
				event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
			}
			// If not allowed, Creator stays nil
		}
	}

	return filteredEvents, total, nil
}

// GetEventsByType retrieves events by event type slug using TeamFilter
func (s *eventService) GetEventsByType(ctx context.Context, eventTypeSlug string, page, pageSize int) ([]*domain.Event, int64, error) {
	if eventTypeSlug == "" {
		return nil, 0, errors.New("event type slug is required")
	}

	eventType, err := s.repo.GetEventTypeBySlug(ctx, eventTypeSlug)
	if err != nil {
		return nil, 0, err
	}
	if eventType == nil {
		return nil, 0, domain.ErrEventTypeNotFound
	}

	limit, offset := s.calculatePagination(page, pageSize)
	userID := s.getUserIDFromContext(ctx)

	// Use ListEvents with TeamFilter and EventTypeID
	filters := ListEventsFilters{
		Team:           domain.TeamFilter{}, // Empty = no team filter
		EventTypeID:    eventType.ID,
		Limit:          limit,
		Offset:         offset,
		Visibility:     string(domain.VisibilityPublic),
		IncludeCreator: false, // Basic list doesn't need creator info by default
	}

	// If user is authenticated, they can see more
	if userID != "" {
		filters.Visibility = ""
	}

	events, total, err := s.ListEvents(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetEventsByInstitution retrieves events by institution ID using TeamFilter
func (s *eventService) GetEventsByInstitution(ctx context.Context, institutionID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if institutionID == "" {
		return nil, 0, errors.New("institution ID is required")
	}

	limit, offset := s.calculatePagination(page, pageSize)
	userID := s.getUserIDFromContext(ctx)

	filters := ListEventsFilters{
		Team: domain.TeamFilter{
			ID:   institutionID,
			Type: "institution",
		},
		Limit:          limit,
		Offset:         offset,
		IncludeCreator: false, // Basic list doesn't need creator info by default
	}

	// Unauthenticated users only see public events
	if userID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
	}

	return s.ListEvents(ctx, filters)
}

// GetEventsByUser retrieves events created by a user using TeamFilter
func (s *eventService) GetEventsByUser(ctx context.Context, userID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if userID == "" {
		return nil, 0, errors.New("user ID is required")
	}

	limit, offset := s.calculatePagination(page, pageSize)
	currentUserID := s.getUserIDFromContext(ctx)

	filters := ListEventsFilters{
		Team: domain.TeamFilter{
			ID:   userID,
			Type: "personal",
		},
		UserID:         userID,
		Limit:          limit,
		Offset:         offset,
		IncludeCreator: false, // Basic list doesn't need creator info by default
	}

	// Unauthenticated users only see public events
	if currentUserID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
	}

	return s.ListEvents(ctx, filters)
}

// GetUpcomingEvents retrieves upcoming events for a team
func (s *eventService) GetUpcomingEvents(ctx context.Context, team domain.TeamFilter, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)
	
	events, err := s.repo.GetUpcomingEvents(ctx, team, limit)
	if err != nil {
		return nil, err
	}

	// ✅ Populate organizer for each event (public-facing)
	userID := s.getUserIDFromContext(ctx)
	for _, event := range events {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer

		// Populate creator if user has permission
		if userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
			event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
		}
	}

	return events, nil
}

// GetPastEvents retrieves past events for a team
func (s *eventService) GetPastEvents(ctx context.Context, team domain.TeamFilter, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)
	
	events, err := s.repo.GetPastEvents(ctx, team, limit)
	if err != nil {
		return nil, err
	}

	// ✅ Populate organizer for each event (public-facing)
	userID := s.getUserIDFromContext(ctx)
	for _, event := range events {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer

		// Populate creator if user has permission
		if userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
			event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
		}
	}

	return events, nil
}


// SearchEvents searches events by query and filters using TeamFilter
func (s *eventService) SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error) {
	if query == "" && filters.EventTypeID == "" && filters.Team.ID == "" {
		return nil, 0, errors.New("search query or filter is required")
	}

	userID := s.getUserIDFromContext(ctx)

	// If user is not authenticated, only search public events
	if userID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
		filters.IncludeCreator = false // Unauthenticated users cannot see creator info
	}

	// If team filter is set and user is authenticated, check permissions
	if !s.isEmptyTeam(filters.Team) && userID != "" {
		scope := s.scopeFromTeamFilter(filters.Team)

		canReadAll, err := s.permChecker.CanReadAllEvents(ctx, userID, scope)
		if err != nil {
			return nil, 0, fmt.Errorf("permission check failed: %w", err)
		}

		if !canReadAll {
			canReadOwn, err := s.permChecker.CanReadOwnEvents(ctx, userID, scope)
			if err != nil {
				return nil, 0, fmt.Errorf("permission check failed: %w", err)
			}
			if !canReadOwn {
				return nil, 0, errors.New("insufficient permissions to search events")
			}
			filters.UserID = userID
		}
	}

	domainFilters := domain.SearchFilters{
		Team:           filters.Team,
		UserID:         filters.UserID,
		EventTypeID:    filters.EventTypeID,
		CategoryID:     filters.CategoryID,
		IncludeDeleted: filters.IncludeDeleted,
		OnlyDeleted:    filters.OnlyDeleted,
		Limit:          filters.Limit,
		Offset:         filters.Offset,
		Visibility:     domain.Visibility(filters.Visibility),
	}

	events, total, err := s.repo.SearchEvents(ctx, query, domainFilters)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions (handles deleted events properly)
	showDeleted := filters.IncludeDeleted || filters.OnlyDeleted
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events, showDeleted)

	// ✅ Populate organizer for ALL events (public-facing)
	for _, event := range filteredEvents {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer

		// ✅ Populate creator info if requested and user has permission
		if filters.IncludeCreator && userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
			event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
		}
	}

	return filteredEvents, total, nil
}

// ============================================================
// EVENT TYPES & STATUSES
// ============================================================

// GetEventTypes retrieves all event types
func (s *eventService) GetEventTypes(ctx context.Context) ([]*domain.EventType, error) {
	return s.repo.GetAllEventTypes(ctx)
}

// GetEventStatuses retrieves all event statuses
func (s *eventService) GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error) {
	return s.repo.GetAllEventStatuses(ctx)
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

// getUserIDFromContext extracts user ID from context
func (s *eventService) getUserIDFromContext(ctx context.Context) string {
	// Get user ID from context (set by handler)
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// scopeFromTeamFilter converts a TeamFilter to a Scope
func (s *eventService) scopeFromTeamFilter(team domain.TeamFilter) domain.Scope {
	if team.Type == "institution" {
		return domain.NewInstitutionTeamScope(team.ID)
	}
	return domain.NewPersonalTeamScope(team.ID)
}



// canViewEvent checks if a user can view an event using Scope
func (s *eventService) canViewEvent(ctx context.Context, userID string, event *domain.Event) bool {
	// Public events are always viewable
	if event.IsPublic() {
		return true
	}

	// Unlisted events - accessible with direct link
	if event.IsUnlisted() {
		return true
	}

	// Private events - only team members can view
	if event.IsPrivate() {
		if userID == "" {
			return false
		}

		// Create scope from event
		scope := s.getScopeFromEvent(event)

		// Check if user can view events in this scope
		allowed, err := s.permChecker.CanViewEvent(ctx, userID, scope)
		if err != nil {
			log.Printf("⚠️ Permission check failed: %v", err)
			return false
		}
		return allowed
	}

	return false
}

// canViewDeletedEvent checks if a user can view a soft-deleted event
func (s *eventService) canViewDeletedEvent(ctx context.Context, userID string, event *domain.Event) bool {
	// Event creator can always view their own deleted events
	if event.CreatedBy == userID {
		return true
	}

	// Create scope from event
	scope := s.getScopeFromEvent(event)

	// Check if user can read ALL events in this scope (Account Admin or Event Manager)
	canReadAll, err := s.permChecker.CanReadAllEvents(ctx, userID, scope)
	if err == nil && canReadAll {
		return true
	}

	// Check if user can read OWN events in this scope (Team Member)
	canReadOwn, err := s.permChecker.CanReadOwnEvents(ctx, userID, scope)
	if err == nil && canReadOwn {
		// Team members can only see their own deleted events
		return event.CreatedBy == userID
	}

	return false
}

// filterEventsByVisibility filters a list of events based on visibility permissions
func (s *eventService) filterEventsByVisibility(ctx context.Context, userID string, events []*domain.Event, showDeleted bool) []*domain.Event {
	if len(events) == 0 {
		return events
	}

	filtered := make([]*domain.Event, 0, len(events))
	for _, event := range events {
		// For deleted events, check if user has permission to view them
		if event.IsDeleted() {
			if showDeleted && s.canViewDeletedEvent(ctx, userID, event) {
				filtered = append(filtered, event)
			}
			continue
		}

		// For non-deleted events, check visibility (public, private, unlisted)
		if s.canViewEvent(ctx, userID, event) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// calculatePagination calculates limit and offset from page and pageSize
func (s *eventService) calculatePagination(page, pageSize int) (limit, offset int) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if page <= 0 {
		page = 1
	}
	return pageSize, (page - 1) * pageSize
}

// sanitizeLimit sanitizes the limit parameter
func (s *eventService) sanitizeLimit(limit, defaultLimit, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}