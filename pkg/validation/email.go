// pkg/validation/email.go
package validation

import (
	"regexp"
	"strings"
)

// Email validates email formats
type Email struct{}

// Validate checks if email is valid
func (e Email) Validate(email string) bool {
	email = strings.TrimSpace(email)
	
	if email == "" {
		return false
	}

	// Simple but robust email validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsValidEmail is a convenience function
func IsValidEmail(email string) bool {
	return Email{}.Validate(email)
}

// GetEmailErrorMessage returns user-friendly error message
func (e Email) ErrorMessage() string {
	return "Please enter a valid email address"
}