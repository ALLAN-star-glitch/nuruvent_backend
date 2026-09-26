// internal/modules/attendance/service/register_attendee_for_external.go

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RegisterAttendeeForExternal registers an attendee for every session
// currently under the given external reference.
//
// Idempotent: an existing status row for a session is left unchanged.
// New rows are created with derived_status = registered.
//
// Called by the events module after a registration is confirmed. The
// attendance module doesn't know what an "event" is — the external
// reference is opaque.
func (s *attendanceService) RegisterAttendeeForExternal(
	ctx context.Context,
	cmd RegisterAttendeeForExternalCommand,
) error {
	if cmd.AttendeeID == "" {
		return fmt.Errorf("attendee_id is required")
	}
	if !cmd.External.IsValid() {
		return fmt.Errorf("external reference is required")
	}

	now := s.deps.Clock.Now()

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Verify the attendee exists.
		if _, err := repos.Attendees.FindByID(ctx, cmd.AttendeeID); err != nil {
			return fmt.Errorf("load attendee: %w", err)
		}

		// Load every session under this external reference.
		sessions, err := repos.Sessions.ListByExternalRef(ctx, cmd.External)
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}

		// For each session, create a status row if one doesn't exist.
		for _, session := range sessions {
			existing, err := repos.SessionStatuses.FindByAttendeeSession(ctx, cmd.AttendeeID, session.ID)
			if err == nil && existing != nil {
				// Already registered for this session; leave as is.
				continue
			}
			if err != nil && !errors.Is(err, attendance.ErrStatusNotFound) {
				return fmt.Errorf("check session status: %w", err)
			}

			st := attendance.NewAttendeeSessionStatus(cmd.AttendeeID, session.ID, now)
			if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
				return fmt.Errorf("upsert status: %w", err)
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	return nil
}

// ensureSessionRoster creates status rows for every attendee already
// registered under a session's external reference.
//
// Called from UpsertSession when a new session is created, so that
// attendees who registered before the session existed don't miss it.
func (s *attendanceService) ensureSessionRoster(
	ctx context.Context,
	repos attendance.Repositories,
	session *attendance.Session,
	now time.Time,
) error {
	// Find all attendees who have any status row under a sibling
	// session of the same external reference. That's the set of
	// "attendees registered for this external entity."
	//
	// For MVP we implement this by asking the rollup status
	// repository — but it doesn't have a "list by external" method
	// that returns all attendees. Instead, we walk siblings.
	siblings, err := repos.Sessions.ListByExternalRef(ctx, session.External)
	if err != nil {
		return fmt.Errorf("list sibling sessions: %w", err)
	}

	// Collect distinct attendee IDs across sibling sessions.
	seen := make(map[string]struct{})
	for _, sib := range siblings {
		if sib.ID == session.ID {
			continue
		}
		statuses, err := repos.SessionStatuses.ListBySession(ctx, sib.ID)
		if err != nil {
			return fmt.Errorf("list sibling statuses: %w", err)
		}
		for _, st := range statuses {
			seen[st.AttendeeID] = struct{}{}
		}
	}

	// Create a status row for each attendee in the new session.
	for attendeeID := range seen {
		// Skip if a row already exists.
		existing, err := repos.SessionStatuses.FindByAttendeeSession(ctx, attendeeID, session.ID)
		if err == nil && existing != nil {
			continue
		}
		if err != nil && !errors.Is(err, attendance.ErrStatusNotFound) {
			return fmt.Errorf("check new session status: %w", err)
		}
		st := attendance.NewAttendeeSessionStatus(attendeeID, session.ID, now)
		if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
			return fmt.Errorf("upsert roster status: %w", err)
		}
	}

	return nil
}