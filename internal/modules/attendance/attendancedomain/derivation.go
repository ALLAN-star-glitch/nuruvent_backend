package attendancedomain

import (
	"fmt"
	"time"
)

// DerivationPolicy captures the rules used to compute attendance
// status from records. Passed as a value so tests can vary it.
type DerivationPolicy struct {
	// FullThresholdPercent is the percent of scheduled duration an
	// attendee must be present to qualify as Full. E.g. 80 means
	// "80% of the scheduled duration".
	FullThresholdPercent int

	// PartialThresholdSeconds is the minimum presence to qualify as
	// Partial. Below this, the attendee is treated as NoShow.
	PartialThresholdSeconds int64

	// GraceSeconds is added to a session's scheduled end when
	// deciding whether a session is final (for NoShow determination).
	GraceSeconds int64
}

// DefaultDerivationPolicy returns the platform default.
func DefaultDerivationPolicy() DerivationPolicy {
	return DerivationPolicy{
		FullThresholdPercent:    80,
		PartialThresholdSeconds: 60,
		GraceSeconds:            30 * 60,
	}
}

// Validate checks the policy is well-formed.
func (p DerivationPolicy) Validate() error {
	if p.FullThresholdPercent <= 0 || p.FullThresholdPercent > 100 {
		return fmt.Errorf("FullThresholdPercent must be in (0, 100]")
	}
	if p.PartialThresholdSeconds < 0 {
		return fmt.Errorf("PartialThresholdSeconds must be >= 0")
	}
	if p.GraceSeconds < 0 {
		return fmt.Errorf("GraceSeconds must be >= 0")
	}
	return nil
}

// DeriveSessionStatus computes the current status for (attendee,
// session) from its records.
//
// The rule is:
//   - No records yet: Registered (session not started) or NoShow
//     (session ended and attendee never joined).
//   - At least one open record: Joined.
//   - At least one closed record but total duration below partial
//     threshold: NoShow.
//   - Total duration between partial and full thresholds: Partial.
//   - Total duration at or above full threshold: Full.
//
// A host confirmation is not represented here — the caller layers it
// on top via AttendeeSessionStatus.Confirm.
func DeriveSessionStatus(
	session *Session,
	records []*AttendanceRecord,
	policy DerivationPolicy,
	now time.Time,
) AttendanceStatus {
	totalDuration := int64(0)
	hasOpen := false

	for _, r := range records {
		if r.IsOpen() {
			hasOpen = true
			if now.After(r.JoinTime) {
				totalDuration += int64(now.Sub(r.JoinTime).Seconds())
			}
		} else {
			totalDuration += r.DurationSeconds
		}
	}

	// Session hasn't started yet.
	if now.Before(session.ScheduledStart) {
		if len(records) == 0 {
			return StatusRegistered
		}
		return StatusJoined
	}

	// Session is live.
	if now.Before(session.ScheduledEnd.Add(time.Duration(policy.GraceSeconds) * time.Second)) {
		if hasOpen {
			return StatusJoined
		}
		if totalDuration == 0 {
			return StatusRegistered
		}
		return StatusJoined
	}

	// Session has ended.
	if totalDuration < policy.PartialThresholdSeconds {
		return StatusNoShow
	}

	thresholdSeconds := int64(session.DurationMinutes) * 60 * int64(policy.FullThresholdPercent) / 100
	if totalDuration >= thresholdSeconds {
		return StatusFull
	}
	return StatusPartial
}

// DeriveRollupStatus rolls up session statuses into a parent-entity
// level status. Sessions with no records yet are treated as
// Registered.
//
// The roll-up rule:
//   - If sessions attended >= FullThresholdPercent of total, Full.
//   - Else if any session was attended, Partial.
//   - Else NoShow (after all sessions have ended).
//
// Returns (status, sessionsAttended, sessionsConfirmed).
func DeriveRollupStatus(
	sessions []*Session,
	statuses map[string]AttendanceStatus, // sessionID -> status
	confirmed map[string]bool,             // sessionID -> hostConfirmed
	policy DerivationPolicy,
	now time.Time,
) (AttendanceStatus, int, int) {
	total := len(sessions)
	attended := 0
	confirmedCount := 0
	anyAttended := false

	for _, s := range sessions {
		st, ok := statuses[s.ID]
		if !ok {
			continue
		}
		if confirmed[s.ID] {
			confirmedCount++
		}
		switch st {
		case StatusFull, StatusConfirmed, StatusPartial, StatusJoined:
			attended++
			anyAttended = true
		case StatusRegistered, StatusNoShow:
			// not counted as attended
		}
	}

	if total == 0 {
		return StatusRegistered, 0, 0
	}

	// Determine if all sessions have ended.
	allEnded := true
	for _, s := range sessions {
		if now.Before(s.ScheduledEnd.Add(time.Duration(policy.GraceSeconds) * time.Second)) {
			allEnded = false
			break
		}
	}

	if !allEnded {
		if anyAttended {
			return StatusJoined, attended, confirmedCount
		}
		return StatusRegistered, 0, confirmedCount
	}

	threshold := total * policy.FullThresholdPercent / 100
	if threshold == 0 {
		threshold = 1
	}
	if attended >= threshold {
		return StatusFull, attended, confirmedCount
	}
	if anyAttended {
		return StatusPartial, attended, confirmedCount
	}
	return StatusNoShow, 0, confirmedCount
}

// SumDuration returns the total seconds across all records, using
// now for open records.
func SumDuration(records []*AttendanceRecord, now time.Time) int64 {
	var total int64
	for _, r := range records {
		if r.IsOpen() {
			if now.After(r.JoinTime) {
				total += int64(now.Sub(r.JoinTime).Seconds())
			}
		} else {
			total += r.DurationSeconds
		}
	}
	return total
}