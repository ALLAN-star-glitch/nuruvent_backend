// internal/modules/attendance/service/confirm.go

package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ConfirmAttendance marks an attendee as confirmed by a host for a
// given session.
//
// Idempotent. Writes an audit record on every state change.
func (s *attendanceService) ConfirmAttendance(
	ctx context.Context,
	cmd ConfirmAttendanceCommand,
) error {
	if cmd.AttendeeID == "" {
		return fmt.Errorf("attendee_id is required")
	}
	if cmd.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if cmd.ActorID == "" {
		return fmt.Errorf("%w: actor_id is required", attendance.ErrUnauthorized)
	}

	now := s.deps.Clock.Now()
	var rollupChange *RollupStatusChange

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		st, err := repos.SessionStatuses.FindByAttendeeSession(ctx, cmd.AttendeeID, cmd.SessionID)
		if err != nil {
			if errors.Is(err, attendance.ErrStatusNotFound) {
				return fmt.Errorf("%w: attendee is not registered for this session", attendance.ErrStatusNotFound)
			}
			return fmt.Errorf("load session status: %w", err)
		}

		// Idempotent: already confirmed to exactly this status.
		if st.HostConfirmed && st.ConfirmedStatus == attendance.StatusConfirmed {
			return nil
		}

		priorStatus := st.EffectiveStatus()
		st.Confirm(cmd.ActorID, cmd.Reason, now)

		if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
			return fmt.Errorf("persist status: %w", err)
		}

		override, err := attendance.NewAttendanceOverride(
			s.deps.IDs.NewID(),
			cmd.AttendeeID,
			cmd.SessionID,
			cmd.ActorID,
			priorStatus,
			attendance.StatusConfirmed,
			cmd.Reason,
			now,
		)
		if err != nil {
			return fmt.Errorf("build override: %w", err)
		}
		if err := repos.Overrides.Create(ctx, override); err != nil {
			return fmt.Errorf("persist override: %w", err)
		}

		session, err := repos.Sessions.FindByID(ctx, cmd.SessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		change, err := s.recomputeRollupTx(ctx, repos, cmd.AttendeeID, session.External, now)
		if err != nil {
			return fmt.Errorf("recompute rollup: %w", err)
		}
		rollupChange = change

		return nil
	})
	if txErr != nil {
		return txErr
	}

	if rollupChange != nil {
		_ = s.deps.Publisher.RollupStatusChanged(ctx, *rollupChange)
	}

	return nil
}