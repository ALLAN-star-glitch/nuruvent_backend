package domain

// ============================================================
// EVENT STATUS - Value Object (Source of Truth)
// ============================================================

// EventStatusValue is a custom typed string for compile-time safety
type EventStatusValue string

const (
	EventStatusDraft     EventStatusValue = "draft"
	EventStatusPublished EventStatusValue = "published"
	EventStatusCancelled EventStatusValue = "cancelled"
	EventStatusCompleted EventStatusValue = "completed"
)

// AllEventStatuses lists all valid event statuses
var AllEventStatuses = []EventStatusValue{
	EventStatusDraft,
	EventStatusPublished,
	EventStatusCancelled,
	EventStatusCompleted,
}

// EventStatusInfo holds metadata for each event status
type EventStatusInfo struct {
	Slug        EventStatusValue
	Name        string
	DisplayName string
	Description string
	Color       string
	Icon        string
	SortOrder   int
	IsFinal     bool
	IsActive    bool
}

// Private registry (prevents external mutation)
var eventStatusRegistry = map[EventStatusValue]EventStatusInfo{
	EventStatusDraft: {
		Slug:        EventStatusDraft,
		Name:        "Draft",
		DisplayName: "Draft",
		Description: "Event is being prepared, not yet published",
		Color:       "#6B7280",
		Icon:        "file-edit",
		SortOrder:   1,
		IsFinal:     false,
		IsActive:    true,
	},
	EventStatusPublished: {
		Slug:        EventStatusPublished,
		Name:        "Published",
		DisplayName: "Published",
		Description: "Event is published and open for registration",
		Color:       "#10B981",
		Icon:        "globe",
		SortOrder:   2,
		IsFinal:     false,
		IsActive:    true,
	},
	EventStatusCancelled: {
		Slug:        EventStatusCancelled,
		Name:        "Cancelled",
		DisplayName: "Cancelled",
		Description: "Event has been cancelled",
		Color:       "#EF4444",
		Icon:        "x-circle",
		SortOrder:   3,
		IsFinal:     true,
		IsActive:    true,
	},
	EventStatusCompleted: {
		Slug:        EventStatusCompleted,
		Name:        "Completed",
		DisplayName: "Completed",
		Description: "Event has been completed",
		Color:       "#3B82F6",
		Icon:        "check-circle",
		SortOrder:   4,
		IsFinal:     true,
		IsActive:    true,
	},
}

// ============================================================
// DOMAIN METHODS (on EventStatusValue)
// ============================================================

func (e EventStatusValue) String() string {
	return string(e)
}

func (e EventStatusValue) IsValid() bool {
	_, ok := eventStatusRegistry[e]
	return ok
}

func (e EventStatusValue) IsActive() bool {
	info, ok := eventStatusRegistry[e]
	if !ok {
		return false
	}
	return info.IsActive
}

func (e EventStatusValue) IsFinal() bool {
	info, ok := eventStatusRegistry[e]
	if !ok {
		return false
	}
	return info.IsFinal
}

func (e EventStatusValue) Info() (EventStatusInfo, bool) {
	info, ok := eventStatusRegistry[e]
	return info, ok
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

func ParseEventStatus(slug string) (EventStatusValue, bool) {
	t := EventStatusValue(slug)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func AllEventStatusInfos() []EventStatusInfo {
	infos := make([]EventStatusInfo, 0, len(eventStatusRegistry))
	for _, info := range eventStatusRegistry {
		infos = append(infos, info)
	}
	return infos
}

func AllEventStatusSlugs() []string {
	slugs := make([]string, 0, len(eventStatusRegistry))
	for slug := range eventStatusRegistry {
		slugs = append(slugs, string(slug))
	}
	return slugs
}

func ActiveEventStatusInfos() []EventStatusInfo {
	infos := make([]EventStatusInfo, 0)
	for _, info := range eventStatusRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func ActiveEventStatusSlugs() []string {
	slugs := make([]string, 0)
	for slug, info := range eventStatusRegistry {
		if info.IsActive {
			slugs = append(slugs, string(slug))
		}
	}
	return slugs
}