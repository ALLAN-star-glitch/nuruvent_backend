// internal/modules/events/domain/errors.go

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
    
    //for validation
    ErrEventValidationFailed = errors.New("event validation failed")
    ErrEventNameRequired = errors.New("event name is required")
    ErrEventDisplayNameRequired = errors.New("event display name is required")
    ErrEventSlugRequired = errors.New("event slug is required")
    ErrEventTypeRequired = errors.New("event type is required")
    ErrEventDateRequired = errors.New("event date is required")
    ErrEventTimeRequired = errors.New("event time is required")
    ErrEventDurationRequired = errors.New("event duration is required")
    ErrEventLocationRequired = errors.New("location is required for in-person events")
    ErrEventMeetingLinkRequired = errors.New("at least one meeting link is required for virtual events")
    ErrEventPastDate = errors.New("cannot publish past event")
    ErrEventDurationTooShort = errors.New("duration must be at least 15 minutes")
    ErrEventDurationTooLong = errors.New("duration cannot exceed 1440 minutes (24 hours)")
)