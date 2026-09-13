// internal/modules/events/domain/errors.go

package domain

import "errors"

var (
	// ============================================================
	// Lookup / lifecycle
	// ============================================================
	ErrEventNotFound       = errors.New("event not found")
	ErrEventFull           = errors.New("event is full")
	ErrInvalidEvent        = errors.New("invalid event data")
	ErrEventNotPublished   = errors.New("event not published")
	ErrCannotEditEvent     = errors.New("event cannot be edited in its current status")
	ErrEventTypeNotFound   = errors.New("event type not found")
	ErrEventStatusNotFound = errors.New("event status not found")
	ErrInstitutionNotFound = errors.New("institution not found")
	ErrInvalidEventType    = errors.New("invalid event type")
	ErrInvalidEventStatus  = errors.New("invalid event status")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrTicketTypeNotFound  = errors.New("ticket type not found")

	// ============================================================
	// Validation
	// ============================================================
	ErrEventValidationFailed    = errors.New("event validation failed")
	ErrEventNameRequired        = errors.New("event name is required")
	ErrEventDisplayNameRequired = errors.New("event display name is required")
	ErrEventSlugRequired        = errors.New("event slug is required")
	ErrEventTypeRequired        = errors.New("event type is required")
	ErrEventDateRequired        = errors.New("event date is required")
	ErrEventTimeRequired        = errors.New("event time is required")
	ErrEventDurationRequired    = errors.New("event duration is required")
	ErrEventLocationRequired    = errors.New("location is required for in-person events")
	ErrEventMeetingLinkRequired = errors.New("at least one meeting link is required for virtual events")
	ErrEventPastDate            = errors.New("cannot publish past event")
	ErrEventDurationTooShort    = errors.New("duration must be at least 15 minutes")
	ErrEventDurationTooLong     = errors.New("duration cannot exceed 1440 minutes (24 hours)")

	// ============================================================
	// Format / recurrence / category / materials / tickets
	// ============================================================
	ErrInvalidEventFormat = errors.New("invalid event format")
	ErrInvalidRecurrence  = errors.New("invalid recurrence configuration")

	// ============================================================
	// Event fields
	// ============================================================
	ErrInvalidEventName        = errors.New("invalid event name")
	ErrInvalidEventDescription = errors.New("invalid event description")
	ErrInvalidEventOwner       = errors.New("invalid event owner")
	ErrEventScheduleRequired   = errors.New("at least one schedule is required")
	ErrEventTicketRequired     = errors.New("at least one ticket is required")
	ErrInvalidCapacity         = errors.New("invalid capacity")

	// ============================================================
	// Certificates
	// ============================================================
	ErrInvalidCertificateTemplate = errors.New("invalid certificate template")
	ErrInvalidCertificateType     = errors.New("invalid certificate type")

	// ============================================================
	// Categories / materials / tickets
	// ============================================================
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidMaterial = errors.New("invalid material")
	ErrInvalidTicket   = errors.New("invalid ticket")

	// ============================================================
	// Authorization
	// ============================================================
	ErrForbidden = errors.New("insufficient permissions to create events for this team")

	// ============================================================
	// Create-path sentinels
	// ============================================================
	//
	// Shape/state validation for the published-event create path.
	// Returned by the service when the incoming command is incomplete
	// or the event cannot be resolved.
	ErrEventTeamRequired      = errors.New("team ID is required")
	ErrEventCreatedByRequired = errors.New("created by is required")
	ErrEventVisibilityInvalid = errors.New("visibility is required (public, private, or unlisted)")
	ErrAccountDomainMissing   = errors.New("cannot resolve account domain: AccountID is empty")
	ErrRepositoryNotReady     = errors.New("repository is not initialized")

	// ============================================================
	// Service-mechanics sentinels
	// ============================================================
	//
	// Wrapped with fmt.Errorf("%w: ...", ...) to attach the raw
	// underlying error. The classifier reads the wrapper; the raw
	// error is stashed in the response's errors map.
	ErrEventStatusLookup     = errors.New("failed to look up event status")
	ErrEventEventTypeLookup  = errors.New("failed to look up event type")
	ErrEventPersistenceWrite = errors.New("failed to persist event")
	ErrEventTicketConversion = errors.New("failed to convert tickets")
)