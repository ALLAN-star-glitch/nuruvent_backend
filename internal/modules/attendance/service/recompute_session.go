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
//
// Rollup recompute is triggered by *either* a status change or a
// duration change. A late joiner can extend their duration without
// changing status — the rollup still needs to reflect the new total.
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
	var statusChanges []SessionStatusChange
	var rollupAttendeeIDs []string

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		session, err := repos.Sessions.FindByID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}
		sessionExternal = session.External

		statuses, err := repos.SessionStatuses.ListBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("list session statuses: %w", err)
		}

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
			oldDuration := st.TotalDurationSeconds

			st.Recompute(derived, totalDuration, now)
			newStatus := st.EffectiveStatus()

			if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
				return fmt.Errorf("upsert status for %s: %w", st.AttendeeID, err)
			}

			statusChanged := oldStatus != newStatus
			durationChanged := oldDuration != st.TotalDurationSeconds

			if statusChanged {
				attendee, err := repos.Attendees.FindByID(ctx, st.AttendeeID)
				if err != nil {
					return fmt.Errorf("load attendee %s: %w", st.AttendeeID, err)
				}
				statusChanges = append(statusChanges, SessionStatusChange{
					AttendeeID:       st.AttendeeID,
					AttendeeExternal: attendee.External,
					SessionID:        session.ID,
					SessionExternal:  session.External,
					OldStatus:        oldStatus,
					NewStatus:        newStatus,
					OccurredAt:       now,
				})
			}

			// Rollup recompute is needed whenever either fact changed.
			if statusChanged || durationChanged {
				rollupAttendeeIDs = append(rollupAttendeeIDs, st.AttendeeID)
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	// Publish status changes after commit. Best-effort.
	for _, change := range statusChanges {
		if err := s.deps.Publisher.SessionStatusChanged(ctx, change); err != nil {
			_ = err
		}
	}

	// Recompute rollups for every attendee whose status OR duration
	// moved. Deduplicate — an attendee may appear in both categories.
	seen := make(map[string]struct{}, len(rollupAttendeeIDs))
	for _, id := range rollupAttendeeIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if err := s.RecomputeRollup(ctx, id, sessionExternal); err != nil {
			_ = err
		}
	}

	return nil
}