// internal/modules/video/videodomain/platform.go

package videodomain


// Platform identifies a video platform the module supports.
//
// New platforms are added here. The rest of the module dispatches
// based on this value; no other file should know the specific names.
type Platform string

const (
	PlatformZoom       Platform = "zoom"
	PlatformGoogleMeet Platform = "google_meet"
	// Future: PlatformTeams, PlatformJitsi, PlatformLiveKit, ...
)

// IsValid reports whether the platform is one we recognize.
func (p Platform) IsValid() bool {
	switch p {
	case PlatformZoom, PlatformGoogleMeet:
		return true
	}
	return false
}

// Capabilities describes what a provider can do. The service
// dispatches based on these flags rather than hard-coding behavior.
type Capabilities struct {
	// RequiresOAuth is true for external platforms where the host
	// connects their own account (Zoom, Google Meet). False for
	// SDK-style platforms (LiveKit) where the app owns credentials.
	RequiresOAuth bool

	// ProvisionRooms is true when the provider can create a new
	// meeting or room on the platform's side.
	ProvisionRooms bool

	// Embeddable is true when the meeting runs inside the Nuruvent
	// UI (LiveKit, Daily) rather than redirecting to the platform.
	Embeddable bool

	// WebhooksSupported is true when the provider sends participant
	// events to our webhook endpoint.
	WebhooksSupported bool
}

