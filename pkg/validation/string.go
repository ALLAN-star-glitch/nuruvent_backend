// pkg/validation/string.go
package validation

import (
	"regexp"
	"strings"
)

// String provides string validation helpers
type String struct{}

// ValidateLength checks if string length is between min and max
func (s String) ValidateLength(str string, min, max int) bool {
	str = strings.TrimSpace(str)
	length := len(str)
	
	if min > 0 && length < min {
		return false
	}
	
	if max > 0 && length > max {
		return false
	}
	
	return true
}

// ValidateNotEmpty checks if string is not empty
func (s String) ValidateNotEmpty(str string) bool {
	return strings.TrimSpace(str) != ""
}

// ValidateAlphanumeric checks if string contains only alphanumeric characters and spaces
func (s String) ValidateAlphanumeric(str string, allowSpaces bool) bool {
	str = strings.TrimSpace(str)
	
	pattern := `^[a-zA-Z0-9`
	if allowSpaces {
		pattern += `\s`
	}
	pattern += `]+$`
	
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// ValidateNoSpecialChars checks if string has no special characters
// Allows business name characters: &, ., ', -, ,, !
func (s String) ValidateNoSpecialChars(str string) bool {
    str = strings.TrimSpace(str)
    
    //  Allow common business name characters
    pattern := `^[a-zA-Z0-9\s\.\&\'\-\,\!]+$`
    matched, _ := regexp.MatchString(pattern, str)
    return matched
}

// GetLengthErrorMessage returns user-friendly error message
func (s String) GetLengthErrorMessage(min, max int) string {
	if min > 0 && max > 0 {
		return "Field must be between " + string(rune(min)) + " and " + string(rune(max)) + " characters"
	}
	if min > 0 {
		return "Field must be at least " + string(rune(min)) + " characters"
	}
	if max > 0 {
		return "Field must be less than " + string(rune(max)) + " characters"
	}
	return "Invalid length"
}