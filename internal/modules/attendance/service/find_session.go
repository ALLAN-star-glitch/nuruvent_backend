package service

import (
	"context"
	"errors"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// internal/modules/attendance/service/find_session.go
func (s *attendanceService) FindSessionIDByProviderSessionID(
    ctx context.Context,
    providerSessionID string,
) (string, error) {
    var sessionID string
    err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
        session, err := repos.Sessions.FindByProviderSessionID(ctx, providerSessionID)
        if err != nil {
            if errors.Is(err, attendance.ErrSessionNotFound) {
                return nil
            }
            return err
        }
        sessionID = session.ID
        return nil
    })
    if err != nil {
        return "", err
    }
    return sessionID, nil
}