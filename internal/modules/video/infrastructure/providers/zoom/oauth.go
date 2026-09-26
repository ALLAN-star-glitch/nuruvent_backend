// internal/modules/video/infrastructure/providers/zoom/oauth.go

package zoom

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// AUTHORIZE URL
// ============================================================

// OAuthAuthorizeURL builds the Zoom authorize URL. The user is
// redirected here to grant access. Zoom echoes the state back on the
// callback.
func (c *Client) OAuthAuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.oauth.ClientID)
	q.Set("redirect_uri", c.oauth.RedirectURI)
	q.Set("state", state)

	return c.oauthBaseURL + "/authorize?" + q.Encode()
}

// ============================================================
// CODE EXCHANGE
// ============================================================

// ExchangeCode trades an authorization code for tokens and identity.
//
// Called once, from the OAuth callback. Does two Zoom calls:
//
//  1. POST /oauth/token — code for tokens.
//  2. GET /v2/users/me — identify the host.
//
// The connection is not persisted here; the service does that after
// receiving the returned ExternalUser and TokenSet.
func (c *Client) ExchangeCode(
	ctx context.Context,
	code string,
) (*videodomain.ExternalUser, *videodomain.TokenSet, error) {
	if strings.TrimSpace(code) == "" {
		return nil, nil, fmt.Errorf("%w: code is required", videodomain.ErrPlatformRejected)
	}

	tokens, err := c.exchangeCodeForTokens(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	user, err := c.fetchUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// exchangeCodeForTokens performs the POST /oauth/token call with the
// authorization_code grant.
func (c *Client) exchangeCodeForTokens(
	ctx context.Context,
	code string,
) (*videodomain.TokenSet, error) {
	q := url.Values{}
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	q.Set("redirect_uri", c.oauth.RedirectURI)

	endpoint := c.oauthBaseURL + "/token?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Authorization", basicAuth(c.oauth.ClientID, c.oauth.ClientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapNetworkError(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	resp.Body.Close()

	if tr.AccessToken == "" {
		return nil, fmt.Errorf("%w: empty access token", videodomain.ErrPlatformRejected)
	}

	return &videodomain.TokenSet{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(tr.ExpiresIn) * time.Second),
		Scopes:       tr.Scope,
	}, nil
}

// ============================================================
// REFRESH
// ============================================================

// RefreshAccessToken trades the stored refresh token for a new access
// token.
//
// Zoom does not rotate refresh tokens; the same refresh token can be
// used repeatedly. The returned TokenSet carries an empty
// RefreshToken field, and the service preserves the existing one.
//
// If the refresh token is expired or revoked, Zoom returns HTTP 400
// with code 1002. That's translated to ErrRefreshTokenExpired, which
// the service interprets as "mark the connection as needing
// reauthorization."
func (c *Client) RefreshAccessToken(
	ctx context.Context,
	conn *videodomain.Connection,
) (*videodomain.TokenSet, error) {
	if conn == nil || strings.TrimSpace(conn.RefreshToken) == "" {
		return nil, videodomain.ErrRefreshTokenExpired
	}

	q := url.Values{}
	q.Set("grant_type", "refresh_token")
	q.Set("refresh_token", conn.RefreshToken)

	endpoint := c.oauthBaseURL + "/token?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build refresh request: %w", err)
	}
	req.Header.Set("Authorization", basicAuth(c.oauth.ClientID, c.oauth.ClientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapNetworkError(err)
	}

	// Zoom signals an expired refresh token with 400 + code 1002.
	if resp.StatusCode == http.StatusBadRequest {
		err := decodeTerminalRefreshError(resp)
		if err != nil {
			return nil, err
		}
		// Fall through to generic handling below if the 400 wasn't
		// the terminal variant.
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}
	resp.Body.Close()

	if tr.AccessToken == "" {
		return nil, fmt.Errorf("%w: empty access token", videodomain.ErrPlatformRejected)
	}

	return &videodomain.TokenSet{
		AccessToken:  tr.AccessToken,
		RefreshToken: "", // Zoom does not rotate; caller preserves existing
		ExpiresAt:    time.Now().UTC().Add(time.Duration(tr.ExpiresIn) * time.Second),
		Scopes:       tr.Scope,
	}, nil
}

// decodeTerminalRefreshError inspects a 400 response from the token
// endpoint. If it's Zoom's "invalid refresh token" (code 1002), it
// returns ErrRefreshTokenExpired. Otherwise it returns nil so the
// caller can fall through to generic handling.
func decodeTerminalRefreshError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	resp.Body.Close()

	var env zoomError
	if err := json.Unmarshal(body, &env); err != nil {
		return nil // not a structured error; let mapHTTPError handle
	}

	// Zoom returns 1002 for "Invalid Refresh Token".
	// See: https://developers.zoom.us/docs/api/#error-codes
	if env.Code == 1002 {
		return videodomain.ErrRefreshTokenExpired
	}
	return nil
}

// ============================================================
// REVOKE
// ============================================================

// RevokeAccess asks Zoom to revoke the access token.
//
// Zoom's /oauth/revoke endpoint takes the access token as a query
// parameter. On success, both the access and refresh tokens are
// invalidated.
//
// Best effort: a network failure or non-2xx response is logged by the
// caller but does not prevent the local connection from being marked
// revoked.
func (c *Client) RevokeAccess(
	ctx context.Context,
	conn *videodomain.Connection,
) error {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return nil // nothing to revoke
	}

	q := url.Values{}
	q.Set("token", conn.AccessToken)

	endpoint := c.oauthBaseURL + "/revoke?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build revoke request: %w", err)
	}
	req.Header.Set("Authorization", basicAuth(c.oauth.ClientID, c.oauth.ClientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mapHTTPError(resp)
	}
	return nil
}

// ============================================================
// USER INFO
// ============================================================

// fetchUserInfo calls GET /v2/users/me to identify the connected
// host.
func (c *Client) fetchUserInfo(
	ctx context.Context,
	accessToken string,
) (*videodomain.ExternalUser, error) {
	endpoint := c.baseURL + "/users/me"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}

	var ur userResponse
	if err := json.NewDecoder(resp.Body).Decode(&ur); err != nil {
		return nil, fmt.Errorf("decode user response: %w", err)
	}
	if ur.ID == "" {
		return nil, fmt.Errorf("%w: empty user id", videodomain.ErrPlatformRejected)
	}

	return &videodomain.ExternalUser{
		ID:     ur.ID,
		Email:  ur.Email,
		OrgID:  ur.AccountID,
		Scopes: "", // populated from the TokenSet by the service
	}, nil
}

// ============================================================
// HELPERS
// ============================================================

// basicAuth returns the base64-encoded "client_id:client_secret"
// string used by Zoom's token endpoints.
func basicAuth(clientID, clientSecret string) string {
	raw := clientID + ":" + clientSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}