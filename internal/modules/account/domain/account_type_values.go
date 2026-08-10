package domain

// ============================================================
// ACCOUNT TYPE - Value Object (Source of Truth)
// ============================================================

// AccountTypeValue is a custom typed string for compile-time safety
type AccountTypeValue string

const (
	AccountTypePersonal    AccountTypeValue = "personal"
	AccountTypeInstitution AccountTypeValue = "institution"
)

// AllAccountTypes lists all valid account types
var AllAccountTypes = []AccountTypeValue{
	AccountTypePersonal,
	AccountTypeInstitution,
}

// AccountTypeInfo holds metadata for each account type
type AccountTypeInfo struct {
	Slug        AccountTypeValue
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	IsActive    bool
}

// Private registry (prevents external mutation)
var accountTypeRegistry = map[AccountTypeValue]AccountTypeInfo{
	AccountTypePersonal: {
		Slug:        AccountTypePersonal,
		Name:        "Personal Account",
		DisplayName: "Personal Account",
		Description: "Individual person (trainer, coach, consultant, freelancer)",
		Icon:        "user",
		Color:       "#4F46E5",
		SortOrder:   1,
		IsActive:    true,
	},
	AccountTypeInstitution: {
		Slug:        AccountTypeInstitution,
		Name:        "Institution Account",
		DisplayName: "Institution Account",
		Description: "Organization, company, institute, or association",
		Icon:        "building",
		Color:       "#7C3AED",
		SortOrder:   2,
		IsActive:    true,
	},
}

// ============================================================
// DOMAIN METHODS (on AccountTypeValue)
// ============================================================

func (a AccountTypeValue) String() string {
	return string(a)
}

func (a AccountTypeValue) IsValid() bool {
	_, ok := accountTypeRegistry[a]
	return ok
}

func (a AccountTypeValue) IsPersonal() bool {
	return a == AccountTypePersonal
}

func (a AccountTypeValue) IsInstitution() bool {
	return a == AccountTypeInstitution
}

func (a AccountTypeValue) IsActive() bool {
	info, ok := accountTypeRegistry[a]
	if !ok {
		return false
	}
	return info.IsActive
}

func (a AccountTypeValue) Info() (AccountTypeInfo, bool) {
	info, ok := accountTypeRegistry[a]
	return info, ok
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

func ParseAccountType(slug string) (AccountTypeValue, bool) {
	t := AccountTypeValue(slug)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func AllAccountTypeInfos() []AccountTypeInfo {
	infos := make([]AccountTypeInfo, 0, len(accountTypeRegistry))
	for _, info := range accountTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

func AllAccountTypeSlugs() []string {
	slugs := make([]string, 0, len(accountTypeRegistry))
	for slug := range accountTypeRegistry {
		slugs = append(slugs, string(slug))
	}
	return slugs
}

func ActiveAccountTypeInfos() []AccountTypeInfo {
	infos := make([]AccountTypeInfo, 0)
	for _, info := range accountTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func ActiveAccountTypeSlugs() []string {
	slugs := make([]string, 0)
	for slug, info := range accountTypeRegistry {
		if info.IsActive {
			slugs = append(slugs, string(slug))
		}
	}
	return slugs
}