// internal/app/adapters/attendance/events_service_resolver.go

package attendance

import (
	"errors"
	"sync"

	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
)

// EventsServiceResolver is the func type that yields the events
// service on demand.
type EventsServiceResolver func() (eventsService.Service, error)

// EventsServiceHolder stores a reference to the events service so
// attendance adapters can reach it after wire finishes.
//
// Wire provides a *EventsServiceHolder, then some code path calls
// holder.Set(svc) once the graph is built. Adapters read through
// holder.Get().
type EventsServiceHolder struct {
	mu  sync.RWMutex
	svc eventsService.Service
}

// NewEventsServiceHolder returns an empty holder.
func NewEventsServiceHolder() *EventsServiceHolder {
	return &EventsServiceHolder{}
}

// Set stores the events service. Safe to call once, after wire.
func (h *EventsServiceHolder) Set(svc eventsService.Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.svc = svc
}


// Get returns the events service or an error if it hasn't been set.
func (h *EventsServiceHolder) Get() (eventsService.Service, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.svc == nil {
		return nil, errors.New("events service not yet initialized")
	}
	return h.svc, nil
}

// Resolver returns a func that satisfies EventsServiceResolver.
func (h *EventsServiceHolder) Resolver() EventsServiceResolver {
	return func() (eventsService.Service, error) {
		return h.Get()
	}
}