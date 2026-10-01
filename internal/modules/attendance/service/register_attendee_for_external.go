// internal/modules/attendance/service/register_attendee_for_external.go

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RegisterAttendeeForExternal registers an attendee for every virtual
// session under the given external reference, and issues one join
// token per session.
//
// Idempotent for the status rows: an existing status row for a session
// is left unchanged. Tokens are always issued fresh, so re-registering
// produces new, valid join URLs.
//
// In-person sessions are skipped entirely: there is no remote meeting
// to join, so no join token is issued and no link is returned. Status
// rows for in-person sessions are also not created here — the host
// records attendance for those manually.
//
// Called by the registration module after a registration is confirmed.
// The attendance module doesn't know what an "event" is — the external
// reference is opaque.
func (s *attendanceService) RegisterAttendeeForExternal(
	ctx context.Context,
	cmd RegisterAttendeeForExternalCommand,
) (*RegisterAttendeeForExternalResult, error) {
	if cmd.AttendeeID == "" {
		return nil, fmt.Errorf("attendee_id is required")
	}
	if !cmd.External.IsValid() {
		return nil, fmt.Errorf("external reference is required")
	}

	now := s.deps.Clock.Now()

	var links []AttendeeSessionLink

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

		// For each virtual session: ensure a status row, then issue a
		// token. In-person sessions are skipped — no join link is
		// meaningful for them.
		for _, session := range sessions {
			if !session.Provider.RequiresMeetingID() {
				// In-person (or unknown provider): no remote join.
				continue
			}

			existing, err := repos.SessionStatuses.FindByAttendeeSession(ctx, cmd.AttendeeID, session.ID)
			if err == nil && existing != nil {
				// Already registered for this session; leave the status
				// row alone but still issue a token so the caller gets a
				// fresh join URL.
			} else if err != nil && !errors.Is(err, attendance.ErrStatusNotFound) {
				return fmt.Errorf("check session status: %w", err)
			} else {
				st := attendance.NewAttendeeSessionStatus(cmd.AttendeeID, session.ID, now)
				if err := repos.SessionStatuses.Upsert(ctx, st); err != nil {
					return fmt.Errorf("upsert status: %w", err)
				}
			}

			// Issue a join token inside this transaction.
			token, err := s.issueTokenInTx(
				ctx,
				repos,
				cmd.AttendeeID,
				session,
				cmd.LinkGrace,
				cmd.PublicBaseURL,
				now,
			)
			if err != nil {
				return fmt.Errorf("issue token for session %s: %w", session.ID, err)
			}

			links = append(links, AttendeeSessionLink{
				SessionID:   session.ID,
				MeetingCode: session.ProviderMeetingID,
				Platform:    session.Provider,
				JoinURL:     token.JoinURL,
				ExpiresAt:   token.ExpiresAt,
			})
		}

		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return &RegisterAttendeeForExternalResult{
		AttendeeID: cmd.AttendeeID,
		Links:      links,
	}, nil
}

// ensureSessionRoster creates status rows for every attendee already
// registered under a session's external reference.
//
// Called from UpsertSession when a new session is created, so that
// attendees who registered before the session existed don't miss it.
//
// In-person sessions are skipped for the same reason as in
// RegisterAttendeeForExternal: attendance is recorded manually.
func (s *attendanceService) ensureSessionRoster(
	ctx context.Context,
	repos attendance.Repositories,
	session *attendance.Session,
	now time.Time,
) error {
	if !session.Provider.RequiresMeetingID() {
		return nil
	}

	siblings, err := repos.Sessions.ListByExternalRef(ctx, session.External)
	if err != nil {
		return fmt.Errorf("list sibling sessions: %w", err)
	}

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

	for attendeeID := range seen {
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