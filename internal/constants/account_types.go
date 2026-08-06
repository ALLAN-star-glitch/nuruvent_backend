// internal/constants/account_types.go

package constants

// AccountTypeInfo holds information about an account type
type AccountTypeInfo struct {
	Slug        string // Database value (e.g., "personal")
	Name        string // Canonical name (e.g., "Personal Account")
	DisplayName string // UI display (e.g., "👤 Personal Account")
	Description string
	Icon        string
	Color       string
	SortOrder   int
}

// AccountTypes - the source of truth for account types
var AccountTypes = []AccountTypeInfo{
	{
		Slug:        "personal",
		Name:        "Personal Account",
		DisplayName: "👤 Personal Account",
		Description: "Individual person (trainer, coach, consultant, freelancer)",
		Icon:        "user",
		Color:       "#4F46E5", // Indigo
		SortOrder:   1,
	},
	{
		Slug:        "institution",
		Name:        "Institution Account",
		DisplayName: "🏛️ Institution Account",
		Description: "Organization, company, institute, or association",
		Icon:        "building",
		Color:       "#7C3AED", // Purple
		SortOrder:   2,
	},
}

// AccountTypeMap for quick lookups
var AccountTypeMap = map[string]AccountTypeInfo{
	"personal":    AccountTypes[0],
	"institution": AccountTypes[1],
}

// AllAccountTypeSlugs returns all valid account type slugs
func AllAccountTypeSlugs() []string {
	values := make([]string, len(AccountTypes))
	for i, at := range AccountTypes {
		values[i] = at.Slug
	}
	return values
}

// GetAccountTypeInfo returns AccountTypeInfo by slug
func GetAccountTypeInfo(slug string) (AccountTypeInfo, bool) {
	info, ok := AccountTypeMap[slug]
	return info, ok
}

// IsValidAccountType checks if an account type slug is valid
func IsValidAccountType(slug string) bool {
	_, ok := AccountTypeMap[slug]
	return ok
}

// IsPersonalType checks if the account type is personal
func IsPersonalType(slug string) bool {
	return slug == "personal"
}

// IsInstitutionType checks if the account type is institution
func IsInstitutionType(slug string) bool {
	return slug == "institution"
}