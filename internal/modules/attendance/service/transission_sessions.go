// internal/modules/attendance/service/transition_sessions.go

package service

import (
	"context"
	"fmt"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// TransitionSessions advances session lifecycle state based on the
// current time.
//
//   - Scheduled sessions whose scheduled start has passed → Live.
//   - Live (or overdue Scheduled) sessions whose scheduled end has
//     passed → Ended.
//
// Called by a background worker on a schedule (e.g. every minute). The
// operation is idempotent: calling it twice with the same time does
// nothing on the second call.
//
// After transitioning sessions to Ended, this method recomputes their
// statuses so NoShow is applied to attendees who never joined.
func (s *attendanceService) TransitionSessions(
	ctx context.Context,
	now time.Time,
) error {
	// 1. Find sessions that need to transition.
	//    - Sessions scheduled to start but not yet live.
	//    - Sessions that should have ended.
	var toTransition []*attendance.Session

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Active (Scheduled or Live) sessions whose scheduled start
		// has already passed. Some of these may transition to Live;
		// some may transition directly to Ended if both start and end
		// are in the past.
		active, err := repos.Sessions.FindActiveBySchedule(ctx, now)
		if err != nil {
			return fmt.Errorf("find active sessions: %w", err)
		}
		toTransition = active
		return nil
	})
	if txErr != nil {
		return txErr
	}

	// 2. Transition each one.
	for _, session := range toTransition {
		var transitioned bool
		var newStatus attendance.SessionStatus

		txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
			// Reload to get the latest state — another worker may have
			// transitioned it in the meantime.
			current, err := repos.Sessions.FindByID(ctx, session.ID)
			if err != nil {
				return fmt.Errorf("reload session %s: %w", session.ID, err)
			}

			before := current.Status
			current.TransitionTo(now)
			if current.Status == before {
				return nil // no change
			}

			if err := repos.Sessions.Update(ctx, current); err != nil {
				return fmt.Errorf("persist session %s: %w", session.ID, err)
			}

			transitioned = true
			newStatus = current.Status
			return nil
		})
		if txErr != nil {
			return txErr
		}

		// 3. When a session transitions to Ended, recompute its
		//    statuses so NoShow is applied to attendees who never
		//    joined. Live transitions don't need recomputation — the
		//    statuses were just updated by whatever triggered the
		//    session to go live.
		if transitioned && newStatus == attendance.SessionStatusEnded {
			if err := s.RecomputeSessionStatuses(ctx, session.ID); err != nil {
				// Best-effort. A subsequent run will retry.
				_ = err
			}
		}
	}

	return nil
}