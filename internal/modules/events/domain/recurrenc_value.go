// internal/modules/events/domain/recurrence_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// RECURRENCE PATTERN - Value Object
// ============================================================

// RecurrencePatternValue is an alias for the shared RecurrencePattern
type RecurrencePatternValue = types.RecurrencePattern

// Pattern constants - Re-exported from shared types for convenience
const (
	RecurrenceDaily   = types.RecurrenceDaily
	RecurrenceWeekly  = types.RecurrenceWeekly
	RecurrenceMonthly = types.RecurrenceMonthly
	RecurrenceCustom  = types.RecurrenceCustom
)

// AllRecurrencePatterns re-exported from shared types
var AllRecurrencePatterns = types.AllRecurrencePatterns

// RecurrencePatternInfo holds metadata for each recurrence pattern
type RecurrencePatternInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	IsActive    bool
}

// ============================================================
// PATTERN REGISTRY - Domain specific wrapper
// ============================================================

var recurrencePatternRegistry = map[types.RecurrencePattern]RecurrencePatternInfo{
	types.RecurrenceDaily: {
		Slug:        types.RecurrenceDailySlug,
		Name:        types.RecurrenceDailyName,
		DisplayName: types.RecurrenceDailyDisplayName,
		Description: types.RecurrenceDailyDescription,
		IsActive:    true,
	},
	types.RecurrenceWeekly: {
		Slug:        types.RecurrenceWeeklySlug,
		Name:        types.RecurrenceWeeklyName,
		DisplayName: types.RecurrenceWeeklyDisplayName,
		Description: types.RecurrenceWeeklyDescription,
		IsActive:    true,
	},
	types.RecurrenceMonthly: {
		Slug:        types.RecurrenceMonthlySlug,
		Name:        types.RecurrenceMonthlyName,
		DisplayName: types.RecurrenceMonthlyDisplayName,
		Description: types.RecurrenceMonthlyDescription,
		IsActive:    true,
	},
	types.RecurrenceCustom: {
		Slug:        types.RecurrenceCustomSlug,
		Name:        types.RecurrenceCustomName,
		DisplayName: types.RecurrenceCustomDisplayName,
		Description: types.RecurrenceCustomDescription,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetRecurrencePatternInfo returns the pattern info for a given recurrence pattern
func GetRecurrencePatternInfo(pattern RecurrencePatternValue) (RecurrencePatternInfo, bool) {
	info, ok := recurrencePatternRegistry[pattern]
	return info, ok
}

// GetRecurrencePatternSlug returns the slug for a recurrence pattern
func GetRecurrencePatternSlug(pattern RecurrencePatternValue) string {
	return pattern.GetSlug()
}

// GetRecurrencePatternName returns the name for a recurrence pattern
func GetRecurrencePatternName(pattern RecurrencePatternValue) string {
	return pattern.GetName()
}

// GetRecurrencePatternDisplayName returns the display name for a recurrence pattern
func GetRecurrencePatternDisplayName(pattern RecurrencePatternValue) string {
	return pattern.GetDisplayName()
}

// IsRecurrencePatternValid checks if the recurrence pattern is valid
func IsRecurrencePatternValid(pattern RecurrencePatternValue) bool {
	return pattern.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllRecurrencePatternInfos returns all pattern infos
func AllRecurrencePatternInfos() []RecurrencePatternInfo {
	infos := make([]RecurrencePatternInfo, 0, len(recurrencePatternRegistry))
	for _, info := range recurrencePatternRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllRecurrencePatternSlugs returns all pattern slugs (with hyphens)
func AllRecurrencePatternSlugs() []string {
	return types.AllRecurrencePatternSlugs()
}

// AllRecurrencePatternNames returns all internal pattern names (with underscores)
func AllRecurrencePatternNames() []string {
	return types.AllRecurrencePatternNames()
}

// ActiveRecurrencePatternInfos returns only active pattern infos
func ActiveRecurrencePatternInfos() []RecurrencePatternInfo {
	infos := make([]RecurrencePatternInfo, 0)
	for _, info := range recurrencePatternRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetRecurrencePatternBySlug returns pattern info by slug
func GetRecurrencePatternBySlug(slug string) (RecurrencePatternInfo, bool) {
	for pattern, info := range recurrencePatternRegistry {
		if pattern.GetSlug() == slug {
			return info, true
		}
	}
	return RecurrencePatternInfo{}, false
}

// GetRecurrencePatternByName returns pattern info by internal name (with underscores)
func GetRecurrencePatternByName(name string) (RecurrencePatternInfo, bool) {
	pattern, ok := types.ParseRecurrencePattern(name)
	if !ok {
		return RecurrencePatternInfo{}, false
	}
	return GetRecurrencePatternInfo(pattern)
}

// GetRecurrencePatternByDisplayName returns pattern info by display name
func GetRecurrencePatternByDisplayName(displayName string) (RecurrencePatternInfo, bool) {
	for _, info := range recurrencePatternRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return RecurrencePatternInfo{}, false
}