// internal/modules/attendance/service/bulk_confirm.go

package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// BulkConfirm confirms several attendees in one call.
//
// Same semantics as ConfirmAttendance, applied to a list. Each
// attendee's status is confirmed inside a single transaction; if any
// fails, the whole batch rolls back.
//
// Returns nil if every attendee in the list was confirmed (or already
// confirmed). Returns an error on the first failure.
func (s *attendanceService) BulkConfirm(
	ctx context.Context,
	cmd BulkConfirmCommand,
) error {
	if cmd.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if cmd.ActorID == "" {
		return fmt.Errorf("%w: actor_id is required", attendance.ErrUnauthorized)
	}
	if len(cmd.AttendeeIDs) == 0 {
		return nil
	}

	now := s.deps.Clock.Now()
	var rollupChanges []RollupStatusChange
	var externalRef attendance.ExternalRef
	var sessionLoaded bool

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Load the session once — we need its external ref for the
		// roll-up recomputes below.
		session, err := repos.Sessions.FindByID(ctx, cmd.SessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}
		externalRef = session.External
		sessionLoaded = true

		for _, attendeeID := range cmd.AttendeeIDs {
			st, err := repos.SessionStatuses.FindByAttendeeSession(ctx, attendeeID, cmd.SessionID)
			if err != nil {
				if errors.Is(err, attendance.ErrStatusNotFound) {
					return fmt.Errorf("%w: attendee %s is not registered for this session",
						attendance.ErrStatusNotFound, attendeeID)
				}
				return fmt.Errorf("load session status for %s: %w", attendeeID, err)
			}

			if st.HostConfirmed && st.ConfirmedStatus == attendance.StatusConfirmed {
				// Already confirmed; skip.
				continue
			}

			priorStatus := st.EffectiveStatus()
			st.Confirm(cmd.ActorID, cmd.Reason, now)

			if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
				return fmt.Errorf("persist status for %s: %w", attendeeID, err)
			}

			override, err := attendance.NewAttendanceOverride(
				s.deps.IDs.NewID(),
				attendeeID,
				cmd.SessionID,
				cmd.ActorID,
				priorStatus,
				attendance.StatusConfirmed,
				cmd.Reason,
				now,
			)
			if err != nil {
				return fmt.Errorf("build override for %s: %w", attendeeID, err)
			}
			if err := repos.Overrides.Create(ctx, override); err != nil {
				return fmt.Errorf("persist override for %s: %w", attendeeID, err)
			}
		}

		// Recompute roll-ups for every attendee in the batch. This is
		// done after all statuses are written so we don't recompute
		// multiple times per attendee when the same list contains them
		// twice.
		seen := make(map[string]struct{}, len(cmd.AttendeeIDs))
		for _, attendeeID := range cmd.AttendeeIDs {
			if _, dup := seen[attendeeID]; dup {
				continue
			}
			seen[attendeeID] = struct{}{}

			change, err := s.recomputeRollupTx(ctx, repos, attendeeID, externalRef, now)
			if err != nil {
				return fmt.Errorf("recompute rollup for %s: %w", attendeeID, err)
			}
			if change != nil {
				rollupChanges = append(rollupChanges, *change)
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	_ = sessionLoaded // silence linter if it complains

	for _, change := range rollupChanges {
		_ = s.deps.Publisher.RollupStatusChanged(ctx, change)
	}

	return nil
}