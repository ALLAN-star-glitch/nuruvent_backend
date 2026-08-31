// internal/modules/events/domain/ticket_type_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// TICKET TYPE - Value Object
// ============================================================

// TicketTypeValue is an alias for the shared TicketType
type TicketTypeValue = types.TicketType

// Type constants - Re-exported from shared types for convenience
const (
	TicketTypeGeneral   = types.TicketTypeGeneral
	TicketTypeEarlyBird = types.TicketTypeEarlyBird
	TicketTypeVIP       = types.TicketTypeVIP
	TicketTypeGroup     = types.TicketTypeGroup
)

// AllTicketTypes re-exported from shared types
var AllTicketTypes = types.AllTicketTypes

// TicketTypeInfo holds metadata for each ticket type
type TicketTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	SortOrder   int
	IsActive    bool
}

// ============================================================
// TICKET TYPE REGISTRY - Domain specific wrapper
// ============================================================

var ticketTypeRegistry = map[types.TicketType]TicketTypeInfo{
	types.TicketTypeGeneral: {
		Slug:        types.TicketTypeGeneralSlug,
		Name:        types.TicketTypeGeneralName,
		DisplayName: types.TicketTypeGeneralDisplayName,
		Description: types.TicketTypeGeneralDescription,
		SortOrder:   types.TicketTypeGeneralSortOrder,
		IsActive:    true,
	},
	types.TicketTypeEarlyBird: {
		Slug:        types.TicketTypeEarlyBirdSlug,
		Name:        types.TicketTypeEarlyBirdName,
		DisplayName: types.TicketTypeEarlyBirdDisplayName,
		Description: types.TicketTypeEarlyBirdDescription,
		SortOrder:   types.TicketTypeEarlyBirdSortOrder,
		IsActive:    true,
	},
	types.TicketTypeVIP: {
		Slug:        types.TicketTypeVIPSlug,
		Name:        types.TicketTypeVIPName,
		DisplayName: types.TicketTypeVIPDisplayName,
		Description: types.TicketTypeVIPDescription,
		SortOrder:   types.TicketTypeVIPSortOrder,
		IsActive:    true,
	},
	types.TicketTypeGroup: {
		Slug:        types.TicketTypeGroupSlug,
		Name:        types.TicketTypeGroupName,
		DisplayName: types.TicketTypeGroupDisplayName,
		Description: types.TicketTypeGroupDescription,
		SortOrder:   types.TicketTypeGroupSortOrder,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// GetTicketTypeInfo returns the ticket type info for a given ticket type
func GetTicketTypeInfo(ticketType TicketTypeValue) (TicketTypeInfo, bool) {
	info, ok := ticketTypeRegistry[ticketType]
	return info, ok
}

// GetTicketTypeSlug returns the slug for a ticket type
func GetTicketTypeSlug(ticketType TicketTypeValue) string {
	return ticketType.GetSlug()
}

// GetTicketTypeName returns the name for a ticket type
func GetTicketTypeName(ticketType TicketTypeValue) string {
	return ticketType.GetName()
}

// GetTicketTypeDisplayName returns the display name for a ticket type
func GetTicketTypeDisplayName(ticketType TicketTypeValue) string {
	return ticketType.GetDisplayName()
}

// IsTicketTypeValid checks if the ticket type is valid
func IsTicketTypeValid(ticketType TicketTypeValue) bool {
	return ticketType.IsValid()
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

// AllTicketTypeInfos returns all ticket type infos
func AllTicketTypeInfos() []TicketTypeInfo {
	infos := make([]TicketTypeInfo, 0, len(ticketTypeRegistry))
	for _, info := range ticketTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

// AllTicketTypeSlugs returns all ticket type slugs (with hyphens)
func AllTicketTypeSlugs() []string {
	return types.AllTicketTypeSlugs()
}

// AllTicketTypeNames returns all internal ticket type names (with underscores)
func AllTicketTypeNames() []string {
	return types.AllTicketTypeNames()
}

// ActiveTicketTypeInfos returns only active ticket type infos
func ActiveTicketTypeInfos() []TicketTypeInfo {
	infos := make([]TicketTypeInfo, 0)
	for _, info := range ticketTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

// GetTicketTypeBySlug returns ticket type info by slug
func GetTicketTypeBySlug(slug string) (TicketTypeInfo, bool) {
	for ticketType, info := range ticketTypeRegistry {
		if ticketType.GetSlug() == slug {
			return info, true
		}
	}
	return TicketTypeInfo{}, false
}

// GetTicketTypeByName returns ticket type info by internal name (with underscores)
func GetTicketTypeByName(name string) (TicketTypeInfo, bool) {
	ticketType, ok := types.ParseTicketType(name)
	if !ok {
		return TicketTypeInfo{}, false
	}
	return GetTicketTypeInfo(ticketType)
}

// GetTicketTypeByDisplayName returns ticket type info by display name
func GetTicketTypeByDisplayName(displayName string) (TicketTypeInfo, bool) {
	for _, info := range ticketTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return TicketTypeInfo{}, false
}