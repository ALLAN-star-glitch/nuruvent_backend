// internal/modules/attendance/service/list_attendance.go

package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ListSessionAttendance returns every attendee status row for a
// session.
//
// Returns a slice ordered by nothing in particular — the caller sorts
// if it wants a specific order. Empty slice if the session has no
// attendees yet.
func (s *attendanceService) ListSessionAttendance(
	ctx context.Context,
	sessionID string,
) ([]*attendance.AttendeeSessionStatus, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	var statuses []*attendance.AttendeeSessionStatus

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Verify the session exists so we can distinguish
		// "no attendees" from "bad session id".
		if _, err := repos.Sessions.FindByID(ctx, sessionID); err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		var err error
		statuses, err = repos.SessionStatuses.ListBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("list statuses: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return statuses, nil
}

// getSessionAttendance loads statuses for a session without the
// existence check. Used internally when the session is already known
// to exist.
func (s *attendanceService) getSessionAttendance(
	ctx context.Context,
	repos attendance.Repositories,
	sessionID string,
) ([]*attendance.AttendeeSessionStatus, error) {
	statuses, err := repos.SessionStatuses.ListBySession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, attendance.ErrStatusNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return statuses, nil
}