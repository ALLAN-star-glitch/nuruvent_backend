// internal/modules/events/domain/repository.go

package domain

import (
	"context"
)

// ============================================================
// LIST OPTIONS
// ============================================================

type ListOptions struct {
	IncludeCreator bool
	IncludeDeleted bool
	OnlyDeleted    bool
	Visibility     Visibility
}

// ============================================================
// TEAM FILTER
// ============================================================

type TeamFilter struct {
	ID   string
	Type string
}

// ============================================================
// ACCOUNT FILTER
// ============================================================

type AccountFilter struct {
	ID   string
	Type string
}

// ============================================================
// REPOSITORY INTERFACE
// ============================================================

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

	ListEvents(ctx context.Context, filters ListEventsFilters) ([]*Event, int64, error)
	ListEventsByAccount(ctx context.Context, account AccountFilter, filters ListEventsFilters) ([]*Event, int64, error)
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

	GetEventByIDIncludingDeleted(ctx context.Context, id string) (*Event, error)

	GetAllCategories(ctx context.Context) ([]*Category, error)

	AccountIDForTeam(ctx context.Context, teamID string) (string, error)

	AdjustAttendeeCount(ctx context.Context, eventID string, delta int) error

	ListEventIDsWithUnprovisionedSchedules(
		ctx context.Context,
		withinDays int,
	) ([]string, error)

	ReorderEventSchedules(ctx context.Context, eventID string, orderedIDs []string) error
}

// ============================================================
// FILTER STRUCTS
// ============================================================

// ListEventsFilters provides comprehensive filtering for ListEvents.
//
// Every field maps to a WHERE / ORDER BY clause in the repository.
// Zero values mean "no filter applied" — the repository skips the
// corresponding clause.
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

	// SearchQuery is a free-text query matched with ILIKE against
	// name, display_name, description, and short_description.
	//
	// Empty string means no text filter. The repository handles
	// case-insensitive matching.
	SearchQuery string

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

	// SortBy specifies the field to sort by. The repository validates
	// this against a whitelist — unknown fields fall back to updated_at.
	SortBy string

	// SortOrder specifies the sort direction ("asc" or "desc").
	// Anything other than "asc" is treated as "desc".
	SortOrder string

	// Visibility filters events by their visibility level.
	Visibility Visibility
}

// SearchFilters provides filtering for the SearchEvents method.
//
// SearchEvents handles a full-text query separately from ListEvents so
// that callers can distinguish "browse" from "search". The individual
// filter fields mirror ListEventsFilters.
type SearchFilters struct {
	Team           TeamFilter
	Account        AccountFilter
	TeamID         string
	UserID         string
	EventTypeID    string
	CategoryID     string
	EventStatusID  string // ← ADDED: filter by draft/published status
	IncludeDeleted bool
	OnlyDeleted    bool
	Limit          int
	Offset         int
	SortBy         string // ← ADDED: consistent sort semantics with ListEvents
	SortOrder      string // ← ADDED: consistent sort semantics with ListEvents
	Visibility     Visibility
}