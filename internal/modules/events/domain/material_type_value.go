// internal/modules/events/domain/material_type_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// MATERIAL TYPE - Value Object
// ============================================================

// MaterialTypeValue is an alias for the shared MaterialType
type MaterialTypeValue = types.MaterialType

// Type constants - Re-exported from shared types for convenience
const (
	MaterialTypePDF      = types.MaterialTypePDF
	MaterialTypeVideo    = types.MaterialTypeVideo
	MaterialTypeLink     = types.MaterialTypeLink
	MaterialTypeDocument = types.MaterialTypeDocument
)

// AllMaterialTypes re-exported from shared types
var AllMaterialTypes = types.AllMaterialTypes

// MaterialTypeInfo holds metadata for each material type
type MaterialTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Icon        string
	IsActive    bool
}

// ============================================================
// MATERIAL TYPE REGISTRY - Domain specific wrapper
// ============================================================

var materialTypeRegistry = map[types.MaterialType]MaterialTypeInfo{
	types.MaterialTypePDF: {
		Slug:        types.MaterialTypePDFSlug,
		Name:        types.MaterialTypePDFName,
		DisplayName: types.MaterialTypePDFDisplayName,
		Description: types.MaterialTypePDFDescription,
		Icon:        types.MaterialTypePDFIcon,
		IsActive:    true,
	},
	types.MaterialTypeVideo: {
		Slug:        types.MaterialTypeVideoSlug,
		Name:        types.MaterialTypeVideoName,
		DisplayName: types.MaterialTypeVideoDisplayName,
		Description: types.MaterialTypeVideoDescription,
		Icon:        types.MaterialTypeVideoIcon,
		IsActive:    true,
	},
	types.MaterialTypeLink: {
		Slug:        types.MaterialTypeLinkSlug,
		Name:        types.MaterialTypeLinkName,
		DisplayName: types.MaterialTypeLinkDisplayName,
		Description: types.MaterialTypeLinkDescription,
		Icon:        types.MaterialTypeLinkIcon,
		IsActive:    true,
	},
	types.MaterialTypeDocument: {
		Slug:        types.MaterialTypeDocumentSlug,
		Name:        types.MaterialTypeDocumentName,
		DisplayName: types.MaterialTypeDocumentDisplayName,
		Description: types.MaterialTypeDocumentDescription,
		Icon:        types.MaterialTypeDocumentIcon,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetMaterialTypeInfo returns the material type info for a given material type
func GetMaterialTypeInfo(materialType MaterialTypeValue) (MaterialTypeInfo, bool) {
	info, ok := materialTypeRegistry[materialType]
	return info, ok
}

// GetMaterialTypeSlug returns the slug for a material type
func GetMaterialTypeSlug(materialType MaterialTypeValue) string {
	return materialType.GetSlug()
}

// GetMaterialTypeName returns the name for a material type
func GetMaterialTypeName(materialType MaterialTypeValue) string {
	return materialType.GetName()
}

// GetMaterialTypeDisplayName returns the display name for a material type
func GetMaterialTypeDisplayName(materialType MaterialTypeValue) string {
	return materialType.GetDisplayName()
}

// IsMaterialTypeValid checks if the material type is valid
func IsMaterialTypeValid(materialType MaterialTypeValue) bool {
	return materialType.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllMaterialTypeInfos returns all material type infos
func AllMaterialTypeInfos() []MaterialTypeInfo {
	infos := make([]MaterialTypeInfo, 0, len(materialTypeRegistry))
	for _, info := range materialTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllMaterialTypeSlugs returns all material type slugs (with hyphens)
func AllMaterialTypeSlugs() []string {
	return types.AllMaterialTypeSlugs()
}

// AllMaterialTypeNames returns all internal material type names (with underscores)
func AllMaterialTypeNames() []string {
	return types.AllMaterialTypeNames()
}

// ActiveMaterialTypeInfos returns only active material type infos
func ActiveMaterialTypeInfos() []MaterialTypeInfo {
	infos := make([]MaterialTypeInfo, 0)
	for _, info := range materialTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetMaterialTypeBySlug returns material type info by slug
func GetMaterialTypeBySlug(slug string) (MaterialTypeInfo, bool) {
	for materialType, info := range materialTypeRegistry {
		if materialType.GetSlug() == slug {
			return info, true
		}
	}
	return MaterialTypeInfo{}, false
}

// GetMaterialTypeByName returns material type info by internal name (with underscores)
func GetMaterialTypeByName(name string) (MaterialTypeInfo, bool) {
	materialType, ok := types.ParseMaterialType(name)
	if !ok {
		return MaterialTypeInfo{}, false
	}
	return GetMaterialTypeInfo(materialType)
}

// GetMaterialTypeByDisplayName returns material type info by display name
func GetMaterialTypeByDisplayName(displayName string) (MaterialTypeInfo, bool) {
	for _, info := range materialTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return MaterialTypeInfo{}, false
}