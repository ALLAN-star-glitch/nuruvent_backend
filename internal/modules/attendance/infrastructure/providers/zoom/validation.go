// internal/modules/attendance/infrastructure/providers/zoom/validation.go

package zoom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// URLValidationError is returned by ParseWebhook when Zoom sends an
// endpoint verification event.
type URLValidationError struct {
	PlainToken string
}

func (e *URLValidationError) Error() string {
	return "zoom: endpoint url validation"
}

// IsURLValidationError reports whether err is a Zoom URL validation
// error.
func (p *Provider) IsURLValidationError(err error) bool {
	var v *URLValidationError
	return errors.As(err, &v)
}

// HandleURLValidation returns the body Zoom expects during endpoint
// verification:
//
//	{ "plainToken": "<plain>", "encryptedToken": "<hmac-sha256(secret, plain)>" }
func (p *Provider) HandleURLValidation(err error) map[string]string {
	var v *URLValidationError
	if !errors.As(err, &v) {
		return nil
	}
	mac := hmac.New(sha256.New, []byte(p.cfg.SecretToken))
	mac.Write([]byte(v.PlainToken))
	return map[string]string{
		"plainToken":     v.PlainToken,
		"encryptedToken": hex.EncodeToString(mac.Sum(nil)),
	}
}

func errURLValidation(plainToken string) error {
	return &URLValidationError{PlainToken: plainToken}
}