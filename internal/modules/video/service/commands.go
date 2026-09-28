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


// UpdateMeetingCommand updates an existing meeting on the host's
// platform account.
//
// The external ID identifies which meeting to update. The spec carries
// the new values.
type UpdateMeetingCommand struct {
	UserID     string
	Platform   videodomain.Platform
	ExternalID string
	Spec       videodomain.MeetingSpec
}


// DeleteMeetingCommand removes a meeting from the platform.
//
// ExternalID is the platform-side meeting ID (Zoom numeric ID, etc.),
// not Nuruvent's internal UUID. Platform must match the value used
// when the meeting was created.
type DeleteMeetingCommand struct {
	UserID     string
	Platform   videodomain.Platform
	ExternalID string
}



// GenerateMeetingSignatureCommand is the input to
// GenerateMeetingSignature.
type GenerateMeetingSignatureCommand struct {
	UserID        string
	Platform      videodomain.Platform
	MeetingNumber string
	Role          int // 0 = attendee, 1 = host
}

// FetchMeetingZAKCommand is the input to FetchMeetingZAK.
type FetchMeetingZAKCommand struct {
	UserID   string
	Platform videodomain.Platform
}

type MeetingJoinInfo struct {
    MeetingNumber string
    Signature     string
    SDKKey        string
    Password      string // Zoom meeting passcode (plaintext, from create response)
	WebEndpoint   string
    ZAK           string // empty for attendees
    Role          int    // 1 = host, 0 = attendee
}


type GetMeetingJoinInfoCommand struct {
    UserID     string
    Platform   videodomain.Platform
    ExternalID string
}