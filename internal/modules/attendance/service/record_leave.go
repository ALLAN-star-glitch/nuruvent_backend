// internal/modules/attendance/service/record_leave.go

package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RecordLeave closes the most recent open record for an attendee in a
// session.
//
// If no open record exists, the call is a no-op and returns nil. This
// makes it safe to call repeatedly — webhook redelivery, out-of-order
// delivery, and duplicate signals all produce the same result.
//
// If the leave time is not after the open record's join time (clock
// skew, out-of-order webhook), the record is closed at its join time
// instead. This preserves the invariant "duration >= 0" without
// rejecting the leave event outright.
//
// Does not trigger recomputation. The caller is responsible for
// calling RecomputeSessionStatuses after this succeeds.
func (s *attendanceService) RecordLeave(
	ctx context.Context,
	cmd RecordLeaveCommand,
) error {
	if cmd.AttendeeID == "" {
		return fmt.Errorf("attendee_id is required")
	}
	if cmd.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if cmd.LeaveTime.IsZero() {
		return fmt.Errorf("leave_time is required")
	}
	if !cmd.Source.IsValid() {
		return fmt.Errorf("%w: invalid source %q", attendance.ErrInvalidAttendance, cmd.Source)
	}

	now := s.deps.Clock.Now()

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		open, err := repos.Records.FindOpenByAttendeeSession(ctx, cmd.AttendeeID, cmd.SessionID)
		if err != nil {
			// No open record — nothing to close.
			if errors.Is(err, attendance.ErrInvalidAttendance) {
				return nil
			}
			return fmt.Errorf("find open record: %w", err)
		}

		// Adjust for clock skew: never let a leave precede its join.
		leaveTime := cmd.LeaveTime
		if !leaveTime.After(open.JoinTime) {
			leaveTime = open.JoinTime
		}

		if err := open.Close(leaveTime, now); err != nil {
			// A failed Close is a domain error — the leave time was
			// rejectable. Log and swallow: an invalid leave should not
			// break the surrounding operation.
			return nil
		}

		if err := repos.Records.Update(ctx, open); err != nil {
			return fmt.Errorf("update record: %w", err)
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	return nil
}