// internal/shared/types/recurrence.go

package types

// ============================================================
// RECURRENCE PATTERN CONSTANTS
// ============================================================

// RecurrencePattern represents the pattern of a recurring event
type RecurrencePattern string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	RecurrenceDailySlug   = "daily"
	RecurrenceWeeklySlug  = "weekly"
	RecurrenceMonthlySlug = "monthly"
	RecurrenceCustomSlug  = "custom"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	RecurrenceDailyName   = "recurrence_daily"
	RecurrenceWeeklyName  = "recurrence_weekly"
	RecurrenceMonthlyName = "recurrence_monthly"
	RecurrenceCustomName  = "recurrence_custom"
)

// Display name constants - Used for UI display
const (
	RecurrenceDailyDisplayName   = "Daily"
	RecurrenceWeeklyDisplayName  = "Weekly"
	RecurrenceMonthlyDisplayName = "Monthly"
	RecurrenceCustomDisplayName  = "Custom"
)

// Description constants
const (
	RecurrenceDailyDescription   = "Repeats every day"
	RecurrenceWeeklyDescription  = "Repeats on selected days of the week"
	RecurrenceMonthlyDescription = "Repeats on selected days of the month"
	RecurrenceCustomDescription  = "Custom recurrence pattern"
)

// ============================================================
// RECURRENCE PATTERN DEFINITIONS
// ============================================================

// RecurrencePattern constants - These are the NAMES (with underscores)
const (
	RecurrenceDaily   RecurrencePattern = RecurrenceDailyName   // "recurrence_daily"
	RecurrenceWeekly  RecurrencePattern = RecurrenceWeeklyName  // "recurrence_weekly"
	RecurrenceMonthly RecurrencePattern = RecurrenceMonthlyName // "recurrence_monthly"
	RecurrenceCustom  RecurrencePattern = RecurrenceCustomName  // "recurrence_custom"
)

// AllRecurrencePatterns lists all valid recurrence patterns for validation
var AllRecurrencePatterns = []RecurrencePattern{
	RecurrenceDaily,
	RecurrenceWeekly,
	RecurrenceMonthly,
	RecurrenceCustom,
}

// ============================================================
// WEEKDAY CONSTANTS
// ============================================================

// Weekday represents a day of the week
type Weekday string

const (
	WeekdayMonday    Weekday = "monday"
	WeekdayTuesday   Weekday = "tuesday"
	WeekdayWednesday Weekday = "wednesday"
	WeekdayThursday  Weekday = "thursday"
	WeekdayFriday    Weekday = "friday"
	WeekdaySaturday  Weekday = "saturday"
	WeekdaySunday    Weekday = "sunday"
)

// AllWeekdays lists all valid weekdays
var AllWeekdays = []Weekday{
	WeekdayMonday,
	WeekdayTuesday,
	WeekdayWednesday,
	WeekdayThursday,
	WeekdayFriday,
	WeekdaySaturday,
	WeekdaySunday,
}

func (w Weekday) String() string { return string(w) }

func (w Weekday) IsValid() bool {
	for _, day := range AllWeekdays {
		if day == w {
			return true
		}
	}
	return false
}

// DisplayName returns the display name for a weekday
func (w Weekday) DisplayName() string {
	switch w {
	case WeekdayMonday:
		return "Monday"
	case WeekdayTuesday:
		return "Tuesday"
	case WeekdayWednesday:
		return "Wednesday"
	case WeekdayThursday:
		return "Thursday"
	case WeekdayFriday:
		return "Friday"
	case WeekdaySaturday:
		return "Saturday"
	case WeekdaySunday:
		return "Sunday"
	default:
		return string(w)
	}
}

// ============================================================
// BASIC METHODS - On RecurrencePattern
// ============================================================

// String returns the string representation (the name with underscores)
func (r RecurrencePattern) String() string {
	return string(r)
}

// IsValid checks if the recurrence pattern is valid
func (r RecurrencePattern) IsValid() bool {
	for _, pattern := range AllRecurrencePatterns {
		if pattern == r {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On RecurrencePattern
// ============================================================

// GetName returns the internal name (with underscores)
func (r RecurrencePattern) GetName() string {
	return string(r)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (r RecurrencePattern) GetSlug() string {
	switch r {
	case RecurrenceDaily:
		return RecurrenceDailySlug
	case RecurrenceWeekly:
		return RecurrenceWeeklySlug
	case RecurrenceMonthly:
		return RecurrenceMonthlySlug
	case RecurrenceCustom:
		return RecurrenceCustomSlug
	default:
		return string(r)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (r RecurrencePattern) GetDisplayName() string {
	switch r {
	case RecurrenceDaily:
		return RecurrenceDailyDisplayName
	case RecurrenceWeekly:
		return RecurrenceWeeklyDisplayName
	case RecurrenceMonthly:
		return RecurrenceMonthlyDisplayName
	case RecurrenceCustom:
		return RecurrenceCustomDisplayName
	default:
		return string(r)
	}
}

// GetDescription returns the description
func (r RecurrencePattern) GetDescription() string {
	switch r {
	case RecurrenceDaily:
		return RecurrenceDailyDescription
	case RecurrenceWeekly:
		return RecurrenceWeeklyDescription
	case RecurrenceMonthly:
		return RecurrenceMonthlyDescription
	case RecurrenceCustom:
		return RecurrenceCustomDescription
	default:
		return ""
	}
}

// ============================================================
// PARSE FUNCTIONS - For RecurrencePattern
// ============================================================

// ParseRecurrencePattern parses a string into a RecurrencePattern
// Expects the name (with underscores), not the slug
func ParseRecurrencePattern(name string) (RecurrencePattern, bool) {
	r := RecurrencePattern(name)
	if r.IsValid() {
		return r, true
	}
	return "", false
}

// ParseRecurrencePatternWithDefault parses a string or returns a default
func ParseRecurrencePatternWithDefault(name string, defaultPattern RecurrencePattern) RecurrencePattern {
	if r, ok := ParseRecurrencePattern(name); ok {
		return r
	}
	return defaultPattern
}

// ParseRecurrencePatternBySlug parses a slug string into a RecurrencePattern
func ParseRecurrencePatternBySlug(slug string) (RecurrencePattern, bool) {
	for _, r := range AllRecurrencePatterns {
		if r.GetSlug() == slug {
			return r, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllRecurrencePatternNames returns all internal recurrence pattern names (with underscores)
func AllRecurrencePatternNames() []string {
	names := make([]string, 0, len(AllRecurrencePatterns))
	for _, r := range AllRecurrencePatterns {
		names = append(names, r.GetName())
	}
	return names
}

// AllRecurrencePatternSlugs returns all recurrence pattern slugs (with hyphens)
func AllRecurrencePatternSlugs() []string {
	slugs := make([]string, 0, len(AllRecurrencePatterns))
	for _, r := range AllRecurrencePatterns {
		slugs = append(slugs, r.GetSlug())
	}
	return slugs
}

// AllRecurrencePatternDisplayNames returns all recurrence pattern display names
func AllRecurrencePatternDisplayNames() []string {
	names := make([]string, 0, len(AllRecurrencePatterns))
	for _, r := range AllRecurrencePatterns {
		names = append(names, r.GetDisplayName())
	}
	return names
}