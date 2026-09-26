// internal/modules/video/service/errors.go

package service

import (
	"errors"
	"fmt"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// WIRING
// ============================================================

// errMissingDependency is returned by New when a required field on
// Dependencies is nil. Failing at construction beats a nil-pointer
// panic at first request.
func errMissingDependency(name string) error {
	return fmt.Errorf("video service: missing dependency %s", name)
}

// ============================================================
// SENTINELS
// ============================================================
//
// These are returned directly or wrapped with %w so callers can
// errors.Is against them. The delivery layer maps them to HTTP
// status codes.

// ErrConnectionNotFound is returned when the host has no active
// connection for the requested platform, or when a lookup by ID
// finds nothing.
var ErrConnectionNotFound = errors.New("video: connection not found")

// ErrConnectionRequiresReconnect is returned when the refresh token
// has expired or been revoked. The row still exists; it just can't
// be used. Distinct from ErrConnectionNotFound because the host
// needs a different UI ("Reconnect Zoom") than a fresh connect.
var ErrConnectionRequiresReconnect = errors.New("video: connection requires reconnect")

// ErrConnectionAlreadyExists is returned when the host already has an
// active connection for the platform they are trying to connect.
// One active connection per (user, platform) — FR-V-003.
var ErrConnectionAlreadyExists = errors.New("video: connection already exists")

// ErrInvalidCommand is returned for malformed or missing input.
// The delivery layer maps it to HTTP 400.
var ErrInvalidCommand = errors.New("video: invalid command")

// ErrInvalidOAuthState is returned when the state parameter is
// unknown, already consumed, or expired. FR-V-034.
var ErrInvalidOAuthState = errors.New("video: invalid or expired oauth state")

// ErrOAuthDenied is returned when the user declined the grant at the
// platform (callback arrives with error=access_denied).
var ErrOAuthDenied = errors.New("video: oauth grant denied by user")

// ErrPlatformUnsupported is returned when the Platform enum value is
// valid but no ProviderClient is registered for it. FR-V-030.
var ErrPlatformUnsupported = errors.New("video: platform not supported")

// ErrMeetingNotFound is returned when a meeting lookup by ID finds
// nothing.
var ErrMeetingNotFound = errors.New("video: meeting not found")

// ============================================================
// HELPERS
// ============================================================

// invalidf wraps ErrInvalidCommand with a formatted reason. Prefer
// this over fmt.Errorf("...: %w", ErrInvalidCommand) so the message
// stays consistent.
func invalidf(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidCommand}, args...)...)
}

// errorsIsNotFound reports whether err is (or wraps) a
// connection-not-found from the domain repository.
func errorsIsNotFound(err error) bool {
	return errors.Is(err, videodomain.ErrConnectionNotFound)
}

// add to internal/modules/video/service/errors.go

// errorsIsMeetingNotFound reports whether err is (or wraps) a
// meeting-not-found from the domain repository.
func errorsIsMeetingNotFound(err error) bool {
	return errors.Is(err, videodomain.ErrMeetingNotFound)
}