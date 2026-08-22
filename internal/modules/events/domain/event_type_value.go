package domain

// ============================================================
// EVENT TYPE - Value Object (Source of Truth)
// ============================================================

// EventTypeValue is a custom typed string for compile-time safety
type EventTypeValue string

const (
	EventTypeUncategorized EventTypeValue = "uncategorized" // ✅ ADD THIS
	EventTypeWorkshop      EventTypeValue = "workshop"
	EventTypeWebinar       EventTypeValue = "webinar"
	EventTypeMeetup        EventTypeValue = "meetup"
	EventTypeBootcamp      EventTypeValue = "bootcamp"
)

// AllEventTypes lists all valid event types
var AllEventTypes = []EventTypeValue{
	EventTypeUncategorized, // ✅ ADD THIS
	EventTypeWorkshop,
	EventTypeWebinar,
	EventTypeMeetup,
	EventTypeBootcamp,
}

// EventTypeInfo holds metadata for each event type
type EventTypeInfo struct {
	Slug                EventTypeValue
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

// Private registry (prevents external mutation)
var eventTypeRegistry = map[EventTypeValue]EventTypeInfo{
	EventTypeUncategorized: { // ✅ ADD THIS
		Slug:                EventTypeUncategorized,
		Name:                "Uncategorized",
		DisplayName:         "Uncategorized",
		Description:         "Default event type for events without a specific category",
		Icon:                "file-text",
		Color:               "#6B7280",
		SortOrder:           0,
		SupportsCertificate: false,
		MinDuration:         0,
		MaxDuration:         1440,
		IsActive:            true,
	},
	EventTypeWorkshop: {
		Slug:                EventTypeWorkshop,
		Name:                "Workshop",
		DisplayName:         "Workshop",
		Description:         "Interactive, hands-on skill-building session",
		Icon:                "tools",
		Color:               "#8B5CF6",
		SortOrder:           1,
		SupportsCertificate: true,
		MinDuration:         120,
		MaxDuration:         480,
		IsActive:            true,
	},
	EventTypeWebinar: {
		Slug:                EventTypeWebinar,
		Name:                "Webinar",
		DisplayName:         "Webinar",
		Description:         "Broadcast-style knowledge sharing",
		Icon:                "presentation",
		Color:               "#3B82F6",
		SortOrder:           2,
		SupportsCertificate: false,
		MinDuration:         30,
		MaxDuration:         120,
		IsActive:            true,
	},
	EventTypeMeetup: {
		Slug:                EventTypeMeetup,
		Name:                "Meetup",
		DisplayName:         "Meetup",
		Description:         "Casual professional networking event",
		Icon:                "users",
		Color:               "#F59E0B",
		SortOrder:           3,
		SupportsCertificate: false,
		MinDuration:         60,
		MaxDuration:         180,
		IsActive:            true,
	},
	EventTypeBootcamp: {
		Slug:                EventTypeBootcamp,
		Name:                "Bootcamp",
		DisplayName:         "Bootcamp",
		Description:         "Intensive multi-session training program",
		Icon:                "rocket",
		Color:               "#EF4444",
		SortOrder:           4,
		SupportsCertificate: true,
		MinDuration:         240,
		MaxDuration:         1440,
		IsActive:            true,
	},
}

// ============================================================
// DOMAIN METHODS (on EventTypeValue)
// ============================================================

func (e EventTypeValue) String() string {
	return string(e)
}

func (e EventTypeValue) IsValid() bool {
	_, ok := eventTypeRegistry[e]
	return ok
}

func (e EventTypeValue) IsActive() bool {
	info, ok := eventTypeRegistry[e]
	if !ok {
		return false
	}
	return info.IsActive
}

func (e EventTypeValue) Info() (EventTypeInfo, bool) {
	info, ok := eventTypeRegistry[e]
	return info, ok
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

func ParseEventType(slug string) (EventTypeValue, bool) {
	t := EventTypeValue(slug)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func AllEventTypeInfos() []EventTypeInfo {
	infos := make([]EventTypeInfo, 0, len(eventTypeRegistry))
	for _, info := range eventTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

func AllEventTypeSlugs() []string {
	slugs := make([]string, 0, len(eventTypeRegistry))
	for slug := range eventTypeRegistry {
		slugs = append(slugs, string(slug))
	}
	return slugs
}

func ActiveEventTypeInfos() []EventTypeInfo {
	infos := make([]EventTypeInfo, 0)
	for _, info := range eventTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func ActiveEventTypeSlugs() []string {
	slugs := make([]string, 0)
	for slug, info := range eventTypeRegistry {
		if info.IsActive {
			slugs = append(slugs, string(slug))
		}
	}
	return slugs
}