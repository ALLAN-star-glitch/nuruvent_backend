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

	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	organizer, err := s.getOrganizerInfo(ctx, event)
	if err != nil {
		log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
	} else {
		event.Organizer = organizer
	}

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

	userID := s.getUserIDFromContext(ctx)
	if !s.canViewEvent(ctx, userID, event) {
		return nil, domain.ErrEventNotFound
	}

	organizer, err := s.getOrganizerInfo(ctx, event)
	if err != nil {
		log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
	} else {
		event.Organizer = organizer
	}

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

	if userID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
		filters.IncludeCreator = false
	}

	if !s.isEmptyTeam(filters.Team) {
		teamDomain := s.teamDomainFromTeamFilter(filters.Team)

		canReadAll, err := s.permChecker.CanReadAllEvents(ctx, userID, teamDomain)
		if err != nil {
			return nil, 0, fmt.Errorf("permission check failed: %w", err)
		}

		if !canReadAll {
			canReadOwn, err := s.permChecker.CanReadOwnEvents(ctx, userID, teamDomain)
			if err != nil {
				return nil, 0, fmt.Errorf("permission check failed: %w", err)
			}
			if !canReadOwn {
				return nil, 0, errors.New("insufficient permissions to view events")
			}
			filters.UserID = userID
		}
	}

	domainFilters := domain.ListEventsFilters{
		Team:           filters.Team,
		TeamID:         filters.TeamID,
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

	log.Printf("🔍 DOMAIN FILTERS: IncludeDeleted=%v, OnlyDeleted=%v",
		domainFilters.IncludeDeleted, domainFilters.OnlyDeleted)

	events, total, err := s.repo.ListEvents(ctx, domainFilters)
	if err != nil {
		return nil, 0, err
	}

	showDeleted := filters.IncludeDeleted || filters.OnlyDeleted
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events, showDeleted)

	for _, event := range filteredEvents {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer
	}

	if filters.IncludeCreator && userID != "" {
		for _, event := range filteredEvents {
			if s.canViewCreatorInfo(ctx, userID, event) {
				event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
			}
		}
	}

	return filteredEvents, total, nil
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

	filters := ListEventsFilters{
		Team:           domain.TeamFilter{},
		EventTypeID:    eventType.ID,
		Limit:          limit,
		Offset:         offset,
		Visibility:     string(domain.VisibilityPublic),
		IncludeCreator: false,
	}

	if userID != "" {
		filters.Visibility = ""
	}

	return s.ListEvents(ctx, filters)
}

// GetEventsByTeam retrieves events by team ID
func (s *eventService) GetEventsByTeam(ctx context.Context, teamID string, page, pageSize int) ([]*domain.Event, int64, error) {
	if teamID == "" {
		return nil, 0, errors.New("team ID is required")
	}

	limit, offset := s.calculatePagination(page, pageSize)
	userID := s.getUserIDFromContext(ctx)

	filters := ListEventsFilters{
		TeamID:         teamID,
		Limit:          limit,
		Offset:         offset,
		IncludeCreator: false,
	}

	if userID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
	}

	return s.ListEvents(ctx, filters)
}

// GetUpcomingEvents retrieves upcoming events for a team
func (s *eventService) GetUpcomingEvents(ctx context.Context, teamID string, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)

	events, err := s.repo.GetUpcomingEvents(ctx, teamID, limit)
	if err != nil {
		return nil, err
	}

	userID := s.getUserIDFromContext(ctx)
	for _, event := range events {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer

		if userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
			event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
		}
	}

	return events, nil
}

// GetPastEvents retrieves past events for a team
func (s *eventService) GetPastEvents(ctx context.Context, teamID string, limit int) ([]*domain.Event, error) {
	limit = s.sanitizeLimit(limit, 10, 50)

	events, err := s.repo.GetPastEvents(ctx, teamID, limit)
	if err != nil {
		return nil, err
	}

	userID := s.getUserIDFromContext(ctx)
	for _, event := range events {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer

		if userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
			event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
		}
	}

	return events, nil
}

// SearchEvents searches events by query and filters
func (s *eventService) SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*domain.Event, int64, error) {
	if query == "" && filters.EventTypeID == "" && filters.TeamID == "" && filters.Team.ID == "" {
		return nil, 0, errors.New("search query or filter is required")
	}

	userID := s.getUserIDFromContext(ctx)

	// Anonymous callers: force public-only, no creator info
	if userID == "" {
		filters.Visibility = string(domain.VisibilityPublic)
		filters.IncludeCreator = false
	}

	// Team-scoped permission check for authenticated callers
	if filters.TeamID != "" || !s.isEmptyTeam(filters.Team) {
		teamID := filters.TeamID
		if teamID == "" {
			teamID = filters.Team.ID
		}

		if userID != "" && teamID != "" {
			// Resolve team type: explicit filter > token context > default
			teamType := filters.Team.Type
			if teamType == "" {
				teamType = domain.GetTeamType(ctx) // from Go context (set by enrichUserContext)
			}

			teamDomain := domain.BuildTeamDomain(teamType, teamID) // ✅ FIXED

			log.Printf("🔍 SEARCH CHECK: user=%s teamID=%s teamType=%s domain=%s",
				userID, teamID, teamType, teamDomain)

			canReadAll, err := s.permChecker.CanReadAllEvents(ctx, userID, teamDomain)
			if err != nil {
				return nil, 0, fmt.Errorf("permission check failed: %w", err)
			}

			if !canReadAll {
				canReadOwn, err := s.permChecker.CanReadOwnEvents(ctx, userID, teamDomain)
				if err != nil {
					return nil, 0, fmt.Errorf("permission check failed: %w", err)
				}
				if !canReadOwn {
					log.Printf("❌ SEARCH DENIED: user=%s domain=%s has neither read_all nor read_own",
						userID, teamDomain)
					return nil, 0, domain.ErrForbidden
				}
				filters.UserID = userID
			}
		}
	}

	domainFilters := domain.SearchFilters{
		Team:           filters.Team,
		TeamID:         filters.TeamID,
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

	showDeleted := filters.IncludeDeleted || filters.OnlyDeleted
	filteredEvents := s.filterEventsByVisibility(ctx, userID, events, showDeleted)

	for _, event := range filteredEvents {
		organizer, err := s.getOrganizerInfo(ctx, event)
		if err != nil {
			log.Printf("⚠️ Failed to get organizer info for event %s: %v", event.ID, err)
			continue
		}
		event.Organizer = organizer

		if filters.IncludeCreator && userID != "" && s.canViewCreatorInfo(ctx, userID, event) {
			event.Creator = s.getCreatorInfo(ctx, event.CreatedBy)
		}
	}

	return filteredEvents, total, nil
}

// ============================================================
// EVENT TYPES, STATUSES, LOOKUPS
// ============================================================

func (s *eventService) GetEventTypes(ctx context.Context) ([]*domain.EventType, error) {
	return s.repo.GetAllEventTypes(ctx)
}

func (s *eventService) GetEventStatuses(ctx context.Context) ([]*domain.EventStatus, error) {
	return s.repo.GetAllEventStatuses(ctx)
}

func (s *eventService) GetTicketTypes(ctx context.Context) ([]*domain.TicketTypeRow, error) {
	return s.repo.GetAllTicketTypes(ctx)
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

// getUserIDFromContext extracts user ID from context (typed key)
func (s *eventService) getUserIDFromContext(ctx context.Context) string {
	return domain.GetUserID(ctx)
}

// teamDomainFromTeamFilter converts a TeamFilter to a Casbin team domain string.
func (s *eventService) teamDomainFromTeamFilter(team domain.TeamFilter) string {
	if team.ID == "" {
		return ""
	}
	return domain.BuildTeamDomain(team.Type, team.ID) // ✅ FIXED
}

// canViewEvent checks if a user can view an event using team domain
func (s *eventService) canViewEvent(ctx context.Context, userID string, event *domain.Event) bool {
	if event.IsPublic() {
		return true
	}

	if event.IsUnlisted() {
		return true
	}

	if event.IsPrivate() {
		if userID == "" {
			return false
		}

		// ✅ FIXED — build domain from event's team type + team id
		teamDomain := domain.BuildTeamDomain(event.TeamType, event.TeamID)

		allowed, err := s.permChecker.CanViewEvent(ctx, userID, teamDomain)
		if err != nil {
			log.Printf("⚠️ Permission check failed (domain=%s): %v", teamDomain, err)
			return false
		}
		return allowed
	}

	return false
}

// canViewDeletedEvent checks if a user can view a soft-deleted event
func (s *eventService) canViewDeletedEvent(ctx context.Context, userID string, event *domain.Event) bool {
	if event.CreatedBy == userID {
		return true
	}

	// ✅ FIXED
	teamDomain := domain.BuildTeamDomain(event.TeamType, event.TeamID)

	canReadAll, err := s.permChecker.CanReadAllEvents(ctx, userID, teamDomain)
	if err == nil && canReadAll {
		return true
	}

	canReadOwn, err := s.permChecker.CanReadOwnEvents(ctx, userID, teamDomain)
	if err == nil && canReadOwn {
		return event.CreatedBy == userID
	}

	return false
}

// filterEventsByVisibility filters events based on visibility permissions
func (s *eventService) filterEventsByVisibility(ctx context.Context, userID string, events []*domain.Event, showDeleted bool) []*domain.Event {
	if len(events) == 0 {
		return events
	}

	filtered := make([]*domain.Event, 0, len(events))
	for _, event := range events {
		if event.IsDeleted() {
			if showDeleted && s.canViewDeletedEvent(ctx, userID, event) {
				filtered = append(filtered, event)
			}
			continue
		}

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