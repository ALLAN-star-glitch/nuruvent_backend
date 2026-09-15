package registrationdomain

import "errors"

// Sentinel errors. Callers use errors.Is to classify.
var (
    ErrRegistrationNotFound     = errors.New("registration not found")
    ErrEventNotOpen             = errors.New("event is not open for registration")
    ErrEventFull                = errors.New("event is full")
    ErrWaitlistFull             = errors.New("waitlist is full")
    ErrDuplicateRegistration    = errors.New("user already has an active registration")
    ErrInvalidStatusTransition  = errors.New("invalid status transition")
    ErrTicketUnavailable        = errors.New("ticket type unavailable")
    ErrTicketLimitExceeded      = errors.New("ticket quantity outside allowed limits")
    ErrNotOwner                 = errors.New("actor does not own this registration")
    ErrNotOrganizer             = errors.New("actor is not the event organizer")
    ErrIdentityRequired         = errors.New("registration requires a user or guest identity")
    ErrPricingMismatch          = errors.New("pricing mismatch")
    ErrRegistrableTypeUnknown   = errors.New("unknown registrable type")
    ErrRegistrationIsTerminal   = errors.New("registration is in a terminal state")
)