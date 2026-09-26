// internal/modules/attendance/service/register_attendee.go

package service

import (
	"context"
	"errors"
	"fmt"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RegisterAttendee registers a new attendee for an external reference.
//
// Idempotent: if an attendee already exists for the given external
// reference, the existing attendee is returned unchanged.
func (s *attendanceService) RegisterAttendee(
	ctx context.Context,
	cmd RegisterAttendeeCommand,
) (*attendance.Attendee, error) {
	if !cmd.External.IsValid() {
		return nil, fmt.Errorf("%w: external reference is required", attendance.ErrInvalidAttendee)
	}

	now := s.deps.Clock.Now()
	var attendee *attendance.Attendee

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		// Idempotency check inside the transaction.
		existing, err := repos.Attendees.FindByExternalRef(ctx, cmd.External)
		if err == nil && existing != nil {
			attendee = existing
			return nil
		}
		if err != nil && !errors.Is(err, attendance.ErrAttendeeNotFound) {
			return fmt.Errorf("check existing attendee: %w", err)
		}

		// Build and persist a new attendee.
		newAttendee, err := attendance.NewAttendee(
			s.deps.IDs.NewID(),
			cmd.External,
			cmd.DisplayName,
			cmd.Email,
			now,
		)
		if err != nil {
			return err
		}

		if err := repos.Attendees.Create(ctx, newAttendee); err != nil {
			// Unique-violation safety net: a concurrent caller won
			// the race. Fetch their row and return it.
			if errors.Is(err, attendance.ErrDuplicateAttendee) {
				winner, findErr := repos.Attendees.FindByExternalRef(ctx, cmd.External)
				if findErr != nil {
					return fmt.Errorf("fetch winner after duplicate: %w", findErr)
				}
				attendee = winner
				return nil
			}
			return fmt.Errorf("persist attendee: %w", err)
		}

		attendee = newAttendee
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return attendee, nil
}