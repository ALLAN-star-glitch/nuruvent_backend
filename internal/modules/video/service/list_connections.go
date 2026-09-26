// internal/modules/video/service/list_connections.go

package service

import (
	"context"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ListConnections returns all of the user's connections, newest first.
//
// Includes revoked connections so the UI can show "previously
// connected to Zoom — reconnect" rather than hiding history. Callers
// filter by RevokedAt if they only want active ones.
func (s *videoService) ListConnections(
	ctx context.Context,
	userID string,
) ([]*videodomain.Connection, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}

	conns, err := s.deps.Connections.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	// Return an empty slice, not nil — JSON marshals nil as `null`,
	// which the frontend then has to guard against.
	if conns == nil {
		conns = []*videodomain.Connection{}
	}
	return conns, nil
}