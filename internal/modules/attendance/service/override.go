// internal/modules/attendance/service/override.go

package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// OverrideAttendance sets a specific host-chosen status for an
// attendee in a session.
//
// Unlike ConfirmAttendance (which always sets "confirmed"),
// Override lets the host pick any status — for example, "partial" for
// a late joiner, or "no_show" for a registration that shouldn't count.
//
// Idempotent: if the effective status already equals the new status,
// no audit record is written.
func (s *attendanceService) OverrideAttendance(
	ctx context.Context,
	cmd OverrideAttendanceCommand,
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
	if !cmd.NewStatus.IsValid() {
		return fmt.Errorf("%w: invalid new_status %q", attendance.ErrInvalidAttendance, cmd.NewStatus)
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

		priorStatus := st.EffectiveStatus()
		if priorStatus == cmd.NewStatus {
			return nil // no-op
		}

		st.Override(cmd.NewStatus, cmd.ActorID, cmd.Reason, now)

		if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
			return fmt.Errorf("persist status: %w", err)
		}

		override, err := attendance.NewAttendanceOverride(
			s.deps.IDs.NewID(),
			cmd.AttendeeID,
			cmd.SessionID,
			cmd.ActorID,
			priorStatus,
			cmd.NewStatus,
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