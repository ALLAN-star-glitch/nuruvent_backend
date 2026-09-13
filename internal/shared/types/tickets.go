// internal/shared/types/tickets.go

package types

// ============================================================
// TICKET TYPE CONSTANTS
// ============================================================

// TicketType represents the type of ticket
type TicketType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	TicketTypeGeneralSlug   = "general"
	TicketTypeEarlyBirdSlug = "early-bird"
	TicketTypeVIPSlug       = "vip"
	TicketTypeGroupSlug     = "group"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	TicketTypeGeneralName   = "ticket_type_general"
	TicketTypeEarlyBirdName = "ticket_type_early_bird"
	TicketTypeVIPName       = "ticket_type_vip"
	TicketTypeGroupName     = "ticket_type_group"
)

// Display name constants - Used for UI display
const (
	TicketTypeGeneralDisplayName   = "General Admission"
	TicketTypeEarlyBirdDisplayName = "Early Bird"
	TicketTypeVIPDisplayName       = "VIP Pass"
	TicketTypeGroupDisplayName     = "Group Discount"
)

// Description constants
const (
	TicketTypeGeneralDescription   = "Standard ticket for general admission"
	TicketTypeEarlyBirdDescription = "Discounted ticket for early registrants"
	TicketTypeVIPDescription       = "Premium ticket with additional benefits"
	TicketTypeGroupDescription     = "Discounted ticket for group registrations"
)

// Sort order constants
const (
	TicketTypeGeneralSortOrder   = 1
	TicketTypeEarlyBirdSortOrder = 2
	TicketTypeVIPSortOrder       = 3
	TicketTypeGroupSortOrder     = 4
)

// ============================================================
// TICKET TYPE DEFINITIONS
// ============================================================

// TicketType constants - These are the NAMES (with underscores)
const (
	TicketTypeGeneral   TicketType = TicketTypeGeneralName   // "ticket_type_general"
	TicketTypeEarlyBird TicketType = TicketTypeEarlyBirdName // "ticket_type_early_bird"
	TicketTypeVIP       TicketType = TicketTypeVIPName       // "ticket_type_vip"
	TicketTypeGroup     TicketType = TicketTypeGroupName     // "ticket_type_group"
)

// AllTicketTypes lists all valid ticket types for validation
var AllTicketTypes = []TicketType{
	TicketTypeGeneral,
	TicketTypeEarlyBird,
	TicketTypeVIP,
	TicketTypeGroup,
}

// ============================================================
// BASIC METHODS - On TicketType
// ============================================================

// String returns the string representation (the name with underscores)
func (t TicketType) String() string {
	return string(t)
}

// IsValid checks if the ticket type is valid
func (t TicketType) IsValid() bool {
	for _, tt := range AllTicketTypes {
		if tt == t {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On TicketType
// ============================================================

// GetName returns the internal name (with underscores)
func (t TicketType) GetName() string {
	return string(t)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (t TicketType) GetSlug() string {
	switch t {
	case TicketTypeGeneral:
		return TicketTypeGeneralSlug
	case TicketTypeEarlyBird:
		return TicketTypeEarlyBirdSlug
	case TicketTypeVIP:
		return TicketTypeVIPSlug
	case TicketTypeGroup:
		return TicketTypeGroupSlug
	default:
		return string(t)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (t TicketType) GetDisplayName() string {
	switch t {
	case TicketTypeGeneral:
		return TicketTypeGeneralDisplayName
	case TicketTypeEarlyBird:
		return TicketTypeEarlyBirdDisplayName
	case TicketTypeVIP:
		return TicketTypeVIPDisplayName
	case TicketTypeGroup:
		return TicketTypeGroupDisplayName
	default:
		return string(t)
	}
}

// GetDescription returns the description
func (t TicketType) GetDescription() string {
	switch t {
	case TicketTypeGeneral:
		return TicketTypeGeneralDescription
	case TicketTypeEarlyBird:
		return TicketTypeEarlyBirdDescription
	case TicketTypeVIP:
		return TicketTypeVIPDescription
	case TicketTypeGroup:
		return TicketTypeGroupDescription
	default:
		return ""
	}
}

// GetSortOrder returns the sort order for this ticket type
func (t TicketType) GetSortOrder() int {
	switch t {
	case TicketTypeGeneral:
		return TicketTypeGeneralSortOrder
	case TicketTypeEarlyBird:
		return TicketTypeEarlyBirdSortOrder
	case TicketTypeVIP:
		return TicketTypeVIPSortOrder
	case TicketTypeGroup:
		return TicketTypeGroupSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For TicketType
// ============================================================

// ParseTicketType parses a string into a TicketType
// Expects the name (with underscores), not the slug
func ParseTicketType(name string) (TicketType, bool) {
	t := TicketType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// ParseTicketTypeWithDefault parses a string or returns a default
func ParseTicketTypeWithDefault(name string, defaultType TicketType) TicketType {
	if t, ok := ParseTicketType(name); ok {
		return t
	}
	return defaultType
}

// ParseTicketTypeBySlug parses a slug string into a TicketType
func ParseTicketTypeBySlug(slug string) (TicketType, bool) {
	for _, t := range AllTicketTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllTicketTypeNames returns all internal ticket type names (with underscores)
func AllTicketTypeNames() []string {
	names := make([]string, 0, len(AllTicketTypes))
	for _, t := range AllTicketTypes {
		names = append(names, t.GetName())
	}
	return names
}

// AllTicketTypeSlugs returns all ticket type slugs (with hyphens)
func AllTicketTypeSlugs() []string {
	slugs := make([]string, 0, len(AllTicketTypes))
	for _, t := range AllTicketTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

// AllTicketTypeDisplayNames returns all ticket type display names
func AllTicketTypeDisplayNames() []string {
	names := make([]string, 0, len(AllTicketTypes))
	for _, t := range AllTicketTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}