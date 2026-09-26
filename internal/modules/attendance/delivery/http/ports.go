// internal/modules/attendance/delivery/http/ports.go

package http

import (
	"context"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// OwnershipResolver maps attendance resources to the authorization
// domain that owns them.
//
// The HTTP layer uses this to know which domain to pass to the
// PermissionChecker. The adapter implementation lives in
// internal/app/adapters/attendance (cross-module boundary), so this
// package doesn't depend on the events or teams modules.
//
// Domain strings are Casbin's: "account:<uuid>". Returning "" means
// "not found" — the handler should respond 404.
type OwnershipResolver interface {
	// ResolveSessionDomain returns the domain that owns the given
	// session.
	ResolveSessionDomain(ctx context.Context, sessionID string) (string, error)

	// ResolveSessionExternalDomain returns the domain for the parent
	// entity of a session (the event or course cohort). Same domain
	// value as ResolveSessionDomain, but resolvable without loading
	// the session itself.
	ResolveSessionExternalDomain(ctx context.Context, ref attendance.ExternalRef) (string, error)
}

// JoinURLBuilder builds the public URL an attendee clicks to redeem a
// join token. The delivery layer needs to know the base URL of the
// frontend (or the API itself, if the frontend proxies).
type JoinURLBuilder interface {
	BuildJoinURL(rawToken string) string
}