// internal/modules/media/domain/media_type_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// MEDIA TYPE - Value Object
// ============================================================

// MediaTypeValue is an alias for the shared MediaType
type MediaTypeValue = types.MediaType

// MediaType constants re-exported from shared types for convenience
const (
	MediaTypeEvent       = types.MediaTypeEvent
	MediaTypeBusiness    = types.MediaTypeBusiness
	MediaTypeProfile     = types.MediaTypeProfile
	MediaTypeCertificate = types.MediaTypeCertificate
	MediaTypeRecording   = types.MediaTypeRecording
)

// AllMediaTypes re-exported from shared types
var AllMediaTypes = types.AllMediaTypes

// MediaTypeInfo holds metadata for each media type
type MediaTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Bucket      string
	Icon        string
	Color       string
	SortOrder   int
	MaxFileSize int64
	IsActive    bool
}

// ============================================================
// MEDIA TYPE REGISTRY - EXPORTED
// ============================================================

// MediaTypeRegistry is the single source of truth for media type metadata
// ✅ Make sure this is exported (capital M)
var MediaTypeRegistry = map[types.MediaType]MediaTypeInfo{
	types.MediaTypeEvent: {
		Slug:        types.MediaTypeEventSlug,
		Name:        types.MediaTypeEventName,
		DisplayName: types.MediaTypeEventDisplayName,
		Description: "Event images and banners",
		Bucket:      "events",
		Icon:        "image",
		Color:       "#3B82F6",
		SortOrder:   1,
		MaxFileSize: 5 * 1024 * 1024,
		IsActive:    true,
	},
	types.MediaTypeBusiness: {
		Slug:        types.MediaTypeBusinessSlug,
		Name:        types.MediaTypeBusinessName,
		DisplayName: types.MediaTypeBusinessDisplayName,
		Description: "Business logos and images",
		Bucket:      "businesses",
		Icon:        "building",
		Color:       "#7C3AED",
		SortOrder:   2,
		MaxFileSize: 5 * 1024 * 1024,
		IsActive:    true,
	},
	types.MediaTypeProfile: {
		Slug:        types.MediaTypeProfileSlug,
		Name:        types.MediaTypeProfileName,
		DisplayName: types.MediaTypeProfileDisplayName,
		Description: "User profile pictures",
		Bucket:      "profiles",
		Icon:        "user",
		Color:       "#10B981",
		SortOrder:   3,
		MaxFileSize: 2 * 1024 * 1024,
		IsActive:    true,
	},
	types.MediaTypeCertificate: {
		Slug:        types.MediaTypeCertificateSlug,
		Name:        types.MediaTypeCertificateName,
		DisplayName: types.MediaTypeCertificateDisplayName,
		Description: "Certificate templates and files",
		Bucket:      "certificates",
		Icon:        "certificate",
		Color:       "#F59E0B",
		SortOrder:   4,
		MaxFileSize: 10 * 1024 * 1024,
		IsActive:    true,
	},
	types.MediaTypeRecording: {
		Slug:        types.MediaTypeRecordingSlug,
		Name:        types.MediaTypeRecordingName,
		DisplayName: types.MediaTypeRecordingDisplayName,
		Description: "Event recordings and videos",
		Bucket:      "recordings",
		Icon:        "video",
		Color:       "#EF4444",
		SortOrder:   5,
		MaxFileSize: 100 * 1024 * 1024,
		IsActive:    true,
	},
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllMediaTypeInfos returns all media type infos from the registry
func AllMediaTypeInfos() []MediaTypeInfo {
	infos := make([]MediaTypeInfo, 0, len(MediaTypeRegistry))
	for _, info := range MediaTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// GetMediaTypeInfo returns media type info from the registry
func GetMediaTypeInfo(mediaType types.MediaType) (MediaTypeInfo, bool) {
	info, ok := MediaTypeRegistry[mediaType]
	return info, ok
}

// GetMediaTypeBySlug returns type info by slug
func GetMediaTypeBySlug(slug string) (MediaTypeInfo, bool) {
	// Parse the slug to get the media type
	for mediaType := range MediaTypeRegistry {
		if mediaType.GetSlug() == slug {
			return GetMediaTypeInfo(mediaType)
		}
	}
	return MediaTypeInfo{}, false
}

// GetMediaTypeByName returns type info by internal name (with underscores)
func GetMediaTypeByName(name string) (MediaTypeInfo, bool) {
	mediaType, ok := types.ParseMediaType(name)
	if !ok {
		return MediaTypeInfo{}, false
	}
	return GetMediaTypeInfo(mediaType)
}