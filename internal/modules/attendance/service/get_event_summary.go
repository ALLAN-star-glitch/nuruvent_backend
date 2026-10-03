// internal/modules/attendance/service/get_event_summary.go

package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// GetEventAttendanceSummaryCommand is the input to
// GetEventAttendanceSummary.
type GetEventAttendanceSummaryCommand struct {
	// UserID is the authenticated caller. Currently used only for
	// logging; ownership enforcement is a TODO.
	UserID string

	// EventID is the events-module UUID. Attendance resolves its
	// sessions via ExternalRef{Type: "event", ID: EventID}.
	EventID string
}

// SessionAttendanceSummary is the per-session aggregate.
type SessionAttendanceSummary struct {
	SessionID          string
	Title              string
	Provider           string
	ScheduledStart     time.Time
	ScheduledEnd       time.Time
	RegisteredCount    int
	AttendedCount      int
	AvgDurationSeconds int64
	HasAttendanceData  bool
	 // VideoMeetingID is the video_meetings.id for this session's
    // meeting, when one exists. Empty for in-person sessions and
    // when the session has no linked video meeting.
    VideoMeetingID     string // NEW
}

// EventAttendanceTotals is the cross-session aggregate.
type EventAttendanceTotals struct {
	TotalSessions          int
	SessionsWithAttendance int
	UniqueAttendees        int
	TotalAttendanceEvents  int
}

// EventAttendanceSummary is what the endpoint returns.
type EventAttendanceSummary struct {
	Sessions []SessionAttendanceSummary
	Totals   EventAttendanceTotals
}

// GetEventAttendanceSummary builds a per-session and total summary
// of attendance for every session under an event.
//
// Sessions are resolved by ExternalRef{Type: "event", ID: eventID}.
// Per-session aggregates come from the derived session statuses.
//
// The N+1 pattern is acceptable: an event typically has one to five
// sessions. If this becomes a hotspot, replace with a single
// aggregate query.
func (s *attendanceService) GetEventAttendanceSummary(
	ctx context.Context,
	cmd GetEventAttendanceSummaryCommand,
) (*EventAttendanceSummary, error) {
	if strings.TrimSpace(cmd.EventID) == "" {
		return nil, fmt.Errorf("%w: event id is required", attendance.ErrInvalidSession)
	}

	ref := attendance.ExternalRef{
		Type: "event",
		ID:   cmd.EventID,
	}

	// 1. Load every session under the event.
	var sessions []*attendance.Session
	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var txErr error
		sessions, txErr = repos.Sessions.ListByExternalRef(ctx, ref)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("get event summary: list sessions: %w", err)
	}

	summary := &EventAttendanceSummary{
		Sessions: make([]SessionAttendanceSummary, 0, len(sessions)),
		Totals: EventAttendanceTotals{
			TotalSessions: len(sessions),
		},
	}

	if len(sessions) == 0 {
		return summary, nil
	}

	// 2. For each session, aggregate its statuses.
	uniqueAttendees := make(map[string]struct{})

	for _, session := range sessions {
		var statuses []*attendance.AttendeeSessionStatus
		err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
			var txErr error
			statuses, txErr = repos.SessionStatuses.ListBySession(ctx, session.ID)
			return txErr
		})
		if err != nil {
			return nil, fmt.Errorf(
				"get event summary: list statuses session=%s: %w",
				session.ID, err,
			)
		}

		row := SessionAttendanceSummary{
			SessionID:      session.ID,
			Title:          session.Title,
			Provider:       string(session.Provider),
			ScheduledStart: session.ScheduledStart,
			ScheduledEnd:   session.ScheduledEnd,
		}

		var (
			attended int
			totalDur int64
			withDur  int
		)

				for _, st := range statuses {
			// Hosts appear in the session roster but are not
			// "registered attendees" — exclude from RegisteredCount
			// and from UniqueAttendees (audience size). They do
			// count toward AttendedCount and average duration, since
			// they were in the room.
			if !st.IsHost {
				row.RegisteredCount++
			}

			if st.DerivedStatus != attendance.StatusRegistered {
				attended++
				if !st.IsHost {
					uniqueAttendees[st.AttendeeID] = struct{}{}
				}
			}

			if st.TotalDurationSeconds > 0 {
				totalDur += st.TotalDurationSeconds
				withDur++
			}
		}

		row.AttendedCount = attended
		if withDur > 0 {
			row.AvgDurationSeconds = totalDur / int64(withDur)
		}



		


		// Resolve the video meeting for this session, if one exists.
		// Non-fatal: in-person sessions and sessions created without
		// a linked video meeting legitimately have none.
		if s.deps.VideoMeetings != nil &&
			session.ProviderMeetingID != "" &&
			session.Provider != "" {

			if id, err := s.deps.VideoMeetings.ResolveMeetingID(
				ctx, session.Provider, session.ProviderMeetingID,
			); err == nil && id != "" {
				row.VideoMeetingID = id
			}
		}
		row.HasAttendanceData = attended > 0

		if row.HasAttendanceData {
			summary.Totals.SessionsWithAttendance++
		}
		summary.Totals.TotalAttendanceEvents += row.AttendedCount
		summary.Sessions = append(summary.Sessions, row)
	}

	summary.Totals.UniqueAttendees = len(uniqueAttendees)

	return summary, nil
}