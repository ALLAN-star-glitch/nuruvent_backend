// internal/modules/events/domain/repository.go

package domain

import "context"

// ============================================================
// LIST OPTIONS
// ============================================================

// ListOptions controls what data is returned in list queries
type ListOptions struct {
    IncludeCreator bool
    IncludeDeleted bool
    OnlyDeleted    bool
    Visibility     Visibility
}

// ============================================================
// TEAM FILTER
// ============================================================

// TeamFilter identifies which team to filter by
// This is a simple data struct - NO methods
type TeamFilter struct {
    // ID is the team identifier
    // - For personal teams: user_id
    // - For institution teams: institution_id
    ID string

    // Type indicates the team type
    // Valid values: "personal" or "institution"
    // If empty, no team filter is applied
    Type string
}

// ============================================================
// REPOSITORY INTERFACE
// ============================================================

// Repository defines the data access interface for the events module
type Repository interface {
    // ============================================================
    // EVENT CRUD OPERATIONS
    // ============================================================

    CreateEvent(ctx context.Context, event *Event) error
    GetEventByID(ctx context.Context, id string) (*Event, error)
    GetEventBySlug(ctx context.Context, slug string) (*Event, error)
    UpdateEvent(ctx context.Context, event *Event) error
    DeleteEvent(ctx context.Context, id string) error
    PermanentlyDeleteEvent(ctx context.Context, id string) error
    RestoreEvent(ctx context.Context, id string) error

    // ============================================================
    // QUERY OPERATIONS
    // ============================================================

    // ListEvents returns a paginated list of events with flexible filtering
    // TeamFilter determines which team's events to return:
    //   - Type="personal", ID=userID → personal team events
    //   - Type="institution", ID=institutionID → institution team events
    //   - Type="" → no team filter (all events user has access to)
    ListEvents(ctx context.Context, filters ListEventsFilters) ([]*Event, int64, error)

    GetUpcomingEvents(ctx context.Context, team TeamFilter, limit int) ([]*Event, error)
    GetPastEvents(ctx context.Context, team TeamFilter, limit int) ([]*Event, error)
    SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*Event, int64, error)

    // ============================================================
    // VALUE OBJECT QUERIES
    // ============================================================

    GetEventTypeByID(ctx context.Context, id string) (*EventType, error)
    GetEventTypeBySlug(ctx context.Context, slug string) (*EventType, error)
    GetAllEventTypes(ctx context.Context) ([]*EventType, error)

    GetEventStatusByID(ctx context.Context, id string) (*EventStatus, error)
    GetEventStatusBySlug(ctx context.Context, slug string) (*EventStatus, error)
    GetAllEventStatuses(ctx context.Context) ([]*EventStatus, error)

     // GetEventByIDIncludingDeleted gets an event by ID including soft-deleted ones
    GetEventByIDIncludingDeleted(ctx context.Context, id string) (*Event, error)
}

// ============================================================
// FILTER STRUCTS
// ============================================================

// ListEventsFilters provides comprehensive filtering for ListEvents
type ListEventsFilters struct {
    // TeamFilter filters events by team (personal or institution)
    Team TeamFilter

    // UserID filters events by the creator (created_by)
    UserID string

    // EventTypeID filters events by their type
    EventTypeID string

    // EventStatusID filters events by their status
    EventStatusID string

    // CategoryID filters events by their category
    CategoryID string

    // IncludeDeleted controls whether soft-deleted events are included
    IncludeDeleted bool

    // OnlyDeleted controls whether ONLY soft-deleted events are returned
    OnlyDeleted bool

    // IncludeCreator controls whether creator user info is populated
    IncludeCreator bool

    // Limit controls the maximum number of events returned
    Limit int

    // Offset controls pagination offset
    Offset int

    // SortBy specifies the field to sort by
    SortBy string

    // SortOrder specifies the sort direction
    SortOrder string

    // Visibility filters events by their visibility level
    Visibility Visibility
}

// SearchFilters provides filtering for the SearchEvents method
type SearchFilters struct {
    Team           TeamFilter
    UserID         string
    EventTypeID    string
    CategoryID     string
    IncludeDeleted bool
    OnlyDeleted    bool
    Limit          int
    Offset         int
    Visibility     Visibility
}