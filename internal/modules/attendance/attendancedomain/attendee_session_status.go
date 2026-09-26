package attendancedomain

import "time"

// AttendeeSessionStatus is the derived state for one (attendee,
// session) pair.
//
// Two independent facts are stored side by side:
//
//   - DerivedStatus — what the raw attendance records imply. This is
//     recomputed whenever records change.
//   - ConfirmedStatus — the final status a host explicitly chose. This
//     overrides the derivation. Empty string when no override is
//     present.
//
// EffectiveStatus returns the host's choice if present, otherwise the
// derivation. Recomputation never touches the host fields, so an
// override survives any number of recomputes.
type AttendeeSessionStatus struct {
	AttendeeID           string
	SessionID            string
	DerivedStatus        AttendanceStatus
	TotalDurationSeconds int64

	HostConfirmed   bool
	ConfirmedStatus AttendanceStatus // "" when no override
	ConfirmedBy     string
	ConfirmedAt     *time.Time
	ConfirmReason   string

	LastDerivedAt time.Time
}

// NewAttendeeSessionStatus constructs a fresh status for a
// (attendee, session) pair with no attendance yet.
func NewAttendeeSessionStatus(
	attendeeID, sessionID string,
	now time.Time,
) *AttendeeSessionStatus {
	return &AttendeeSessionStatus{
		AttendeeID:    attendeeID,
		SessionID:     sessionID,
		DerivedStatus: StatusRegistered,
		LastDerivedAt: now,
	}
}

// EffectiveStatus returns the status that consumers should trust.
//
//   - If a host override is present, that status wins.
//   - Otherwise, the derived status.
func (s *AttendeeSessionStatus) EffectiveStatus() AttendanceStatus {
	if s.HostConfirmed && s.ConfirmedStatus != "" {
		return s.ConfirmedStatus
	}
	return s.DerivedStatus
}

// CertEligible reports whether this status qualifies for a certificate.
func (s *AttendeeSessionStatus) CertEligible() bool {
	switch s.EffectiveStatus() {
	case StatusFull, StatusConfirmed:
		return true
	}
	return false
}

// Confirm records a generic host confirmation: "this attendee
// attended". The effective status becomes StatusConfirmed.
//
// Idempotent for the same actor.
func (s *AttendeeSessionStatus) Confirm(actorID, reason string, now time.Time) {
	s.HostConfirmed = true
	s.ConfirmedStatus = StatusConfirmed
	s.ConfirmedBy = actorID
	s.ConfirmedAt = &now
	s.ConfirmReason = reason
	s.LastDerivedAt = now
}

// Override records a specific host-chosen status. The effective status
// becomes the one the host chose.
//
// Use this when the host wants to say something more specific than
// "confirmed" — for example, "partial" for a late joiner, or "no_show"
// for a registration that shouldn't count.
func (s *AttendeeSessionStatus) Override(
	status AttendanceStatus,
	actorID, reason string,
	now time.Time,
) {
	s.HostConfirmed = true
	s.ConfirmedStatus = status
	s.ConfirmedBy = actorID
	s.ConfirmedAt = &now
	s.ConfirmReason = reason
	s.LastDerivedAt = now
}

// Recompute updates the derived fields from fresh evidence.
//
// Host override fields (HostConfirmed, ConfirmedStatus, ConfirmedBy,
// ConfirmedAt, ConfirmReason) are intentionally preserved — a host
// override is not a derived fact, and recomputation must not undo it.
func (s *AttendeeSessionStatus) Recompute(
	derived AttendanceStatus,
	totalDurationSeconds int64,
	now time.Time,
) {
	s.DerivedStatus = derived
	s.TotalDurationSeconds = totalDurationSeconds
	s.LastDerivedAt = now
}