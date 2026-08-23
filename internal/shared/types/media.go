// internal/shared/types/media.go

package types

// ============================================================
// MEDIA TYPE CONSTANTS
// ============================================================

// MediaType represents the type of media being stored
type MediaType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	MediaTypeEventSlug       = "media-type-event"
	MediaTypeBusinessSlug    = "media-type-business"
	MediaTypeProfileSlug     = "media-type-profile"
	MediaTypeCertificateSlug = "media-type-certificate"
	MediaTypeRecordingSlug   = "media-type-recording"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	MediaTypeEventName       = "media_type_event"
	MediaTypeBusinessName    = "media_type_business"
	MediaTypeProfileName     = "media_type_profile"
	MediaTypeCertificateName = "media_type_certificate"
	MediaTypeRecordingName   = "media_type_recording"
)

// Display name constants - Used for UI display
const (
	MediaTypeEventDisplayName       = "Event"
	MediaTypeBusinessDisplayName    = "Business"
	MediaTypeProfileDisplayName     = "Profile"
	MediaTypeCertificateDisplayName = "Certificate"
	MediaTypeRecordingDisplayName   = "Recording"
)

// ============================================================
// MEDIA TYPE DEFINITIONS
// ============================================================

// MediaType constants - These are the NAMES (with underscores)
const (
	MediaTypeEvent       MediaType = MediaTypeEventName       // "media_type_event"
	MediaTypeBusiness    MediaType = MediaTypeBusinessName    // "media_type_business"
	MediaTypeProfile     MediaType = MediaTypeProfileName     // "media_type_profile"
	MediaTypeCertificate MediaType = MediaTypeCertificateName // "media_type_certificate"
	MediaTypeRecording   MediaType = MediaTypeRecordingName   // "media_type_recording"
)

// AllMediaTypes lists all valid media types for validation
var AllMediaTypes = []MediaType{
	MediaTypeEvent,
	MediaTypeBusiness,
	MediaTypeProfile,
	MediaTypeCertificate,
	MediaTypeRecording,
}

// ============================================================
// BASIC METHODS - On MediaType
// ============================================================

// String returns the string representation (the name with underscores)
func (m MediaType) String() string {
	return string(m)
}

// IsValid checks if the media type is valid
func (m MediaType) IsValid() bool {
	for _, t := range AllMediaTypes {
		if t == m {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS
// ============================================================

// GetName returns the internal name (with underscores)
// Since MediaType IS the name, just return string(m)
func (m MediaType) GetName() string {
	return string(m) // ✅ "media_type_event"
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
// Need to map from name to slug
func (m MediaType) GetSlug() string {
	switch m {
	case MediaTypeEvent:
		return MediaTypeEventSlug // "media-type-event"
	case MediaTypeBusiness:
		return MediaTypeBusinessSlug // "media-type-business"
	case MediaTypeProfile:
		return MediaTypeProfileSlug // "media-type-profile"
	case MediaTypeCertificate:
		return MediaTypeCertificateSlug // "media-type-certificate"
	case MediaTypeRecording:
		return MediaTypeRecordingSlug // "media-type-recording"
	default:
		return string(m)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (m MediaType) GetDisplayName() string {
	switch m {
	case MediaTypeEvent:
		return MediaTypeEventDisplayName // "Event"
	case MediaTypeBusiness:
		return MediaTypeBusinessDisplayName // "Business"
	case MediaTypeProfile:
		return MediaTypeProfileDisplayName // "Profile"
	case MediaTypeCertificate:
		return MediaTypeCertificateDisplayName // "Certificate"
	case MediaTypeRecording:
		return MediaTypeRecordingDisplayName // "Recording"
	default:
		return string(m)
	}
}

// ============================================================
// PARSE FUNCTIONS
// ============================================================

// ParseMediaType parses a string into a MediaType
// Expects the name (with underscores), not the slug
func ParseMediaType(name string) (MediaType, bool) {
	t := MediaType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// ParseMediaTypeWithDefault parses a string or returns a default
func ParseMediaTypeWithDefault(name string, defaultType MediaType) MediaType {
	if t, ok := ParseMediaType(name); ok {
		return t
	}
	return defaultType
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllMediaTypeNames returns all internal type names (with underscores)
func AllMediaTypeNames() []string {
	names := make([]string, 0, len(AllMediaTypes))
	for _, t := range AllMediaTypes {
		names = append(names, t.GetName())
	}
	return names
}

// AllMediaTypeSlugs returns all type slugs (with hyphens)
func AllMediaTypeSlugs() []string {
	slugs := make([]string, 0, len(AllMediaTypes))
	for _, t := range AllMediaTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

// AllMediaTypeDisplayNames returns all display names
func AllMediaTypeDisplayNames() []string {
	names := make([]string, 0, len(AllMediaTypes))
	for _, t := range AllMediaTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}