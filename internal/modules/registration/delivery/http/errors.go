// internal/modules/registration/delivery/http/errors.go

package http

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// mapDomainError translates a domain error into the correct HTTP response.
// Unknown errors map to 500 with a generic message.
func mapDomainError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, registrationdomain.ErrRegistrationNotFound):
		return response.NotFound(c, "Registration not found", nil)

	case errors.Is(err, registrationdomain.ErrEventNotOpen):
		return response.Conflict(c, "Registration is not open for this event", nil)

	case errors.Is(err, registrationdomain.ErrEventFull):
		return response.Conflict(c, "Event is full", nil)

	case errors.Is(err, registrationdomain.ErrWaitlistFull):
		return response.Conflict(c, "Waitlist is full", nil)

	case errors.Is(err, registrationdomain.ErrDuplicateRegistration):
		return response.Conflict(c, "You already have an active registration for this event", nil)

	case errors.Is(err, registrationdomain.ErrTicketUnavailable):
		return response.Conflict(c, "One or more selected tickets are unavailable", nil)

	case errors.Is(err, registrationdomain.ErrTicketLimitExceeded):
		return response.UnprocessableEntity(c, "Ticket quantity outside allowed limits", nil)

	case errors.Is(err, registrationdomain.ErrInvalidStatusTransition):
		return response.Conflict(c, "This registration can no longer be changed", nil)

	case errors.Is(err, registrationdomain.ErrRegistrationIsTerminal):
		return response.Conflict(c, "This registration is in a final state", nil)

	case errors.Is(err, registrationdomain.ErrNotOwner):
		return response.Forbidden(c, "You do not have access to this registration", nil)

	case errors.Is(err, registrationdomain.ErrNotOrganizer):
		return response.Forbidden(c, "You must be the event organizer", nil)

	case errors.Is(err, registrationdomain.ErrIdentityRequired):
		return response.BadRequest(c, "Registration requires a user or guest identity", nil)

	case errors.Is(err, registrationdomain.ErrPricingMismatch):
		return response.InternalError(c, "Pricing could not be resolved", nil)

	case errors.Is(err, registrationdomain.ErrRegistrableTypeUnknown):
		return response.InternalError(c, "Unknown registration target", nil)
	}

	return response.InternalError(c, "Something went wrong", nil)
}