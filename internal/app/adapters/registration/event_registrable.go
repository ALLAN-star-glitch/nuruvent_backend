// internal/app/adapters/registration/event_registrable.go

package registration

import (
	"context"
	"errors"
	"math"
	"time"

	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	registrationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

type EventRegistrable struct {
	event *eventsDomain.Event
	svc   eventsService.Service
}



func NewEventRegistrable(event *eventsDomain.Event, svc eventsService.Service) *EventRegistrable {
	return &EventRegistrable{event: event, svc: svc}
}


func (r *EventRegistrable) ID() string   { return r.event.ID }
func (r *EventRegistrable) Type() string { return "event" }

func (r *EventRegistrable) Capacity() int {
	if r.event.Capacity == nil {
		return math.MaxInt
	}
	return *r.event.Capacity
}

func (r *EventRegistrable) CurrentRegistrations() int {
	return r.event.CurrentAttendees
}

func (r *EventRegistrable) IsRegistrationOpen() bool {
	if r.event.EventStatus == nil {
		return false
	}
	if r.event.EventStatus.Slug != eventsDomain.EventStatusPublished.GetSlug() {
		return false
	}

	now := time.Now().UTC()

	// Registration closes at the end of the event's start day (23:59:59),
	// not at the beginning. This allows same-day registration.
	if !r.event.StartDate.IsZero() {
		endOfStartDay := time.Date(
			r.event.StartDate.Year(),
			r.event.StartDate.Month(),
			r.event.StartDate.Day(),
			23, 59, 59, 0,
			r.event.StartDate.Location(),
		)
		if now.After(endOfStartDay) {
			return false
		}
	}

	if r.event.TicketSalesStart != nil && now.Before(*r.event.TicketSalesStart) {
		return false
	}
	if r.event.TicketSalesEnd != nil && now.After(*r.event.TicketSalesEnd) {
		return false
	}

	return true
}

func (r *EventRegistrable) RequiresPayment() bool {
	return !r.event.IsFreeEvent
}

func (r *EventRegistrable) AdjustCount(delta int) error {
	if delta == 0 {
		return nil
	}
	return r.svc.AdjustAttendeeCount(context.Background(), r.event.ID, delta)
}

func (r *EventRegistrable) TicketAvailability() (map[string]int, error) {
	out := make(map[string]int, len(r.event.Tickets))
	for _, t := range r.event.Tickets {
		if !t.IsActive {
			continue
		}
		out[t.TicketTypeID] = t.Quantity
	}
	if len(out) == 0 {
		return nil, errors.New("event has no active ticket types")
	}
	return out, nil
}

func (r *EventRegistrable) TicketPricing() (map[string]registrationdomain.TicketPrice, error) {
	out := make(map[string]registrationdomain.TicketPrice, len(r.event.Tickets))
	for _, t := range r.event.Tickets {
		if !t.IsActive {
			continue
		}
		out[t.TicketTypeID] = registrationdomain.TicketPrice{
			UnitPrice: toMinorUnits(t.Price),
			Discount:  0,
			Remaining: t.Quantity,
		}
	}
	if len(out) == 0 {
		return nil, errors.New("event has no active ticket types")
	}
	return out, nil
}

func toMinorUnits(major float64) int64 {
	return int64(math.Round(major * 100))
}

var _ registrationdomain.Registrable = (*EventRegistrable)(nil)