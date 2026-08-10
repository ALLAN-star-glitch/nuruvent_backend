package domain

// ============================================================
// INSTITUTION TYPE - Value Object
// ============================================================

type InstitutionTypeValue string

const (
	InstitutionTypeCompany    InstitutionTypeValue = "company"
	InstitutionTypeInstitute  InstitutionTypeValue = "institute"
	InstitutionTypeAssociation InstitutionTypeValue = "association"
	InstitutionTypeSchool     InstitutionTypeValue = "school"
	InstitutionTypeUniversity InstitutionTypeValue = "university"
)

// InstitutionTypeInfo holds metadata for each institution type
type InstitutionTypeInfo struct {
	Slug        InstitutionTypeValue
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	IsActive    bool
}

// Private registry (prevents external mutation)
var institutionTypeRegistry = map[InstitutionTypeValue]InstitutionTypeInfo{
	InstitutionTypeCompany: {
		Slug:        InstitutionTypeCompany,
		Name:        "Company",
		DisplayName: "🏢 Company",
		Description: "Corporate organization",
		Icon:        "building",
		Color:       "#7C3AED",
		SortOrder:   1,
		IsActive:    true,
	},
	InstitutionTypeInstitute: {
		Slug:        InstitutionTypeInstitute,
		Name:        "Institute",
		DisplayName: "🏛️ Institute",
		Description: "Educational or research institute",
		Icon:        "landmark",
		Color:       "#4F46E5",
		SortOrder:   2,
		IsActive:    true,
	},
	InstitutionTypeAssociation: {
		Slug:        InstitutionTypeAssociation,
		Name:        "Association",
		DisplayName: "🤝 Association",
		Description: "Professional or trade association",
		Icon:        "users",
		Color:       "#0EA5E9",
		SortOrder:   3,
		IsActive:    true,
	},
	InstitutionTypeSchool: {
		Slug:        InstitutionTypeSchool,
		Name:        "School",
		DisplayName: "📚 School",
		Description: "Primary or secondary educational institution",
		Icon:        "school",
		Color:       "#F59E0B",
		SortOrder:   4,
		IsActive:    true,
	},
	InstitutionTypeUniversity: {
		Slug:        InstitutionTypeUniversity,
		Name:        "University",
		DisplayName: "🎓 University",
		Description: "Higher education institution",
		Icon:        "graduation-cap",
		Color:       "#10B981",
		SortOrder:   5,
		IsActive:    true,
	},
}

// ============================================================
// DOMAIN METHODS (on InstitutionTypeValue)
// ============================================================

// String returns the string representation
func (i InstitutionTypeValue) String() string {
	return string(i)
}

// IsValid checks if the institution type exists in the registry
func (i InstitutionTypeValue) IsValid() bool {
	_, ok := institutionTypeRegistry[i]
	return ok
}

// Info returns the InstitutionTypeInfo for this type
func (i InstitutionTypeValue) Info() (InstitutionTypeInfo, bool) {
	info, ok := institutionTypeRegistry[i]
	return info, ok
}

// IsActive returns whether the institution type is active
func (i InstitutionTypeValue) IsActive() bool {
	info, ok := institutionTypeRegistry[i]
	if !ok {
		return false
	}
	return info.IsActive
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// ParseInstitutionType parses a string into an InstitutionTypeValue
func ParseInstitutionType(slug string) (InstitutionTypeValue, bool) {
	t := InstitutionTypeValue(slug)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// AllInstitutionTypeInfos returns all institution type infos (read-only copy)
func AllInstitutionTypeInfos() []InstitutionTypeInfo {
	infos := make([]InstitutionTypeInfo, 0, len(institutionTypeRegistry))
	for _, info := range institutionTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllInstitutionTypeSlugs returns all valid slugs
func AllInstitutionTypeSlugs() []string {
	slugs := make([]string, 0, len(institutionTypeRegistry))
	for slug := range institutionTypeRegistry {
		slugs = append(slugs, string(slug))
	}
	return slugs
}

// ActiveInstitutionTypeInfos returns only active institution types
func ActiveInstitutionTypeInfos() []InstitutionTypeInfo {
	infos := make([]InstitutionTypeInfo, 0)
	for _, info := range institutionTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// ActiveInstitutionTypeSlugs returns only active institution type slugs
func ActiveInstitutionTypeSlugs() []string {
	slugs := make([]string, 0)
	for slug, info := range institutionTypeRegistry {
		if info.IsActive {
			slugs = append(slugs, string(slug))
		}
	}
	return slugs
}