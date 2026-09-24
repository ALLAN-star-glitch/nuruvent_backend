// internal/modules/payment/infrastructure/providers/paystack/errors.go

package paystack

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// paystackError is the shape Paystack returns on non-2xx responses.
//
//	{ "status": false, "message": "Invalid key" }
type paystackError struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// translateHTTPError converts a non-2xx response from Paystack into a
// payment-domain error.
func translateHTTPError(statusCode int, body []byte) error {
	var payload paystackError
	_ = json.Unmarshal(body, &payload)

	msg := payload.Message
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}

	switch {
	case statusCode == 400 || statusCode == 422:
		return fmt.Errorf("%w: %s", paymentdomain.ErrProviderRejected, msg)

	case statusCode == 401 || statusCode == 403:
		return fmt.Errorf("%w: authentication failed: %s",
			paymentdomain.ErrProviderRejected, msg)

	case statusCode == 404:
		return fmt.Errorf("%w: not found: %s",
			paymentdomain.ErrProviderRejected, msg)

	case statusCode == 429:
		return fmt.Errorf("%w: rate limited: %s",
			paymentdomain.ErrProviderUnavailable, msg)

	case statusCode >= 500:
		return fmt.Errorf("%w: provider error %d: %s",
			paymentdomain.ErrProviderUnavailable, statusCode, msg)

	default:
		return fmt.Errorf("%w: unexpected status %d: %s",
			paymentdomain.ErrProviderUnavailable, statusCode, msg)
	}
}