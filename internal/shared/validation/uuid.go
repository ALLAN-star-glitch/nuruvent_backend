// pkg/validation/uuid.go
package validation

import (
	"regexp"
	"strings"
)

// UUID validates UUID formats
type UUID struct{}

// Validate checks if string is a valid UUID
func (u UUID) Validate(id string) bool {
	id = strings.TrimSpace(id)
	
	if id == "" {
		return false
	}

	// UUID v4 pattern
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, id)
	return matched
}

// IsValidUUID is a convenience function
func IsValidUUID(id string) bool {
	return UUID{}.Validate(id)
}

// GetUUIDErrorMessage returns user-friendly error message
func (u UUID) ErrorMessage() string {
	return "Invalid UUID format. Expected format: 123e4567-e89b-12d3-a456-426614174000"
}