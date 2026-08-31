// internal/shared/types/event_format.go

package types

// ============================================================
// EVENT FORMAT CONSTANTS
// ============================================================

// EventFormat represents the format of an event
type EventFormat string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	EventFormatVirtualSlug  = "virtual"
	EventFormatInPersonSlug = "in-person"
	EventFormatHybridSlug   = "hybrid"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	EventFormatVirtualName  = "event_format_virtual"
	EventFormatInPersonName = "event_format_in_person"
	EventFormatHybridName   = "event_format_hybrid"
)

// Display name constants - Used for UI display
const (
	EventFormatVirtualDisplayName  = "Virtual"
	EventFormatInPersonDisplayName = "In-Person"
	EventFormatHybridDisplayName   = "Hybrid"
)

// Description constants
const (
	EventFormatVirtualDescription  = "Fully online event via video conferencing"
	EventFormatInPersonDescription = "In-person event at a physical venue"
	EventFormatHybridDescription   = "Both in-person and virtual attendance"
)

// Icon constants
const (
	EventFormatVirtualIcon  = "globe"
	EventFormatInPersonIcon = "map-pin"
	EventFormatHybridIcon   = "globe-2"
)

// ============================================================
// EVENT FORMAT DEFINITIONS
// ============================================================

// EventFormat constants - These are the NAMES (with underscores)
const (
	EventFormatVirtual  EventFormat = EventFormatVirtualName  // "event_format_virtual"
	EventFormatInPerson EventFormat = EventFormatInPersonName // "event_format_in_person"
	EventFormatHybrid   EventFormat = EventFormatHybridName   // "event_format_hybrid"
)

// AllEventFormats lists all valid event formats for validation
var AllEventFormats = []EventFormat{
	EventFormatVirtual,
	EventFormatInPerson,
	EventFormatHybrid,
}

// ============================================================
// BASIC METHODS - On EventFormat
// ============================================================

// String returns the string representation (the name with underscores)
func (f EventFormat) String() string {
	return string(f)
}

// IsValid checks if the event format is valid
func (f EventFormat) IsValid() bool {
	for _, format := range AllEventFormats {
		if format == f {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On EventFormat
// ============================================================

// GetName returns the internal name (with underscores)
func (f EventFormat) GetName() string {
	return string(f)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (f EventFormat) GetSlug() string {
	switch f {
	case EventFormatVirtual:
		return EventFormatVirtualSlug
	case EventFormatInPerson:
		return EventFormatInPersonSlug
	case EventFormatHybrid:
		return EventFormatHybridSlug
	default:
		return string(f)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (f EventFormat) GetDisplayName() string {
	switch f {
	case EventFormatVirtual:
		return EventFormatVirtualDisplayName
	case EventFormatInPerson:
		return EventFormatInPersonDisplayName
	case EventFormatHybrid:
		return EventFormatHybridDisplayName
	default:
		return string(f)
	}
}

// GetDescription returns the description
func (f EventFormat) GetDescription() string {
	switch f {
	case EventFormatVirtual:
		return EventFormatVirtualDescription
	case EventFormatInPerson:
		return EventFormatInPersonDescription
	case EventFormatHybrid:
		return EventFormatHybridDescription
	default:
		return ""
	}
}

// GetIcon returns the icon name for this event format
func (f EventFormat) GetIcon() string {
	switch f {
	case EventFormatVirtual:
		return EventFormatVirtualIcon
	case EventFormatInPerson:
		return EventFormatInPersonIcon
	case EventFormatHybrid:
		return EventFormatHybridIcon
	default:
		return "globe"
	}
}

// ============================================================
// PARSE FUNCTIONS - For EventFormat
// ============================================================

// ParseEventFormat parses a string into an EventFormat
// Expects the name (with underscores), not the slug
func ParseEventFormat(name string) (EventFormat, bool) {
	f := EventFormat(name)
	if f.IsValid() {
		return f, true
	}
	return "", false
}

// ParseEventFormatWithDefault parses a string or returns a default
func ParseEventFormatWithDefault(name string, defaultFormat EventFormat) EventFormat {
	if f, ok := ParseEventFormat(name); ok {
		return f
	}
	return defaultFormat
}

// ParseEventFormatBySlug parses a slug string into an EventFormat
func ParseEventFormatBySlug(slug string) (EventFormat, bool) {
	for _, f := range AllEventFormats {
		if f.GetSlug() == slug {
			return f, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllEventFormatNames returns all internal format names (with underscores)
func AllEventFormatNames() []string {
	names := make([]string, 0, len(AllEventFormats))
	for _, f := range AllEventFormats {
		names = append(names, f.GetName())
	}
	return names
}

// AllEventFormatSlugs returns all format slugs (with hyphens)
func AllEventFormatSlugs() []string {
	slugs := make([]string, 0, len(AllEventFormats))
	for _, f := range AllEventFormats {
		slugs = append(slugs, f.GetSlug())
	}
	return slugs
}

// AllEventFormatDisplayNames returns all format display names
func AllEventFormatDisplayNames() []string {
	names := make([]string, 0, len(AllEventFormats))
	for _, f := range AllEventFormats {
		names = append(names, f.GetDisplayName())
	}
	return names
}