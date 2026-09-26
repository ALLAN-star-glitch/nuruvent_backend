package attendancedomain

import "errors"

// ============================================================
// DOMAIN ERRORS
// ============================================================

var (
	// Input validation
	ErrInvalidAttendee   = errors.New("invalid attendee")
	ErrInvalidSession    = errors.New("invalid session")
	ErrInvalidAttendance = errors.New("invalid attendance record")
	ErrInvalidToken      = errors.New("invalid join token")

	// Not found
	ErrAttendeeNotFound = errors.New("attendee not found")
	ErrSessionNotFound  = errors.New("session not found")
	ErrTokenNotFound    = errors.New("join token not found")
	ErrStatusNotFound   = errors.New("attendee session status not found")

	// Token state
	ErrTokenExpired = errors.New("join token expired")
	ErrTokenRevoked = errors.New("join token revoked")

	// Authorization
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")

	// Conflicts
	ErrDuplicateAttendee = errors.New("duplicate attendee for external reference")
	ErrDuplicateSession  = errors.New("duplicate session for external reference")

	// Status derivation
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)