// internal/modules/video/service/cleanup_states.go

package service

import (
	"context"
	"fmt"
)

// CleanupExpiredOAuthStates removes oauth_state rows that have passed
// their expiry window. Called periodically by a background job.
//
// Returns the number of rows deleted. A zero return is a healthy
// result, not an error.
func (s *videoService) CleanupExpiredOAuthStates(ctx context.Context) (int, error) {
	now := s.deps.Clock.Now()
	n, err := s.deps.OAuthStates.DeleteExpired(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("cleanup oauth states: %w", err)
	}
	return n, nil
}