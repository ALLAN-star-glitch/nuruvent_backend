package validation

import (
	"regexp"
)

// Password validates password strength
type Password struct{}

// Validate checks password strength
// Returns (isValid bool, message string)
func (p Password) Validate(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return false, "Password must contain at least one uppercase letter"
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return false, "Password must contain at least one lowercase letter"
	}

	if !regexp.MustCompile(`\d`).MatchString(password) {
		return false, "Password must contain at least one number"
	}

	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return false, "Password must contain at least one special character (!@#$%^&*(),.?\":{}|<>)"
	}

	return true, ""
}

// Score returns password strength score (0-4)
func (p Password) Score(password string) int {
	score := 0

	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[a-z]`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`\d`).MatchString(password) {
		score++
	}
	if regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		score++
	}

	if score <= 2 {
		return 0 // Weak
	}
	if score <= 3 {
		return 1 // Fair
	}
	if score <= 4 {
		return 2 // Good
	}
	if score <= 5 {
		return 3 // Strong
	}
	return 4 // Very Strong
}

// StrengthLabel returns human-readable strength label
func (p Password) StrengthLabel(score int) string {
	labels := []string{"Weak", "Fair", "Good", "Strong", "Very Strong"}
	if score < 0 || score > 4 {
		return "Unknown"
	}
	return labels[score]
}