// internal/modules/video/service/get_connection.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// GetConnection returns the user's active connection for the given
// platform.
//
// Returns ErrConnectionNotFound if none exists, or
// ErrConnectionRevoked if the connection exists but has been revoked.
// The distinction matters: the first means "never connected", the
// second means "was connected, needs reconnect".
func (s *videoService) GetConnection(
	ctx context.Context,
	userID string,
	platform videodomain.Platform,
) (*videodomain.Connection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform %q",
			videodomain.ErrUnsupportedPlatform, platform)
	}

	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(ctx, userID, platform)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrConnectionNotFound
		}
		return nil, fmt.Errorf("get connection: %w", err)
	}
	if conn == nil {
		return nil, videodomain.ErrConnectionNotFound
	}
	if conn.RevokedAt != nil {
		return nil, videodomain.ErrConnectionRevoked
	}
	return conn, nil
}

// IsConnected reports whether the user has an active connection for
// the given platform. Convenience wrapper over GetConnection for
// callers that only need a boolean.
func (s *videoService) IsConnected(
	ctx context.Context,
	userID string,
	platform videodomain.Platform,
) (bool, error) {
	_, err := s.GetConnection(ctx, userID, platform)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, videodomain.ErrConnectionNotFound) ||
		errors.Is(err, videodomain.ErrConnectionRevoked) {
		return false, nil
	}
	return false, err
}