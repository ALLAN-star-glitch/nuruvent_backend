package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// GetEventAttendeeDetail returns one attendee's rollup plus their
// per-session statuses for an event.
func (s *attendanceService) GetEventAttendeeDetail(
	ctx context.Context,
	cmd GetEventAttendeeDetailCommand,
) (*EventAttendeeDetail, error) {

	if cmd.EventID == "" || cmd.AttendeeID == "" {
		return nil, attendance.ErrInvalidSession
	}

	ref := attendance.ExternalRef{Type: "event", ID: cmd.EventID}

	var (
		rollup   *attendance.EventAttendeeRow
		statuses []*attendance.AttendeeSessionStatus
		sessions []*attendance.Session
	)

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		r, err := repos.RollupStatuses.FindEventAttendee(ctx, cmd.EventID, cmd.AttendeeID)
		if err != nil {
			return fmt.Errorf("find event attendee: %w", err)
		}
		rollup = r

		st, err := repos.SessionStatuses.ListByAttendee(ctx, cmd.AttendeeID)
		if err != nil {
			return fmt.Errorf("list session statuses: %w", err)
		}
		statuses = st

		ss, err := repos.Sessions.ListByExternalRef(ctx, ref)
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}
		sessions = ss
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Index sessions so we can join statuses to their session details.
	sessionByID := make(map[string]*attendance.Session, len(sessions))
	for _, sess := range sessions {
		sessionByID[sess.ID] = sess
	}

	detail := &EventAttendeeDetail{
		AttendeeID:           rollup.AttendeeID,
		DisplayName:          rollup.DisplayName,
		Email:                rollup.Email,
		EffectiveStatus:      attendance.AttendanceStatus(rollup.DerivedStatus),
		SessionsTotal:        rollup.SessionsTotal,
		SessionsAttended:     rollup.SessionsAttended,
		SessionsConfirmed:    rollup.SessionsConfirmed,
		TotalDurationSeconds: rollup.TotalDurationSeconds,
		RegisteredAt:         rollup.RegisteredAt,
		LastActivityAt:       rollup.LastDerivedAt,
		Sessions:             make([]EventAttendeeSessionDetail, 0, len(statuses)),
	}

	for _, st := range statuses {
		sess, ok := sessionByID[st.SessionID]
		if !ok {
			// Status row for a session not under this event — skip.
			continue
		}
		detail.Sessions = append(detail.Sessions, EventAttendeeSessionDetail{
			SessionID:        st.SessionID,
			Title:            sess.Title,
			Provider:         sess.Provider,
			ScheduledStart:   sess.ScheduledStart,
			ScheduledEnd:     sess.ScheduledEnd,
			DerivedStatus:    st.DerivedStatus,
			HostConfirmed:    st.HostConfirmed,
			TotalDurationSec: int(st.TotalDurationSeconds),
			LastDerivedAt:    st.LastDerivedAt,
		})
	}

	return detail, nil
}