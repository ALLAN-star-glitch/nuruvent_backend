// internal/modules/video/infrastructure/providers/googlemeet/client.go

package googlemeet

import (
	"net/http"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// Google API endpoints.
//
// Not configurable per environment; the APIs are global. Overridable
// in tests via NewClientWithEndpoints.
const (
	defaultBaseURL      = "https://meet.googleapis.com/v2"
	defaultOAuthBaseURL = "https://accounts.google.com/o/oauth2/v2"
	defaultTokenBaseURL = "https://oauth2.googleapis.com"
	defaultUserInfoURL  = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// Client talks to Google Meet on behalf of hosts.
//
// The same client instance is shared across requests; it holds no
// per-user state. Host credentials travel in the Connection passed to
// each method.
type Client struct {
	oauth        config.VideoOAuthPlatformConfig
	baseURL      string
	oauthBaseURL string
	tokenBaseURL string
	userInfoURL  string
	http         *http.Client
}

// NewClient constructs a Google Meet provider client.
func NewClient(oauth config.VideoOAuthPlatformConfig) *Client {
	return &Client{
		oauth:        oauth,
		baseURL:      defaultBaseURL,
		oauthBaseURL: defaultOAuthBaseURL,
		tokenBaseURL: defaultTokenBaseURL,
		userInfoURL:  defaultUserInfoURL,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// NewClientWithEndpoints is used only by tests to point at fake
// Google servers.
func NewClientWithEndpoints(
	oauth config.VideoOAuthPlatformConfig,
	baseURL, oauthBaseURL, tokenBaseURL, userInfoURL string,
) *Client {
	c := NewClient(oauth)
	if baseURL != "" {
		c.baseURL = baseURL
	}
	if oauthBaseURL != "" {
		c.oauthBaseURL = oauthBaseURL
	}
	if tokenBaseURL != "" {
		c.tokenBaseURL = tokenBaseURL
	}
	if userInfoURL != "" {
		c.userInfoURL = userInfoURL
	}
	return c
}

// Platform returns the platform this client serves.
func (c *Client) Platform() videodomain.Platform {
	return videodomain.PlatformGoogleMeet
}

// Capabilities describes what this Google Meet client can do.
func (c *Client) Capabilities() videodomain.Capabilities {
	return videodomain.Capabilities{
		RequiresOAuth:     true,
		ProvisionRooms:    true,
		Embeddable:        false,
		WebhooksSupported: false,
	}
}

// Compile-time assertions.
var (
	_ videodomain.ProviderClient     = (*Client)(nil)
	_ videodomain.OAuthProvider      = (*Client)(nil)
	_ videodomain.MeetingProvisioner = (*Client)(nil)
)