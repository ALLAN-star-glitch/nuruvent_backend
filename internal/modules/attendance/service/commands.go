// internal/modules/attendance/service/commands.go

package service

import (
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ============================================================
// CONSUMER-FACING
// ============================================================

// RegisterAttendeeCommand registers a new attendee for an external
// reference.
type RegisterAttendeeCommand struct {
	External    attendance.ExternalRef
	DisplayName string
	Email       string
	Phone 	 string
	Username    string
}

// UpsertSessionCommand creates or updates a session for an external
// reference.
type UpsertSessionCommand struct {
	External          attendance.ExternalRef
	ProviderSessionID string
	Title             string
	ScheduledStart    time.Time
	ScheduledEnd      time.Time
	Provider          attendance.SessionProvider
	ProviderMeetingID string
	ProviderURL       string

	// Denormalized from the parent event at sync time. Used to build
	// the frontend redirect URL on join-token redemption.
	EventDisplayName     string
	OrganizerDisplayName string
}

// RegisterAttendeeForExternalCommand registers an attendee for every
// session under an external reference.
//
// When PublicBaseURL is set, the service issues a fresh join token
// per session and builds a full join URL from it. The raw token is
// only available at issue time — it is never persisted, only its
// hash. The result carries the URLs back to the caller.
type RegisterAttendeeForExternalCommand struct {
	AttendeeID string
	External   attendance.ExternalRef

	// PublicBaseURL is the scheme+host used to build join URLs
	// ("https://nuruvent.com"). When empty, no URL is built; links
	// in the result still carry the platform meeting code so the
	// caller can construct one.
	PublicBaseURL string

	// LinkGrace is how long after a session ends the issued join
	// token remains valid. Zero means use the service default
	// (JoinTokenGrace).
	LinkGrace time.Duration
	// Phone is not on this command — the attendee already exists.
	// Phone is set when the attendee is first created.
}

// ============================================================
// JOIN TOKENS
// ============================================================

// IssueJoinTokenCommand issues a join token for an attendee in a
// session.
type IssueJoinTokenCommand struct {
	AttendeeID string
	SessionID  string

	// Grace is how long after the session's scheduled end the token
	// remains valid. Zero means use the configured default.
	Grace time.Duration

	// PublicBaseURL, when non-empty, is used to build a full join
	// URL. When empty, only the raw token is returned and the caller
	// is responsible for building the URL.
	PublicBaseURL string
}

// RevokeJoinTokensCommand revokes all tokens for a
// (attendee, session) pair.
type RevokeJoinTokensCommand struct {
	AttendeeID string
	SessionID  string
}

// ============================================================
// ATTENDANCE RECORDING
// ============================================================

// RecordJoinCommand records a join event.
type RecordJoinCommand struct {
	AttendeeID string
	SessionID  string
	JoinTime   time.Time
	Source     attendance.AttendanceSource
}

// RecordLeaveCommand records a leave event.
type RecordLeaveCommand struct {
	AttendeeID string
	SessionID  string
	LeaveTime  time.Time
	Source     attendance.AttendanceSource
}

// ============================================================
// HOST OPERATIONS
// ============================================================

// ConfirmAttendanceCommand marks an attendee as confirmed.
type ConfirmAttendanceCommand struct {
	AttendeeID string
	SessionID  string
	ActorID    string
	Reason     string
}

// OverrideAttendanceCommand corrects an attendee's derived status.
type OverrideAttendanceCommand struct {
	AttendeeID string
	SessionID  string
	ActorID    string
	NewStatus  attendance.AttendanceStatus
	Reason     string
}

// BulkConfirmCommand confirms several attendees in one call.
type BulkConfirmCommand struct {
	SessionID   string
	AttendeeIDs []string
	ActorID     string
	Reason      string
}

// ============================================================
// RESULTS
// ============================================================

// AttendeeSessionLink is one personalized join URL for one
// (attendee, session) pair. Returned to the caller at registration
// time so the confirmation email can include it.
//
// JoinURL is empty when the command did not carry a PublicBaseURL,
// or when the session has no provider meeting (in-person sessions).
// Callers that need a URL can always build one from MeetingCode +
// Platform.
type AttendeeSessionLink struct {
	SessionID   string
	MeetingCode string // bare platform code, e.g. "abc-defg-hij"
	Platform    attendance.SessionProvider
	JoinURL     string // https://nuruvent.com/join/<raw-token>
	ExpiresAt   time.Time
}

// RegisterAttendeeForExternalResult is what
// RegisterAttendeeForExternal returns. The caller uses AttendeeID for
// later lookups and Links for the confirmation email.
//
// Links is empty when the external ref has no sessions.
type RegisterAttendeeForExternalResult struct {
	AttendeeID string
	Links      []AttendeeSessionLink
}

// IssueJoinTokenResult is what IssueJoinToken returns. The raw token
// is only available here — it is never persisted, only its hash.
//
// JoinURL is set when the command carried a PublicBaseURL.
type IssueJoinTokenResult struct {
	RawToken  string
	ExpiresAt time.Time
	JoinURL   string // e.g. https://nuruvent.com/join/nrt_abc...
}

// RedeemResult is the outcome of redeeming a join token.
type RedeemResult struct {
	AttendeeID  string
	SessionID   string
	MeetingCode string // bare platform code
	Platform    attendance.SessionProvider
	RedeemedAt  time.Time
	RedirectTo  string // Nuruvent-hosted meeting URL
}

// AttendeeSummary aggregates an attendee's status across sessions and
// their roll-up.
type AttendeeSummary struct {
	Attendee *attendance.Attendee
	Statuses []*attendance.AttendeeSessionStatus
	Rollups  []*attendance.AttendeeRollupStatus
}


// ============================================================
// ATTENDEE DIRECTORY — commands and results
// ============================================================

type ListEventAttendeesCommand struct {
	UserID    string
	EventID   string
	Search    string
	Statuses  []string
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

type ListEventAttendeesResult struct {
	Attendees []*EventAttendeeListItem
	Total     int
	Page      int
	PageSize  int
}

// EventAttendeeListItem is one row in the event attendee list.
type EventAttendeeListItem struct {
	AttendeeID           string
	DisplayName          string
	Email                string
	EffectiveStatus      attendance.AttendanceStatus
	SessionsTotal        int
	SessionsAttended     int
	SessionsConfirmed    int
	TotalDurationSeconds int
	RegisteredAt         time.Time
	LastActivityAt       time.Time
	IsHost               bool   // ← add
	Phone				string
}

type GetEventAttendeeDetailCommand struct {
	UserID     string
	EventID    string
	AttendeeID string
}

type EventAttendeeDetail struct {
	AttendeeID           string
	DisplayName          string
	Email                string
	EffectiveStatus      attendance.AttendanceStatus
	SessionsTotal        int
	SessionsAttended     int
	SessionsConfirmed    int
	TotalDurationSeconds int
	RegisteredAt         time.Time
	LastActivityAt       time.Time
	IsHost               bool   // ← add
	Phone				string
	Sessions             []EventAttendeeSessionDetail
}

type EventAttendeeSessionDetail struct {
	SessionID        string
	Title            string
	Provider         attendance.SessionProvider
	ScheduledStart   time.Time
	ScheduledEnd     time.Time
	DerivedStatus    attendance.AttendanceStatus
	HostConfirmed    bool
	TotalDurationSec int
	LastDerivedAt    time.Time
}


// ============================================================
// CROSS-EVENT ATTENDEE DIRECTORY
// ============================================================

type ListAttendeesCommand struct {
	UserID    string
	EventID   string
	Search    string
	Statuses  []string
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

type ListAttendeesResult struct {
	Attendees []*CrossEventAttendeeItem
	Total     int
	Page      int
	PageSize  int
}

type CrossEventAttendeeItem struct {
	AttendeeID           string
	DisplayName          string
	Email                string
	EventID              string
	EventName            string
	EventSlug            string
	EventStartDate       time.Time
	EffectiveStatus      attendance.AttendanceStatus
	SessionsTotal        int
	SessionsAttended     int
	SessionsConfirmed    int
	TotalDurationSeconds int64
	RegisteredAt         time.Time
	LastActivityAt       time.Time
	IsHost               bool   // ← add
	Phone 				string
}


// RegisterHostAttendeeCommand creates or updates an attendee row for
// the host of an event.
//
// Hosts have no registration. Their row is created at event publish
// time so attendance for the host is tracked like any other
// participant.
type RegisterHostAttendeeCommand struct {
	EventID         string
	HostUserID      string
	HostDisplayName string
	HostEmail       string
	HostUsername    string
	HostPhone            string

	// HostGoogleMeetUserID is the host's Google user resource name
	// ("users/<id>"), pulled from their active video connection at
	// publish time. Empty when the host hasn't connected Google Meet.
	HostGoogleMeetUserID string
}