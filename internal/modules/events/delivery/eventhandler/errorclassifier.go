// internal/modules/events/handler/errorclassifier.go

package eventhandler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// ERROR CLASSIFIER
// ============================================================
//
// Single place where every service-layer error becomes an HTTP
// response. Callers replace their ad-hoc `if errors.Is(...)` chains
// with:
//
//     return respondClassifiedError(c, "publish event", err)
//
// All sentinels live in the domain package so the classifier only
// imports `domain`. Raw errors from lower layers are wrapped with
// fmt.Errorf("%w: ...", ...) and unwrapped here via errors.Is/As.

// classifiedError is the internal shape the classifier produces.
type classifiedError struct {
	Status  int
	Message string
	Details fiber.Map
}

// ============================================================
// MAIN CLASSIFIER
// ============================================================

func classifyEventError(action string, err error) classifiedError {
	if err == nil {
		return classifiedError{
			Status:  fiber.StatusInternalServerError,
			Message: "Unexpected error",
		}
	}

	// ---------- AUTH / PERMISSION ----------
	if errors.Is(err, domain.ErrForbidden) {
		return classifiedError{
			Status:  fiber.StatusForbidden,
			Message: "You don't have permission to " + action,
		}
	}

	// ---------- NOT FOUND ----------
	if errors.Is(err, domain.ErrEventNotFound) {
		return classifiedError{
			Status:  fiber.StatusNotFound,
			Message: "Event not found",
		}
	}
	if errors.Is(err, domain.ErrEventTypeNotFound) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Event type not found",
			Details: fiber.Map{
				"field":  "event_type_id",
				"reason": "the selected event type does not exist",
			},
		}
	}
	if errors.Is(err, domain.ErrEventStatusNotFound) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Event status not found",
			Details: fiber.Map{
				"field":  "event_status_id",
				"reason": "the event status does not exist in the system",
			},
		}
	}
	if errors.Is(err, domain.ErrInstitutionNotFound) {
		return classifiedError{
			Status:  fiber.StatusNotFound,
			Message: "Institution not found",
		}
	}
	if errors.Is(err, domain.ErrCategoryNotFound) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Category not found",
			Details: fiber.Map{
				"field":  "category_id",
				"reason": "the selected category does not exist",
			},
		}
	}
	if errors.Is(err, domain.ErrTicketTypeNotFound) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Ticket type not found",
			Details: fiber.Map{
				"field":  "ticket_type_ids",
				"reason": cleanReason(err.Error()),
			},
		}
	}

	// ---------- VALIDATION ----------
	if ce, ok := classifyValidationError(err); ok {
		return ce
	}

	// ---------- CREATE-PATH SENTINELS ----------
	if errors.Is(err, domain.ErrEventTeamRequired) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Team is required",
			Details: fiber.Map{
				"field":  "team_id",
				"reason": "cannot create an event without a team context",
			},
		}
	}
	if errors.Is(err, domain.ErrEventCreatedByRequired) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Creator is required",
			Details: fiber.Map{
				"field":  "created_by",
				"reason": "cannot create an event without an authenticated user",
			},
		}
	}
	if errors.Is(err, domain.ErrEventVisibilityInvalid) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "Visibility must be public, private, or unlisted",
			Details: fiber.Map{
				"field":  "visibility",
				"reason": "missing or invalid visibility value",
			},
		}
	}
	if errors.Is(err, domain.ErrAccountDomainMissing) {
		return classifiedError{
			Status:  fiber.StatusUnprocessableEntity,
			Message: "Event ownership could not be resolved",
			Details: fiber.Map{
				"reason": "account domain could not be determined for this team",
			},
		}
	}
	if errors.Is(err, domain.ErrEventTicketConversion) {
		return classifiedError{
			Status:  fiber.StatusBadRequest,
			Message: "One or more tickets are invalid",
			Details: fiber.Map{
				"field":  "tickets",
				"reason": cleanReason(err.Error()),
			},
		}
	}

	// ---------- SERVER-SIDE LOOKUP FAILURES ----------
	if errors.Is(err, domain.ErrEventStatusLookup) {
		return classifiedError{
			Status:  fiber.StatusInternalServerError,
			Message: "Could not look up the event status",
			Details: fiber.Map{"error": err.Error()},
		}
	}
	if errors.Is(err, domain.ErrEventEventTypeLookup) {
		return classifiedError{
			Status:  fiber.StatusInternalServerError,
			Message: "Could not look up the event type",
			Details: fiber.Map{"error": err.Error()},
		}
	}
	if errors.Is(err, domain.ErrRepositoryNotReady) {
		return classifiedError{
			Status:  fiber.StatusInternalServerError,
			Message: "Server is not ready to handle this request",
			Details: fiber.Map{"error": err.Error()},
		}
	}

	// ---------- PERSISTENCE ----------
	if errors.Is(err, domain.ErrEventPersistenceWrite) {
		if isDuplicateKey(err) {
			return classifiedError{
				Status:  fiber.StatusConflict,
				Message: "An event with this name already exists",
				Details: fiber.Map{
					"field":  "name",
					"reason": "please choose a different name",
				},
			}
		}
		return classifiedError{
			Status:  fiber.StatusInternalServerError,
			Message: "Could not save the event",
			Details: fiber.Map{"error": err.Error()},
		}
	}

	// ---------- PUBLISH VALIDATION SUMMARY ----------
	// `ValidateForPublish` returns `fmt.Errorf("validation failed: %s", ...)`.
	// Extract the joined list so the user sees every failing rule.
	if summary := extractValidationSummary(err); summary != "" {
		return classifiedError{
			Status:  fiber.StatusUnprocessableEntity,
			Message: "Event is not ready to publish: " + summary,
			Details: fiber.Map{"error": err.Error()},
		}
	}

	// ---------- FALLBACK ----------
	return classifiedError{
		Status:  fiber.StatusInternalServerError,
		Message: "Failed to " + action,
		Details: fiber.Map{"error": err.Error()},
	}
}

// ============================================================
// VALIDATION SENTINELS
// ============================================================

func classifyValidationError(err error) (classifiedError, bool) {
	type rule struct {
		target  error
		field   string
		message string
	}

	rules := []rule{
		{domain.ErrEventNameRequired, "name", "Event name is required"},
		{domain.ErrInvalidEventName, "name", "Event name is invalid"},
		{domain.ErrEventDisplayNameRequired, "display_name", "Display name is required"},
		{domain.ErrEventSlugRequired, "slug", "Slug is required"},
		{domain.ErrInvalidEventDescription, "description", "Event description is required"},
		{domain.ErrEventTypeRequired, "event_type_id", "Event type is required"},
		{domain.ErrInvalidEventType, "event_type_id", "Event type is invalid"},
		{domain.ErrEventDateRequired, "start_date", "Event date is required"},
		{domain.ErrEventTimeRequired, "start_time", "Event time is required"},
		{domain.ErrEventDurationRequired, "duration", "Event duration is required"},
		{domain.ErrEventDurationTooShort, "duration", "Duration must be at least 15 minutes"},
		{domain.ErrEventDurationTooLong, "duration", "Duration cannot exceed 24 hours"},
		{domain.ErrEventLocationRequired, "location", "Location is required for in-person events"},
		{domain.ErrEventMeetingLinkRequired, "zoom_link", "At least one meeting link is required for virtual events"},
		{domain.ErrEventPastDate, "start_date", "Cannot publish an event with a past date"},
		{domain.ErrEventScheduleRequired, "schedules", "At least one schedule is required"},
		{domain.ErrEventTicketRequired, "tickets", "At least one ticket is required"},
		{domain.ErrInvalidCapacity, "capacity", "Event capacity is invalid"},
		{domain.ErrInvalidCategory, "category_id", "Category is invalid"},
		{domain.ErrInvalidEventFormat, "event_format_id", "Event format is invalid"},
		{domain.ErrInvalidRecurrence, "recurrence", "Recurrence configuration is invalid"},
		{domain.ErrInvalidTicket, "tickets", "One or more tickets are invalid"},
		{domain.ErrInvalidMaterial, "materials", "One or more materials are invalid"},
		{domain.ErrInvalidCertificateTemplate, "certificate_template_id", "Certificate template is invalid"},
		{domain.ErrInvalidCertificateType, "certificate_type_id", "Certificate type is invalid"},
	}

	for _, r := range rules {
		if errors.Is(err, r.target) {
			return classifiedError{
				Status:  fiber.StatusBadRequest,
				Message: r.message,
				Details: fiber.Map{
					"field":  r.field,
					"reason": r.message,
				},
			}, true
		}
	}
	return classifiedError{}, false
}

// ============================================================
// HELPERS
// ============================================================

// extractValidationSummary pulls the `fmt.Errorf("validation failed: %s", ...)`
// payload from ValidateForPublish into a single-line summary.
func extractValidationSummary(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	const marker = "validation failed: "
	idx := strings.Index(msg, marker)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(msg[idx+len(marker):])
}

// isDuplicateKey reports whether the error is a Postgres unique-violation.
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint")
}

// cleanReason strips sentinel wrapper prefixes so the user sees only
// the reason, not "ticket type not found: ticket type not found: <id>".
func cleanReason(raw string) string {
	for {
		idx := strings.Index(raw, ": ")
		if idx < 0 || idx > 40 {
			break
		}
		raw = raw[idx+2:]
	}
	return raw
}

// ============================================================
// RESPONDER
// ============================================================

// respondClassifiedError converts a service error into an HTTP response.
//
// `action` is a verb phrase like "publish event" — used only when the
// classifier falls through to the generic case.
func respondClassifiedError(c fiber.Ctx, action string, err error) error {
	ce := classifyEventError(action, err)

	details := ce.Details
	if details == nil {
		details = fiber.Map{}
	}

	switch ce.Status {
	case fiber.StatusBadRequest:
		return response.BadRequest(c, ce.Message, details)
	case fiber.StatusUnauthorized:
		return response.Unauthorized(c, ce.Message, details)
	case fiber.StatusForbidden:
		return response.Forbidden(c, ce.Message, details)
	case fiber.StatusNotFound:
		return response.NotFound(c, ce.Message, details)
	case fiber.StatusConflict:
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": ce.Message,
			"errors":  details,
		})
	case fiber.StatusUnprocessableEntity:
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": ce.Message,
			"errors":  details,
		})
	default:
		return response.InternalError(c, ce.Message, details)
	}
}