// internal/constants/event_types.go

package constants

// EventTypeInfo holds information about an event type
type EventTypeInfo struct {
	Slug                string // Database value (e.g., "workshop")
	Name                string // Canonical name (e.g., "Workshop")
	DisplayName         string // UI display (e.g., "🛠️ Workshop")
	Description         string
	Icon                string
	Color               string
	SortOrder           int
	SupportsCertificate bool
	MinDuration         int
	MaxDuration         int
	MetaTitle           string
	MetaDescription     string
}

// EventTypes - the source of truth for event types
var EventTypes = []EventTypeInfo{
	{
		Slug:                "workshop",
		Name:                "Workshop",
		DisplayName:         "🛠️ Workshop",
		Description:         "Interactive, hands-on skill-building sessions. Participants actively engage and practice skills.",
		Icon:                "graduation-cap",
		Color:               "#4F46E5", // Indigo
		SortOrder:           1,
		SupportsCertificate: true,
		MinDuration:         120, // 2 hours
		MaxDuration:         480, // 8 hours
		MetaTitle:           "Professional Workshops in Kenya | Nuruvent",
		MetaDescription:     "Join interactive workshops in Kenya. Hands-on skill-building sessions with professional trainers. Book now on Nuruvent.",
	},
	{
		Slug:                "webinar",
		Name:                "Webinar",
		DisplayName:         "🎥 Webinar",
		Description:         "Broadcast-style knowledge sharing. One or more speakers present to a larger audience.",
		Icon:                "video",
		Color:               "#7C3AED", // Purple
		SortOrder:           2,
		SupportsCertificate: true,
		MinDuration:         60, // 1 hour
		MaxDuration:         240, // 4 hours
		MetaTitle:           "Professional Webinars in Kenya | Nuruvent",
		MetaDescription:     "Attend professional webinars in Kenya. Learn from industry experts from anywhere. Register now on Nuruvent.",
	},
	{
		Slug:                "meetup",
		Name:                "Meetup",
		DisplayName:         "🤝 Meetup",
		Description:         "Casual community networking events. Professionals connect and share ideas.",
		Icon:                "users",
		Color:               "#059669", // Emerald
		SortOrder:           3,
		SupportsCertificate: false,
		MinDuration:         60, // 1 hour
		MaxDuration:         180, // 3 hours
		MetaTitle:           "Professional Meetups in Kenya | Nuruvent",
		MetaDescription:     "Connect with professionals at meetups in Kenya. Network, share ideas, and grow your career. Join now on Nuruvent.",
	},
	{
		Slug:                "bootcamp",
		Name:                "Bootcamp",
		DisplayName:         "💻 Bootcamp",
		Description:         "Intensive multi-session training programs. Deep skill development over days or weeks.",
		Icon:                "laptop",
		Color:               "#DC2626", // Red
		SortOrder:           4,
		SupportsCertificate: true,
		MinDuration:         480, // 8 hours
		MaxDuration:         2880, // 48 hours (multiple days)
		MetaTitle:           "Professional Bootcamps in Kenya | Nuruvent",
		MetaDescription:     "Join intensive bootcamps in Kenya. Deep skill development with expert trainers. Enroll now on Nuruvent.",
	},
}

// EventTypeMap for quick lookups
var EventTypeMap = map[string]EventTypeInfo{
	"workshop": EventTypes[0],
	"webinar":  EventTypes[1],
	"meetup":   EventTypes[2],
	"bootcamp": EventTypes[3],
}

// AllEventTypeSlugs returns all valid event type slugs
func AllEventTypeSlugs() []string {
	values := make([]string, len(EventTypes))
	for i, et := range EventTypes {
		values[i] = et.Slug
	}
	return values
}

// GetEventTypeInfo returns EventTypeInfo by slug
func GetEventTypeInfo(slug string) (EventTypeInfo, bool) {
	info, ok := EventTypeMap[slug]
	return info, ok
}

// IsValidEventType checks if an event type is valid
func IsValidEventType(slug string) bool {
	_, ok := EventTypeMap[slug]
	return ok
}

// GetEventTypesByCertificateSupport returns event types that support certificates
func GetEventTypesByCertificateSupport(supportsCertificate bool) []EventTypeInfo {
	var result []EventTypeInfo
	for _, et := range EventTypes {
		if et.SupportsCertificate == supportsCertificate {
			result = append(result, et)
		}
	}
	return result
}