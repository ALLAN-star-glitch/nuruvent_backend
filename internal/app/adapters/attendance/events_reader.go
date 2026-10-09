// internal/app/adapters/attendance/events_reader.go

package attendance

import (
	"context"
	"errors"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// ATTENDANCE → EVENTS ADAPTER
//
// Implements attendanceDomain.EventsReader on top of the events
// service, resolved lazily to break the wire-time cycle.
//
// This is the module boundary. Cross-module error translation
// happens here so the attendance service never imports events.
// ============================================================

type eventsReaderAdapter struct {
	resolve EventsServiceResolver
}

// NewEventsReaderAdapter constructs the EventsReader adapter.
func NewEventsReaderAdapter(
	resolve EventsServiceResolver,
) attendanceDomain.EventsReader {
	return &eventsReaderAdapter{resolve: resolve}
}

func (a *eventsReaderAdapter) ListEventIDsByTeam(
	ctx context.Context,
	userID string,
	teamID string,
) ([]string, error) {
	svc, err := a.resolve()
	if err != nil {
		return nil, err
	}

	ids, err := svc.ListEventIDsByTeam(ctx, userID, teamID)
	if err != nil {
		if errors.Is(err, eventsDomain.ErrForbidden) {
			return nil, attendanceDomain.ErrTeamAccessDenied
		}
		return nil, err
	}
	return ids, nil
}

// Compile-time assertion.
var _ attendanceDomain.EventsReader = (*eventsReaderAdapter)(nil)