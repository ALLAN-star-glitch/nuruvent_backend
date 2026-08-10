package validation

import (
	"regexp"
	"strings"
)

// Phone validates and normalizes Kenyan phone numbers
type Phone struct{}

// ValidatePhoneNumber validates Kenyan phone numbers
// Supports formats:
// - 254XXXXXXXXX (11 digits)
// - +254XXXXXXXXX (12 digits with +)
// - 07XXXXXXXX (10 digits)
// - 01XXXXXXXX (10 digits)
// - 02XXXXXXXX (10 digits)
func (p Phone) Validate(phone string) bool {
	phone = normalizePhone(phone)

	patterns := []string{
		`^\+254\d{9}$`,   // +254XXXXXXXXX
		`^254\d{9}$`,     // 254XXXXXXXXX
		`^07\d{8}$`,      // 07XXXXXXXX
		`^01\d{8}$`,      // 01XXXXXXXX
		`^02\d{8}$`,      // 02XXXXXXXX
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, phone)
		if matched {
			return true
		}
	}

	return false
}

// Normalize converts any Kenyan phone number to E.164 format
// Returns: +254XXXXXXXXX
func (p Phone) Normalize(phone string) string {
	phone = normalizePhone(phone)

	if phone == "" {
		return ""
	}

	// Remove leading '0' and add '254'
	if strings.HasPrefix(phone, "0") && len(phone) == 10 {
		return "+254" + phone[1:]
	}

	// Already has 254 prefix
	if strings.HasPrefix(phone, "254") && len(phone) == 12 {
		return "+" + phone
	}

	// Just numbers, assume it's a local number
	if len(phone) == 9 {
		return "+254" + phone
	}

	return phone
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")
	return phone
}