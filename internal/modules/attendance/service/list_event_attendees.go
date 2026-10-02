package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ListEventAttendees returns a paginated attendee list for an event.
//
// Reads from the pre-computed attendee_rollup_statuses table — one
// row per (attendee, event). The rollup is maintained by
// RecomputeRollup whenever a session status changes.
func (s *attendanceService) ListEventAttendees(
	ctx context.Context,
	cmd ListEventAttendeesCommand,
) (*ListEventAttendeesResult, error) {

	if cmd.EventID == "" {
		return nil, attendance.ErrInvalidSession
	}

	q := attendance.ListEventAttendeesQuery{
		EventID:   cmd.EventID,
		Search:    cmd.Search,
		Statuses:  cmd.Statuses,
		SortBy:    cmd.SortBy,
		SortOrder: cmd.SortOrder,
		Page:      cmd.Page,
		PageSize:  cmd.PageSize,
	}

	var res *attendance.ListEventAttendeesResult
	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		r, err := repos.RollupStatuses.ListEventAttendees(ctx, q)
		if err != nil {
			return fmt.Errorf("list event attendees: %w", err)
		}
		res = r
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := &ListEventAttendeesResult{
		Attendees: make([]*EventAttendeeListItem, 0, len(res.Attendees)),
		Total:     res.Total,
		Page:      res.Page,
		PageSize:  res.PageSize,
	}
	for _, a := range res.Attendees {
		out.Attendees = append(out.Attendees, &EventAttendeeListItem{
			AttendeeID:           a.AttendeeID,
			DisplayName:          a.DisplayName,
			Email:                a.Email,
			EffectiveStatus:      attendance.AttendanceStatus(a.DerivedStatus),
			SessionsTotal:        a.SessionsTotal,
			SessionsAttended:     a.SessionsAttended,
			SessionsConfirmed:    a.SessionsConfirmed,
			TotalDurationSeconds: a.TotalDurationSeconds,
			RegisteredAt:         a.RegisteredAt,
			LastActivityAt:       a.LastDerivedAt,
		})
	}
	return out, nil
}