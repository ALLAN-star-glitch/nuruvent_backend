package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ReorderSchedules assigns session_number = 1..N to the given schedule
// IDs, in the order supplied. Every schedule belonging to the event
// must be present — partial reorders are rejected so numbering stays
// gapless and unique.
func (s *eventService) ReorderSchedules(
	ctx context.Context,
	eventID, userID string,
	orderedIDs []string,
) (*domain.Event, error) {
	if eventID == "" {
		return nil, errors.New("event ID is required")
	}
	if len(orderedIDs) == 0 {
		return nil, errors.New("at least one schedule ID is required")
	}

	event, err := s.getEventAndCheckUpdatePermission(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	// Every ID must belong to this event, and the incoming set must
	// cover all of the event's schedules exactly.
	owned := make(map[string]struct{}, len(event.Schedules))
	for _, sch := range event.Schedules {
		owned[sch.ID] = struct{}{}
	}
	if len(orderedIDs) != len(owned) {
		return nil, fmt.Errorf(
			"reorder requires all %d schedules, got %d",
			len(owned), len(orderedIDs),
		)
	}
	seen := make(map[string]struct{}, len(orderedIDs))
	for _, id := range orderedIDs {
		if _, ok := owned[id]; !ok {
			return nil, fmt.Errorf("schedule %s does not belong to this event", id)
		}
		if _, dup := seen[id]; dup {
			return nil, fmt.Errorf("duplicate schedule id %s in ordered_ids", id)
		}
		seen[id] = struct{}{}
	}

	if err := s.repo.ReorderEventSchedules(ctx, eventID, orderedIDs); err != nil {
		return nil, fmt.Errorf("reorder schedules: %w", err)
	}

	// Reload so the returned event reflects the new order.
	reloaded, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if reloaded == nil {
		return nil, domain.ErrEventNotFound
	}
	return reloaded, nil
}