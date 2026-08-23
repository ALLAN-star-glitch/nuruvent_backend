// internal/shared/types/event.go

package types

// ============================================================
// EVENT TYPE CONSTANTS
// ============================================================

// EventType represents the type of event being stored
type EventType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	EventTypeUncategorizedSlug = "event-type-uncategorized"
	EventTypeWorkshopSlug      = "event-type-workshop"
	EventTypeWebinarSlug       = "event-type-webinar"
	EventTypeMeetupSlug        = "event-type-meetup"
	EventTypeBootcampSlug      = "event-type-bootcamp"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	EventTypeUncategorizedName = "event_type_uncategorized"
	EventTypeWorkshopName      = "event_type_workshop"
	EventTypeWebinarName       = "event_type_webinar"
	EventTypeMeetupName        = "event_type_meetup"
	EventTypeBootcampName      = "event_type_bootcamp"
)

// Display name constants - Used for UI display
const (
	EventTypeUncategorizedDisplayName = "Uncategorized"
	EventTypeWorkshopDisplayName      = "Workshop"
	EventTypeWebinarDisplayName       = "Webinar"
	EventTypeMeetupDisplayName        = "Meetup"
	EventTypeBootcampDisplayName      = "Bootcamp"
)

// Description constants
const (
	EventTypeUncategorizedDescription = "Default event type for events without a specific category"
	EventTypeWorkshopDescription      = "Interactive, hands-on skill-building session"
	EventTypeWebinarDescription       = "Broadcast-style knowledge sharing"
	EventTypeMeetupDescription        = "Casual professional networking event"
	EventTypeBootcampDescription      = "Intensive multi-session training program"
)

// Color constants
const (
	EventTypeUncategorizedColor = "#6B7280" // Gray
	EventTypeWorkshopColor      = "#8B5CF6" // Purple
	EventTypeWebinarColor       = "#3B82F6" // Blue
	EventTypeMeetupColor        = "#F59E0B" // Amber
	EventTypeBootcampColor      = "#EF4444" // Red
)

// Icon constants
const (
	EventTypeUncategorizedIcon = "file-text"
	EventTypeWorkshopIcon      = "tools"
	EventTypeWebinarIcon       = "presentation"
	EventTypeMeetupIcon        = "users"
	EventTypeBootcampIcon      = "rocket"
)

// Sort order constants
const (
	EventTypeUncategorizedSortOrder = 0
	EventTypeWorkshopSortOrder      = 1
	EventTypeWebinarSortOrder       = 2
	EventTypeMeetupSortOrder        = 3
	EventTypeBootcampSortOrder      = 4
)

// ============================================================
// EVENT TYPE DEFINITIONS
// ============================================================

// EventType constants - These are the NAMES (with underscores)
const (
	EventTypeUncategorized EventType = EventTypeUncategorizedName // "event_type_uncategorized"
	EventTypeWorkshop      EventType = EventTypeWorkshopName      // "event_type_workshop"
	EventTypeWebinar       EventType = EventTypeWebinarName       // "event_type_webinar"
	EventTypeMeetup        EventType = EventTypeMeetupName        // "event_type_meetup"
	EventTypeBootcamp      EventType = EventTypeBootcampName      // "event_type_bootcamp"
)

// AllEventTypes lists all valid event types for validation
var AllEventTypes = []EventType{
	EventTypeUncategorized,
	EventTypeWorkshop,
	EventTypeWebinar,
	EventTypeMeetup,
	EventTypeBootcamp,
}

// ============================================================
// BASIC METHODS - On EventType
// ============================================================

// String returns the string representation (the name with underscores)
func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the event type is valid
func (e EventType) IsValid() bool {
	for _, t := range AllEventTypes {
		if t == e {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS
// ============================================================

// GetName returns the internal name (with underscores)
func (e EventType) GetName() string {
	return string(e)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (e EventType) GetSlug() string {
	switch e {
	case EventTypeUncategorized:
		return EventTypeUncategorizedSlug
	case EventTypeWorkshop:
		return EventTypeWorkshopSlug
	case EventTypeWebinar:
		return EventTypeWebinarSlug
	case EventTypeMeetup:
		return EventTypeMeetupSlug
	case EventTypeBootcamp:
		return EventTypeBootcampSlug
	default:
		return string(e)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (e EventType) GetDisplayName() string {
	switch e {
	case EventTypeUncategorized:
		return EventTypeUncategorizedDisplayName
	case EventTypeWorkshop:
		return EventTypeWorkshopDisplayName
	case EventTypeWebinar:
		return EventTypeWebinarDisplayName
	case EventTypeMeetup:
		return EventTypeMeetupDisplayName
	case EventTypeBootcamp:
		return EventTypeBootcampDisplayName
	default:
		return string(e)
	}
}

// GetDescription returns the description
func (e EventType) GetDescription() string {
	switch e {
	case EventTypeUncategorized:
		return EventTypeUncategorizedDescription
	case EventTypeWorkshop:
		return EventTypeWorkshopDescription
	case EventTypeWebinar:
		return EventTypeWebinarDescription
	case EventTypeMeetup:
		return EventTypeMeetupDescription
	case EventTypeBootcamp:
		return EventTypeBootcampDescription
	default:
		return ""
	}
}

// GetColor returns the color for this event type
func (e EventType) GetColor() string {
	switch e {
	case EventTypeUncategorized:
		return EventTypeUncategorizedColor
	case EventTypeWorkshop:
		return EventTypeWorkshopColor
	case EventTypeWebinar:
		return EventTypeWebinarColor
	case EventTypeMeetup:
		return EventTypeMeetupColor
	case EventTypeBootcamp:
		return EventTypeBootcampColor
	default:
		return "#6B7280"
	}
}

// GetIcon returns the icon name for this event type
func (e EventType) GetIcon() string {
	switch e {
	case EventTypeUncategorized:
		return EventTypeUncategorizedIcon
	case EventTypeWorkshop:
		return EventTypeWorkshopIcon
	case EventTypeWebinar:
		return EventTypeWebinarIcon
	case EventTypeMeetup:
		return EventTypeMeetupIcon
	case EventTypeBootcamp:
		return EventTypeBootcampIcon
	default:
		return "file-text"
	}
}

// GetSortOrder returns the sort order for this event type
func (e EventType) GetSortOrder() int {
	switch e {
	case EventTypeUncategorized:
		return EventTypeUncategorizedSortOrder
	case EventTypeWorkshop:
		return EventTypeWorkshopSortOrder
	case EventTypeWebinar:
		return EventTypeWebinarSortOrder
	case EventTypeMeetup:
		return EventTypeMeetupSortOrder
	case EventTypeBootcamp:
		return EventTypeBootcampSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS
// ============================================================

// ParseEventType parses a string into an EventType
// Expects the name (with underscores), not the slug
func ParseEventType(name string) (EventType, bool) {
	t := EventType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// ParseEventTypeWithDefault parses a string or returns a default
func ParseEventTypeWithDefault(name string, defaultType EventType) EventType {
	if t, ok := ParseEventType(name); ok {
		return t
	}
	return defaultType
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllEventTypeNames returns all internal type names (with underscores)
func AllEventTypeNames() []string {
	names := make([]string, 0, len(AllEventTypes))
	for _, t := range AllEventTypes {
		names = append(names, t.GetName())
	}
	return names
}

// AllEventTypeSlugs returns all type slugs (with hyphens)
func AllEventTypeSlugs() []string {
	slugs := make([]string, 0, len(AllEventTypes))
	for _, t := range AllEventTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

// AllEventTypeDisplayNames returns all display names
func AllEventTypeDisplayNames() []string {
	names := make([]string, 0, len(AllEventTypes))
	for _, t := range AllEventTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}

// ============================================================
// EVENT STATUS CONSTANTS
// ============================================================

// EventStatus represents the status of an event
type EventStatus string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	EventStatusDraftSlug     = "event-status-draft"
	EventStatusPublishedSlug = "event-status-published"
	EventStatusCancelledSlug = "event-status-cancelled"
	EventStatusCompletedSlug = "event-status-completed"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	EventStatusDraftName     = "event_status_draft"
	EventStatusPublishedName = "event_status_published"
	EventStatusCancelledName = "event_status_cancelled"
	EventStatusCompletedName = "event_status_completed"
)

// Display name constants - Used for UI display
const (
	EventStatusDraftDisplayName     = "Draft"
	EventStatusPublishedDisplayName = "Published"
	EventStatusCancelledDisplayName = "Cancelled"
	EventStatusCompletedDisplayName = "Completed"
)

// Description constants
const (
	EventStatusDraftDescription     = "Event is being prepared, not yet published"
	EventStatusPublishedDescription = "Event is published and open for registration"
	EventStatusCancelledDescription = "Event has been cancelled"
	EventStatusCompletedDescription = "Event has been completed"
)

// Color constants
const (
	EventStatusDraftColor     = "#6B7280" // Gray
	EventStatusPublishedColor = "#10B981" // Green
	EventStatusCancelledColor = "#EF4444" // Red
	EventStatusCompletedColor = "#3B82F6" // Blue
)

// Icon constants
const (
	EventStatusDraftIcon     = "file-edit"
	EventStatusPublishedIcon = "globe"
	EventStatusCancelledIcon = "x-circle"
	EventStatusCompletedIcon = "check-circle"
)

// Sort order constants
const (
	EventStatusDraftSortOrder     = 1
	EventStatusPublishedSortOrder = 2
	EventStatusCancelledSortOrder = 3
	EventStatusCompletedSortOrder = 4
)

// ============================================================
// EVENT STATUS DEFINITIONS
// ============================================================

// EventStatus constants - These are the NAMES (with underscores)
const (
	EventStatusDraft     EventStatus = EventStatusDraftName     // "event_status_draft"
	EventStatusPublished EventStatus = EventStatusPublishedName // "event_status_published"
	EventStatusCancelled EventStatus = EventStatusCancelledName // "event_status_cancelled"
	EventStatusCompleted EventStatus = EventStatusCompletedName // "event_status_completed"
)

// AllEventStatuses lists all valid event statuses for validation
var AllEventStatuses = []EventStatus{
	EventStatusDraft,
	EventStatusPublished,
	EventStatusCancelled,
	EventStatusCompleted,
}

// ============================================================
// BASIC METHODS - On EventStatus
// ============================================================

// String returns the string representation (the name with underscores)
func (e EventStatus) String() string {
	return string(e)
}

// IsValid checks if the event status is valid
func (e EventStatus) IsValid() bool {
	for _, s := range AllEventStatuses {
		if s == e {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On EventStatus
// ============================================================

// GetName returns the internal name (with underscores)
func (e EventStatus) GetName() string {
	return string(e)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (e EventStatus) GetSlug() string {
	switch e {
	case EventStatusDraft:
		return EventStatusDraftSlug
	case EventStatusPublished:
		return EventStatusPublishedSlug
	case EventStatusCancelled:
		return EventStatusCancelledSlug
	case EventStatusCompleted:
		return EventStatusCompletedSlug
	default:
		return string(e)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (e EventStatus) GetDisplayName() string {
	switch e {
	case EventStatusDraft:
		return EventStatusDraftDisplayName
	case EventStatusPublished:
		return EventStatusPublishedDisplayName
	case EventStatusCancelled:
		return EventStatusCancelledDisplayName
	case EventStatusCompleted:
		return EventStatusCompletedDisplayName
	default:
		return string(e)
	}
}

// GetDescription returns the description
func (e EventStatus) GetDescription() string {
	switch e {
	case EventStatusDraft:
		return EventStatusDraftDescription
	case EventStatusPublished:
		return EventStatusPublishedDescription
	case EventStatusCancelled:
		return EventStatusCancelledDescription
	case EventStatusCompleted:
		return EventStatusCompletedDescription
	default:
		return ""
	}
}

// GetColor returns the color for this event status
func (e EventStatus) GetColor() string {
	switch e {
	case EventStatusDraft:
		return EventStatusDraftColor
	case EventStatusPublished:
		return EventStatusPublishedColor
	case EventStatusCancelled:
		return EventStatusCancelledColor
	case EventStatusCompleted:
		return EventStatusCompletedColor
	default:
		return "#6B7280"
	}
}

// GetIcon returns the icon name for this event status
func (e EventStatus) GetIcon() string {
	switch e {
	case EventStatusDraft:
		return EventStatusDraftIcon
	case EventStatusPublished:
		return EventStatusPublishedIcon
	case EventStatusCancelled:
		return EventStatusCancelledIcon
	case EventStatusCompleted:
		return EventStatusCompletedIcon
	default:
		return "file-text"
	}
}

// GetSortOrder returns the sort order for this event status
func (e EventStatus) GetSortOrder() int {
	switch e {
	case EventStatusDraft:
		return EventStatusDraftSortOrder
	case EventStatusPublished:
		return EventStatusPublishedSortOrder
	case EventStatusCancelled:
		return EventStatusCancelledSortOrder
	case EventStatusCompleted:
		return EventStatusCompletedSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For EventStatus
// ============================================================

// ParseEventStatus parses a string into an EventStatus
// Expects the name (with underscores), not the slug
func ParseEventStatus(name string) (EventStatus, bool) {
	s := EventStatus(name)
	if s.IsValid() {
		return s, true
	}
	return "", false
}

// ParseEventStatusWithDefault parses a string or returns a default
func ParseEventStatusWithDefault(name string, defaultStatus EventStatus) EventStatus {
	if s, ok := ParseEventStatus(name); ok {
		return s
	}
	return defaultStatus
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllEventStatusNames returns all internal status names (with underscores)
func AllEventStatusNames() []string {
	names := make([]string, 0, len(AllEventStatuses))
	for _, s := range AllEventStatuses {
		names = append(names, s.GetName())
	}
	return names
}

// AllEventStatusSlugs returns all status slugs (with hyphens)
func AllEventStatusSlugs() []string {
	slugs := make([]string, 0, len(AllEventStatuses))
	for _, s := range AllEventStatuses {
		slugs = append(slugs, s.GetSlug())
	}
	return slugs
}

// AllEventStatusDisplayNames returns all display names
func AllEventStatusDisplayNames() []string {
	names := make([]string, 0, len(AllEventStatuses))
	for _, s := range AllEventStatuses {
		names = append(names, s.GetDisplayName())
	}
	return names
}


// ParseEventTypeBySlug parses a slug string into an EventType
func ParseEventTypeBySlug(slug string) (EventType, bool) {
	for _, t := range AllEventTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ParseEventStatusBySlug parses a slug string into an EventStatus
func ParseEventStatusBySlug(slug string) (EventStatus, bool) {
	for _, s := range AllEventStatuses {
		if s.GetSlug() == slug {
			return s, true
		}
	}
	return "", false
}