// internal/modules/auth/authdomain/institution_type.go

package authdomain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// INSTITUTION TYPE - Value Object
// ============================================================

// InstitutionTypeValue is an alias for the shared InstitutionType
type InstitutionTypeValue = types.InstitutionType

// Type constants - Re-exported from shared types for convenience
const (
	InstitutionTypeCompany     = types.InstitutionTypeCompany
	InstitutionTypeInstitute   = types.InstitutionTypeInstitute
	InstitutionTypeAssociation = types.InstitutionTypeAssociation
	InstitutionTypeSchool      = types.InstitutionTypeSchool
	InstitutionTypeUniversity  = types.InstitutionTypeUniversity
	InstitutionTypeCollege     = types.InstitutionTypeCollege      // NEW
	InstitutionTypeHighSchool  = types.InstitutionTypeHighSchool   // NEW
	InstitutionTypeNgo         = types.InstitutionTypeNgo
	InstitutionTypeGovernment  = types.InstitutionTypeGovernment
	InstitutionTypeReligious   = types.InstitutionTypeReligious
	InstitutionTypeCommunity   = types.InstitutionTypeCommunity
	InstitutionTypeOther       = types.InstitutionTypeOther
)

// AllInstitutionTypes re-exported from shared types
var AllInstitutionTypes = types.AllInstitutionTypes

// InstitutionTypeInfo holds metadata for each institution type
type InstitutionTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	IsActive    bool
}

// ============================================================
// INSTITUTION TYPE REGISTRY - Domain specific wrapper
// ============================================================

var institutionTypeRegistry = map[types.InstitutionType]InstitutionTypeInfo{
	types.InstitutionTypeCompany: {
		Slug:        types.InstitutionTypeCompanySlug,
		Name:        types.InstitutionTypeCompanyName,
		DisplayName: types.InstitutionTypeCompanyDisplayName,
		Description: types.InstitutionTypeCompanyDescription,
		Icon:        types.InstitutionTypeCompanyIcon,
		Color:       types.InstitutionTypeCompanyColor,
		SortOrder:   types.InstitutionTypeCompanySortOrder,
		IsActive:    true,
	},
	types.InstitutionTypeInstitute: {
		Slug:        types.InstitutionTypeInstituteSlug,
		Name:        types.InstitutionTypeInstituteName,
		DisplayName: types.InstitutionTypeInstituteDisplayName,
		Description: types.InstitutionTypeInstituteDescription,
		Icon:        types.InstitutionTypeInstituteIcon,
		Color:       types.InstitutionTypeInstituteColor,
		SortOrder:   types.InstitutionTypeInstituteSortOrder,
		IsActive:    true,
	},
	types.InstitutionTypeAssociation: {
		Slug:        types.InstitutionTypeAssociationSlug,
		Name:        types.InstitutionTypeAssociationName,
		DisplayName: types.InstitutionTypeAssociationDisplayName,
		Description: types.InstitutionTypeAssociationDescription,
		Icon:        types.InstitutionTypeAssociationIcon,
		Color:       types.InstitutionTypeAssociationColor,
		SortOrder:   types.InstitutionTypeAssociationSortOrder,
		IsActive:    true,
	},
	types.InstitutionTypeSchool: {
		Slug:        types.InstitutionTypeSchoolSlug,
		Name:        types.InstitutionTypeSchoolName,
		DisplayName: types.InstitutionTypeSchoolDisplayName,
		Description: types.InstitutionTypeSchoolDescription,
		Icon:        types.InstitutionTypeSchoolIcon,
		Color:       types.InstitutionTypeSchoolColor,
		SortOrder:   types.InstitutionTypeSchoolSortOrder,
		IsActive:    true,
	},
	types.InstitutionTypeUniversity: {
		Slug:        types.InstitutionTypeUniversitySlug,
		Name:        types.InstitutionTypeUniversityName,
		DisplayName: types.InstitutionTypeUniversityDisplayName,
		Description: types.InstitutionTypeUniversityDescription,
		Icon:        types.InstitutionTypeUniversityIcon,
		Color:       types.InstitutionTypeUniversityColor,
		SortOrder:   types.InstitutionTypeUniversitySortOrder,
		IsActive:    true,
	},
	// NEW: College
	types.InstitutionTypeCollege: {
		Slug:        types.InstitutionTypeCollegeSlug,
		Name:        types.InstitutionTypeCollegeName,
		DisplayName: types.InstitutionTypeCollegeDisplayName,
		Description: types.InstitutionTypeCollegeDescription,
		Icon:        types.InstitutionTypeCollegeIcon,
		Color:       types.InstitutionTypeCollegeColor,
		SortOrder:   types.InstitutionTypeCollegeSortOrder,
		IsActive:    true,
	},
	// NEW: High School
	types.InstitutionTypeHighSchool: {
		Slug:        types.InstitutionTypeHighSchoolSlug,
		Name:        types.InstitutionTypeHighSchoolName,
		DisplayName: types.InstitutionTypeHighSchoolDisplayName,
		Description: types.InstitutionTypeHighSchoolDescription,
		Icon:        types.InstitutionTypeHighSchoolIcon,
		Color:       types.InstitutionTypeHighSchoolColor,
		SortOrder:   types.InstitutionTypeHighSchoolSortOrder,
		IsActive:    true,
	},
	// NGO / Non-Profit
	types.InstitutionTypeNgo: {
		Slug:        types.InstitutionTypeNgoSlug,
		Name:        types.InstitutionTypeNgoName,
		DisplayName: types.InstitutionTypeNgoDisplayName,
		Description: types.InstitutionTypeNgoDescription,
		Icon:        types.InstitutionTypeNgoIcon,
		Color:       types.InstitutionTypeNgoColor,
		SortOrder:   types.InstitutionTypeNgoSortOrder,
		IsActive:    true,
	},
	// Government Agency
	types.InstitutionTypeGovernment: {
		Slug:        types.InstitutionTypeGovernmentSlug,
		Name:        types.InstitutionTypeGovernmentName,
		DisplayName: types.InstitutionTypeGovernmentDisplayName,
		Description: types.InstitutionTypeGovernmentDescription,
		Icon:        types.InstitutionTypeGovernmentIcon,
		Color:       types.InstitutionTypeGovernmentColor,
		SortOrder:   types.InstitutionTypeGovernmentSortOrder,
		IsActive:    true,
	},
	// Religious Organization
	types.InstitutionTypeReligious: {
		Slug:        types.InstitutionTypeReligiousSlug,
		Name:        types.InstitutionTypeReligiousName,
		DisplayName: types.InstitutionTypeReligiousDisplayName,
		Description: types.InstitutionTypeReligiousDescription,
		Icon:        types.InstitutionTypeReligiousIcon,
		Color:       types.InstitutionTypeReligiousColor,
		SortOrder:   types.InstitutionTypeReligiousSortOrder,
		IsActive:    true,
	},
	// Community Organization
	types.InstitutionTypeCommunity: {
		Slug:        types.InstitutionTypeCommunitySlug,
		Name:        types.InstitutionTypeCommunityName,
		DisplayName: types.InstitutionTypeCommunityDisplayName,
		Description: types.InstitutionTypeCommunityDescription,
		Icon:        types.InstitutionTypeCommunityIcon,
		Color:       types.InstitutionTypeCommunityColor,
		SortOrder:   types.InstitutionTypeCommunitySortOrder,
		IsActive:    true,
	},
	// Other
	types.InstitutionTypeOther: {
		Slug:        types.InstitutionTypeOtherSlug,
		Name:        types.InstitutionTypeOtherName,
		DisplayName: types.InstitutionTypeOtherDisplayName,
		Description: types.InstitutionTypeOtherDescription,
		Icon:        types.InstitutionTypeOtherIcon,
		Color:       types.InstitutionTypeOtherColor,
		SortOrder:   types.InstitutionTypeOtherSortOrder,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetInstitutionTypeInfo returns the type info for a given institution type
func GetInstitutionTypeInfo(institutionType InstitutionTypeValue) (InstitutionTypeInfo, bool) {
	info, ok := institutionTypeRegistry[institutionType]
	return info, ok
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllInstitutionTypeInfos returns all type infos
func AllInstitutionTypeInfos() []InstitutionTypeInfo {
	infos := make([]InstitutionTypeInfo, 0, len(institutionTypeRegistry))
	for _, info := range institutionTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllInstitutionTypeSlugs returns all type slugs (with hyphens)
func AllInstitutionTypeSlugs() []string {
	return types.AllInstitutionTypeSlugs()
}

// AllInstitutionTypeNames returns all internal type names (with underscores)
func AllInstitutionTypeNames() []string {
	return types.AllInstitutionTypeNames()
}

// AllInstitutionTypeDisplayNames returns all display names
func AllInstitutionTypeDisplayNames() []string {
	return types.AllInstitutionTypeDisplayNames()
}

// ActiveInstitutionTypeInfos returns only active type infos
func ActiveInstitutionTypeInfos() []InstitutionTypeInfo {
	infos := make([]InstitutionTypeInfo, 0)
	for _, info := range institutionTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetInstitutionTypeBySlug returns type info by slug
func GetInstitutionTypeBySlug(slug string) (InstitutionTypeInfo, bool) {
	for _, info := range institutionTypeRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return InstitutionTypeInfo{}, false
}

// GetInstitutionTypeByName returns type info by internal name (with underscores)
func GetInstitutionTypeByName(name string) (InstitutionTypeInfo, bool) {
	institutionType, ok := types.ParseInstitutionType(name)
	if !ok {
		return InstitutionTypeInfo{}, false
	}
	return GetInstitutionTypeInfo(institutionType)
}

// GetInstitutionTypeByDisplayName returns type info by display name
func GetInstitutionTypeByDisplayName(displayName string) (InstitutionTypeInfo, bool) {
	for _, info := range institutionTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return InstitutionTypeInfo{}, false
}

// IsInstitutionTypeValid checks if the institution type is valid
func IsInstitutionTypeValid(institutionType InstitutionTypeValue) bool {
	return institutionType.IsValid()
}