// internal/modules/events/domain/event_format_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// EVENT FORMAT - Value Object
// ============================================================

// EventFormatValue is an alias for the shared EventFormat
type EventFormatValue = types.EventFormat

// Format constants - Re-exported from shared types for convenience
const (
	EventFormatVirtual  = types.EventFormatVirtual
	EventFormatInPerson = types.EventFormatInPerson
	EventFormatHybrid   = types.EventFormatHybrid
)

// AllEventFormats re-exported from shared types
var AllEventFormats = types.AllEventFormats

// EventFormatInfo holds metadata for each event format
type EventFormatInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	IsActive    bool
}

// ============================================================
// FORMAT REGISTRY - Domain specific wrapper
// ============================================================

var eventFormatRegistry = map[types.EventFormat]EventFormatInfo{
	types.EventFormatVirtual: {
		Slug:        types.EventFormatVirtualSlug,
		Name:        types.EventFormatVirtualName,
		DisplayName: types.EventFormatVirtualDisplayName,
		Description: types.EventFormatVirtualDescription,
		Icon:        types.EventFormatVirtualIcon,
		IsActive:    true,
	},
	types.EventFormatInPerson: {
		Slug:        types.EventFormatInPersonSlug,
		Name:        types.EventFormatInPersonName,
		DisplayName: types.EventFormatInPersonDisplayName,
		Description: types.EventFormatInPersonDescription,
		Icon:        types.EventFormatInPersonIcon,
		IsActive:    true,
	},
	types.EventFormatHybrid: {
		Slug:        types.EventFormatHybridSlug,
		Name:        types.EventFormatHybridName,
		DisplayName: types.EventFormatHybridDisplayName,
		Description: types.EventFormatHybridDescription,
		Icon:        types.EventFormatHybridIcon,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetEventFormatInfo returns the format info for a given event format
func GetEventFormatInfo(format EventFormatValue) (EventFormatInfo, bool) {
	info, ok := eventFormatRegistry[format]
	return info, ok
}

// GetEventFormatSlug returns the slug for an event format
func GetEventFormatSlug(format EventFormatValue) string {
	return format.GetSlug()
}

// GetEventFormatName returns the name for an event format
func GetEventFormatName(format EventFormatValue) string {
	return format.GetName()
}

// GetEventFormatDisplayName returns the display name for an event format
func GetEventFormatDisplayName(format EventFormatValue) string {
	return format.GetDisplayName()
}

// IsEventFormatValid checks if the event format is valid
func IsEventFormatValid(format EventFormatValue) bool {
	return format.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllEventFormatInfos returns all format infos
func AllEventFormatInfos() []EventFormatInfo {
	infos := make([]EventFormatInfo, 0, len(eventFormatRegistry))
	for _, info := range eventFormatRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllEventFormatSlugs returns all format slugs (with hyphens)
func AllEventFormatSlugs() []string {
	return types.AllEventFormatSlugs()
}

// AllEventFormatNames returns all internal format names (with underscores)
func AllEventFormatNames() []string {
	return types.AllEventFormatNames()
}

// ActiveEventFormatInfos returns only active format infos
func ActiveEventFormatInfos() []EventFormatInfo {
	infos := make([]EventFormatInfo, 0)
	for _, info := range eventFormatRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetEventFormatBySlug returns format info by slug
func GetEventFormatBySlug(slug string) (EventFormatInfo, bool) {
	for format, info := range eventFormatRegistry {
		if format.GetSlug() == slug {
			return info, true
		}
	}
	return EventFormatInfo{}, false
}

// GetEventFormatByName returns format info by internal name (with underscores)
func GetEventFormatByName(name string) (EventFormatInfo, bool) {
	format, ok := types.ParseEventFormat(name)
	if !ok {
		return EventFormatInfo{}, false
	}
	return GetEventFormatInfo(format)
}

// GetEventFormatByDisplayName returns format info by display name
func GetEventFormatByDisplayName(displayName string) (EventFormatInfo, bool) {
	for _, info := range eventFormatRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return EventFormatInfo{}, false
}