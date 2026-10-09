package attendancedomain

import "time"

// ============================================================
// CROSS-EVENT ATTENDEE DIRECTORY
// ============================================================

// ListAttendeesQuery is the input to the cross-event attendee list.
//
// AccountIDs is required — it scopes the query to events owned by
// accounts the caller belongs to. An empty slice returns zero rows;
// the service is responsible for refusing to call the repo when the
// caller has no accounts.
type ListAttendeesQuery struct {
	AccountIDs []string  // caller's account scope (security boundary)
	EventIDs   []string  // optional team filter (UX filter)
	EventID    string
	Search     string
	Statuses   []string
	SortBy     string
	SortOrder  string
	Page       int
	PageSize   int
}

// ListAttendeesResult is the paginated output.
type ListAttendeesResult struct {
	Attendees []*CrossEventAttendeeRow
	Total     int
	Page      int
	PageSize  int
}

// CrossEventAttendeeRow is the read model for one attendee's rollup
// in one event, denormalized with the event's display fields so the
// list can render the event column without a second lookup.
type CrossEventAttendeeRow struct {
	AttendeeID           string
	DisplayName          string
	Email                string
	EventID              string
	EventName            string
	EventSlug            string
	EventStartDate       time.Time
	DerivedStatus        string
	SessionsTotal        int
	SessionsAttended     int
	SessionsConfirmed    int
	TotalDurationSeconds int64
	LastDerivedAt        time.Time
	RegisteredAt         time.Time
	IsHost               bool   // ← add
	Phone  string
}