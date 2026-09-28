// internal/modules/video/videodomain/meeting_sdk.go

package videodomain

import "context"

// MeetingSDKSigner is implemented by OAuth providers that support
// embedding their meeting experience via a client-side SDK.
//
// The Zoom provider implements this. Future providers (Teams, Webex)
// implement their own variants, and the service resolves the signer
// through the same registry it uses for OAuth and provisioning.
//
// Meeting SDK embedding is separate from OAuth and from meeting
// provisioning. A provider may implement all three, or only some.
type MeetingSDKSigner interface {
	// SignedMeetingJWT returns a signed JWT that authorizes a join
	// to the given meeting. role is 0 for attendee, 1 for host.
	//
	// The signature is short-lived and bound to the meeting number.
	// It is safe to send to the browser.
	SignedMeetingJWT(meetingNumber string, role int) (string, error)

	// FetchZAK returns an Access Key token for the given connection.
	// The ZAK authenticates the user as the meeting host inside the
	// embedded SDK.
	//
	// Requires the user:read:zak scope on the OAuth app. Returns
	// ErrPlatformRejected if the scope is missing, or
	// ErrRefreshTokenExpired if the underlying access token is
	// stale and could not be refreshed.
	FetchZAK(ctx context.Context, conn *Connection) (string, error)

	// ClientID returns the public client identifier used as the SDK
	// key on the frontend. For Zoom, this is the OAuth Client ID.
	ClientID() string
}

// MeetingSignature is the result of GenerateMeetingSignature. The
// frontend passes both fields to the embedded SDK's join call.
type MeetingSignature struct {
	Signature string
	SDKKey    string
}

// MeetingZAK is the result of FetchMeetingZAK. The token is passed
// to the embedded SDK only for the host.
type MeetingZAK struct {
	ZAK string
}