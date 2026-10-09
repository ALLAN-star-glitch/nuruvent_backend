// internal/shared/validation/text_sanitize.go

package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Sanitize provides text sanitization methods
type Sanitize struct{}

// ============================================================
// TEXT SANITIZATION
// ============================================================

// Name sanitizes a name for database storage (lowercase, underscores instead of spaces)
// Keeps: letters, numbers, underscores, hyphens, apostrophes
// Use cases: Event names, Certificate names, Category names, Account names (internal)
func (s Sanitize) Name(name string) string {
	if name == "" {
		return ""
	}

	name = strings.TrimSpace(name)
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s\-']`)
	name = reg.ReplaceAllString(name, "")

	name = strings.ReplaceAll(name, " ", "_")

	underscoreReg := regexp.MustCompile(`_+`)
	name = underscoreReg.ReplaceAllString(name, "_")

	name = strings.ToLower(name)
	name = strings.Trim(name, "_")

	if len(name) > 100 {
		name = name[:100]
	}
	return name
}

// DisplayName sanitizes a display name - PRESERVES emojis and special characters
// Only removes control characters and trims spaces
// Use cases: Event display names, Certificate titles, User display names
func (s Sanitize) DisplayName(name string) string {
	if name == "" {
		return ""
	}

	name = strings.TrimSpace(name)

	spaceReg := regexp.MustCompile(`\s+`)
	name = spaceReg.ReplaceAllString(name, " ")

	name = strings.TrimSpace(name)

	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

// Description sanitizes a description (preserves case and basic punctuation)
// Keeps: letters, numbers, spaces, common punctuation
func (s Sanitize) Description(description string) string {
	if description == "" {
		return ""
	}

	description = strings.TrimSpace(description)
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s\-'.,!?&()]`)
	description = reg.ReplaceAllString(description, "")
	spaceReg := regexp.MustCompile(`\s+`)
	description = spaceReg.ReplaceAllString(description, " ")
	description = strings.TrimSpace(description)

	if len(description) > 5000 {
		description = description[:5000]
	}
	return description
}

// Slug generates a URL-friendly slug with hyphens
func (s Sanitize) Slug(text string) string {
	if text == "" {
		return "untitled"
	}

	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")

	hyphenReg := regexp.MustCompile(`\-+`)
	slug = hyphenReg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "untitled"
	}
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

// GenerateSlugFromName generates a slug from a name with proper hyphenation
func (s Sanitize) GenerateSlugFromName(name string) string {
	if name == "" {
		return "untitled"
	}

	slug := strings.TrimSpace(name)
	slug = strings.ToLower(slug)

	slug = strings.ReplaceAll(slug, "-", " ")
	slug = strings.ReplaceAll(slug, "_", " ")
	slug = strings.ReplaceAll(slug, ":", " ")
	slug = strings.ReplaceAll(slug, "|", " ")

	reg := regexp.MustCompile(`[^a-z0-9\s]`)
	slug = reg.ReplaceAllString(slug, "")

	spaceReg := regexp.MustCompile(`\s+`)
	slug = spaceReg.ReplaceAllString(slug, " ")
	slug = strings.TrimSpace(slug)

	slug = strings.ReplaceAll(slug, " ", "-")

	hyphenReg := regexp.MustCompile(`\-+`)
	slug = hyphenReg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "untitled"
	}
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

// GenerateUniqueSlug generates a unique slug with a counter suffix if needed
func (s Sanitize) GenerateUniqueSlug(baseSlug string, excludeID string, existsFunc func(slug string, excludeID string) bool) string {
	if baseSlug == "" {
		baseSlug = "untitled"
	}

	if !existsFunc(baseSlug, excludeID) {
		return baseSlug
	}

	maxAttempts := 100
	for i := 1; i <= maxAttempts; i++ {
		candidate := fmt.Sprintf("%s-%d", baseSlug, i)
		if !existsFunc(candidate, excludeID) {
			return candidate
		}
	}

	timestamp := time.Now().UnixNano() % 100000
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

// GenerateUniqueSlugWithTimestamp generates a unique slug with timestamp suffix
func (s Sanitize) GenerateUniqueSlugWithTimestamp(baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "untitled"
	}
	timestamp := time.Now().UnixNano() % 100000
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

// GenerateUniqueSlugWithRandom generates a unique slug with random suffix
func (s Sanitize) GenerateUniqueSlugWithRandom(baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "untitled"
	}
	randStr := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	return fmt.Sprintf("%s-%d", baseSlug, randStr)
}

// GenerateSlugWithID generates a slug with an ID suffix
func (s Sanitize) GenerateSlugWithID(name string, id string) string {
	baseSlug := s.GenerateSlugFromName(name)
	if id == "" {
		return baseSlug
	}
	return baseSlug + "-" + id
}

// GenerateColorSlug generates a slug for color names
func (s Sanitize) GenerateColorSlug(colorName string) string {
	if colorName == "" {
		return "unknown-color"
	}

	slug := strings.ToLower(colorName)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")

	hyphenReg := regexp.MustCompile(`\-+`)
	slug = hyphenReg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "unknown-color"
	}
	return slug
}

// Identifier sanitizes text for identifiers (lowercase, underscores)
//
// Keeps: lowercase letters, numbers, hyphens, underscores.
//
// WARNING: This strips `@`, `.`, `+`, and other syntax characters.
// NEVER use Identifier for emails or phone numbers — use Email or Phone.
// Identifier is for usernames, slugs, and machine-readable keys only.
func (s Sanitize) Identifier(text string) string {
	if text == "" {
		return ""
	}

	identifier := strings.ToLower(text)
	reg := regexp.MustCompile(`[^a-z0-9\-_]`)
	identifier = reg.ReplaceAllString(identifier, "")
	identifier = strings.ReplaceAll(identifier, " ", "_")
	underscoreReg := regexp.MustCompile(`_+`)
	identifier = underscoreReg.ReplaceAllString(identifier, "_")
	identifier = strings.Trim(identifier, "_")

	if len(identifier) > 100 {
		identifier = identifier[:100]
	}
	return identifier
}

// ============================================================
// EMAIL AND PHONE SANITIZATION
// ============================================================
//
// These are intentionally permissive. They preserve syntax characters
// (@, ., +, -, _) because stripping them corrupts the value.
//
// Deep format validation happens in the Email/Phone validators
// (validator.Email / validator.Phone in the handler layer), not here.

// Email performs minimal sanitization on an email address.
//
// - Trims leading and trailing whitespace
// - Lowercases the entire string
// - Removes internal whitespace (spaces are never valid in an email)
//
// Does NOT strip `@`, `.`, `+`, `-`, or `_`.
func (s Sanitize) Email(email string) string {
	if email == "" {
		return ""
	}

	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	// Remove any internal whitespace.
	spaceReg := regexp.MustCompile(`\s+`)
	email = spaceReg.ReplaceAllString(email, "")

	if len(email) > 254 {
		email = email[:254]
	}
	return email
}

// Phone performs minimal sanitization on a phone number.
//
// - Trims leading and trailing whitespace
// - Removes whitespace and common formatting characters (parentheses, dashes)
// - PRESERVES a leading `+` for E.164 numbers
//
// Deep format validation happens in the Phone validator.
func (s Sanitize) Phone(phone string) string {
	if phone == "" {
		return ""
	}

	phone = strings.TrimSpace(phone)

	// Remove whitespace, parentheses, and dashes — but keep `+`.
	cleaner := regexp.MustCompile(`[\s()\-]`)
	phone = cleaner.ReplaceAllString(phone, "")

	if len(phone) > 20 {
		phone = phone[:20]
	}
	return phone
}

// ============================================================
// VALIDATION
// ============================================================

// ValidateName validates a name
func (s Sanitize) ValidateName(name string) (bool, string) {
	if name == "" {
		return false, "name is required"
	}
	if len(name) < 3 {
		return false, "name must be at least 3 characters"
	}
	if len(name) > 100 {
		return false, "name must be less than 100 characters"
	}

	validChars := regexp.MustCompile(`^[a-zA-Z0-9_\-']+$`)
	if !validChars.MatchString(name) {
		return false, "name can only contain letters, numbers, underscores, hyphens, and apostrophes"
	}

	for _, r := range name {
		if unicode.IsSymbol(r) || unicode.IsMark(r) {
			return false, "name cannot contain emojis or special symbols"
		}
	}
	return true, ""
}

// ValidateDisplayName validates a display name
func (s Sanitize) ValidateDisplayName(name string) (bool, string) {
	if name == "" {
		return false, "display name is required"
	}
	if len(name) < 1 {
		return false, "display name must be at least 1 character"
	}
	if len(name) > 200 {
		return false, "display name must be less than 200 characters"
	}
	return true, ""
}

// ValidateSlug validates a slug
func (s Sanitize) ValidateSlug(slug string) (bool, string) {
	if slug == "" {
		return false, "slug is required"
	}
	if len(slug) < 3 {
		return false, "slug must be at least 3 characters"
	}
	if len(slug) > 100 {
		return false, "slug must be less than 100 characters"
	}

	validChars := regexp.MustCompile(`^[a-z0-9\-]+$`)
	if !validChars.MatchString(slug) {
		return false, "slug can only contain lowercase letters, numbers, and hyphens"
	}
	return true, ""
}

// ============================================================
// UTILITY FUNCTIONS
// ============================================================

// RemoveEmojis removes all emojis and special symbols from text
func (s Sanitize) RemoveEmojis(text string) string {
	var result strings.Builder
	for _, r := range text {
		if !unicode.IsSymbol(r) && !unicode.IsMark(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Truncate truncates text to the specified length
func (s Sanitize) Truncate(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength]
}

// DefaultName returns a default name if empty
func (s Sanitize) DefaultName() string {
	return "untitled"
}

// DefaultDisplayName returns a default display name
func (s Sanitize) DefaultDisplayName() string {
	return "Untitled"
}

// DefaultSlug returns a default slug
func (s Sanitize) DefaultSlug() string {
	return "untitled"
}



// URL performs minimal sanitization on a URL.
//
// - Trims leading and trailing whitespace
// - Ensures a scheme exists (prepends https:// if missing)
//
// Preserves `:`, `/`, `.`, `?`, `=`, `&`, `#`, `-`, `_`.
// Deep format validation happens in the URL validator.
func (s Sanitize) URL(url string) string {
	if url == "" {
		return ""
	}

	url = strings.TrimSpace(url)

	// If no scheme, default to https://
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}

	if len(url) > 500 {
		url = url[:500]
	}
	return url
}