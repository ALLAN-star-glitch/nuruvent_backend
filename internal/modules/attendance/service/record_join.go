// internal/modules/attendance/service/record_join.go

package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RecordJoin records a join event for an attendee in a session.
//
// This is the primitive. Callers include:
//   - RedeemJoinToken (source = SourceJoinLink)
//   - IngestWebhook (source = SourceZoomWebhook / SourceGoogleEvent)
//   - Host manual marking (source = SourceHostManual)
//
// Does not trigger recomputation. The caller is responsible for
// calling RecomputeSessionStatuses after this succeeds, once the
// outer transaction (if any) has committed.
//
// Idempotency is not enforced here — every call creates a new record.
// That's correct: an attendee can join multiple times (reconnects,
// dual devices) and each interval should be recorded.
func (s *attendanceService) RecordJoin(
	ctx context.Context,
	cmd RecordJoinCommand,
) (*attendance.AttendanceRecord, error) {
	if cmd.AttendeeID == "" {
		return nil, fmt.Errorf("attendee_id is required")
	}
	if cmd.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if cmd.JoinTime.IsZero() {
		return nil, fmt.Errorf("join_time is required")
	}
	if !cmd.Source.IsValid() {
		return nil, fmt.Errorf("%w: invalid source %q", attendance.ErrInvalidAttendance, cmd.Source)
	}

	now := s.deps.Clock.Now()

	var record *attendance.AttendanceRecord

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Verify both entities exist. This is a cheap existence check
		// (primary key lookup) and prevents orphaned records if the
		// caller passed bad IDs.
		if _, err := repos.Attendees.FindByID(ctx, cmd.AttendeeID); err != nil {
			return fmt.Errorf("load attendee: %w", err)
		}
		if _, err := repos.Sessions.FindByID(ctx, cmd.SessionID); err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		newRecord, err := attendance.NewAttendanceRecord(
			s.deps.IDs.NewID(),
			cmd.AttendeeID,
			cmd.SessionID,
			cmd.JoinTime,
			cmd.Source,
			now,
		)
		if err != nil {
			return err
		}

		if err := repos.Records.Create(ctx, newRecord); err != nil {
			return fmt.Errorf("persist record: %w", err)
		}

		record = newRecord
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return record, nil
}