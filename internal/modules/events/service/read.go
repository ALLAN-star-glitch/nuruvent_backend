// internal/modules/events/service/read.go

package service

import (
	"context"
	"errors"

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

	// Check if user can view this event
	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
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

	// Check if user can view this event
	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

// ListEvents lists events with filters
func (s *eventService) ListEvents(ctx context.Context, filters ListEventsFilters) ([]*domain.Event, int64, error) {
	userID := s.getUserIDFromContext(ctx)

	// If user is not authenticated, only show public events
	if userID == "" {
		filters.Visibility = domain.VisibilityPublic
	}

	domainFilters := domain.ListEventsFilters{
		InstitutionID:  filters.InstitutionID,
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
		Visibility:     filters.Visibility,
	}

	events, _, err := s.repo.ListEvents(ctx, domainFilters)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
}

// GetEventsByType retrieves events by event type slug
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

	events, _, err := s.repo.GetEventsByType(ctx, eventType.ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
}

// GetEventsByInstitution retrieves events by institution ID
func (s *eventService) GetEventsByInstitution(ctx context.Context, institutionID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if institutionID == "" {
		return nil, 0, errors.New("institution ID is required")
	}

	limit, offset := s.calculatePagination(page, pageSize)
	userID := s.getUserIDFromContext(ctx)

	events, _, err := s.repo.GetEventsByInstitution(ctx, institutionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
}

// GetEventsByUser retrieves events created by a user
func (s *eventService) GetEventsByUser(ctx context.Context, userID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if userID == "" {
		return nil, 0, errors.New("user ID is required")
	}

	limit, offset := s.calculatePagination(page, pageSize)
	currentUserID := s.getUserIDFromContext(ctx)

	events, _, err := s.repo.GetEventsByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions
	filteredEvents := s.filterEventsByVisibility(ctx, currentUserID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
}

// GetUpcomingEvents retrieves upcoming events (public only)
func (s *eventService) GetUpcomingEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)
	return s.repo.GetUpcomingEvents(ctx, limit)
}

// GetPastEvents retrieves past events (public only)
func (s *eventService) GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)
	return s.repo.GetPastEvents(ctx, limit)
}

// SearchEvents searches events by query and filters
func (s *eventService) SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error) {
	if query == "" && filters.EventTypeID == "" && filters.InstitutionID == "" {
		return nil, 0, errors.New("search query or filter is required")
	}

	userID := s.getUserIDFromContext(ctx)

	// If user is not authenticated, only search public events
	if userID == "" {
		filters.Visibility = domain.VisibilityPublic
	}

	domainFilters := domain.SearchFilters{
		InstitutionID:  filters.InstitutionID,
		UserID:         filters.UserID,
		EventTypeID:    filters.EventTypeID,
		CategoryID:     filters.CategoryID,
		IncludeDeleted: filters.IncludeDeleted,
		OnlyDeleted:    filters.OnlyDeleted,
		Limit:          filters.Limit,
		Offset:         filters.Offset,
		Visibility:     filters.Visibility,
	}

	events, _, err := s.repo.SearchEvents(ctx, query, domainFilters)
	if err != nil {
		return nil, 0, err
	}

	// Filter events by visibility permissions
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
}

// ============================================================
// READ - With Creator Info
// ============================================================

// GetUpcomingEventsWithCreator retrieves upcoming events with creator info (public only)
func (s *eventService) GetUpcomingEventsWithCreator(ctx context.Context, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)
	return s.repo.GetUpcomingEventsWithCreator(ctx, limit)
}

// GetEventBySlugWithCreator retrieves an event by slug with creator info
func (s *eventService) GetEventBySlugWithCreator(ctx context.Context, slug string) (*domain.Event, error) {
	if slug == "" {
		return nil, errors.New("event slug is required")
	}

	event, err := s.repo.GetEventBySlugWithCreator(ctx, slug)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

// GetEventByIDWithCreator retrieves an event by ID with creator info
func (s *eventService) GetEventByIDWithCreator(ctx context.Context, id string) (*domain.Event, error) {
	if id == "" {
		return nil, errors.New("event ID is required")
	}

	event, err := s.repo.GetEventByIDWithCreator(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

// GetEventsByInstitutionWithCreator retrieves institution events with creator info
func (s *eventService) GetEventsByInstitutionWithCreator(ctx context.Context, institutionID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if institutionID == "" {
		return nil, 0, errors.New("institution ID is required")
	}

	limit, offset := s.calculatePaginationWithDefaults(page, pageSize, 20, 100)
	userID := s.getUserIDFromContext(ctx)

	events, _, err := s.repo.GetEventsByInstitutionWithCreator(ctx, institutionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	filteredEvents := s.filterEventsByVisibility(ctx, userID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
}

// GetEventsByUserWithCreator retrieves user events with creator info
func (s *eventService) GetEventsByUserWithCreator(ctx context.Context, userID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if userID == "" {
		return nil, 0, errors.New("user ID is required")
	}

	limit, offset := s.calculatePaginationWithDefaults(page, pageSize, 20, 100)
	currentUserID := s.getUserIDFromContext(ctx)

	events, _, err := s.repo.GetEventsByUserWithCreator(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	filteredEvents := s.filterEventsByVisibility(ctx, currentUserID, events)
	return filteredEvents, int64(len(filteredEvents)), nil
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
	// TODO: Implement based on your auth system
	return ""
}

// canViewEvent checks if a user can view an event
func (s *eventService) canViewEvent(ctx context.Context, userID string, event *domain.Event) bool {
	// Public events are always viewable
	if event.Visibility == domain.VisibilityPublic {
		return true
	}

	// Unlisted events - accessible with direct link
	if event.Visibility == domain.VisibilityUnlisted {
		return true
	}

	// Private events - only team members can view
	if event.Visibility == domain.VisibilityPrivate {
		if userID == "" {
			return false
		}

		// Check if user is a member of the team that owns this event
		institutionID := s.getInstitutionIDFromEvent(event)
		if institutionID == "" {
			// Personal event - user must be the creator
			return userID == event.CreatedBy
		}

		// Institution event - check if user is a member of the institution
		return s.permChecker.CanViewEvent(ctx, userID, institutionID)
	}

	return false
}

// filterEventsByVisibility filters a list of events based on visibility permissions
func (s *eventService) filterEventsByVisibility(ctx context.Context, userID string, events []*domain.Event) []*domain.Event {
	if len(events) == 0 {
		return events
	}

	filtered := make([]*domain.Event, 0, len(events))
	for _, event := range events {
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

// calculatePaginationWithDefaults calculates pagination with custom defaults
func (s *eventService) calculatePaginationWithDefaults(page, pageSize, defaultLimit, maxLimit int) (limit, offset int) {
	if pageSize <= 0 {
		pageSize = defaultLimit
	}
	if pageSize > maxLimit {
		pageSize = maxLimit
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