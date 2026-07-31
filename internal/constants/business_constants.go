package constants

// BusinessTypeInfo holds information about a business type
type BusinessTypeInfo struct {
	Name        string
	DisplayName string
	Description string
	Icon        string
	SortOrder   int
	Category    string // "organization" or "individual"
}

// BusinessTypes - the source of truth for business types
var BusinessTypes = []BusinessTypeInfo{
	// ============================================
	// ORGANIZATION TYPES
	// ============================================
	{
		Name:        "training_institute",
		DisplayName: "Training Institute",
		Description: "Private training providers, bootcamps, and skill development centers",
		Icon:        "school",
		SortOrder:   1,
		Category:    "organization",
	},
	{
		Name:        "college",
		DisplayName: "College/University",
		Description: "Higher education institutions offering professional courses",
		Icon:        "university",
		SortOrder:   2,
		Category:    "organization",
	},
	{
		Name:        "professional_body",
		DisplayName: "Professional Body",
		Description: "Professional associations offering CPD and certification events",
		Icon:        "briefcase",
		SortOrder:   3,
		Category:    "organization",
	},
	{
		Name:        "ngo",
		DisplayName: "NGO / Non-Profit",
		Description: "Non-governmental organizations conducting training and community events",
		Icon:        "heart",
		SortOrder:   4,
		Category:    "organization",
	},
	{
		Name:        "corporate",
		DisplayName: "Corporate",
		Description: "Corporate training departments and HR teams",
		Icon:        "building",
		SortOrder:   5,
		Category:    "organization",
	},
	{
		Name:        "government",
		DisplayName: "Government Agency",
		Description: "Government departments and agencies offering training",
		Icon:        "landmark",
		SortOrder:   6,
		Category:    "organization",
	},

	// ============================================
	// INDIVIDUAL TYPES
	// ============================================
	{
		Name:        "individual_formal",
		DisplayName: "Individual Professional (Registered)",
		Description: "Registered individual trainer, coach, or consultant with a business name",
		Icon:        "user-check",
		SortOrder:   7,
		Category:    "individual",
	},
	{
		Name:        "individual_informal",
		DisplayName: "Individual Professional (Freelance)",
		Description: "Freelance or informal trainer, coach, or mentor",
		Icon:        "user",
		SortOrder:   8,
		Category:    "individual",
	},
}

// BusinessTypeMap for quick lookups
var BusinessTypeMap = map[string]BusinessTypeInfo{
	"training_institute":   BusinessTypes[0],
	"college":              BusinessTypes[1],
	"professional_body":    BusinessTypes[2],
	"ngo":                  BusinessTypes[3],
	"corporate":            BusinessTypes[4],
	"government":           BusinessTypes[5],
	"individual_formal":    BusinessTypes[6],
	"individual_informal":  BusinessTypes[7],
}

// AllBusinessTypeValues returns all valid business type values
func AllBusinessTypeValues() []string {
	values := make([]string, len(BusinessTypes))
	for i, bt := range BusinessTypes {
		values[i] = bt.Name
	}
	return values
}

// GetBusinessTypeInfo returns BusinessTypeInfo by name
func GetBusinessTypeInfo(name string) (BusinessTypeInfo, bool) {
	info, ok := BusinessTypeMap[name]
	return info, ok
}

// IsValidBusinessType checks if a business type is valid
func IsValidBusinessType(name string) bool {
	_, ok := BusinessTypeMap[name]
	return ok
}

// ================================================
// HELPER FUNCTIONS
// ================================================

// IsOrganizationType checks if a business type is an organization
func IsOrganizationType(name string) bool {
	info, ok := BusinessTypeMap[name]
	if !ok {
		return false
	}
	return info.Category == "organization"
}

// IsIndividualType checks if a business type is an individual
func IsIndividualType(name string) bool {
	info, ok := BusinessTypeMap[name]
	if !ok {
		return false
	}
	return info.Category == "individual"
}

// IsFormalIndividual checks if a business type is a formal individual
func IsFormalIndividual(name string) bool {
	return name == "individual_formal"
}

// IsInformalIndividual checks if a business type is an informal individual
func IsInformalIndividual(name string) bool {
	return name == "individual_informal"
}

// GetOrganizationTypes returns all organization business types
func GetOrganizationTypes() []BusinessTypeInfo {
	var orgTypes []BusinessTypeInfo
	for _, bt := range BusinessTypes {
		if bt.Category == "organization" {
			orgTypes = append(orgTypes, bt)
		}
	}
	return orgTypes
}

// GetIndividualTypes returns all individual business types
func GetIndividualTypes() []BusinessTypeInfo {
	var indTypes []BusinessTypeInfo
	for _, bt := range BusinessTypes {
		if bt.Category == "individual" {
			indTypes = append(indTypes, bt)
		}
	}
	return indTypes
}