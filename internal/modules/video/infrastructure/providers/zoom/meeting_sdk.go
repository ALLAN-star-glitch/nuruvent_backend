// internal/modules/video/infrastructure/providers/zoom/meeting_sdk.go

package zoom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// MEETING SDK SUPPORT
// ============================================================
//
// The Zoom Meeting SDK lets the Nuruvent frontend embed a Zoom
// meeting inside its own page. Two credentials make that possible:
//
//  1. A signed JWT. Generated on the backend with the app's
//     Client ID and Client Secret. It authorizes the browser to
//     join a specific meeting with a specific role.
//
//  2. A ZAK (Zoom Access Key) token. Fetched from Zoom with the
//     host's OAuth access token. It authenticates the browser as
//     the meeting host, granting host controls.
//
// Attendees join with only the signature. Hosts join with both
// the signature and the ZAK.

// sdkJWTLifetime is how long the signature is valid. Zoom requires
// exp - iat to be less than 48 hours; two hours is well inside that
// and matches the typical meeting length.
const sdkJWTLifetime = 2 * time.Hour

// sdkJWTClockSkew is subtracted from iat to tolerate small clock
// differences between the Nuruvent backend and Zoom's servers.
const sdkJWTClockSkew = 30 * time.Second

// SignedMeetingJWT returns a signed JWT that authorizes the browser
// to join the given meeting with the given role.
//
// role is 0 for attendee, 1 for host. Any other value is rejected.
//
// The JWT is signed with HS256 using the app's Client Secret. The
// payload includes:
//
//   - appKey: the Client ID, so Zoom can identify the app.
//   - mn: the meeting number.
//   - role: 0 or 1.
//   - iat, exp: issued-at and expiry, in Unix seconds.
//   - tokenExp: must equal exp; Zoom requires this field.
//
// The signature is safe to send to the browser. It cannot be used
// to call the Zoom REST API; only the Meeting SDK accepts it.
func (c *Client) SignedMeetingJWT(meetingNumber string, role int) (string, error) {
	if meetingNumber == "" {
		return "", fmt.Errorf("meeting number is required")
	}
	if role != 0 && role != 1 {
		return "", fmt.Errorf("role must be 0 or 1")
	}
	if c.oauth.ClientID == "" || c.oauth.ClientSecret == "" {
		return "", fmt.Errorf("zoom client id or secret is not configured")
	}

	now := time.Now().Add(-sdkJWTClockSkew)
	exp := now.Add(sdkJWTLifetime).Unix()

	claims := jwt.MapClaims{
		"appKey":   c.oauth.ClientID,
		"mn":       meetingNumber,
		"role":     role,
		"iat":      now.Unix(),
		"exp":      exp,
		"tokenExp": exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(c.oauth.ClientSecret))
	if err != nil {
		return "", fmt.Errorf("sign meeting jwt: %w", err)
	}
	return signed, nil
}

// FetchZAK returns a Zoom Access Key token for the given connection.
//
// The ZAK is issued by Zoom for a specific user. It is used by the
// embedded Meeting SDK to authenticate the browser as that user, so
// the user gets host controls instead of attendee controls.
//
// Requires the user:read:zak scope on the OAuth app. If the scope is
// missing, Zoom returns 400 and the call maps to ErrPlatformRejected.
//
// The caller is responsible for ensuring the connection's access
// token is fresh before calling this method. The service's
// ensureFreshToken does that.
func (c *Client) FetchZAK(
	ctx context.Context,
	conn *videodomain.Connection,
) (string, error) {
	if conn == nil || conn.AccessToken == "" {
		return "", videodomain.ErrNotConnected
	}

	endpoint := c.baseURL + "/users/me/token?type=zak"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build zak request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", mapHTTPError(resp)
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode zak response: %w", err)
	}
	if body.Token == "" {
		return "", fmt.Errorf("%w: empty zak token", videodomain.ErrPlatformRejected)
	}

	return body.Token, nil
}

// ClientID returns the public client identifier for this Zoom app.
//
// The frontend passes this value as the sdkKey parameter when
// initializing the Meeting SDK. It is not secret.
func (c *Client) ClientID() string {
	return c.oauth.ClientID
}

// Compile-time assertion that *Client satisfies the MeetingSDKSigner
// port. If the interface changes, the build fails here.
var _ videodomain.MeetingSDKSigner = (*Client)(nil)