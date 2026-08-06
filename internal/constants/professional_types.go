// internal/constants/professional_types.go

package constants

// ProfessionalTypeInfo holds information about a professional type
type ProfessionalTypeInfo struct {
	Slug        string // Database value (e.g., "trainer")
	Name        string // Canonical name (e.g., "Trainer / Coach / Consultant")
	DisplayName string // UI display (e.g., "👨‍🏫 Trainer / Coach / Consultant")
	Description string
	Icon        string
	Color       string
	SortOrder   int
	CanHost     bool // Can this professional type host events?
}

// ProfessionalTypes - the source of truth for professional types
var ProfessionalTypes = []ProfessionalTypeInfo{
	{
		Slug:        "trainer",
		Name:        "Trainer / Coach / Consultant",
		DisplayName: "👨‍🏫 Trainer / Coach / Consultant",
		Description: "Professional who trains, coaches, or consults others",
		Icon:        "graduation-cap",
		Color:       "#059669", // Emerald
		SortOrder:   1,
		CanHost:     true,
	},
	{
		Slug:        "professional",
		Name:        "Professional / Employee",
		DisplayName: "💼 Professional / Employee",
		Description: "Working professional or employee",
		Icon:        "briefcase",
		Color:       "#2563EB", // Blue
		SortOrder:   2,
		CanHost:     false,
	},
	{
		Slug:        "student",
		Name:        "Student / Learner",
		DisplayName: "📚 Student / Learner",
		Description: "Student or lifelong learner",
		Icon:        "book-open",
		Color:       "#D97706", // Amber
		SortOrder:   3,
		CanHost:     false,
	},
	{
		Slug:        "other",
		Name:        "Other / Just browsing",
		DisplayName: "🔍 Other / Just browsing",
		Description: "Exploring the platform",
		Icon:        "user",
		Color:       "#6B7280", // Gray
		SortOrder:   4,
		CanHost:     false,
	},
}

// ProfessionalTypeMap for quick lookups
var ProfessionalTypeMap = map[string]ProfessionalTypeInfo{
	"trainer":      ProfessionalTypes[0],
	"professional": ProfessionalTypes[1],
	"student":      ProfessionalTypes[2],
	"other":        ProfessionalTypes[3],
}

// AllProfessionalTypeSlugs returns all valid professional type slugs
func AllProfessionalTypeSlugs() []string {
	values := make([]string, len(ProfessionalTypes))
	for i, pt := range ProfessionalTypes {
		values[i] = pt.Slug
	}
	return values
}

// GetProfessionalTypeInfo returns ProfessionalTypeInfo by slug
func GetProfessionalTypeInfo(slug string) (ProfessionalTypeInfo, bool) {
	info, ok := ProfessionalTypeMap[slug]
	return info, ok
}

// IsValidProfessionalType checks if a professional type slug is valid
func IsValidProfessionalType(slug string) bool {
	_, ok := ProfessionalTypeMap[slug]
	return ok
}

// GetProfessionalTypesThatCanHost returns professional types that can host events
func GetProfessionalTypesThatCanHost() []ProfessionalTypeInfo {
	var result []ProfessionalTypeInfo
	for _, pt := range ProfessionalTypes {
		if pt.CanHost {
			result = append(result, pt)
		}
	}
	return result
}

// GetProfessionalTypeBySlug returns professional type info by slug (alias for GetProfessionalTypeInfo)
func GetProfessionalTypeBySlug(slug string) (ProfessionalTypeInfo, bool) {
	return GetProfessionalTypeInfo(slug)
}

// GetProfessionalTypeNames returns all professional type names
func GetProfessionalTypeNames() []string {
	names := make([]string, len(ProfessionalTypes))
	for i, pt := range ProfessionalTypes {
		names[i] = pt.Name
	}
	return names
}

// GetProfessionalTypeSlugsMap returns a map of slug to name
func GetProfessionalTypeSlugsMap() map[string]string {
	result := make(map[string]string)
	for _, pt := range ProfessionalTypes {
		result[pt.Slug] = pt.Name
	}
	return result
}