package domain

import "context"

// ============================================================
// OUTBOUND PORT: VideoIdentityProvider
// ============================================================
//
// Resolves a user's external identity on a video platform so the
// events module's attendance sync can pre-link the event host's
// attendee row.
//
// The host is registered as an attendee at publish time (see
// RegisterHostAttendee). Filling in the host's Google Meet user id
// here means the first Meet fetch after publish matches the host
// without a manual roster link.
//
// Only Google Meet needs this. Zoom matches participants by username
// (the Meeting SDK's customerKey), which the host attendee row
// already carries via HostUsername — no platform-specific id is
// required for Zoom.

type VideoIdentityProvider interface {
	// ExternalUserIDForPlatform returns the user's external user
	// resource id on the given platform.
	//
	// For google_meet: "users/<id>" or "" when not connected.
	// For zoom: the numeric user id, or "" when not connected.
	//
	// Returns ("", nil) when the user has no active connection.
	// Errors indicate infrastructure failure only.
	ExternalUserIDForPlatform(
		ctx context.Context,
		userID string,
		platform string,
	) (string, error)
}