// internal/modules/attendance/service/issue_token.go

package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// IssueJoinToken generates a new join token for an attendee in a
// session and returns the raw token.
//
// The raw token is only returned once — it is never persisted, only
// its SHA-256 hash. Multiple tokens per (attendee, session) are
// allowed and all remain valid until expiry.
//
// Expiry:
//   - If cmd.Grace > 0, expires at session.ScheduledEnd + cmd.Grace.
//   - Otherwise, expires at session.ScheduledEnd + deps.JoinTokenGrace.
func (s *attendanceService) IssueJoinToken(
	ctx context.Context,
	cmd IssueJoinTokenCommand,
) (string, error) {
	if cmd.AttendeeID == "" {
		return "", fmt.Errorf("attendee_id is required")
	}
	if cmd.SessionID == "" {
		return "", fmt.Errorf("session_id is required")
	}

	grace := cmd.Grace
	if grace <= 0 {
		grace = s.deps.JoinTokenGrace
	}

	now := s.deps.Clock.Now()

	// Load session and attendee. Reads only — no transaction needed
	// for the lookups.
	var session *attendance.Session
	var attendee *attendance.Attendee

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var txErr error
		session, txErr = repos.Sessions.FindByID(ctx, cmd.SessionID)
		if txErr != nil {
			return fmt.Errorf("load session: %w", txErr)
		}
		attendee, txErr = repos.Attendees.FindByID(ctx, cmd.AttendeeID)
		if txErr != nil {
			return fmt.Errorf("load attendee: %w", txErr)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Guard against issuing tokens for sessions that can no longer be
	// attended.
	if session.Status == attendance.SessionStatusCancelled {
		return "", fmt.Errorf("%w: session is cancelled", attendance.ErrInvalidToken)
	}
	if session.Status == attendance.SessionStatusEnded {
		return "", fmt.Errorf("%w: session has ended", attendance.ErrInvalidToken)
	}

	// Compute expiry from session end + grace.
	expiresAt := attendance.JoinTokenLifetime(session, grace)

	// Generate the raw token + hash.
	rawToken, tokenHash, err := s.deps.TokenGenerator.Generate()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	// Build the domain object.
	token, err := attendance.NewJoinToken(
		s.deps.IDs.NewID(),
		attendee.ID,
		session.ID,
		tokenHash,
		now,
		expiresAt,
	)
	if err != nil {
		return "", err
	}

	// Persist inside a transaction.
	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		if err := repos.JoinTokens.Create(ctx, token); err != nil {
			return fmt.Errorf("persist join token: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return "", txErr
	}

	return rawToken, nil
}

// RevokeJoinTokens revokes every active token for a (attendee, session)
// pair.
func (s *attendanceService) RevokeJoinTokens(
	ctx context.Context,
	cmd RevokeJoinTokensCommand,
) error {
	if cmd.AttendeeID == "" || cmd.SessionID == "" {
		return fmt.Errorf("%w: attendee_id and session_id are required", attendance.ErrInvalidToken)
	}
	now := s.deps.Clock.Now()
	return s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		if err := repos.JoinTokens.RevokeByAttendeeSession(ctx, cmd.AttendeeID, cmd.SessionID, now); err != nil {
			return fmt.Errorf("revoke tokens: %w", err)
		}
		return nil
	})
}