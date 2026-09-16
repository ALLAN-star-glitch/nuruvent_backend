// internal/modules/registration/service/expire.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// ExpirePending transitions a pending registration to expired.
// Called by a background job after the payment window elapses.
func (s *service) ExpirePending(ctx context.Context, registrationID string) error {
	now := s.deps.Clock.Now()

	reg, err := s.deps.Registrations.FindByID(ctx, registrationID)
	if err != nil {
		return err
	}

	// Idempotent: not pending → nothing to do.
	if reg.Status != registrationdomain.StatusPending {
		return nil
	}

	if err := reg.Expire(now); err != nil {
		return err
	}

	if err := s.deps.Registrations.Update(ctx, reg); err != nil {
		return fmt.Errorf("update registration: %w", err)
	}

	return nil
}