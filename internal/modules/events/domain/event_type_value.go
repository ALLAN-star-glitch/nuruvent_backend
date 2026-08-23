// internal/modules/events/domain/event_type_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// EVENT TYPE - Value Object
// ============================================================

// EventTypeValue is an alias for the shared EventType
type EventTypeValue = types.EventType

// Type constants - Re-exported from shared types for convenience
const (
	EventTypeUncategorized = types.EventTypeUncategorized
	EventTypeWorkshop      = types.EventTypeWorkshop
	EventTypeWebinar       = types.EventTypeWebinar
	EventTypeMeetup        = types.EventTypeMeetup
	EventTypeBootcamp      = types.EventTypeBootcamp
)

// AllEventTypes re-exported from shared types
var AllEventTypes = types.AllEventTypes

// EventTypeInfo holds metadata for each event type
type EventTypeInfo struct {
	Slug                string
	Name                string
	DisplayName         string
	Description         string
	Icon                string
	Color               string
	SortOrder           int
	SupportsCertificate bool
	MinDuration         int
	MaxDuration         int
	IsActive            bool
}

// ============================================================
// TYPE REGISTRY - Domain specific wrapper
// ============================================================

var eventTypeRegistry = map[types.EventType]EventTypeInfo{
	types.EventTypeUncategorized: {
		Slug:                types.EventTypeUncategorizedSlug,
		Name:                types.EventTypeUncategorizedName,
		DisplayName:         types.EventTypeUncategorizedDisplayName,
		Description:         types.EventTypeUncategorizedDescription,
		Icon:                types.EventTypeUncategorizedIcon,
		Color:               types.EventTypeUncategorizedColor,
		SortOrder:           types.EventTypeUncategorizedSortOrder,
		SupportsCertificate: false,
		MinDuration:         0,
		MaxDuration:         1440,
		IsActive:            true,
	},
	types.EventTypeWorkshop: {
		Slug:                types.EventTypeWorkshopSlug,
		Name:                types.EventTypeWorkshopName,
		DisplayName:         types.EventTypeWorkshopDisplayName,
		Description:         types.EventTypeWorkshopDescription,
		Icon:                types.EventTypeWorkshopIcon,
		Color:               types.EventTypeWorkshopColor,
		SortOrder:           types.EventTypeWorkshopSortOrder,
		SupportsCertificate: true,
		MinDuration:         120,
		MaxDuration:         480,
		IsActive:            true,
	},
	types.EventTypeWebinar: {
		Slug:                types.EventTypeWebinarSlug,
		Name:                types.EventTypeWebinarName,
		DisplayName:         types.EventTypeWebinarDisplayName,
		Description:         types.EventTypeWebinarDescription,
		Icon:                types.EventTypeWebinarIcon,
		Color:               types.EventTypeWebinarColor,
		SortOrder:           types.EventTypeWebinarSortOrder,
		SupportsCertificate: false,
		MinDuration:         30,
		MaxDuration:         120,
		IsActive:            true,
	},
	types.EventTypeMeetup: {
		Slug:                types.EventTypeMeetupSlug,
		Name:                types.EventTypeMeetupName,
		DisplayName:         types.EventTypeMeetupDisplayName,
		Description:         types.EventTypeMeetupDescription,
		Icon:                types.EventTypeMeetupIcon,
		Color:               types.EventTypeMeetupColor,
		SortOrder:           types.EventTypeMeetupSortOrder,
		SupportsCertificate: false,
		MinDuration:         60,
		MaxDuration:         180,
		IsActive:            true,
	},
	types.EventTypeBootcamp: {
		Slug:                types.EventTypeBootcampSlug,
		Name:                types.EventTypeBootcampName,
		DisplayName:         types.EventTypeBootcampDisplayName,
		Description:         types.EventTypeBootcampDescription,
		Icon:                types.EventTypeBootcampIcon,
		Color:               types.EventTypeBootcampColor,
		SortOrder:           types.EventTypeBootcampSortOrder,
		SupportsCertificate: true,
		MinDuration:         240,
		MaxDuration:         1440,
		IsActive:            true,
	},
}

// ============================================================
// HELPER FUNCTIONS (not methods on the type)
// ============================================================

// GetEventTypeInfo returns the type info for a given event type
func GetEventTypeInfo(eventType EventTypeValue) (EventTypeInfo, bool) {
	info, ok := eventTypeRegistry[eventType]
	return info, ok
}

// GetEventTypeSlug returns the slug for an event type
func GetEventTypeSlug(eventType EventTypeValue) string {
	return eventType.GetSlug()
}

// GetEventTypeName returns the name for an event type
func GetEventTypeName(eventType EventTypeValue) string {
	return eventType.GetName()
}

// GetEventTypeDisplayName returns the display name for an event type
func GetEventTypeDisplayName(eventType EventTypeValue) string {
	return eventType.GetDisplayName()
}

// IsEventTypeValid checks if the event type is valid
func IsEventTypeValid(eventType EventTypeValue) bool {
	return eventType.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllEventTypeInfos returns all type infos
func AllEventTypeInfos() []EventTypeInfo {
	infos := make([]EventTypeInfo, 0, len(eventTypeRegistry))
	for _, info := range eventTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllEventTypeSlugs returns all type slugs (with hyphens)
func AllEventTypeSlugs() []string {
	return types.AllEventTypeSlugs()
}

// AllEventTypeNames returns all internal type names (with underscores)
func AllEventTypeNames() []string {
	return types.AllEventTypeNames()
}

// AllEventTypeDisplayNames returns all display names
func AllEventTypeDisplayNames() []string {
	return types.AllEventTypeDisplayNames()
}

// ActiveEventTypeInfos returns only active type infos
func ActiveEventTypeInfos() []EventTypeInfo {
	infos := make([]EventTypeInfo, 0)
	for _, info := range eventTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// ActiveEventTypeSlugs returns only active type slugs (with hyphens)
func ActiveEventTypeSlugs() []string {
	slugs := make([]string, 0)
	for _, info := range eventTypeRegistry {
		if info.IsActive {
			slugs = append(slugs, info.Slug)
		}
	}
	return slugs
}

// ActiveEventTypeNames returns only active internal type names (with underscores)
func ActiveEventTypeNames() []string {
	names := make([]string, 0)
	for _, info := range eventTypeRegistry {
		if info.IsActive {
			names = append(names, info.Name)
		}
	}
	return names
}

// GetEventTypeBySlug returns type info by slug
func GetEventTypeBySlug(slug string) (EventTypeInfo, bool) {
	for _, info := range eventTypeRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return EventTypeInfo{}, false
}

// GetEventTypeByName returns type info by internal name (with underscores)
func GetEventTypeByName(name string) (EventTypeInfo, bool) {
	eventType, ok := types.ParseEventType(name)
	if !ok {
		return EventTypeInfo{}, false
	}
	return GetEventTypeInfo(eventType)
}

// GetEventTypeByDisplayName returns type info by display name
func GetEventTypeByDisplayName(displayName string) (EventTypeInfo, bool) {
	for _, info := range eventTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return EventTypeInfo{}, false
}

// GetEventTypeBySlugString returns type info by slug string (convenience)
func GetEventTypeBySlugString(slug string) (EventTypeInfo, bool) {
	for _, info := range eventTypeRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return EventTypeInfo{}, false
}