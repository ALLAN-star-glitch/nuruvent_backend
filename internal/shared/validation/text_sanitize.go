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
	
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	
	// Collapse multiple underscores
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
// Use cases: Event display names, Certificate titles, User display names (what users see)
func (s Sanitize) DisplayName(name string) string {
	if name == "" {
		return ""
	}

	// ✅ Only trim spaces - DO NOT remove emojis or special characters
	name = strings.TrimSpace(name)
	
	// ✅ Collapse multiple spaces into one
	spaceReg := regexp.MustCompile(`\s+`)
	name = spaceReg.ReplaceAllString(name, " ")
	
	// ✅ Trim again
	name = strings.TrimSpace(name)

	// Limit length
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
// Examples: "Tertiary Green" -> "tertiary-green", "Event 2024" -> "event-2024"
func (s Sanitize) Slug(text string) string {
	if text == "" {
		return "untitled"
	}

	// Convert to lowercase
	slug := strings.ToLower(text)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Replace underscores with hyphens (for consistency)
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove special characters (keep only letters, numbers, and hyphens)
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")

	// Collapse multiple hyphens into one
	hyphenReg := regexp.MustCompile(`\-+`)
	slug = hyphenReg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
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
// This is the recommended method for generating slugs for events, categories, etc.
// Examples: 
//   - "Tertiary Green" -> "tertiary-green"
//   - "Nuruvent Conference 2024" -> "nuruvent-conference-2024"
//   - "Workshop: Advanced Go" -> "workshop-advanced-go"
func (s Sanitize) GenerateSlugFromName(name string) string {
	if name == "" {
		return "untitled"
	}

	// Trim spaces
	slug := strings.TrimSpace(name)

	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Replace common separators with spaces
	slug = strings.ReplaceAll(slug, "-", " ")
	slug = strings.ReplaceAll(slug, "_", " ")
	slug = strings.ReplaceAll(slug, ":", " ")
	slug = strings.ReplaceAll(slug, "|", " ")

	// Remove special characters but keep letters, numbers, and spaces
	reg := regexp.MustCompile(`[^a-z0-9\s]`)
	slug = reg.ReplaceAllString(slug, "")

	// Collapse multiple spaces
	spaceReg := regexp.MustCompile(`\s+`)
	slug = spaceReg.ReplaceAllString(slug, " ")

	// Trim spaces
	slug = strings.TrimSpace(slug)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Collapse multiple hyphens
	hyphenReg := regexp.MustCompile(`\-+`)
	slug = hyphenReg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "untitled"
	}
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

// ✅ NEW: GenerateUniqueSlug generates a unique slug with a counter suffix if needed
// This should be used when creating or updating entities that require unique slugs
// The checker function should return true if the slug already exists
func (s Sanitize) GenerateUniqueSlug(baseSlug string, excludeID string, existsFunc func(slug string, excludeID string) bool) string {
	if baseSlug == "" {
		baseSlug = "untitled"
	}

	// Check if the base slug already exists
	if !existsFunc(baseSlug, excludeID) {
		return baseSlug
	}

	// Try with counter
	maxAttempts := 100
	for i := 1; i <= maxAttempts; i++ {
		candidate := fmt.Sprintf("%s-%d", baseSlug, i)
		if !existsFunc(candidate, excludeID) {
			return candidate
		}
	}

	// Fallback: use timestamp
	timestamp := time.Now().UnixNano() % 100000
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

// ✅ NEW: GenerateUniqueSlugWithTimestamp generates a unique slug with timestamp suffix
// Simpler version that just adds a timestamp to ensure uniqueness
func (s Sanitize) GenerateUniqueSlugWithTimestamp(baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "untitled"
	}
	timestamp := time.Now().UnixNano() % 100000
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

// ✅ NEW: GenerateUniqueSlugWithRandom generates a unique slug with random suffix
func (s Sanitize) GenerateUniqueSlugWithRandom(baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "untitled"
	}
	// Use a short random string
	randStr := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	return fmt.Sprintf("%s-%d", baseSlug, randStr)
}

// GenerateSlugWithID generates a slug with an ID suffix
// Example: "tertiary-green-123"
func (s Sanitize) GenerateSlugWithID(name string, id string) string {
	baseSlug := s.GenerateSlugFromName(name)
	if id == "" {
		return baseSlug
	}
	return baseSlug + "-" + id
}

// GenerateColorSlug generates a slug for color names
// Specifically for color-based slugs like "tertiary-green", "primary-blue"
// Examples: "Tertiary Green" -> "tertiary-green", "Primary Blue" -> "primary-blue"
func (s Sanitize) GenerateColorSlug(colorName string) string {
	if colorName == "" {
		return "unknown-color"
	}

	slug := strings.ToLower(colorName)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")
	
	// Collapse multiple hyphens
	hyphenReg := regexp.MustCompile(`\-+`)
	slug = hyphenReg.ReplaceAllString(slug, "-")
	
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "unknown-color"
	}
	return slug
}

// Identifier sanitizes text for identifiers (lowercase, underscores)
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