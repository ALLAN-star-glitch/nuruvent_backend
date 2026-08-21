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
	GetEventBySlug(ctx context.Context, slug string) (*Event, error)
	UpdateEvent(ctx context.Context, event *Event) error
	DeleteEvent(ctx context.Context, id string) error

	// ✅ NEW: Hard delete - permanently removes from database
	PermanentlyDeleteEvent(ctx context.Context, id string) error

	// ============================================================
	// QUERIES
	// ============================================================

	ListEvents(ctx context.Context, filters ListEventsFilters) ([]*Event, int64, error)
	GetEventsByType(ctx context.Context, eventTypeID string, limit, offset int) ([]*Event, int64, error)
	GetEventsByAccount(ctx context.Context, accountID string, limit, offset int) ([]*Event, int64, error)
	GetUpcomingEvents(ctx context.Context, limit int) ([]*Event, error)
	GetPastEvents(ctx context.Context, limit int) ([]*Event, error)
	SearchEvents(ctx context.Context, query string, filters SearchFilters) ([]*Event, int64, error)

	// ============================================================
	// EVENT TYPES (Value Objects)
	// ============================================================

	GetEventTypeByID(ctx context.Context, id string) (*EventType, error)
	GetEventTypeBySlug(ctx context.Context, slug string) (*EventType, error)
	GetAllEventTypes(ctx context.Context) ([]*EventType, error)

	// ============================================================
	// EVENT STATUSES (Value Objects)
	// ============================================================

	GetEventStatusByID(ctx context.Context, id string) (*EventStatus, error)
	GetEventStatusBySlug(ctx context.Context, slug string) (*EventStatus, error)
	GetAllEventStatuses(ctx context.Context) ([]*EventStatus, error)
}

// ============================================================
// FILTERS
// ============================================================

type ListEventsFilters struct {
	AccountID      string
	EventTypeID    string
	EventStatusID  string
	IncludeDeleted bool // ✅ NEW: Include soft-deleted events
	OnlyDeleted    bool   // ✅ NEW: Show ONLY deleted events
	Limit          int
	Offset         int
}

type SearchFilters struct {
	AccountID      string
	EventTypeID    string
	IncludeDeleted bool // ✅ NEW: Include soft-deleted events
	OnlyDeleted    bool   // ✅ NEW: Show ONLY deleted events
	Limit          int
	Offset         int
}