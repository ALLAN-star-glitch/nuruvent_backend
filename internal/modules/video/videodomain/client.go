// internal/modules/video/videodomain/client.go

package videodomain

import (
	"context"
	"time"
)

// ============================================================
// PROVIDER CLIENT
// ============================================================

// ProviderClient is the interface every platform integration
// implements.
//
// The service layer depends on this interface, never on a concrete
// platform. Adding a new platform means writing a new implementation
// and registering it — no changes to service or HTTP handlers.
//
// The interface is intentionally split into small capability
// interfaces (below). ProviderClient itself is the base that every
// provider implements. Providers that support OAuth also implement
// OAuthProvider. Providers that provision meetings also implement
// MeetingProvisioner. And so on.
type ProviderClient interface {
	// Platform returns the platform this client serves.
	Platform() Platform

	// Capabilities describes what this client can do. The service
	// uses these flags to decide which methods to call.
	Capabilities() Capabilities
}

// ============================================================
// CAPABILITY INTERFACES
// ============================================================

// OAuthProvider is implemented by platforms where each host connects
// their own account (Zoom, Google Meet, Teams).
//
// Platforms that use app-level credentials (LiveKit, self-hosted
// Jitsi) do not implement this interface.
type OAuthProvider interface {
	ProviderClient

	// OAuthAuthorizeURL returns the URL the user is redirected to in
	// order to grant access. The state is included so the platform
	// echoes it back on the callback.
	OAuthAuthorizeURL(state string) string

	// ExchangeCode trades an authorization code for tokens and
	// identity. Called once, on the OAuth callback.
	ExchangeCode(ctx context.Context, code string) (*ExternalUser, *TokenSet, error)

	// RefreshAccessToken exchanges the stored refresh token for a
	// new access token. Called whenever the current access token is
	// near expiry.
	//
	// Returns ErrRefreshTokenExpired if the platform rejects the
	// refresh — the caller must then mark the connection as
	// requiring reauthorization.
	RefreshAccessToken(
		ctx context.Context,
		conn *Connection,
	) (*TokenSet, error)

	// RevokeAccess asks the platform to revoke the tokens. Best
	// effort: if the platform call fails, the local connection is
	// still marked revoked.
	RevokeAccess(ctx context.Context, conn *Connection) error
}

// MeetingProvisioner is implemented by platforms that can create
// meetings or rooms on demand.
//
// Zoom, Google Meet, and LiveKit all implement this. A platform that
// only accepts pasted URLs (self-hosted Jitsi with static rooms)
// would not.
type MeetingProvisioner interface {
	ProviderClient

	// CreateMeeting asks the platform to create a new meeting (or
	// room). The connection is passed so the client can use the
	// host's credentials. For app-credential platforms (LiveKit),
	// the connection may be nil.
	CreateMeeting(
		ctx context.Context,
		conn *Connection,
		spec MeetingSpec,
	) (*Meeting, error)

	// DeleteMeeting removes a meeting from the platform.
	DeleteMeeting(
		ctx context.Context,
		conn *Connection,
		externalID string,
	) error
}

// RoomParticipantMinter is implemented by embeddable providers
// (LiveKit, Daily, 100ms). It issues per-participant tokens that the
// frontend uses to connect to a room.
//
// External platforms (Zoom, Google Meet) do not implement this
// interface — attendees simply open the join URL.
type RoomParticipantMinter interface {
	ProviderClient

	// IssueParticipantToken returns a short-lived token that the
	// Nuruvent frontend uses to connect to the given room as the
	// given participant.
	IssueParticipantToken(
		ctx context.Context,
		roomExternalID string,
		identity ParticipantIdentity,
	) (string, error)
}

// ============================================================
// SUPPORTING TYPES
// ============================================================

// TokenSet is what a provider returns after a successful OAuth
// exchange or token refresh.
type TokenSet struct {
	AccessToken  string
	RefreshToken string // may be empty if the provider keeps the same one
	ExpiresAt    time.Time
	Scopes       string // space-separated; may be empty
}

// ParticipantIdentity describes who is joining a room. Used only by
// embeddable providers.
type ParticipantIdentity struct {
	// ExternalUserID is the Nuruvent-side identifier we want the
	// provider to associate with this participant. Usually the
	// attendance attendee ID.
	ExternalUserID string

	// DisplayName is what other participants will see.
	DisplayName string

	// Email is optional. Some providers use it for moderation.
	Email string

	// Metadata is arbitrary key/value the provider stores with the
	// participant. Useful for correlation.
	Metadata map[string]string
}

// ============================================================
// REGISTRY
// ============================================================

// ClientRegistry resolves a ProviderClient by platform.
type ClientRegistry interface {
	// For returns the client registered for the platform.
	//
	// Returns ErrUnsupportedPlatform if the platform is valid but no
	// client has been registered.
	For(platform Platform) (ProviderClient, error)

	// OAuthFor is a convenience for retrieving a client that
	// implements OAuthProvider. Returns ErrCapabilityMissing if the
	// platform's client doesn't support OAuth.
	OAuthFor(platform Platform) (OAuthProvider, error)

	// ProvisionerFor is a convenience for retrieving a client that
	// implements MeetingProvisioner.
	ProvisionerFor(platform Platform) (MeetingProvisioner, error)

	// MinterFor is a convenience for retrieving a client that
	// implements RoomParticipantMinter.
	MinterFor(platform Platform) (RoomParticipantMinter, error)
}