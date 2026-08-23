// internal/modules/events/domain/event_status_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// EVENT STATUS - Value Object
// ============================================================

// EventStatusValue is an alias for the shared EventStatus
type EventStatusValue = types.EventStatus

// Status constants - Re-exported from shared types for convenience
const (
	EventStatusDraft     = types.EventStatusDraft
	EventStatusPublished = types.EventStatusPublished
	EventStatusCancelled = types.EventStatusCancelled
	EventStatusCompleted = types.EventStatusCompleted
)

// AllEventStatuses re-exported from shared types
var AllEventStatuses = types.AllEventStatuses

// EventStatusInfo holds metadata for each event status
type EventStatusInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Color       string
	Icon        string
	SortOrder   int
	IsFinal     bool
	IsActive    bool
}

// ============================================================
// STATUS REGISTRY - Domain specific wrapper
// ============================================================

// eventStatusRegistry is the domain's source of truth for status metadata
// ✅ Make this exported if needed elsewhere
var eventStatusRegistry = map[types.EventStatus]EventStatusInfo{
	types.EventStatusDraft: {
		Slug:        types.EventStatusDraftSlug,
		Name:        types.EventStatusDraftName,
		DisplayName: types.EventStatusDraftDisplayName,
		Description: types.EventStatusDraftDescription,
		Color:       types.EventStatusDraftColor,
		Icon:        types.EventStatusDraftIcon,
		SortOrder:   types.EventStatusDraftSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
	types.EventStatusPublished: {
		Slug:        types.EventStatusPublishedSlug,
		Name:        types.EventStatusPublishedName,
		DisplayName: types.EventStatusPublishedDisplayName,
		Description: types.EventStatusPublishedDescription,
		Color:       types.EventStatusPublishedColor,
		Icon:        types.EventStatusPublishedIcon,
		SortOrder:   types.EventStatusPublishedSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
	types.EventStatusCancelled: {
		Slug:        types.EventStatusCancelledSlug,
		Name:        types.EventStatusCancelledName,
		DisplayName: types.EventStatusCancelledDisplayName,
		Description: types.EventStatusCancelledDescription,
		Color:       types.EventStatusCancelledColor,
		Icon:        types.EventStatusCancelledIcon,
		SortOrder:   types.EventStatusCancelledSortOrder,
		IsFinal:     true,
		IsActive:    true,
	},
	types.EventStatusCompleted: {
		Slug:        types.EventStatusCompletedSlug,
		Name:        types.EventStatusCompletedName,
		DisplayName: types.EventStatusCompletedDisplayName,
		Description: types.EventStatusCompletedDescription,
		Color:       types.EventStatusCompletedColor,
		Icon:        types.EventStatusCompletedIcon,
		SortOrder:   types.EventStatusCompletedSortOrder,
		IsFinal:     true,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS (not methods on the type)
// ============================================================

// GetEventStatusInfo returns the status info for a given status
func GetEventStatusInfo(status EventStatusValue) (EventStatusInfo, bool) {
	info, ok := eventStatusRegistry[status]
	return info, ok
}

// GetEventStatusSlug returns the slug for a status
func GetEventStatusSlug(status EventStatusValue) string {
	return status.GetSlug()
}

// GetEventStatusName returns the name for a status
func GetEventStatusName(status EventStatusValue) string {
	return status.GetName()
}

// GetEventStatusDisplayName returns the display name for a status
func GetEventStatusDisplayName(status EventStatusValue) string {
	return status.GetDisplayName()
}

// IsEventStatusValid checks if the status is valid
func IsEventStatusValid(status EventStatusValue) bool {
	return status.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllEventStatusInfos returns all status infos
func AllEventStatusInfos() []EventStatusInfo {
	infos := make([]EventStatusInfo, 0, len(eventStatusRegistry))
	for _, info := range eventStatusRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllEventStatusSlugs returns all status slugs (with hyphens)
func AllEventStatusSlugs() []string {
	return types.AllEventStatusSlugs()
}

// AllEventStatusNames returns all internal status names (with underscores)
func AllEventStatusNames() []string {
	return types.AllEventStatusNames()
}

// ActiveEventStatusInfos returns only active status infos
func ActiveEventStatusInfos() []EventStatusInfo {
	infos := make([]EventStatusInfo, 0)
	for _, info := range eventStatusRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetEventStatusBySlug returns status info by slug
func GetEventStatusBySlug(slug string) (EventStatusInfo, bool) {
	for status, info := range eventStatusRegistry {
		if status.GetSlug() == slug {
			return info, true
		}
	}
	return EventStatusInfo{}, false
}

// GetEventStatusByName returns status info by internal name (with underscores)
func GetEventStatusByName(name string) (EventStatusInfo, bool) {
	status, ok := types.ParseEventStatus(name)
	if !ok {
		return EventStatusInfo{}, false
	}
	return GetEventStatusInfo(status)
}

// GetEventStatusByDisplayName returns status info by display name
func GetEventStatusByDisplayName(displayName string) (EventStatusInfo, bool) {
	for _, info := range eventStatusRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return EventStatusInfo{}, false
}