// internal/modules/attendance/service/recompute_rollup.go

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RecomputeRollup recomputes the roll-up status for one attendee and
// one external reference (an event, a course cohort, etc.).
//
// The external ref identifies the parent entity. Every session whose
// External matches the given ref participates in the roll-up.
//
// Idempotent.
func (s *attendanceService) RecomputeRollup(
	ctx context.Context,
	attendeeID string,
	ref attendance.ExternalRef,
) error {
	if attendeeID == "" {
		return fmt.Errorf("attendee_id is required")
	}
	if !ref.IsValid() {
		return fmt.Errorf("external reference is required")
	}

	now := s.deps.Clock.Now()
	var change *RollupStatusChange

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var err error
		change, err = s.recomputeRollupTx(ctx, repos, attendeeID, ref, now)
		return err
	})
	if txErr != nil {
		return txErr
	}

	if change != nil {
		_ = s.deps.Publisher.RollupStatusChanged(ctx, *change)
	}

	return nil
}

// recomputeRollupTx recomputes the roll-up for one attendee and one
// external reference inside an existing transaction.
//
// Returns the change, if any, so the caller can publish after commit.
func (s *attendanceService) recomputeRollupTx(
	ctx context.Context,
	repos attendance.Repositories,
	attendeeID string,
	ref attendance.ExternalRef,
	now time.Time,
) (*RollupStatusChange, error) {
	sessions, err := repos.Sessions.ListByExternalRef(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	statusMap := make(map[string]attendance.AttendanceStatus, len(sessions))
	confirmedMap := make(map[string]bool, len(sessions))
	var totalDuration int64

	for _, session := range sessions {
		st, err := repos.SessionStatuses.FindByAttendeeSession(ctx, attendeeID, session.ID)
		if err != nil {
			if errors.Is(err, attendance.ErrStatusNotFound) {
				statusMap[session.ID] = attendance.StatusRegistered
				continue
			}
			return nil, fmt.Errorf("load status: %w", err)
		}
		statusMap[session.ID] = st.DerivedStatus
		confirmedMap[session.ID] = st.HostConfirmed
		totalDuration += st.TotalDurationSeconds
	}

	derived, attended, confirmed := attendance.DeriveRollupStatus(
		sessions, statusMap, confirmedMap, s.deps.DerivationPolicy, now,
	)

	existing, err := repos.RollupStatuses.FindByAttendeeExternal(ctx, attendeeID, ref)
	if err != nil && !errors.Is(err, attendance.ErrStatusNotFound) {
		return nil, fmt.Errorf("load rollup: %w", err)
	}

	var oldStatus attendance.AttendanceStatus
	if existing != nil {
		oldStatus = existing.DerivedStatus
	} else {
		oldStatus = attendance.StatusRegistered
		existing = attendance.NewAttendeeRollupStatus(attendeeID, ref, now)
	}

	existing.Recompute(derived, len(sessions), attended, confirmed, totalDuration, now)
	if err := repos.RollupStatuses.Upsert(ctx, existing); err != nil {
		return nil, fmt.Errorf("upsert rollup: %w", err)
	}

	if oldStatus == derived {
		return nil, nil
	}
	return &RollupStatusChange{
		AttendeeID: attendeeID,
		External:   ref,
		OldStatus:  oldStatus,
		NewStatus:  derived,
		OccurredAt: now,
	}, nil
}