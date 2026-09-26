// internal/modules/video/infrastructure/providers/zoom/errors.go

package zoom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// mapHTTPError converts a Zoom HTTP response into a domain error.
//
// Zoom returns errors as `{ "code": N, "message": "..." }` with a
// non-2xx HTTP status. The code and status together determine what
// the caller should do.
//
// The body is read but bounded — Zoom's error responses are small,
// but we don't trust that.
func mapHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
	defer resp.Body.Close()

	var envelope zoomError
	_ = json.Unmarshal(body, &envelope)

	msg := strings.TrimSpace(envelope.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w: %s", videodomain.ErrUnauthorized, msg)
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformRejected, msg)
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", videodomain.ErrMeetingNotFound, msg)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: rate limited: %s", videodomain.ErrPlatformUnavailable, msg)
	case http.StatusRequestTimeout:
		return fmt.Errorf("%w: %s", videodomain.ErrPlatformTimeout, msg)
	default:
		if resp.StatusCode >= 500 {
			return fmt.Errorf("%w: %d %s", videodomain.ErrPlatformUnavailable, resp.StatusCode, msg)
		}
		return fmt.Errorf("%w: %d %s", videodomain.ErrPlatformRejected, resp.StatusCode, msg)
	}
}

// mapNetworkError converts a transport-level error into a domain
// error.
func mapNetworkError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", videodomain.ErrPlatformTimeout, err)
	}
	return fmt.Errorf("%w: %v", videodomain.ErrPlatformUnavailable, err)
}