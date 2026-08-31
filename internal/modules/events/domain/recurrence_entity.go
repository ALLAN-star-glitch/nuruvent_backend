// internal/modules/events/domain/recurrence_entity.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// RECURRENCE - Domain Entity (Database Representation)
// ============================================================

// Recurrence represents the recurrence configuration for a recurring event
type Recurrence struct {
	ID           string
	EventID      string // Foreign key to Event
	Pattern      string // "daily", "weekly", "monthly", "custom"
	Interval     int    // Every N days/weeks/months
	DaysOfWeek   []string // For weekly: ["monday", "wednesday"]
	DayOfMonth   *int    // For monthly: 15 = 15th of month
	WeekOfMonth  *string // "first", "second", "third", "fourth", "last"
	EndsOn       *time.Time // Recurrence end date
	Occurrences  *int       // Number of occurrences (if no end date)
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// ============================================================
// FACTORY
// ============================================================

// NewRecurrence creates a new Recurrence entity
func NewRecurrence(
	eventID string,
	pattern string,
	interval int,
	daysOfWeek []string,
	dayOfMonth *int,
	weekOfMonth *string,
	endsOn *time.Time,
	occurrences *int,
) (*Recurrence, error) {
	if eventID == "" {
		return nil, ErrInvalidRecurrence
	}
	if pattern == "" {
		return nil, ErrInvalidRecurrence
	}
	if interval < 1 {
		interval = 1
	}

	// Validate pattern
	if !isValidPattern(pattern) {
		return nil, ErrInvalidRecurrence
	}

	// Validate based on pattern
	switch pattern {
	case "weekly":
		if len(daysOfWeek) == 0 {
			return nil, ErrInvalidRecurrence
		}
	case "monthly":
		if dayOfMonth == nil && weekOfMonth == nil {
			return nil, ErrInvalidRecurrence
		}
	}

	now := time.Now()
	return &Recurrence{
		ID:          generateRecurrenceID(),
		EventID:     eventID,
		Pattern:     pattern,
		Interval:    interval,
		DaysOfWeek:  daysOfWeek,
		DayOfMonth:  dayOfMonth,
		WeekOfMonth: weekOfMonth,
		EndsOn:      endsOn,
		Occurrences: occurrences,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ============================================================
// DOMAIN BEHAVIORS
// ============================================================

// IsValid checks if the recurrence entity is valid
func (r *Recurrence) IsValid() bool {
	return r.ID != "" && r.EventID != "" && r.Pattern != ""
}

// IsActiveType returns whether the recurrence is active
func (r *Recurrence) IsActiveType() bool {
	return r.IsActive
}

// Activate activates the recurrence
func (r *Recurrence) Activate() {
	r.IsActive = true
	r.UpdatedAt = time.Now()
}

// Deactivate deactivates the recurrence
func (r *Recurrence) Deactivate() {
	r.IsActive = false
	r.UpdatedAt = time.Now()
}

// UpdatePattern updates the recurrence pattern
func (r *Recurrence) UpdatePattern(pattern string) error {
	if pattern == "" || !isValidPattern(pattern) {
		return ErrInvalidRecurrence
	}
	r.Pattern = pattern
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateInterval updates the interval
func (r *Recurrence) UpdateInterval(interval int) {
	if interval < 1 {
		interval = 1
	}
	r.Interval = interval
	r.UpdatedAt = time.Now()
}

// UpdateDaysOfWeek updates the days of week (for weekly pattern)
func (r *Recurrence) UpdateDaysOfWeek(daysOfWeek []string) error {
	if len(daysOfWeek) == 0 {
		return ErrInvalidRecurrence
	}
	r.DaysOfWeek = daysOfWeek
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateDayOfMonth updates the day of month (for monthly pattern)
func (r *Recurrence) UpdateDayOfMonth(dayOfMonth *int) {
	r.DayOfMonth = dayOfMonth
	r.UpdatedAt = time.Now()
}

// UpdateWeekOfMonth updates the week of month (for monthly pattern)
func (r *Recurrence) UpdateWeekOfMonth(weekOfMonth *string) {
	r.WeekOfMonth = weekOfMonth
	r.UpdatedAt = time.Now()
}

// UpdateEndsOn updates the end date
func (r *Recurrence) UpdateEndsOn(endsOn *time.Time) {
	r.EndsOn = endsOn
	r.UpdatedAt = time.Now()
}

// UpdateOccurrences updates the number of occurrences
func (r *Recurrence) UpdateOccurrences(occurrences *int) {
	r.Occurrences = occurrences
	r.UpdatedAt = time.Now()
}

// HasEndDate checks if the recurrence has an end date
func (r *Recurrence) HasEndDate() bool {
	return r.EndsOn != nil
}

// HasOccurrenceLimit checks if the recurrence has an occurrence limit
func (r *Recurrence) HasOccurrenceLimit() bool {
	return r.Occurrences != nil && *r.Occurrences > 0
}

// IsInfinite checks if the recurrence never ends
func (r *Recurrence) IsInfinite() bool {
	return !r.HasEndDate() && !r.HasOccurrenceLimit()
}

// ============================================================
// VALIDATION HELPERS
// ============================================================

func isValidPattern(pattern string) bool {
	validPatterns := []string{"daily", "weekly", "monthly", "custom"}
	for _, p := range validPatterns {
		if p == pattern {
			return true
		}
	}
	return false
}

// IsValidWeekday checks if the weekday is valid
func IsValidWeekday(day string) bool {
	validDays := []string{
		"monday", "tuesday", "wednesday", "thursday",
		"friday", "saturday", "sunday",
	}
	for _, d := range validDays {
		if d == day {
			return true
		}
	}
	return false
}

// IsValidWeekOfMonth checks if the week of month is valid
func IsValidWeekOfMonth(week string) bool {
	validWeeks := []string{"first", "second", "third", "fourth", "last"}
	for _, w := range validWeeks {
		if w == week {
			return true
		}
	}
	return false
}

// ============================================================
// HELPERS
// ============================================================

func generateRecurrenceID() string {
	return uuid.New().String()
}