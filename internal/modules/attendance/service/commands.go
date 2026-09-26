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

// RedeemResult is the outcome of redeeming a join token.
type RedeemResult struct {
	AttendeeID string
	SessionID  string
	RedirectTo string // provider URL to redirect the attendee to
}

// AttendeeSummary aggregates an attendee's status across sessions and
// their roll-up.
type AttendeeSummary struct {
	Attendee *attendance.Attendee
	Statuses []*attendance.AttendeeSessionStatus
	Rollups  []*attendance.AttendeeRollupStatus
}


// RegisterAttendeeForExternalCommand registers an attendee for every
// session under an external reference.
type RegisterAttendeeForExternalCommand struct {
	AttendeeID string
	External   attendance.ExternalRef
}