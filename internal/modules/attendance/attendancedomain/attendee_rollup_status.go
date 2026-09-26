package attendancedomain

import "time"

// AttendeeRollupStatus is the derived state across all sessions that
// share a common external reference.
//
// The reference is opaque to the attendance module: consumers use
// whatever (type, id) pair makes sense for their domain — an event,
// a course cohort, a certification track, or something else. The
// attendance module never interprets these strings.
//
// One row per (attendee, external reference). Recomputed whenever any
// of the reference's session statuses change.
type AttendeeRollupStatus struct {
	AttendeeID           string
	External             ExternalRef
	DerivedStatus        AttendanceStatus
	SessionsTotal        int
	SessionsAttended     int
	SessionsConfirmed    int
	TotalDurationSeconds int64
	LastDerivedAt        time.Time
}

// NewAttendeeRollupStatus constructs a fresh roll-up with no sessions
// yet.
func NewAttendeeRollupStatus(
	attendeeID string,
	external ExternalRef,
	now time.Time,
) *AttendeeRollupStatus {
	return &AttendeeRollupStatus{
		AttendeeID:    attendeeID,
		External:      external,
		DerivedStatus: StatusRegistered,
		LastDerivedAt: now,
	}
}

// Recompute updates the roll-up from fresh aggregation input.
func (s *AttendeeRollupStatus) Recompute(
	derived AttendanceStatus,
	sessionsTotal, sessionsAttended, sessionsConfirmed int,
	totalDurationSeconds int64,
	now time.Time,
) {
	s.DerivedStatus = derived
	s.SessionsTotal = sessionsTotal
	s.SessionsAttended = sessionsAttended
	s.SessionsConfirmed = sessionsConfirmed
	s.TotalDurationSeconds = totalDurationSeconds
	s.LastDerivedAt = now
}