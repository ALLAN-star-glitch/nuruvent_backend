// internal/modules/video/infrastructure/providers/googlemeet/errors.go

package googlemeet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// NETWORK ERRORS
// ============================================================

// mapNetworkError translates transport-layer failures into domain
// sentinels.
//
// A network error is any failure that occurs before an HTTP response
// is received: DNS resolution, TLS handshake, connection reset,
// timeout, context cancellation.
//
// All of these map to ErrPlatformUnavailable, except for context
// cancellation, which is propagated unchanged so the caller can
// distinguish "the platform is down" from "we gave up waiting."
func mapNetworkError(err error) error {
	if err == nil {
		return nil
	}

	// Context cancellation is not a platform failure. Propagate
	// it so the caller sees the original reason.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	// A net.Error with Timeout() true indicates a deadline or
	// read timeout. Treat as unavailable.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return fmt.Errorf("%w: %v", videodomain.ErrPlatformUnavailable, err)
	}

	return fmt.Errorf("%w: %v", videodomain.ErrPlatformUnavailable, err)
}

// ============================================================
// HTTP ERRORS
// ============================================================

// mapHTTPError translates a non-2xx HTTP response into a domain
// sentinel.
//
// Google's error responses are JSON with this shape:
//
//	{
//	  "error": "invalid_grant",
//	  "error_description": "Token has been expired or revoked."
//	}
//
// or, for API errors:
//
//	{
//	  "error": {
//	    "code": 429,
//	    "message": "Quota exceeded",
//	    "status": "RESOURCE_EXHAUSTED"
//	  }
//	}
//
// This function inspects the status code first, then the body for
// a more specific sentinel.
//
// The response body is closed. Callers must not read it again.
func mapHTTPError(resp *http.Response) error {
	body := readLimitedBody(resp)
	detail := extractErrorDetail(body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		// 401 on a token-bearing request means the access
		// token was rejected. This is not the same as an
		// expired refresh token — the service layer will
		// refresh and retry.
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformRejected, detail)

	case http.StatusForbidden:
		// 403 means the token is valid but lacks permission.
		// Most often: the user revoked access in their Google
		// account, or an admin blocked the app.
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformRejected, detail)

	case http.StatusBadRequest:
		// 400 with "invalid_grant" is the terminal refresh
		// error. But this function is not only called from
		// refresh; callers that need that distinction check
		// for it before calling here (see RefreshAccessToken).
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformRejected, detail)

	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", videodomain.ErrMeetingNotFound, detail)

	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformRateLimited, detail)

	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformUnavailable, detail)
	}

	// Any other status code is treated as a platform rejection.
	return fmt.Errorf("%w: status %d: %s", videodomain.ErrPlatformRejected, resp.StatusCode, detail)
}

// ============================================================
// HELPERS
// ============================================================

// readLimitedBody reads up to 8 KiB of the response body and
// closes it. The 8 KiB cap prevents a misbehaving server from
// streaming an unbounded body into memory.
func readLimitedBody(resp *http.Response) []byte {
	if resp.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	resp.Body.Close()
	return body
}

// extractErrorDetail pulls a human-readable message from Google's
// error response body.
//
// Google returns two shapes:
//
//	string form: {"error": "invalid_grant", "error_description": "..."}
//	object form: {"error": {"code": 429, "message": "...", "status": "..."}}
//
// This function handles both. If the body is not valid JSON or
// does not match either shape, it returns a truncated raw body so
// the caller still has something to log.
func extractErrorDetail(body []byte) string {
	if len(body) == 0 {
		return "(empty body)"
	}

	// Try the object form first: {"error": {"code": ..., "message": ...}}
	var objForm struct {
		Error struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &objForm); err == nil && objForm.Error.Message != "" {
		if objForm.Error.Status != "" {
			return objForm.Error.Status + ": " + objForm.Error.Message
		}
		return objForm.Error.Message
	}

	// Try the string form: {"error": "invalid_grant", "error_description": "..."}
	var strForm struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &strForm); err == nil && strForm.Error != "" {
		if strForm.ErrorDescription != "" {
			return strForm.Error + ": " + strForm.ErrorDescription
		}
		return strForm.Error
	}

	// Fall back to a truncated raw body.
	raw := strings.TrimSpace(string(body))
	if len(raw) > 200 {
		raw = raw[:200] + "..."
	}
	return raw
}