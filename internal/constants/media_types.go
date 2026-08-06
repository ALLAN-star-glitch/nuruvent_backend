// internal/constants/media_types.go

package constants

// MediaTypeInfo holds information about a media type
type MediaTypeInfo struct {
	Slug              string   // Database value (e.g., "event")
	Name              string   // Canonical name (e.g., "Event Image")
	DisplayName       string   // UI display (e.g., "🖼️ Event Image")
	Description       string
	Bucket            string   // Supabase bucket name
	Icon              string
	SortOrder         int
	AllowedMimeTypes  []string
	MaxFileSize       int64
}

// MediaTypes - the source of truth for media types
var MediaTypes = []MediaTypeInfo{
	{
		Slug:        "event",
		Name:        "Event Image",
		DisplayName: "🖼️ Event Image",
		Description: "Images for events (cover photos, thumbnails)",
		Bucket:      "events",
		Icon:        "image",
		SortOrder:   1,
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		MaxFileSize: 5 * 1024 * 1024, // 5MB
	},
	{
		Slug:        "institution",
		Name:        "Institution Logo",
		DisplayName: "🏛️ Institution Logo",
		Description: "Logos and images for institutions",
		Bucket:      "institutions",
		Icon:        "building",
		SortOrder:   2,
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"image/svg+xml",
		},
		MaxFileSize: 2 * 1024 * 1024, // 2MB
	},
	{
		Slug:        "profile",
		Name:        "Profile Picture",
		DisplayName: "👤 Profile Picture",
		Description: "Profile pictures for accounts",
		Bucket:      "profiles",
		Icon:        "user",
		SortOrder:   3,
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		MaxFileSize: 2 * 1024 * 1024, // 2MB
	},
	{
		Slug:        "certificate",
		Name:        "Certificate Template",
		DisplayName: "📜 Certificate Template",
		Description: "Certificate templates uploaded by hosts",
		Bucket:      "certificates",
		Icon:        "certificate",
		SortOrder:   4,
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/webp",
			"application/pdf",
		},
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	},
	{
		Slug:        "recording",
		Name:        "Event Recording",
		DisplayName: "🎥 Event Recording",
		Description: "Recordings of events for replays",
		Bucket:      "recordings",
		Icon:        "video",
		SortOrder:   5,
		AllowedMimeTypes: []string{
			"video/mp4",
			"video/webm",
			"video/ogg",
		},
		MaxFileSize: 500 * 1024 * 1024, // 500MB
	},
}

// MediaTypeMap for quick lookups
var MediaTypeMap = map[string]MediaTypeInfo{
	"event":       MediaTypes[0],
	"institution": MediaTypes[1],
	"profile":     MediaTypes[2],
	"certificate": MediaTypes[3],
	"recording":   MediaTypes[4],
}

// AllMediaTypeSlugs returns all valid media type slugs
func AllMediaTypeSlugs() []string {
	values := make([]string, len(MediaTypes))
	for i, mt := range MediaTypes {
		values[i] = mt.Slug
	}
	return values
}

// GetMediaTypeInfo returns MediaTypeInfo by slug
func GetMediaTypeInfo(slug string) (MediaTypeInfo, bool) {
	info, ok := MediaTypeMap[slug]
	return info, ok
}

// IsValidMediaType checks if a media type is valid
func IsValidMediaType(slug string) bool {
	_, ok := MediaTypeMap[slug]
	return ok
}

// GetBucketForMediaType returns the bucket name for a media type
func GetBucketForMediaType(slug string) (string, error) {
	info, ok := MediaTypeMap[slug]
	if !ok {
		return "", nil
	}
	return info.Bucket, nil
}

// GetMaxFileSizeForMediaType returns the max file size for a media type
func GetMaxFileSizeForMediaType(slug string) (int64, error) {
	info, ok := MediaTypeMap[slug]
	if !ok {
		return 0, nil
	}
	return info.MaxFileSize, nil
}

// GetAllowedMimeTypesForMediaType returns allowed mime types for a media type
func GetAllowedMimeTypesForMediaType(slug string) ([]string, error) {
	info, ok := MediaTypeMap[slug]
	if !ok {
		return nil, nil
	}
	return info.AllowedMimeTypes, nil
}