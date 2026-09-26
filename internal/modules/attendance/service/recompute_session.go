// internal/modules/attendance/service/recompute_session.go

package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RecomputeSessionStatuses recomputes the derived status for every
// attendee in a session, and refreshes their roll-ups.
//
// Called:
//   - After a join or leave is recorded (by the HTTP handler).
//   - By a background job that scans recently-active sessions.
//   - Manually by an admin tool.
//
// Idempotent: calling it twice with the same state produces the same
// result and no spurious publish events.
func (s *attendanceService) RecomputeSessionStatuses(
	ctx context.Context,
	sessionID string,
) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	now := s.deps.Clock.Now()

	// Captured from the transaction, used after commit.
	var sessionExternal attendance.ExternalRef
	var changes []SessionStatusChange

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// 1. Load the session.
		session, err := repos.Sessions.FindByID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}
		sessionExternal = session.External

		// 2. Load every status row for the session.
		statuses, err := repos.SessionStatuses.ListBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("list session statuses: %w", err)
		}

		// 3. Re-derive each attendee's status.
		for _, st := range statuses {
			records, err := repos.Records.FindByAttendeeSession(ctx, st.AttendeeID, st.SessionID)
			if err != nil {
				return fmt.Errorf("load records for %s: %w", st.AttendeeID, err)
			}

			derived := attendance.DeriveSessionStatus(
				session,
				records,
				s.deps.DerivationPolicy,
				now,
			)
			totalDuration := attendance.SumDuration(records, now)

			oldStatus := st.EffectiveStatus()
			st.Recompute(derived, totalDuration, now)
			newStatus := st.EffectiveStatus()

			if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
				return fmt.Errorf("upsert status for %s: %w", st.AttendeeID, err)
			}

			if oldStatus != newStatus {
				attendee, err := repos.Attendees.FindByID(ctx, st.AttendeeID)
				if err != nil {
					return fmt.Errorf("load attendee %s: %w", st.AttendeeID, err)
				}
				changes = append(changes, SessionStatusChange{
					AttendeeID:       st.AttendeeID,
					AttendeeExternal: attendee.External,
					SessionID:        session.ID,
					SessionExternal:  session.External,
					OldStatus:        oldStatus,
					NewStatus:        newStatus,
					OccurredAt:       now,
				})
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	// Publish changes after commit. Publish failures do not roll back
	// the status changes — the changes are authoritative, and the
	// publisher is best-effort. Real deployments would use an outbox
	// for reliable delivery.
	for _, change := range changes {
		if err := s.deps.Publisher.SessionStatusChanged(ctx, change); err != nil {
			_ = err // swallow; see comment above
		}
	}

	// Recompute roll-ups for every attendee whose status moved.
	for _, change := range changes {
		if err := s.RecomputeRollup(ctx, change.AttendeeID, sessionExternal); err != nil {
			_ = err // best-effort; a background job can re-run
		}
	}

	return nil
}