// internal/app/adapters/registration/registrable_resolver.go

package registration

import (
	"context"
	"errors"
	"fmt"

	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	registrationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// RegistrableResolver resolves (type, id) pairs into a Registrable by
// delegating to the events module.
type RegistrableResolver struct {
	events eventsService.Service
}

func NewRegistrableResolver(events eventsService.Service) *RegistrableResolver {
	return &RegistrableResolver{events: events}
}

// Resolve returns a Registrable for the given type and ID.
// Currently only "event" is supported.
func (r *RegistrableResolver) Resolve(typeName, id string) (registrationdomain.Registrable, error) {
	switch typeName {
	case "event":
		event, err := r.events.GetEventByID(context.Background(), id)
		if err != nil {
			return nil, fmt.Errorf("fetch event: %w", err)
		}
		if event == nil {
			return nil, errors.New("event not found")
		}
		return NewEventRegistrable(event, r.events), nil
	default:
		return nil, fmt.Errorf("unsupported registrable type: %s", typeName)
	}
}

var _ registrationdomain.RegistrableResolver = (*RegistrableResolver)(nil)