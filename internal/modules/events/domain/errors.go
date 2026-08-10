package domain

import "errors"

var (
    ErrEventNotFound  = errors.New("event not found")
    ErrEventFull      = errors.New("event is full")
    ErrInvalidEvent   = errors.New("invalid event data")
    ErrEventNotPublished = errors.New("event not published")
    ErrCannotEditEvent = errors.New("event cannot be edited in its current status")
    ErrEventTypeNotFound = errors.New("event type not found")
    ErrEventStatusNotFound = errors.New("event status not found")
    ErrInstitutionNotFound = errors.New("institution not found")
    ErrInvalidEventType = errors.New("invalid event type")
    ErrInvalidEventStatus = errors.New("invalid event status")
)