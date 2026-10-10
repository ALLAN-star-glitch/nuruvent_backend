package registration

import (
	"context"
	"errors"

	attendanceadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/attendance"
	eventsdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

type eventsReader struct {
	resolver attendanceadapters.EventsServiceResolver
}

func NewEventsReader(
	holder *attendanceadapters.EventsServiceHolder,
) registrationdomain.EventsReader {
	return &eventsReader{resolver: holder.Resolver()}
}

func (r *eventsReader) ListEventIDsByTeam(
	ctx context.Context,
	userID, teamID string,
) ([]string, error) {
	svc, err := r.resolver()
	if err != nil {
		return nil, err
	}
	ids, err := svc.ListEventIDsByTeam(ctx, userID, teamID)
	if err != nil {
		if errors.Is(err, eventsdomain.ErrForbidden) {
			return nil, registrationdomain.ErrTeamAccessDenied
		}
		return nil, err
	}
	return ids, nil
}