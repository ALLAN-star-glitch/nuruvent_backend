// internal/modules/attendance/service/redeem_token.go

package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RedeemJoinToken validates a raw join token, records a join event for
// the (attendee, session) pair the token belongs to, and returns the
// session's provider URL for the redirect.
//
// This is the handler for the public /join/:token endpoint. It must
// be safe to call multiple times: if the same attendee clicks their
// link twice, two join records are created. That's the correct
// behavior — the attendee might have disconnected and rejoined.
func (s *attendanceService) RedeemJoinToken(
	ctx context.Context,
	rawToken string,
) (*RedeemResult, error) {
	if rawToken == "" {
		return nil, fmt.Errorf("%w: raw token is required", attendance.ErrInvalidToken)
	}

	now := s.deps.Clock.Now()
	hash := s.deps.TokenGenerator.Hash(rawToken)

	var result *RedeemResult

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// 1. Look up the token by hash.
		token, err := repos.JoinTokens.FindByHash(ctx, hash)
		if err != nil {
			return fmt.Errorf("lookup token: %w", err)
		}

		// 2. Validate the token's state.
		if !token.IsActive(now) {
			if token.RevokedAt != nil {
				return attendance.ErrTokenRevoked
			}
			return attendance.ErrTokenExpired
		}

		// 3. Load the session.
		session, err := repos.Sessions.FindByID(ctx, token.SessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		// 4. Guard: session must not be cancelled.
		if session.Status == attendance.SessionStatusCancelled {
			return fmt.Errorf("%w: session is cancelled", attendance.ErrTokenRevoked)
		}

		// 5. Record the join.
		record, err := attendance.NewAttendanceRecord(
			s.deps.IDs.NewID(),
			token.AttendeeID,
			token.SessionID,
			now,
			attendance.SourceJoinLink,
			now,
		)
		if err != nil {
			return fmt.Errorf("build record: %w", err)
		}
		if err := repos.Records.Create(ctx, record); err != nil {
			return fmt.Errorf("persist record: %w", err)
		}

		result = &RedeemResult{
			AttendeeID: token.AttendeeID,
			SessionID:  token.SessionID,
			RedirectTo: session.ProviderURL,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return result, nil
}