// internal/modules/attendance/delivery/http/errors.go

package http

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// mapDomainError translates attendance domain errors into HTTP
// responses.
//
// Unknown errors fall through to a 500 with a generic message. The
// real error is logged by the middleware.
func mapDomainError(c fiber.Ctx, err error) error {
	switch {
	// ---- 400 Bad Request ----
	case errors.Is(err, attendance.ErrInvalidAttendee),
		errors.Is(err, attendance.ErrInvalidSession),
		errors.Is(err, attendance.ErrInvalidAttendance),
		errors.Is(err, attendance.ErrInvalidToken):
		return response.BadRequest(c, err.Error(), nil)

	// ---- 401 Unauthorized ----
	case errors.Is(err, attendance.ErrUnauthorized):
		return response.Unauthorized(c, "Authentication required", nil)

	// ---- 403 Forbidden ----
	case errors.Is(err, attendance.ErrForbidden):
		return response.Forbidden(c, "You do not have access to this resource", nil)

	// ---- 404 Not Found ----
	case errors.Is(err, attendance.ErrAttendeeNotFound),
		errors.Is(err, attendance.ErrSessionNotFound),
		errors.Is(err, attendance.ErrTokenNotFound),
		errors.Is(err, attendance.ErrStatusNotFound):
		return response.NotFound(c, err.Error(), nil)

	// ---- 409 Conflict ----
	case errors.Is(err, attendance.ErrDuplicateAttendee),
		errors.Is(err, attendance.ErrDuplicateSession):
		return response.Conflict(c, err.Error(), nil)

	// ---- 410 Gone ----
	case errors.Is(err, attendance.ErrTokenExpired),
		errors.Is(err, attendance.ErrTokenRevoked):
		return response.Gone(c, err.Error(), nil)

	// ---- 422 Unprocessable ----
	case errors.Is(err, attendance.ErrInvalidStatusTransition):
		return response.UnprocessableEntity(c, err.Error(), nil)

	default:
		return response.InternalError(c, "Something went wrong", err)
	}
}