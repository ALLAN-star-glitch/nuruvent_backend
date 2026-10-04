package registrationdomain

import "context"

// ListFilter carries pagination and filtering options for list queries.
type ListFilter struct {
    Statuses  []Status
    EventID   string   // optional
    Search    string   // optional, name or email
    SortBy    string   // optional, e.g. "created_at"
    SortOrder string   // optional, "asc" | "desc"
    Page      int
    PageSize  int
}

// ListAllFilter carries options for the cross-event registration list.
type ListAllFilter struct {
    AccountIDs []string // scope: caller's accounts
    EventID    string   // optional filter
    Search     string   // optional, matches attendee name or email
    Statuses   []Status // optional
    SortBy     string   // optional
    SortOrder  string   // optional
    Page       int
    PageSize   int
}


// ListMineFilter carries options for the current user's own
// registration list (GET /me/registrations).
type ListMineFilter struct {
	UserID    string   // required; caller's user ID
	EventID   string   // optional; filter to one event
	Search    string   // optional; matches guest name/email (no user name join — see notes)
	Statuses  []Status // optional
	SortBy    string   // "created_at" (default) | "event_name" | "status"
	SortOrder string   // "asc" | "desc" (default)
	Page      int
	PageSize  int
}

// RegistrationRepository persists the base Registration entity.
type RegistrationRepository interface {
    Create(ctx context.Context, r *Registration) error
    Update(ctx context.Context, r *Registration) error
    FindByID(ctx context.Context, id string) (*Registration, error)
    FindActiveByUserAndEvent(ctx context.Context, userID, eventID string) (*Registration, error)

     

    WithTx(
        ctx context.Context,
        fn func(
            RegistrationRepository,
            EventRegistrationRepository,
            WaitlistRepository,
        ) error,
    ) error
}

// EventRegistrationRepository persists event-specific registrations.
type EventRegistrationRepository interface {
    Create(ctx context.Context, e *EventRegistration) error
    Update(ctx context.Context, e *EventRegistration) error
    FindByID(ctx context.Context, id string) (*EventRegistration, error)
    ListByEvent(ctx context.Context, eventID string, f ListFilter) ([]*EventRegistration, int, error)
    ListByUser(ctx context.Context, userID string, f ListFilter) ([]*EventRegistration, int, error)
    Deactivate(ctx context.Context, registrationID string) error
// ListAll returns registrations across every event owned by any of
// the given account IDs. Empty AccountIDs returns an empty page.
//
// Returns a flattened read model — not the full EventRegistration
// entity — because the caller only needs table columns.
ListAll(
    ctx context.Context,
    f ListAllFilter,
) ([]*CrossEventRegistrationRow, int, error)

// ListMine returns the caller's own registrations across every event
// they've registered for. Same flattened read model as ListAll —
// different filter.
ListMine(
	ctx context.Context,
	f ListMineFilter,
) ([]*CrossEventRegistrationRow, int, error)
}

// WaitlistRepository persists waitlist entries.
type WaitlistRepository interface {
    Create(ctx context.Context, w *WaitlistEntry) error
    NextPosition(ctx context.Context, eventID string) (int, error)
    PeekNext(ctx context.Context, eventID string) (*WaitlistEntry, error)
    ListByEvent(ctx context.Context, eventID string) ([]*WaitlistEntry, error)
    Delete(ctx context.Context, id string) error
}