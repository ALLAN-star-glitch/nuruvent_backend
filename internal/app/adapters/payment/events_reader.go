// internal/app/adapters/payment/events_reader.go

package payment

import (
	"context"
	"errors"

	attendanceadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/attendance"
	eventsdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

type eventsReader struct {
	resolver attendanceadapters.EventsServiceResolver
}

func NewEventsReader(
	holder *attendanceadapters.EventsServiceHolder,
) paymentdomain.EventsReader {
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
			return nil, paymentdomain.ErrTeamAccessDenied
		}
		return nil, err
	}
	return ids, nil
}