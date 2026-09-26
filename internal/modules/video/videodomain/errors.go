// internal/modules/video/videodomain/errors.go

package videodomain

import "errors"

// ============================================================
// VALIDATION
// ============================================================

var (
	ErrInvalidConnection = errors.New("invalid connection")
	ErrInvalidMeeting    = errors.New("invalid meeting")
	ErrInvalidOAuthState = errors.New("invalid oauth state")
)

// ============================================================
// NOT FOUND
// ============================================================

var (
	ErrConnectionNotFound = errors.New("connection not found")
	ErrMeetingNotFound    = errors.New("meeting not found")
	ErrOAuthStateNotFound = errors.New("oauth state not found")
)

// ============================================================
// LIFECYCLE
// ============================================================

var (
	// ErrNotConnected is returned when a host tries to create a
	// meeting on a platform they haven't connected.
	ErrNotConnected = errors.New("host not connected to platform")

	// ErrConnectionRevoked is returned when the connection exists but
	// has been revoked.
	ErrConnectionRevoked = errors.New("connection revoked")

	// ErrOAuthStateExpired is returned when a callback arrives with a
	// state that has passed its expiry window.
	ErrOAuthStateExpired = errors.New("oauth state expired")

	// ErrOAuthStateConsumed is returned when a state has already been
	// used. OAuth states are single-use.
	ErrOAuthStateConsumed = errors.New("oauth state already consumed")

	// ErrRefreshTokenExpired is returned when the provider rejects a
	// refresh attempt. The host must reconnect.
	ErrRefreshTokenExpired = errors.New("refresh token expired; reconnect required")
)

// ============================================================
// AUTHORIZATION
// ============================================================

var (
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
)

// ============================================================
// PLATFORM
// ============================================================

var (
	// ErrUnsupportedPlatform is returned when a platform is valid per
	// the enum but no client is registered for it.
	ErrUnsupportedPlatform = errors.New("unsupported platform")

	// ErrPlatformRejected is returned when the provider rejects a
	// request. Wraps provider-specific messages.
	ErrPlatformRejected = errors.New("platform rejected request")

	// ErrPlatformUnavailable is returned on network failures or 5xx
	// responses from the platform.
	ErrPlatformUnavailable = errors.New("platform unavailable")

	// ErrPlatformTimeout is returned when a platform call exceeds the
	// configured timeout.
	ErrPlatformTimeout = errors.New("platform timeout")

	// ErrCapabilityMissing is returned when a caller asks a provider
	// to do something it doesn't support (e.g. OAuth on a platform
	// that uses app-level credentials).
	ErrCapabilityMissing = errors.New("provider does not support requested capability")



	ErrInvalidCiphertext = errors.New("invalid ciphertext")


		// under LIFECYCLE
	ErrOAuthDenied = errors.New("oauth grant denied by host")

	// under PLATFORM
	ErrOAuthBadScope     = errors.New("oauth scope rejected")
	ErrOAuthPlatformDown = errors.New("oauth platform unavailable")

	// under VALIDATION
	ErrInvalidCallback = errors.New("invalid oauth callback")
)