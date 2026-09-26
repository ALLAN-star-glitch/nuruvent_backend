// internal/modules/attendance/service/get_summary.go

package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// GetAttendeeSummary returns an attendee's status across every session
// they're registered for, plus their roll-ups for each parent entity.
//
// Used by:
//   - The learner dashboard ("show me all my attendance").
//   - The certificate module (checking eligibility before issuing).
//   - Admin tools (audit trail).
//
// Roll-ups for parents with no sessions yet are omitted — they carry
// no information.
func (s *attendanceService) GetAttendeeSummary(
	ctx context.Context,
	attendeeID string,
) (*AttendeeSummary, error) {
	if attendeeID == "" {
		return nil, fmt.Errorf("attendee_id is required")
	}

	var summary *AttendeeSummary

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Load the attendee.
		attendee, err := repos.Attendees.FindByID(ctx, attendeeID)
		if err != nil {
			return fmt.Errorf("load attendee: %w", err)
		}

		// Load every session status for this attendee.
		statuses, err := repos.SessionStatuses.ListByAttendee(ctx, attendeeID)
		if err != nil {
			return fmt.Errorf("list session statuses: %w", err)
		}

		// Collect the distinct external references across the
		// attendee's sessions. Each distinct ref is one roll-up to
		// load.
		seen := make(map[attendance.ExternalRef]struct{})
		for _, st := range statuses {
			session, err := repos.Sessions.FindByID(ctx, st.SessionID)
			if err != nil {
				return fmt.Errorf("load session for status: %w", err)
			}
			seen[session.External] = struct{}{}
		}

		rollups := make([]*attendance.AttendeeRollupStatus, 0, len(seen))
		for ref := range seen {
			rollup, err := repos.RollupStatuses.FindByAttendeeExternal(ctx, attendeeID, ref)
			if err != nil {
				// Missing roll-up is not an error — the roll-up may
				// not have been computed yet. Skip.
				continue
			}
			rollups = append(rollups, rollup)
		}

		summary = &AttendeeSummary{
			Attendee: attendee,
			Statuses: statuses,
			Rollups:  rollups,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return summary, nil
}