// internal/modules/attendance/service/issue_token.go

package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// IssueJoinToken generates a new join token for an attendee in a
// session and returns the raw token plus an optional join URL.
//
// The raw token is only returned once — it is never persisted, only
// its SHA-256 hash. Multiple tokens per (attendee, session) are
// allowed and all remain valid until expiry.
//
// Expiry:
//   - If cmd.Grace > 0, expires at session.ScheduledEnd + cmd.Grace.
//   - Otherwise, expires at session.ScheduledEnd + deps.JoinTokenGrace.
//
// In-person sessions are rejected: there is no remote meeting for a
// token to resolve to, so issuing one would create a dead link.
func (s *attendanceService) IssueJoinToken(
	ctx context.Context,
	cmd IssueJoinTokenCommand,
) (*IssueJoinTokenResult, error) {
	if cmd.AttendeeID == "" {
		return nil, fmt.Errorf("attendee_id is required")
	}
	if cmd.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	now := s.deps.Clock.Now()

	// Load session and attendee. Reads only.
	var session *attendance.Session

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var txErr error
		session, txErr = repos.Sessions.FindByID(ctx, cmd.SessionID)
		if txErr != nil {
			return fmt.Errorf("load session: %w", txErr)
		}
		if _, txErr = repos.Attendees.FindByID(ctx, cmd.AttendeeID); txErr != nil {
			return fmt.Errorf("load attendee: %w", txErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if session.Status == attendance.SessionStatusCancelled {
		return nil, fmt.Errorf("%w: session is cancelled", attendance.ErrInvalidToken)
	}
	if session.Status == attendance.SessionStatusEnded {
		return nil, fmt.Errorf("%w: session has ended", attendance.ErrInvalidToken)
	}
	// In-person sessions have no remote join — reject before
	// persisting a dead token.
	if !session.Provider.RequiresMeetingID() {
		return nil, fmt.Errorf(
			"%w: session %s is in-person; no join token",
			attendance.ErrInvalidToken, cmd.SessionID,
		)
	}

	// Issue within a transaction (persists the token hash).
	var result *IssueJoinTokenResult
	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var txErr error
		result, txErr = s.issueTokenInTx(
			ctx,
			repos,
			cmd.AttendeeID,
			session,
			cmd.Grace,
			cmd.PublicBaseURL,
			now,
		)
		return txErr
	})
	if txErr != nil {
		return nil, txErr
	}

	return result, nil
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

// ============================================================
// INTERNAL — token issuance without opening a transaction
// ============================================================

// issueTokenInTx generates a join token for a (attendee, session) pair
// and persists its hash.
//
// It does NOT open a transaction — the caller owns the transaction
// context. This lets IssueJoinToken and RegisterAttendeeForExternal
// issue tokens inside their own transactional scopes without nesting
// UnitOfWork.Do calls.
//
// Returns the raw token, its expiry, and a join URL when publicBaseURL
// is non-empty.
func (s *attendanceService) issueTokenInTx(
	ctx context.Context,
	repos attendance.Repositories,
	attendeeID string,
	session *attendance.Session,
	grace time.Duration,
	publicBaseURL string,
	now time.Time,
) (*IssueJoinTokenResult, error) {
	if grace <= 0 {
		grace = s.deps.JoinTokenGrace
	}

	expiresAt := attendance.JoinTokenLifetime(session, grace)

	rawToken, tokenHash, err := s.deps.TokenGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	token, err := attendance.NewJoinToken(
		s.deps.IDs.NewID(),
		attendeeID,
		session.ID,
		tokenHash,
		now,
		expiresAt,
	)
	if err != nil {
		return nil, err
	}

	if err := repos.JoinTokens.Create(ctx, token); err != nil {
		return nil, fmt.Errorf("persist join token: %w", err)
	}

	var joinURL string
	if publicBaseURL != "" {
		joinURL = strings.TrimRight(publicBaseURL, "/") + "/join/" + rawToken
	}

	return &IssueJoinTokenResult{
		RawToken:  rawToken,
		ExpiresAt: expiresAt,
		JoinURL:   joinURL,
	}, nil
}