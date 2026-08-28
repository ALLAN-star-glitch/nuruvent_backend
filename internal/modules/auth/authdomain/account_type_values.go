

package authdomain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// ACCOUNT TYPE - Value Object
// ============================================================

// AccountTypeValue is an alias for the shared AccountType
type AccountTypeValue = types.AccountType

// Type constants - Re-exported from shared types for convenience
const (
	AccountTypePersonal    = types.AccountTypePersonal
	AccountTypeInstitution = types.AccountTypeInstitution
)

// AllAccountTypes re-exported from shared types
var AllAccountTypes = types.AllAccountTypes

// AccountTypeInfo holds metadata for each account type
type AccountTypeInfo struct {
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
// ACCOUNT TYPE REGISTRY - Domain specific wrapper
// ============================================================

var accountTypeRegistry = map[types.AccountType]AccountTypeInfo{
	types.AccountTypePersonal: {
		Slug:        types.AccountTypePersonalSlug,
		Name:        types.AccountTypePersonalName,
		DisplayName: types.AccountTypePersonalDisplayName,
		Description: types.AccountTypePersonalDescription,
		Icon:        types.AccountTypePersonalIcon,
		Color:       types.AccountTypePersonalColor,
		SortOrder:   types.AccountTypePersonalSortOrder,
		IsActive:    true,
	},
	types.AccountTypeInstitution: {
		Slug:        types.AccountTypeInstitutionSlug,
		Name:        types.AccountTypeInstitutionName,
		DisplayName: types.AccountTypeInstitutionDisplayName,
		Description: types.AccountTypeInstitutionDescription,
		Icon:        types.AccountTypeInstitutionIcon,
		Color:       types.AccountTypeInstitutionColor,
		SortOrder:   types.AccountTypeInstitutionSortOrder,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetAccountTypeInfo returns the type info for a given account type
func GetAccountTypeInfo(accountType AccountTypeValue) (AccountTypeInfo, bool) {
	info, ok := accountTypeRegistry[accountType]
	return info, ok
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllAccountTypeInfos returns all type infos
func AllAccountTypeInfos() []AccountTypeInfo {
	infos := make([]AccountTypeInfo, 0, len(accountTypeRegistry))
	for _, info := range accountTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllAccountTypeSlugs returns all type slugs (with hyphens)
func AllAccountTypeSlugs() []string {
	return types.AllAccountTypeSlugs()
}

// AllAccountTypeNames returns all internal type names (with underscores)
func AllAccountTypeNames() []string {
	return types.AllAccountTypeNames()
}

// AllAccountTypeDisplayNames returns all display names
func AllAccountTypeDisplayNames() []string {
	return types.AllAccountTypeDisplayNames()
}

// ActiveAccountTypeInfos returns only active type infos
func ActiveAccountTypeInfos() []AccountTypeInfo {
	infos := make([]AccountTypeInfo, 0)
	for _, info := range accountTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetAccountTypeBySlug returns type info by slug
func GetAccountTypeBySlug(slug string) (AccountTypeInfo, bool) {
	for _, info := range accountTypeRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return AccountTypeInfo{}, false
}

// GetAccountTypeByName returns type info by internal name (with underscores)
func GetAccountTypeByName(name string) (AccountTypeInfo, bool) {
	accountType, ok := types.ParseAccountType(name)
	if !ok {
		return AccountTypeInfo{}, false
	}
	return GetAccountTypeInfo(accountType)
}

// GetAccountTypeByDisplayName returns type info by display name
func GetAccountTypeByDisplayName(displayName string) (AccountTypeInfo, bool) {
	for _, info := range accountTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return AccountTypeInfo{}, false
}

// IsAccountTypeValid checks if the account type is valid
func IsAccountTypeValid(accountType AccountTypeValue) bool {
	return accountType.IsValid()
}