// internal/modules/events/domain/category_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// CATEGORY - Value Object
// ============================================================

// CategoryValue is an alias for the shared Category
type CategoryValue = types.Category

// Category constants - Re-exported from shared types for convenience
const (
	CategoryTechnology = types.CategoryTechnology
	CategoryBusiness   = types.CategoryBusiness
	CategoryHealth     = types.CategoryHealth
	CategoryEducation  = types.CategoryEducation
	CategoryFinance    = types.CategoryFinance
	CategoryScience    = types.CategoryScience
	CategoryArts       = types.CategoryArts
	CategorySports     = types.CategorySports
)

// AllCategories re-exported from shared types
var AllCategories = types.AllCategories

// CategoryInfo holds metadata for each category
type CategoryInfo struct {
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
// CATEGORY REGISTRY - Domain specific wrapper
// ============================================================

var categoryRegistry = map[types.Category]CategoryInfo{
	types.CategoryTechnology: {
		Slug:        types.CategoryTechnologySlug,
		Name:        types.CategoryTechnologyName,
		DisplayName: types.CategoryTechnologyDisplayName,
		Description: types.CategoryTechnologyDescription,
		Icon:        types.CategoryTechnologyIcon,
		Color:       types.CategoryTechnologyColor,
		SortOrder:   types.CategoryTechnologySortOrder,
		IsActive:    true,
	},
	types.CategoryBusiness: {
		Slug:        types.CategoryBusinessSlug,
		Name:        types.CategoryBusinessName,
		DisplayName: types.CategoryBusinessDisplayName,
		Description: types.CategoryBusinessDescription,
		Icon:        types.CategoryBusinessIcon,
		Color:       types.CategoryBusinessColor,
		SortOrder:   types.CategoryBusinessSortOrder,
		IsActive:    true,
	},
	types.CategoryHealth: {
		Slug:        types.CategoryHealthSlug,
		Name:        types.CategoryHealthName,
		DisplayName: types.CategoryHealthDisplayName,
		Description: types.CategoryHealthDescription,
		Icon:        types.CategoryHealthIcon,
		Color:       types.CategoryHealthColor,
		SortOrder:   types.CategoryHealthSortOrder,
		IsActive:    true,
	},
	types.CategoryEducation: {
		Slug:        types.CategoryEducationSlug,
		Name:        types.CategoryEducationName,
		DisplayName: types.CategoryEducationDisplayName,
		Description: types.CategoryEducationDescription,
		Icon:        types.CategoryEducationIcon,
		Color:       types.CategoryEducationColor,
		SortOrder:   types.CategoryEducationSortOrder,
		IsActive:    true,
	},
	types.CategoryFinance: {
		Slug:        types.CategoryFinanceSlug,
		Name:        types.CategoryFinanceName,
		DisplayName: types.CategoryFinanceDisplayName,
		Description: types.CategoryFinanceDescription,
		Icon:        types.CategoryFinanceIcon,
		Color:       types.CategoryFinanceColor,
		SortOrder:   types.CategoryFinanceSortOrder,
		IsActive:    true,
	},
	types.CategoryScience: {
		Slug:        types.CategoryScienceSlug,
		Name:        types.CategoryScienceName,
		DisplayName: types.CategoryScienceDisplayName,
		Description: types.CategoryScienceDescription,
		Icon:        types.CategoryScienceIcon,
		Color:       types.CategoryScienceColor,
		SortOrder:   types.CategoryScienceSortOrder,
		IsActive:    true,
	},
	types.CategoryArts: {
		Slug:        types.CategoryArtsSlug,
		Name:        types.CategoryArtsName,
		DisplayName: types.CategoryArtsDisplayName,
		Description: types.CategoryArtsDescription,
		Icon:        types.CategoryArtsIcon,
		Color:       types.CategoryArtsColor,
		SortOrder:   types.CategoryArtsSortOrder,
		IsActive:    true,
	},
	types.CategorySports: {
		Slug:        types.CategorySportsSlug,
		Name:        types.CategorySportsName,
		DisplayName: types.CategorySportsDisplayName,
		Description: types.CategorySportsDescription,
		Icon:        types.CategorySportsIcon,
		Color:       types.CategorySportsColor,
		SortOrder:   types.CategorySportsSortOrder,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetCategoryInfo returns the category info for a given category
func GetCategoryInfo(category CategoryValue) (CategoryInfo, bool) {
	info, ok := categoryRegistry[category]
	return info, ok
}

// GetCategorySlug returns the slug for a category
func GetCategorySlug(category CategoryValue) string {
	return category.GetSlug()
}

// GetCategoryName returns the name for a category
func GetCategoryName(category CategoryValue) string {
	return category.GetName()
}

// GetCategoryDisplayName returns the display name for a category
func GetCategoryDisplayName(category CategoryValue) string {
	return category.GetDisplayName()
}

// IsCategoryValid checks if the category is valid
func IsCategoryValid(category CategoryValue) bool {
	return category.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllCategoryInfos returns all category infos
func AllCategoryInfos() []CategoryInfo {
	infos := make([]CategoryInfo, 0, len(categoryRegistry))
	for _, info := range categoryRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllCategorySlugs returns all category slugs (with hyphens)
func AllCategorySlugs() []string {
	return types.AllCategorySlugs()
}

// AllCategoryNames returns all internal category names (with underscores)
func AllCategoryNames() []string {
	return types.AllCategoryNames()
}

// ActiveCategoryInfos returns only active category infos
func ActiveCategoryInfos() []CategoryInfo {
	infos := make([]CategoryInfo, 0)
	for _, info := range categoryRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetCategoryBySlug returns category info by slug
func GetCategoryBySlug(slug string) (CategoryInfo, bool) {
	for category, info := range categoryRegistry {
		if category.GetSlug() == slug {
			return info, true
		}
	}
	return CategoryInfo{}, false
}

// GetCategoryByName returns category info by internal name (with underscores)
func GetCategoryByName(name string) (CategoryInfo, bool) {
	category, ok := types.ParseCategory(name)
	if !ok {
		return CategoryInfo{}, false
	}
	return GetCategoryInfo(category)
}

// GetCategoryByDisplayName returns category info by display name
func GetCategoryByDisplayName(displayName string) (CategoryInfo, bool) {
	for _, info := range categoryRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return CategoryInfo{}, false
}