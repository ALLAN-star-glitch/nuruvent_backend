package flutterwave

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// translateHTTPError converts a non-2xx response from Flutterwave into
// a payment-domain error. The service layer can then use errors.Is to
// classify.
func translateHTTPError(statusCode int, body []byte) error {
	var payload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &payload)

	msg := payload.Message
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}

	switch {
	case statusCode == 400 || statusCode == 422:
		// Client error — not retryable.
		return fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, msg)
	case statusCode == 401 || statusCode == 403:
		// Auth failure.
		return fmt.Errorf("%w: authentication failed: %s", paymentdomain.ErrProviderRejected, msg)
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