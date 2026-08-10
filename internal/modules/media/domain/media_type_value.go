package domain

// ============================================================
// MEDIA TYPE - Value Object (Source of Truth)
// ============================================================

// MediaTypeValue is a custom typed string for compile-time safety
type MediaTypeValue string

const (
	MediaTypeEvent       MediaTypeValue = "event"
	MediaTypeBusiness    MediaTypeValue = "business"
	MediaTypeProfile     MediaTypeValue = "profile"
	MediaTypeCertificate MediaTypeValue = "certificate"
	MediaTypeRecording   MediaTypeValue = "recording"
)

// AllMediaTypes lists all valid media types
var AllMediaTypes = []MediaTypeValue{
	MediaTypeEvent,
	MediaTypeBusiness,
	MediaTypeProfile,
	MediaTypeCertificate,
	MediaTypeRecording,
}

// MediaTypeInfo holds metadata for each media type
type MediaTypeInfo struct {
	Slug        MediaTypeValue
	Name        string
	DisplayName string
	Description string
	Bucket      string
	Icon        string
	SortOrder   int
	MaxFileSize int64
	IsActive    bool
}

// Private registry (prevents external mutation)
var mediaTypeRegistry = map[MediaTypeValue]MediaTypeInfo{
	MediaTypeEvent: {
		Slug:        MediaTypeEvent,
		Name:        "Event",
		DisplayName: "Event",
		Description: "Event images and banners",
		Bucket:      "events",
		Icon:        "image",
		SortOrder:   1,
		MaxFileSize: 5 * 1024 * 1024, // 5MB
		IsActive:    true,
	},
	MediaTypeBusiness: {
		Slug:        MediaTypeBusiness,
		Name:        "Business",
		DisplayName: "Business",
		Description: "Business logos and images",
		Bucket:      "businesses",
		Icon:        "building",
		SortOrder:   2,
		MaxFileSize: 5 * 1024 * 1024, // 5MB
		IsActive:    true,
	},
	MediaTypeProfile: {
		Slug:        MediaTypeProfile,
		Name:        "Profile",
		DisplayName: "Profile",
		Description: "User profile pictures",
		Bucket:      "profiles",
		Icon:        "user",
		SortOrder:   3,
		MaxFileSize: 2 * 1024 * 1024, // 2MB
		IsActive:    true,
	},
	MediaTypeCertificate: {
		Slug:        MediaTypeCertificate,
		Name:        "Certificate",
		DisplayName: "Certificate",
		Description: "Certificate templates and files",
		Bucket:      "certificates",
		Icon:        "certificate",
		SortOrder:   4,
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		IsActive:    true,
	},
	MediaTypeRecording: {
		Slug:        MediaTypeRecording,
		Name:        "Recording",
		DisplayName: "Recording",
		Description: "Event recordings and videos",
		Bucket:      "recordings",
		Icon:        "video",
		SortOrder:   5,
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		IsActive:    true,
	},
}

// ============================================================
// DOMAIN METHODS (on MediaTypeValue)
// ============================================================

func (m MediaTypeValue) String() string {
	return string(m)
}

func (m MediaTypeValue) IsValid() bool {
	_, ok := mediaTypeRegistry[m]
	return ok
}

func (m MediaTypeValue) IsActive() bool {
	info, ok := mediaTypeRegistry[m]
	if !ok {
		return false
	}
	return info.IsActive
}

func (m MediaTypeValue) Info() (MediaTypeInfo, bool) {
	info, ok := mediaTypeRegistry[m]
	return info, ok
}

func (m MediaTypeValue) MaxFileSize() int64 {
	info, ok := mediaTypeRegistry[m]
	if !ok {
		return 0
	}
	return info.MaxFileSize
}

func (m MediaTypeValue) Bucket() string {
	info, ok := mediaTypeRegistry[m]
	if !ok {
		return ""
	}
	return info.Bucket
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

func ParseMediaType(slug string) (MediaTypeValue, bool) {
	t := MediaTypeValue(slug)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func AllMediaTypeInfos() []MediaTypeInfo {
	infos := make([]MediaTypeInfo, 0, len(mediaTypeRegistry))
	for _, info := range mediaTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

func AllMediaTypeSlugs() []string {
	slugs := make([]string, 0, len(mediaTypeRegistry))
	for slug := range mediaTypeRegistry {
		slugs = append(slugs, string(slug))
	}
	return slugs
}

func ActiveMediaTypeInfos() []MediaTypeInfo {
	infos := make([]MediaTypeInfo, 0)
	for _, info := range mediaTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func ActiveMediaTypeSlugs() []string {
	slugs := make([]string, 0)
	for slug, info := range mediaTypeRegistry {
		if info.IsActive {
			slugs = append(slugs, string(slug))
		}
	}
	return slugs
}