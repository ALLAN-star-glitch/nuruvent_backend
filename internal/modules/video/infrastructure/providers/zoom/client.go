// internal/modules/video/infrastructure/providers/zoom/client.go

package zoom

import (
	"net/http"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// Zoom API endpoints.
//
// Not configurable per environment; the API is global. Overridable in
// tests via NewClientWithEndpoints.
const (
	defaultBaseURL      = "https://api.zoom.us/v2"
	defaultOAuthBaseURL = "https://zoom.us/oauth"
)

// Client talks to Zoom on behalf of hosts.
//
// The same client instance is shared across requests; it holds no
// per-user state. Host credentials travel in the Connection passed to
// each method.
type Client struct {
	oauth        config.VideoOAuthPlatformConfig
	baseURL      string
	oauthBaseURL string
	http         *http.Client
}

// NewClient constructs a Zoom provider client.
func NewClient(oauth config.VideoOAuthPlatformConfig) *Client {
	return &Client{
		oauth:        oauth,
		baseURL:      defaultBaseURL,
		oauthBaseURL: defaultOAuthBaseURL,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// NewClientWithEndpoints is used only by tests to point at a fake
// Zoom server.
func NewClientWithEndpoints(
	oauth config.VideoOAuthPlatformConfig,
	baseURL, oauthBaseURL string,
) *Client {
	c := NewClient(oauth)
	if baseURL != "" {
		c.baseURL = baseURL
	}
	if oauthBaseURL != "" {
		c.oauthBaseURL = oauthBaseURL
	}
	return c
}

// Platform returns the platform this client serves.
func (c *Client) Platform() videodomain.Platform {
	return videodomain.PlatformZoom
}

// Capabilities describes what this Zoom client can do.
func (c *Client) Capabilities() videodomain.Capabilities {
	return videodomain.Capabilities{
		RequiresOAuth:     true,
		ProvisionRooms:    true,
		Embeddable:        false,
		WebhooksSupported: true,
	}
}

// Compile-time assertions.
var (
	_ videodomain.ProviderClient     = (*Client)(nil)
	_ videodomain.OAuthProvider      = (*Client)(nil)
	_ videodomain.MeetingProvisioner = (*Client)(nil)
)