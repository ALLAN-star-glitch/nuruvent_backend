// internal/modules/video/service/commands.go

package service

import (
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// CONNECT
// ============================================================

// BeginConnectCommand starts the OAuth flow.
type BeginConnectCommand struct {
	// UserID is the Nuruvent user initiating the connection.
	UserID string

	// Platform identifies which video platform to connect to.
	Platform videodomain.Platform

	// ReturnURL is where the user should be redirected after the
	// OAuth flow completes. Typically the URL they were on when
	// they clicked "Connect Zoom" — e.g. an event editor.
	// Optional; if empty, the frontend decides where to go.
	ReturnURL string
}

// ConnectResult is what BeginConnect returns.
type ConnectResult struct {
	// AuthorizeURL is where the user's browser should be redirected
	// to complete the OAuth grant.
	AuthorizeURL string

	// State is informational. The service has already persisted it;
	// the caller normally doesn't need it, but it's exposed for
	// logging or testing.
	State string
}

// CallbackCommand is the input when the platform redirects back.
type CallbackCommand struct {
	// State is the value the platform echoed back. Must match a
	// persisted, unconsumed, unexpired state.
	State string

	// Code is the authorization code from the platform.
	Code string

	// Error is set when the platform returned an error instead of a
	// code. Common values: "access_denied" (user declined).
	Error string
}

// ConnectionResult is what HandleCallback returns.
type ConnectionResult struct {
	// Connection is the newly-persisted connection.
	Connection *videodomain.Connection

	// ReturnURL is where the frontend should redirect the user. It's
	// whatever was passed to BeginConnect, or empty.
	ReturnURL string
}

// ============================================================
// DISCONNECT
// ============================================================

// DisconnectCommand revokes a user's connection to a platform.
type DisconnectCommand struct {
	UserID   string
	Platform videodomain.Platform
}

// ============================================================
// MEETINGS
// ============================================================

// CreateMeetingCommand is the input to CreateMeeting.
type CreateMeetingCommand struct {
	// UserID identifies the host on whose account the meeting is
	// created. Must have an active connection to Platform.
	UserID string

	// Platform is which video platform to create the meeting on.
	Platform videodomain.Platform

	// Spec describes the meeting.
	Spec videodomain.MeetingSpec
}



// DeleteMeetingCommand is the input to DeleteMeeting.
type DeleteMeetingCommand struct {
	// MeetingID is the Nuruvent-side meeting ID.
	MeetingID string

	// UserID is the requesting host. The service verifies that the
	// meeting belongs to them before deleting (FR-V-047/048/049/308).
	UserID string
}