// internal/constants/event_statuses.go

package constants

// EventStatusInfo holds information about an event status
type EventStatusInfo struct {
	Slug        string // Database value (e.g., "draft")
	Name        string // Canonical name (e.g., "Draft")
	DisplayName string // UI display (e.g., "📝 Draft")
	Description string
	Color       string
	Icon        string
	SortOrder   int
	IsFinal     bool
	CanEdit     bool
	CanRegister bool
}

// EventStatuses - the source of truth for event statuses
var EventStatuses = []EventStatusInfo{
	{
		Slug:        "draft",
		Name:        "Draft",
		DisplayName: "📝 Draft",
		Description: "Event is being created and not yet visible to the public",
		Color:       "#6B7280", // Gray
		Icon:        "pencil",
		SortOrder:   1,
		IsFinal:     false,
		CanEdit:     true,
		CanRegister: false,
	},
	{
		Slug:        "published",
		Name:        "Published",
		DisplayName: "🌐 Published",
		Description: "Event is live and visible to the public",
		Color:       "#10B981", // Green
		Icon:        "globe",
		SortOrder:   2,
		IsFinal:     false,
		CanEdit:     true,
		CanRegister: true,
	},
	{
		Slug:        "cancelled",
		Name:        "Cancelled",
		DisplayName: "🚫 Cancelled",
		Description: "Event has been cancelled",
		Color:       "#EF4444", // Red
		Icon:        "x-circle",
		SortOrder:   3,
		IsFinal:     true,
		CanEdit:     false,
		CanRegister: false,
	},
	{
		Slug:        "completed",
		Name:        "Completed",
		DisplayName: "✅ Completed",
		Description: "Event has been completed",
		Color:       "#3B82F6", // Blue
		Icon:        "check-circle",
		SortOrder:   4,
		IsFinal:     true,
		CanEdit:     false,
		CanRegister: false,
	},
}

// EventStatusMap for quick lookups
var EventStatusMap = map[string]EventStatusInfo{
	"draft":     EventStatuses[0],
	"published": EventStatuses[1],
	"cancelled": EventStatuses[2],
	"completed": EventStatuses[3],
}

// AllEventStatusSlugs returns all valid event status slugs
func AllEventStatusSlugs() []string {
	values := make([]string, len(EventStatuses))
	for i, es := range EventStatuses {
		values[i] = es.Slug
	}
	return values
}

// GetEventStatusInfo returns EventStatusInfo by slug
func GetEventStatusInfo(slug string) (EventStatusInfo, bool) {
	info, ok := EventStatusMap[slug]
	return info, ok
}

// IsValidEventStatus checks if an event status is valid
func IsValidEventStatus(slug string) bool {
	_, ok := EventStatusMap[slug]
	return ok
}

// GetEditableStatuses returns statuses where events can be edited
func GetEditableStatuses() []EventStatusInfo {
	var result []EventStatusInfo
	for _, es := range EventStatuses {
		if es.CanEdit {
			result = append(result, es)
		}
	}
	return result
}

// GetRegisterableStatuses returns statuses where attendees can register
func GetRegisterableStatuses() []EventStatusInfo {
	var result []EventStatusInfo
	for _, es := range EventStatuses {
		if es.CanRegister {
			result = append(result, es)
		}
	}
	return result
}