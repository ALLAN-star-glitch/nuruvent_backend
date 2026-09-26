// internal/modules/attendance/service/list_sessions.go

package service

import (
	"context"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ListSessionsForExternal returns every session under the given
// external reference, ordered by scheduled start.
func (s *attendanceService) ListSessionsForExternal(
	ctx context.Context,
	ref attendance.ExternalRef,
) ([]*attendance.Session, error) {
	if !ref.IsValid() {
		return nil, fmt.Errorf("external reference is required")
	}

	var sessions []*attendance.Session

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		var e error
		sessions, e = repos.Sessions.ListByExternalRef(ctx, ref)
		return e
	})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}