// internal/modules/video/infrastructure/providers/googlemeet/oauth.go

package googlemeet

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

// OAuthAuthorizeURL builds the Google authorize URL.
//
// Google differs from Zoom in three ways here:
//
//  1. Scopes must be explicit. Zoom configures them on the
//     Marketplace app; Google requires a scope parameter.
//
//  2. access_type=offline requests a refresh token. Without
//     it, Google issues only an access token.
//
//  3. prompt=consent forces the consent screen on every
//     authorization. This matters for re-authorization:
//     without it, Google will not re-issue a refresh token
//     if the user has already granted access.
func (c *Client) OAuthAuthorizeURL(state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.oauth.ClientID)
	q.Set("redirect_uri", c.oauth.RedirectURI)
	q.Set("state", state)

	// Explicit scopes. "openid" and "email" are for identity
	// (ID token). The Meet scope is for space creation.
	q.Set("scope", "openid email https://www.googleapis.com/auth/meetings.space.created")

	// Request a refresh token. Required for long-lived access.
	q.Set("access_type", "offline")

	// Force consent on every authorization. Without this, a
	// returning user who already granted access will not receive
	// a new refresh token.
	q.Set("prompt", "consent")

	return c.oauthBaseURL + "/auth?" + q.Encode()
}

// ============================================================
// CODE EXCHANGE
// ============================================================

// ExchangeCode trades an authorization code for tokens and identity.
//
// Google differs from Zoom here in a significant way: the token
// response includes an ID token (a signed JWT) containing the
// user's identity. There is no need for a separate /users/me
// call.
//
// If the ID token is absent (unusual for a well-formed request
// with the "openid" scope), we fall back to the userinfo endpoint.
func (c *Client) ExchangeCode(
	ctx context.Context,
	code string,
) (*videodomain.ExternalUser, *videodomain.TokenSet, error) {
	if strings.TrimSpace(code) == "" {
		return nil, nil, fmt.Errorf("%w: code is required", videodomain.ErrPlatformRejected)
	}

	tokens, idToken, err := c.exchangeCodeForTokens(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	user, err := c.identityFromIDToken(ctx, tokens.AccessToken, idToken)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// exchangeCodeForTokens performs the POST /token call.
//
// Google's token endpoint accepts parameters as a form body,
// not a query string. This differs from Zoom, which uses query
// parameters.
//
// Client authentication is client_secret_post: the client_id and
// client_secret are included in the form body, not HTTP Basic.
func (c *Client) exchangeCodeForTokens(
	ctx context.Context,
	code string,
) (*videodomain.TokenSet, string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.oauth.RedirectURI)
	form.Set("client_id", c.oauth.ClientID)
	form.Set("client_secret", c.oauth.ClientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.tokenBaseURL+"/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", mapHTTPError(resp)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, "", fmt.Errorf("decode token response: %w", err)
	}

	if tr.AccessToken == "" {
		return nil, "", fmt.Errorf("%w: empty access token", videodomain.ErrPlatformRejected)
	}

	tokens := &videodomain.TokenSet{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken, // May be empty on re-authorization
		ExpiresAt:    time.Now().UTC().Add(time.Duration(tr.ExpiresIn) * time.Second),
		Scopes:       tr.Scope,
	}

	return tokens, tr.IDToken, nil
}

// identityFromIDToken extracts the user's identity.
//
// Google's token response includes an ID token (JWT) with the
// user's "sub" (stable subject ID) and "email" claims. We decode
// the JWT payload without verification here, because the token
// arrived directly from Google over TLS in response to our own
// request. Full signature verification is only necessary when
// the token travels through an untrusted channel (e.g. from a
// client).
//
// If the ID token is missing, fall back to the userinfo endpoint.
func (c *Client) identityFromIDToken(
	ctx context.Context,
	accessToken string,
	idToken string,
) (*videodomain.ExternalUser, error) {
	if idToken == "" {
		return c.fetchUserInfo(ctx, accessToken)
	}

	claims, err := parseIDTokenClaims(idToken)
	if err != nil {
		return nil, fmt.Errorf("%w: parse id token: %v", videodomain.ErrPlatformRejected, err)
	}

	if claims.Sub == "" {
		return nil, fmt.Errorf("%w: empty subject in id token", videodomain.ErrPlatformRejected)
	}

	return &videodomain.ExternalUser{
		ID:     claims.Sub,
		Email:  claims.Email,
		OrgID:  "", // Google does not return an org ID in the ID token
		Scopes: "",
	}, nil
}

// parseIDTokenClaims decodes the payload of a JWT without verifying
// the signature.
//
// This is safe only when the token arrived over TLS directly from
// Google in response to our own request. Do not use this function
// for tokens received from clients.
func parseIDTokenClaims(raw string) (*idTokenClaims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT: expected 3 parts, got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}

	var claims idTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal JWT claims: %w", err)
	}
	return &claims, nil
}

// ============================================================
// REFRESH
// ============================================================

// RefreshAccessToken trades the stored refresh token for a new
// access token.
//
// Google may rotate refresh tokens. When it does, the response
// contains a new refresh_token value that must be persisted by
// the service layer. When Google does not rotate, the field is
// empty and the service preserves the existing one.
//
// This differs from Zoom, which never rotates.
func (c *Client) RefreshAccessToken(
	ctx context.Context,
	conn *videodomain.Connection,
) (*videodomain.TokenSet, error) {
	if conn == nil || strings.TrimSpace(conn.RefreshToken) == "" {
		return nil, videodomain.ErrRefreshTokenExpired
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", conn.RefreshToken)
	form.Set("client_id", c.oauth.ClientID)
	form.Set("client_secret", c.oauth.ClientSecret)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.tokenBaseURL+"/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("build refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapNetworkError(err)
	}

	// Google signals an invalid refresh token with 400 and
	// error "invalid_grant".
	if resp.StatusCode == http.StatusBadRequest {
		if isInvalidGrant(resp) {
			return nil, videodomain.ErrRefreshTokenExpired
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode refresh response: %w", err)
	}

	if tr.AccessToken == "" {
		return nil, fmt.Errorf("%w: empty access token", videodomain.ErrPlatformRejected)
	}

	return &videodomain.TokenSet{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken, // Empty if Google didn't rotate; new token if it did
		ExpiresAt:    time.Now().UTC().Add(time.Duration(tr.ExpiresIn) * time.Second),
		Scopes:       tr.Scope,
	}, nil
}

// isInvalidGrant inspects a 400 response for Google's
// "invalid_grant" error, which indicates an expired or revoked
// refresh token.
//
// The response body is read and then re-wrapped so that the
// caller's fall-through to mapHTTPError can still read it.
func isInvalidGrant(resp *http.Response) bool {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	resp.Body.Close()
	resp.Body = io.NopCloser(strings.NewReader(string(body)))

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false
	}
	return errResp.Error == "invalid_grant"
}

// ============================================================
// REVOKE
// ============================================================

// RevokeAccess asks Google to revoke the token.
//
// Google's revocation endpoint accepts the token as a form
// parameter. Both access and refresh tokens can be revoked;
// revoking an access token also revokes its paired refresh
// token.
//
// Best effort: a failure does not prevent local disconnect.
func (c *Client) RevokeAccess(
	ctx context.Context,
	conn *videodomain.Connection,
) error {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return nil
	}

	form := url.Values{}
	form.Set("token", conn.AccessToken)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.tokenBaseURL+"/revoke",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("build revoke request: %w", err)
	}
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
// USER INFO (fallback)
// ============================================================

// fetchUserInfo is the fallback when no ID token is present.
// Google's userinfo endpoint returns the user's identity using
// the access token.
func (c *Client) fetchUserInfo(
	ctx context.Context,
	accessToken string,
) (*videodomain.ExternalUser, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.userInfoURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}

	var ur struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ur); err != nil {
		return nil, fmt.Errorf("decode user response: %w", err)
	}
	if ur.Sub == "" {
		return nil, fmt.Errorf("%w: empty subject", videodomain.ErrPlatformRejected)
	}

	return &videodomain.ExternalUser{
		ID:     ur.Sub,
		Email:  ur.Email,
		OrgID:  "",
		Scopes: "",
	}, nil
}