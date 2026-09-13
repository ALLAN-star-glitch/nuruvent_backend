

package authdomain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// PROFESSIONAL TYPE - Value Object
// ============================================================

// ProfessionalTypeValue is an alias for the shared ProfessionalType
type ProfessionalTypeValue = types.ProfessionalType

// Type constants - Re-exported from shared types for convenience
const (
	ProfessionalTypeTrainer    = types.ProfessionalTypeTrainer
	ProfessionalTypeCoach      = types.ProfessionalTypeCoach
	ProfessionalTypeConsultant = types.ProfessionalTypeConsultant
	ProfessionalTypeFreelancer = types.ProfessionalTypeFreelancer
)

// AllProfessionalTypes re-exported from shared types
var AllProfessionalTypes = types.AllProfessionalTypes

// ProfessionalTypeInfo holds metadata for each professional type
type ProfessionalTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	CanHost     bool
	IsActive    bool
}

// ============================================================
// PROFESSIONAL TYPE REGISTRY - Domain specific wrapper
// ============================================================

var professionalTypeRegistry = map[types.ProfessionalType]ProfessionalTypeInfo{
	types.ProfessionalTypeTrainer: {
		Slug:        types.ProfessionalTypeTrainerSlug,
		Name:        types.ProfessionalTypeTrainerName,
		DisplayName: types.ProfessionalTypeTrainerDisplayName,
		Description: types.ProfessionalTypeTrainerDescription,
		Icon:        types.ProfessionalTypeTrainerIcon,
		Color:       types.ProfessionalTypeTrainerColor,
		SortOrder:   types.ProfessionalTypeTrainerSortOrder,
		CanHost:     true,
		IsActive:    true,
	},
	types.ProfessionalTypeCoach: {
		Slug:        types.ProfessionalTypeCoachSlug,
		Name:        types.ProfessionalTypeCoachName,
		DisplayName: types.ProfessionalTypeCoachDisplayName,
		Description: types.ProfessionalTypeCoachDescription,
		Icon:        types.ProfessionalTypeCoachIcon,
		Color:       types.ProfessionalTypeCoachColor,
		SortOrder:   types.ProfessionalTypeCoachSortOrder,
		CanHost:     true,
		IsActive:    true,
	},
	types.ProfessionalTypeConsultant: {
		Slug:        types.ProfessionalTypeConsultantSlug,
		Name:        types.ProfessionalTypeConsultantName,
		DisplayName: types.ProfessionalTypeConsultantDisplayName,
		Description: types.ProfessionalTypeConsultantDescription,
		Icon:        types.ProfessionalTypeConsultantIcon,
		Color:       types.ProfessionalTypeConsultantColor,
		SortOrder:   types.ProfessionalTypeConsultantSortOrder,
		CanHost:     true,
		IsActive:    true,
	},
	types.ProfessionalTypeFreelancer: {
		Slug:        types.ProfessionalTypeFreelancerSlug,
		Name:        types.ProfessionalTypeFreelancerName,
		DisplayName: types.ProfessionalTypeFreelancerDisplayName,
		Description: types.ProfessionalTypeFreelancerDescription,
		Icon:        types.ProfessionalTypeFreelancerIcon,
		Color:       types.ProfessionalTypeFreelancerColor,
		SortOrder:   types.ProfessionalTypeFreelancerSortOrder,
		CanHost:     false,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetProfessionalTypeInfo returns the type info for a given professional type
func GetProfessionalTypeInfo(professionalType ProfessionalTypeValue) (ProfessionalTypeInfo, bool) {
	info, ok := professionalTypeRegistry[professionalType]
	return info, ok
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllProfessionalTypeInfos returns all type infos
func AllProfessionalTypeInfos() []ProfessionalTypeInfo {
	infos := make([]ProfessionalTypeInfo, 0, len(professionalTypeRegistry))
	for _, info := range professionalTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllProfessionalTypeSlugs returns all type slugs (with hyphens)
func AllProfessionalTypeSlugs() []string {
	return types.AllProfessionalTypeSlugs()
}

// AllProfessionalTypeNames returns all internal type names (with underscores)
func AllProfessionalTypeNames() []string {
	return types.AllProfessionalTypeNames()
}

// AllProfessionalTypeDisplayNames returns all display names
func AllProfessionalTypeDisplayNames() []string {
	return types.AllProfessionalTypeDisplayNames()
}

// ActiveProfessionalTypeInfos returns only active type infos
func ActiveProfessionalTypeInfos() []ProfessionalTypeInfo {
	infos := make([]ProfessionalTypeInfo, 0)
	for _, info := range professionalTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// HostProfessionalTypeInfos returns only professional types that can host events
func HostProfessionalTypeInfos() []ProfessionalTypeInfo {
	infos := make([]ProfessionalTypeInfo, 0)
	for _, info := range professionalTypeRegistry {
		if info.CanHost && info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetProfessionalTypeBySlug returns type info by slug
func GetProfessionalTypeBySlug(slug string) (ProfessionalTypeInfo, bool) {
	for _, info := range professionalTypeRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return ProfessionalTypeInfo{}, false
}

// GetProfessionalTypeByName returns type info by internal name (with underscores)
func GetProfessionalTypeByName(name string) (ProfessionalTypeInfo, bool) {
	professionalType, ok := types.ParseProfessionalType(name)
	if !ok {
		return ProfessionalTypeInfo{}, false
	}
	return GetProfessionalTypeInfo(professionalType)
}

// GetProfessionalTypeByDisplayName returns type info by display name
func GetProfessionalTypeByDisplayName(displayName string) (ProfessionalTypeInfo, bool) {
	for _, info := range professionalTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return ProfessionalTypeInfo{}, false
}

// IsProfessionalTypeValid checks if the professional type is valid
func IsProfessionalTypeValid(professionalType ProfessionalTypeValue) bool {
	return professionalType.IsValid()
}