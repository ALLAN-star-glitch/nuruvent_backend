// pkg/validation/url.go
package validation

import (
	"net/url"
	"strings"
)

// URL validates URLs
type URL struct{}

// Validate checks if string is a valid URL
func (u URL) Validate(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	
	if rawURL == "" {
		return false
	}

	// Must have http:// or https://
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Must have a host
	if parsed.Host == "" {
		return false
	}

	return true
}

// IsValidURL is a convenience function
func IsValidURL(rawURL string) bool {
	return URL{}.Validate(rawURL)
}

// GetURLErrorMessage returns user-friendly error message
func (u URL) ErrorMessage() string {
	return "Invalid URL format. Include http:// or https://"
}