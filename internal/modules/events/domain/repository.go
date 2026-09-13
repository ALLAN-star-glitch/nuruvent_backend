// internal/modules/events/domain/repository.go

package domain

import (
	"context"
)

// ============================================================
// LIST OPTIONS
// ============================================================

// ListOptions controls what data is returned in list queries.
type ListOptions struct {
	IncludeCreator bool
	IncludeDeleted bool
	OnlyDeleted    bool
	Visibility     Visibility
}

// ============================================================
// TEAM FILTER
// ============================================================

// TeamFilter identifies which team to filter events by.
//
// This is a DATA filter, not an authorization boundary. Authorization is
// enforced by the service layer using the event's parent account domain.
//
// Post-revamp note: teams are not Casbin domains. Filtering by team ID
// narrows the result set; it does not gate access.
type TeamFilter struct {
	// ID is the team identifier (UUID from teams table).
	ID string

	// Type filters by team type for organizational grouping.
	// Valid values: "personal" or "institution".
	// If empty, no type filter is applied.
	//
	// This field does NOT affect authorization.
	Type string
}

// ============================================================
// ACCOUNT FILTER
// ============================================================

// AccountFilter identifies which account to filter events by.
//
// Filtering by account returns all events across all teams belonging to
// that account. This is the natural grouping for an account admin who
// needs to see everything under their account.
type AccountFilter struct {
	// ID is the account identifier (UUID from accounts table).
	ID string

	// Type filters by account type for organizational grouping.
	// Valid values: "personal" or "institution".
	// If empty, no type filter is applied.
	Type string
}

// ============================================================
// REPOSITORY INTERFACE
// ============================================================

// Repository defines the data access interface for the events module.
//
// The repository does NOT perform authorization. It returns data that the
// service layer then filters based on the caller's permissions. Authz
// decisions belong to the service, enforced against the event's parent
// account domain.
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

	// ListEvents returns a paginated list of events with flexible filtering.
	//
	// TeamFilter and AccountFilter narrow the result set by data scope.
	// They do NOT enforce authorization — the caller (service layer) is
	// responsible for checking that the authenticated user may see these
	// events.
	//
	// Typical usage:
	//   - Account-scoped listing (admin views all events in their account):
	//       Account.ID = <account-uuid>
	//   - Team-scoped listing (member views one team's events):
	//       Team.ID = <team-uuid>
	//   - Creator-scoped listing (user views events they created):
	//       UserID = <user-uuid>
	ListEvents(ctx context.Context, filters ListEventsFilters) ([]*Event, int64, error)

	// ListEventsByAccount returns all events across all teams under an account.
	ListEventsByAccount(ctx context.Context, account AccountFilter, filters ListEventsFilters) ([]*Event, int64, error)

	// ListEventsByTeam returns all events for a specific team.
	ListEventsByTeam(ctx context.Context, teamID string, filters ListEventsFilters) ([]*Event, int64, error)

	GetUpcomingEvents(ctx context.Context, teamID string, limit int) ([]*Event, error)
	GetPastEvents(ctx context.Context, teamID string, limit int) ([]*Event, error)
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
	GetAllTicketTypes(ctx context.Context) ([]*TicketTypeRow, error)

	// GetEventByIDIncludingDeleted gets an event by ID including soft-deleted ones.
	GetEventByIDIncludingDeleted(ctx context.Context, id string) (*Event, error)

	GetAllCategories(ctx context.Context) ([]*Category, error)

	// Recurrence pattern lookups.
	GetRecurrencePatternBySlug(ctx context.Context, slug string) (*RecurrencePattern, error)

	AccountIDForTeam(ctx context.Context, teamID string) (string, error)
}

// ============================================================
// FILTER STRUCTS
// ============================================================

// ListEventsFilters provides comprehensive filtering for ListEvents.
type ListEventsFilters struct {
	// Team filters events by team (data filter, not authz).
	Team TeamFilter

	// Account filters events by account (all teams under account).
	Account AccountFilter

	// TeamID filters events by a specific team ID (direct).
	TeamID string

	// UserID filters events by the creator (created_by).
	UserID string

	// EventTypeID filters events by their type.
	EventTypeID string

	// EventStatusID filters events by their status.
	EventStatusID string

	// CategoryID filters events by their category.
	CategoryID string

	// IncludeDeleted controls whether soft-deleted events are included.
	IncludeDeleted bool

	// OnlyDeleted controls whether ONLY soft-deleted events are returned.
	OnlyDeleted bool

	// IncludeCreator controls whether creator user info is populated.
	IncludeCreator bool

	// Limit controls the maximum number of events returned.
	Limit int

	// Offset controls pagination offset.
	Offset int

	// SortBy specifies the field to sort by.
	SortBy string

	// SortOrder specifies the sort direction.
	SortOrder string

	// Visibility filters events by their visibility level.
	Visibility Visibility
}

// SearchFilters provides filtering for the SearchEvents method.
type SearchFilters struct {
	Team           TeamFilter
	Account        AccountFilter
	TeamID         string
	UserID         string
	EventTypeID    string
	CategoryID     string
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
	Visibility     Visibility
}