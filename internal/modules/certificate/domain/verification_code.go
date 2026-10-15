// internal/modules/certificate/domain/verification_code.go

package domain

import "strings"

// VerificationCode is the public identifier embedded in a certificate
// and encoded in its QR code. It must be long enough to resist
// enumeration (BRD FR-CV-007: >= 12 chars).
const (
	// MinVerificationCodeLength is the minimum acceptable code length.
	MinVerificationCodeLength = 12
	// MaxVerificationCodeLength bounds what we accept from callers.
	MaxVerificationCodeLength = 64
)

// VerificationCode is a value object wrapping the code string.
type VerificationCode string

// NewVerificationCode validates a code produced by a generator.
func NewVerificationCode(raw string) (VerificationCode, error) {
	c := VerificationCode(strings.TrimSpace(raw))
	if err := c.Validate(); err != nil {
		return "", err
	}
	return c, nil
}

// Validate enforces length and character rules.
func (c VerificationCode) Validate() error {
	s := string(c)
	if s == "" {
		return ErrVerificationCodeRequired
	}
	if len(s) < MinVerificationCodeLength {
		return ErrVerificationCodeTooShort
	}
	if len(s) > MaxVerificationCodeLength {
		return ErrVerificationCodeTooLong
	}
	// URL-safe alphabet only; QR codes link to HTTPS URLs.
	for _, r := range s {
		if !isURLSafe(r) {
			return ErrVerificationCodeInvalidChars
		}
	}
	return nil
}

// String returns the raw code.
func (c VerificationCode) String() string { return string(c) }

func isURLSafe(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '-' || r == '_':
		return true
	default:
		return false
	}
}