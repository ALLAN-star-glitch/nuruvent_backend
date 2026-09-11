package domain

// ============================================================
// RECURRENCE PATTERN - Lookup Entity
// ============================================================

// RecurrencePattern represents a row in the `recurrence_patterns` lookup table.
//
// Distinct from a "recurrence rule": this entity is the pattern KIND
// (daily/weekly/monthly/custom), not the event's specific configuration.
// An event references a pattern via events.recurrence_pattern_id (UUID FK),
// and stores its own interval/days/ends_on columns alongside.
type RecurrencePattern struct {
	ID          string
	Slug        string // "daily" | "weekly" | "monthly" | "custom"
	Name        string // "recurrence_daily" etc.
	DisplayName string // "Daily" etc.
	Description string
	IsActive    bool
}

// IsValidRecurrencePattern checks if a pattern slug is recognized.
func IsValidRecurrencePattern(pattern string) bool {
	switch pattern {
	case "daily", "weekly", "monthly", "custom":
		return true
	}
	return false
}

// IsValidWeekday checks if a weekday name is valid.
func IsValidWeekday(day string) bool {
	switch day {
	case "monday", "tuesday", "wednesday", "thursday",
		"friday", "saturday", "sunday":
		return true
	}
	return false
}

// IsValidWeekOfMonth checks if a week-of-month descriptor is valid.
func IsValidWeekOfMonth(week string) bool {
	switch week {
	case "first", "second", "third", "fourth", "last":
		return true
	}
	return false
}