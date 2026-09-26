package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)





// UpsertSession creates a session, or updates the mutable fields of an
// existing session.
//
// Identity: (external_type, external_id, provider_session_id).
// A parent entity may have many sessions; the provider_session_id
// selects one.
func (s *attendanceService) UpsertSession(
	ctx context.Context,
	cmd UpsertSessionCommand,
) (*attendance.Session, error) {
	if !cmd.External.IsValid() {
		return nil, fmt.Errorf("%w: external reference is required", attendance.ErrInvalidSession)
	}

	now := s.deps.Clock.Now()
	var session *attendance.Session

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		existing, err := repos.Sessions.FindByExternalAndProviderSession(
			ctx, cmd.External, cmd.ProviderSessionID,
		)
		if err == nil && existing != nil {
			// Update mutable fields on the existing session.
			existing.Title = cmd.Title
			existing.ScheduledStart = cmd.ScheduledStart
			existing.ScheduledEnd = cmd.ScheduledEnd
			existing.DurationMinutes = int(cmd.ScheduledEnd.Sub(cmd.ScheduledStart).Minutes())
			existing.Provider = cmd.Provider
			existing.ProviderMeetingID = cmd.ProviderMeetingID
			existing.ProviderURL = cmd.ProviderURL
			existing.UpdatedAt = now

			if err := validateUpdatedSession(existing); err != nil {
				return err
			}

			if err := repos.Sessions.Update(ctx, existing); err != nil {
				return fmt.Errorf("update session: %w", err)
			}
			session = existing
			return nil
		}
		if err != nil && !errors.Is(err, attendance.ErrSessionNotFound) {
			return fmt.Errorf("check existing session: %w", err)
		}

		newSession, err := attendance.NewSession(
			s.deps.IDs.NewID(),
			cmd.External,
			cmd.ProviderSessionID,
			cmd.Title,
			cmd.ScheduledStart,
			cmd.ScheduledEnd,
			cmd.Provider,
			cmd.ProviderMeetingID,
			cmd.ProviderURL,
			now,
		)
		if err != nil {
			return err
		}

		if err := repos.Sessions.Create(ctx, newSession); err != nil {
			if errors.Is(err, attendance.ErrDuplicateSession) {
				winner, findErr := repos.Sessions.FindByExternalAndProviderSession(
					ctx, cmd.External, cmd.ProviderSessionID,
				)
				if findErr != nil {
					return fmt.Errorf("fetch winner after duplicate: %w", findErr)
				}
				session = winner
				return nil
			}
			return fmt.Errorf("persist session: %w", err)
		}

		if err := s.ensureSessionRoster(ctx, repos, newSession, now); err != nil {
			return fmt.Errorf("ensure roster: %w", err)
		}

		session = newSession
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return session, nil
}

func validateUpdatedSession(s *attendance.Session) error {
	if s.ScheduledEnd.Before(s.ScheduledStart) {
		return fmt.Errorf("%w: end must be at or after start", attendance.ErrInvalidSession)
	}
	if !s.Provider.IsValid() {
		return fmt.Errorf("%w: invalid provider %q", attendance.ErrInvalidSession, s.Provider)
	}
	if s.Provider.RequiresMeetingID() && s.ProviderMeetingID == "" {
		return fmt.Errorf("%w: meeting id required for %s", attendance.ErrInvalidSession, s.Provider)
	}
	return nil
}