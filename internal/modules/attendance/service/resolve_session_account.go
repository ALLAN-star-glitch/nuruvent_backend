package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ResolveSessionAccount returns the account UUID that owns the
// session's parent event.
//
// Chain: sessions.external_id (local) → events.team_id → teams.account_id
// (via the events port).
//
// Returns attendance.ErrSessionNotFound when the session doesn't
// exist or doesn't belong to an event-typed parent.
func (s *attendanceService) ResolveSessionAccount(
	ctx context.Context,
	sessionID string,
) (string, error) {
	if sessionID == "" {
		return "", attendance.ErrSessionNotFound
	}
	if s.deps.EventsReader == nil {
		return "", fmt.Errorf("events reader port is not configured")
	}

	var eventID string
	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		session, err := repos.Sessions.FindByID(ctx, sessionID)
		if err != nil {
			if errors.Is(err, attendance.ErrSessionNotFound) {
				return err
			}
			return fmt.Errorf("load session: %w", err)
		}

		// Sessions can belong to any external parent (event, course
		// cohort, ...). This resolver only knows how to walk the
		// event branch; anything else defers.
		if session.External.Type != "event" {
			return attendance.ErrSessionNotFound
		}

		eventID = session.External.ID
		return nil
	})
	if err != nil {
		return "", err
	}
	if eventID == "" {
		return "", attendance.ErrSessionNotFound
	}

	accountID, err := s.deps.EventsReader.ResolveEventAccount(ctx, eventID)
	if err != nil {
		return "", fmt.Errorf("resolve event account: %w", err)
	}
	return accountID, nil
}