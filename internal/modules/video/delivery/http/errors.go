// internal/modules/video/delivery/http/errors.go

package http

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// mapDomainError translates video domain errors into HTTP responses.
//
// Unknown errors fall through to a 500 with a generic message. The
// real error is logged by the middleware.
func mapDomainError(c fiber.Ctx, err error) error {
	switch {
	// ---- 400 Bad Request ----
	case errors.Is(err, videodomain.ErrInvalidConnection),
		errors.Is(err, videodomain.ErrInvalidMeeting),
		errors.Is(err, videodomain.ErrInvalidOAuthState),
		errors.Is(err, videodomain.ErrInvalidCallback):
		return response.BadRequest(c, err.Error(), nil)

	// ---- 401 Unauthorized ----
	case errors.Is(err, videodomain.ErrUnauthorized):
		return response.Unauthorized(c, "Authentication required", nil)

	// ---- 403 Forbidden ----
	case errors.Is(err, videodomain.ErrForbidden):
		return response.Forbidden(c, "You do not have access to this resource", nil)

	// ---- 404 Not Found ----
	case errors.Is(err, videodomain.ErrConnectionNotFound),
		errors.Is(err, videodomain.ErrMeetingNotFound),
		errors.Is(err, videodomain.ErrOAuthStateNotFound):
		return response.NotFound(c, err.Error(), nil)

	// ---- 409 Conflict ----
	case errors.Is(err, videodomain.ErrConnectionRevoked):
		return response.Conflict(c, "This connection has been revoked; reconnect to continue", nil)

	// ---- 410 Gone ----
	// The OAuth state was valid once but can no longer be used.
	case errors.Is(err, videodomain.ErrOAuthStateExpired),
		errors.Is(err, videodomain.ErrOAuthStateConsumed):
		return response.Gone(c, "This authorization link has expired. Please start again.", nil)

	// ---- 422 Unprocessable ----
	case errors.Is(err, videodomain.ErrNotConnected):
		return response.UnprocessableEntity(c, "Connect a video platform account before creating meetings", nil)

	case errors.Is(err, videodomain.ErrRefreshTokenExpired):
		return response.UnprocessableEntity(c, "Your session with the platform expired; please reconnect", nil)

	// ---- 502 Bad Gateway ----
	case errors.Is(err, videodomain.ErrPlatformUnavailable),
		errors.Is(err, videodomain.ErrPlatformRejected),
		errors.Is(err, videodomain.ErrPlatformTimeout):
		return response.BadGateway(c, "The video platform is unavailable; please try again shortly", err)

	// ---- 501 Not Implemented ----
	case errors.Is(err, videodomain.ErrUnsupportedPlatform),
		errors.Is(err, videodomain.ErrCapabilityMissing):
		return response.NotImplemented(c, err.Error(), nil)

	default:
		return response.InternalError(c, "Something went wrong", err)
	}
}