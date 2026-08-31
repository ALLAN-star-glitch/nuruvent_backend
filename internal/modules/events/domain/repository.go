// internal/modules/events/domain/repository.go

package domain

import "context"

// ============================================================
// OUTBOUND PORT: Repository Interface
// ============================================================

type Repository interface {
    // ============================================================
    // EVENT CRUD
    // ============================================================

    CreateEvent(ctx context.Context, event *Event) error
    GetEventByID(ctx context.Context, id string) (*Event, error)
    GetEventByIDIncludingDeleted(ctx context.Context, id string) (*Event, error)
    GetEventByName(ctx context.Context, name string) (*Event, error)
    GetEventBySlug(ctx context.Context, slug string) (*Event, error)
    UpdateEvent(ctx context.Context, event *Event) error
    DeleteEvent(ctx context.Context, id string) error

    // Hard delete - permanently removes from database
    PermanentlyDeleteEvent(ctx context.Context, id string) error

    // ============================================================
    // QUERIES
    // ============================================================

    ListEvents(ctx context.Context, filters ListEventsFilters) ([]*Event, int64, error)
    GetEventsByType(ctx context.Context, eventTypeID string, limit, offset int) ([]*Event, int64, error)
    
    // Get events by institution (all events for an organization)
    GetEventsByInstitution(ctx context.Context, institutionID string, limit, offset int) ([]*Event, int64, error)
    
    // Get events by user (all events created by a specific user)
    GetEventsByUser(ctx context.Context, userID string, limit, offset int) ([]*Event, int64, error)
    
    GetUpcomingEvents(ctx context.Context, limit int) ([]*Event, error)
    GetPastEvents(ctx context.Context, limit int) ([]*Event, error)
    SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*Event, int64, error)

    // ============================================================
    // EVENT QUERIES WITH CREATOR INFO
    // ============================================================

    // GetUpcomingEventsWithCreator returns upcoming events with creator info populated
    GetUpcomingEventsWithCreator(ctx context.Context, limit int) ([]*Event, error)

    // GetEventByNameWithCreator returns an event by name with creator info populated
    GetEventByNameWithCreator(ctx context.Context, name string) (*Event, error)

    // GetEventBySlugWithCreator returns an event by slug with creator info populated
    GetEventBySlugWithCreator(ctx context.Context, slug string) (*Event, error)

    // GetEventByIDWithCreator returns an event by ID with creator info populated
    GetEventByIDWithCreator(ctx context.Context, id string) (*Event, error)

    // GetEventsByInstitutionWithCreator returns events for an institution with creator info populated
    GetEventsByInstitutionWithCreator(ctx context.Context, institutionID string, limit, offset int) ([]*Event, int64, error)

    // GetEventsByUserWithCreator returns events created by a user with creator info populated
    GetEventsByUserWithCreator(ctx context.Context, userID string, limit, offset int) ([]*Event, int64, error)

    // ============================================================
    // EVENT QUERIES WITH CREATOR INFO AND FILTERS (NEW)
    // ============================================================

    // GetEventsByUserWithCreatorFiltered returns events with creator info, with privacy/status filters
    // includePrivate: true to include private events (for team members/owners)
    GetEventsByUserWithCreatorFiltered(ctx context.Context, userID string, includePrivate bool, limit, offset int) ([]*Event, int64, error)

    // GetEventsByUserWithCreatorPublic returns only public events (published, non-private)
    GetEventsByUserWithCreatorPublic(ctx context.Context, userID string, limit, offset int) ([]*Event, int64, error)

    // ============================================================
    // USER INFO
    // ============================================================

    // GetUserInfoByID returns user info for a given user ID
    GetUserInfoByID(ctx context.Context, userID string) (*UserInfo, error)

    // GetUserInfoByIDs returns user info for multiple user IDs
    GetUserInfoByIDs(ctx context.Context, userIDs []string) ([]*UserInfo, error)

    // ============================================================
    // EVENT TYPES (Value Objects)
    // ============================================================

    GetEventTypeByID(ctx context.Context, id string) (*EventType, error)
    GetEventTypeByName(ctx context.Context, name string) (*EventType, error)
    GetEventTypeBySlug(ctx context.Context, slug string) (*EventType, error)
    GetAllEventTypes(ctx context.Context) ([]*EventType, error)

    // ============================================================
    // EVENT STATUSES (Value Objects)
    // ============================================================

    GetEventStatusByID(ctx context.Context, id string) (*EventStatus, error)
    GetEventStatusByName(ctx context.Context, name string) (*EventStatus, error)
    GetEventStatusBySlug(ctx context.Context, slug string) (*EventStatus, error)
    GetAllEventStatuses(ctx context.Context) ([]*EventStatus, error)
}

// ============================================================
// FILTERS
// ============================================================

type ListEventsFilters struct {
	InstitutionID  string // Filter by institution
	UserID         string // Filter by user (creator)
	EventTypeID    string
	EventStatusID  string
	CategoryID     string // Filter by category
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
	SortBy         string // Field to sort by (e.g., "created_at", "start_date", "name")
	SortOrder      string // "asc" or "desc"
	Visibility     string // ✅ ADDED: Filter by visibility ("public", "private", "unlisted")
}

type SearchFilters struct {
	InstitutionID  string // Filter by institution
	UserID         string // Filter by user (creator)
	EventTypeID    string
	CategoryID     string // Filter by category
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
	Visibility     string // ✅ ADDED: Filter by visibility ("public", "private", "unlisted")
}