// internal/modules/video/service/refresh.go

package service

import (
	"context"
	"errors"
	"fmt"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ensureFreshToken refreshes the connection's access token when it is
// near expiry. Returns the same connection if no refresh was needed,
// or the updated connection after a successful refresh.
//
// On terminal refresh failure (refresh token expired or revoked) the
// connection is marked as revoked and persisted, and the caller sees
// ErrRefreshTokenExpired. Transient failures (network, 5xx) pass
// through unwrapped so the caller can retry.
//
// Called by every use case that needs a valid access token:
// CreateMeeting, DeleteMeeting, and any future read-through calls.
func (s *videoService) ensureFreshToken(
	ctx context.Context,
	conn *videodomain.Connection,
) (*videodomain.Connection, error) {
	if !conn.IsTokenExpired(s.deps.Clock.Now()) {
		return conn, nil
	}

	oauth, err := s.deps.Clients.OAuthFor(conn.Platform)
	if err != nil {
		return nil, fmt.Errorf("refresh token: resolve client: %w", err)
	}

	tokens, err := oauth.RefreshAccessToken(ctx, conn)
	if err != nil {
		if errors.Is(err, videodomain.ErrRefreshTokenExpired) {
			// The refresh token is dead. Mark the connection revoked
			// so future lookups stop returning it as active.
			//
			// We deliberately do NOT expose a separate
			// "requires_reconnect" flag: RevokedAt is the single
			// source of truth for "this connection can no longer be
			// used". The frontend distinguishes "never connected"
			// (no row) from "was connected, now revoked" (row with
			// RevokedAt set) and prompts a reconnect in both cases.
			conn.Revoke(s.deps.Clock.Now())
			_ = s.deps.Connections.Update(ctx, conn)
			return nil, videodomain.ErrRefreshTokenExpired
		}
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	if tokens == nil {
		return nil, errors.New("refresh token: client returned nil token set")
	}

	// Apply the new tokens. Connection.Refresh validates its inputs
	// and keeps the existing refresh token if the provider did not
	// rotate it.
	if err := conn.Refresh(tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresAt); err != nil {
		return nil, fmt.Errorf("refresh token: apply: %w", err)
	}
	if tokens.Scopes != "" {
		conn.Scopes = tokens.Scopes
	}

	if err := s.deps.Connections.Update(ctx, conn); err != nil {
		return nil, fmt.Errorf("refresh token: persist: %w", err)
	}
	return conn, nil
}