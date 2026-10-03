package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RegisterHostAttendee creates or updates the host attendee row for
// an event, and ensures a session-status row exists for every session
// under it.
//
// The host attendee is distinguished by:
//   - external_type = "event_host"
//   - external_id   = event_id
//   - is_host       = true
//
// Idempotent. On subsequent calls the existing row is updated in
// place — display name, email, username, and google_meet_user_id
// are refreshed so the row tracks profile and connection changes.
func (s *attendanceService) RegisterHostAttendee(
	ctx context.Context,
	cmd RegisterHostAttendeeCommand,
) error {
	if cmd.EventID == "" {
		return fmt.Errorf("%w: event_id is required", attendance.ErrInvalidAttendee)
	}
	if cmd.HostUserID == "" {
		return fmt.Errorf("%w: host_user_id is required", attendance.ErrInvalidAttendee)
	}

	now := s.deps.Clock.Now()

	ref := attendance.ExternalRef{
		Type: "event_host",
		ID:   cmd.EventID,
	}

	return s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// 1. Upsert the host attendee row.
		existing, err := repos.Attendees.FindByExternalRef(ctx, ref)
		if err != nil && !errors.Is(err, attendance.ErrAttendeeNotFound) {
			return fmt.Errorf("find host attendee: %w", err)
		}

		if existing == nil {
			host, err := attendance.NewAttendee(
				s.deps.IDs.NewID(),
				ref,
				cmd.HostDisplayName,
				cmd.HostEmail,
				cmd.HostUsername,
				true, // isHost
				now,
				cmd.HostGoogleMeetUserID,
			)
			if err != nil {
				return fmt.Errorf("construct host attendee: %w", err)
			}
			if err := repos.Attendees.Create(ctx, host); err != nil {
				return fmt.Errorf("create host attendee: %w", err)
			}
			existing = host
		} else {
			// Refresh mutable fields, keep identity.
			existing.DisplayName = cmd.HostDisplayName
			existing.Email = cmd.HostEmail
			existing.Username = cmd.HostUsername
			existing.GoogleMeetUserID = cmd.HostGoogleMeetUserID
			existing.IsHost = true
			existing.UpdatedAt = now
			if err := repos.Attendees.Update(ctx, existing); err != nil {
				return fmt.Errorf("update host attendee: %w", err)
			}
		}

		// 2. Ensure a session-status row for every session under the
		//    event. Hosts are "registered" for all their sessions by
		//    default — the join is what upgrades the status.
		sessions, err := repos.Sessions.ListByExternalRef(ctx, attendance.ExternalRef{
			Type: "event",
			ID:   cmd.EventID,
		})
		if err != nil {
			return fmt.Errorf("list sessions for host: %w", err)
		}

		for _, session := range sessions {
			_, err := repos.SessionStatuses.FindByAttendeeSession(
				ctx, existing.ID, session.ID,
			)
			if err == nil {
				continue // already registered
			}
			if !errors.Is(err, attendance.ErrStatusNotFound) {
				return fmt.Errorf("check host session status: %w", err)
			}

			st := attendance.NewAttendeeSessionStatus(existing.ID, session.ID, now)
			if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
				return fmt.Errorf("upsert host session status: %w", err)
			}
		}

		return nil
	})
}