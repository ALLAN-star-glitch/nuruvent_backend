package intasend

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// intasendError is the shape IntaSend returns on non-2xx responses.
//
// IntaSend uses two main shapes:
//   - { "detail": "Invalid API key" }          — general API errors
//   - { "errors": { "field": ["msg1","msg2"] } } — validation errors
//
// Some endpoints return { "error": "..." }. We try all three.
type intasendError struct {
	Detail string              `json:"detail"`
	Error  string              `json:"error"`
	Errors map[string][]string `json:"errors"`
}

// ErrorMessage returns the best human-readable message from an IntaSend
// error body. Tries detail, error, then validation errors, then the raw body.
func (e intasendError) ErrorMessage(raw []byte) string {
	if e.Detail != "" {
		return e.Detail
	}
	if e.Error != "" {
		return e.Error
	}
	if len(e.Errors) > 0 {
		// Flatten {"field": ["msg1","msg2"]} into "field: msg1, msg2"
		parts := make([]string, 0, len(e.Errors))
		for field, msgs := range e.Errors {
			parts = append(parts, fmt.Sprintf("%s: %s", field, strings.Join(msgs, ", ")))
		}
		return strings.Join(parts, "; ")
	}
	return strings.TrimSpace(string(raw))
}

// translateHTTPError converts a non-2xx response from IntaSend into
// a payment-domain error. The service layer can then use errors.Is to
// classify.
func translateHTTPError(statusCode int, body []byte) error {
	var payload intasendError
	_ = json.Unmarshal(body, &payload)

	msg := payload.ErrorMessage(body)
	if msg == "" {
		msg = fmt.Sprintf("IntaSend returned status %d with no body", statusCode)
	}

	switch {
	case statusCode == 400 || statusCode == 422:
		// Client error — bad request or validation failure. Not retryable.
		return fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, msg)

	case statusCode == 401 || statusCode == 403:
		// Auth failure — bad secret key or insufficient permissions.
		return fmt.Errorf("%w: authentication failed: %s", paymentdomain.ErrProviderRejected, msg)

	case statusCode == 404:
		// Resource not found (e.g. invoice_id not found on status check).
		return fmt.Errorf("%w: not found: %s", paymentdomain.ErrProviderRejected, msg)

	case statusCode == 409:
		// Conflict — often "transaction already in progress" or
		// "api_ref already used". Not retryable.
		return fmt.Errorf("%w: conflict: %s", paymentdomain.ErrProviderRejected, msg)

	case statusCode == 429:
		// Rate limited — retryable.
		return fmt.Errorf("%w: rate limited: %s", paymentdomain.ErrProviderUnavailable, msg)

	case statusCode >= 500:
		// Server error — retryable.
		return fmt.Errorf("%w: provider error %d: %s", paymentdomain.ErrProviderUnavailable, statusCode, msg)

	default:
		return fmt.Errorf("%w: unexpected status %d: %s", paymentdomain.ErrProviderUnavailable, statusCode, msg)
	}
}