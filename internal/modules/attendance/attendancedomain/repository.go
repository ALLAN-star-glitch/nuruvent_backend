package attendancedomain

import (
	"context"
	"time"
)

// ============================================================
// REPOSITORY INTERFACES
// ============================================================

// AttendeeRepository persists attendees.
type AttendeeRepository interface {
	Create(ctx context.Context, a *Attendee) error
	Update(ctx context.Context, a *Attendee) error
	FindByID(ctx context.Context, id string) (*Attendee, error)
	FindByExternalRef(ctx context.Context, ref ExternalRef) (*Attendee, error)
	FindByEmail(ctx context.Context, email string) ([]*Attendee, error)
	FindByUsername(ctx context.Context, username string) ([]*Attendee, error)
}

// SessionRepository persists sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	Update(ctx context.Context, s *Session) error
	FindByID(ctx context.Context, id string) (*Session, error)
	FindByExternalAndProviderSession(ctx context.Context, ref ExternalRef, providerSessionID string) (*Session, error)
	ListByExternalRef(ctx context.Context, ref ExternalRef) ([]*Session, error)
	FindByProviderMeetingID(ctx context.Context, provider SessionProvider, meetingID string) (*Session, error)
	FindActiveBySchedule(ctx context.Context, before time.Time) ([]*Session, error)
	FindEndedBefore(ctx context.Context, before time.Time) ([]*Session, error)

   UpdateStatus(ctx context.Context, sessionID string, status SessionStatus) error
}

// JoinTokenRepository persists join tokens.
type JoinTokenRepository interface {
	Create(ctx context.Context, t *JoinToken) error
	FindByHash(ctx context.Context, hash string) (*JoinToken, error)
	RevokeByAttendeeSession(ctx context.Context, attendeeID, sessionID string, now time.Time) error
	RevokeByID(ctx context.Context, id string, now time.Time) error
	ListByAttendee(ctx context.Context, attendeeID string) ([]*JoinToken, error)
}

// AttendanceRecordRepository persists attendance records.
type AttendanceRecordRepository interface {
	Create(ctx context.Context, r *AttendanceRecord) error
	Update(ctx context.Context, r *AttendanceRecord) error
	FindByID(ctx context.Context, id string) (*AttendanceRecord, error)
	FindByAttendeeSession(ctx context.Context, attendeeID, sessionID string) ([]*AttendanceRecord, error)
	FindOpenByAttendeeSession(ctx context.Context, attendeeID, sessionID string) (*AttendanceRecord, error)
	ListOpenBySession(ctx context.Context, sessionID string) ([]*AttendanceRecord, error)
	FindBySession(ctx context.Context, sessionID string) ([]*AttendanceRecord, error)
}

// AttendeeSessionStatusRepository persists the derived per-session
// statuses.
type AttendeeSessionStatusRepository interface {
	Upsert(ctx context.Context, s *AttendeeSessionStatus) error
	FindByAttendeeSession(ctx context.Context, attendeeID, sessionID string) (*AttendeeSessionStatus, error)
	ListBySession(ctx context.Context, sessionID string) ([]*AttendeeSessionStatus, error)
	ListByAttendee(ctx context.Context, attendeeID string) ([]*AttendeeSessionStatus, error)
	ListCertEligibleBySession(ctx context.Context, sessionID string) ([]*AttendeeSessionStatus, error)
}

// AttendeeRollupStatusRepository persists the roll-up statuses.
// The roll-up is the attendee's overall status across all sessions
// under a given external reference (event, course, etc.).
type AttendeeRollupStatusRepository interface {
	Upsert(ctx context.Context, s *AttendeeRollupStatus) error
	FindByAttendeeExternal(ctx context.Context, attendeeID string, ref ExternalRef) (*AttendeeRollupStatus, error)
	ListByExternal(ctx context.Context, ref ExternalRef) ([]*AttendeeRollupStatus, error)
	ListEventAttendees(ctx context.Context, q ListEventAttendeesQuery) (*ListEventAttendeesResult, error)
	FindEventAttendee(ctx context.Context, eventID, attendeeID string) (*EventAttendeeRow, error)

	// ListAttendees is the cross-event directory. Scoped to the given
	// account IDs — the caller's accounts. Returns zero rows whe
	// AccountIDs is empty.
	ListAttendees(ctx context.Context, q ListAttendeesQuery) (*ListAttendeesResult, error)
}

// AttendanceOverrideRepository persists the audit trail of manual
// status changes.
type AttendanceOverrideRepository interface {
	Create(ctx context.Context, o *AttendanceOverride) error
	ListByAttendeeSession(ctx context.Context, attendeeID, sessionID string) ([]*AttendanceOverride, error)
}

// ============================================================
// UNIT OF WORK
// ============================================================

// Repositories groups all repositories for use inside a transaction.
type Repositories struct {
	Attendees       AttendeeRepository
	Sessions        SessionRepository
	JoinTokens      JoinTokenRepository
	Records         AttendanceRecordRepository
	SessionStatuses AttendeeSessionStatusRepository
	RollupStatuses  AttendeeRollupStatusRepository
	Overrides       AttendanceOverrideRepository
}

// UnitOfWork executes a function inside a database transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(Repositories) error) error
}




// ============================================================
// EVENT ATTENDEE QUERIES
// ============================================================

type ListEventAttendeesQuery struct {
	EventID   string
	Search    string     // matches display_name or email (ILIKE)
	Statuses  []string   // filter on derived_status; empty = all
	SortBy    string     // "name" | "registered_at" | "status" | "duration"
	SortOrder string     // "asc" | "desc"
	Page      int        // 1-indexed
	PageSize  int        // clamped server-side
}

type ListEventAttendeesResult struct {
	Attendees []*EventAttendeeRow
	Total     int
	Page      int
	PageSize  int
}

// EventAttendeeRow is the read model for the attendee list and the
// attendee detail endpoint.
type EventAttendeeRow struct {
	AttendeeID            string
	DisplayName           string
	Email                 string
	DerivedStatus         string
	SessionsTotal         int
	SessionsAttended      int
	SessionsConfirmed     int
	TotalDurationSeconds  int
	LastDerivedAt         time.Time
	RegisteredAt          time.Time
	IsHost                bool   // ← add
	Phone  string
}