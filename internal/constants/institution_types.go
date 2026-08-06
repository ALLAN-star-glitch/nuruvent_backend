// internal/constants/institution_types.go

package constants

// InstitutionTypeInfo holds information about an institution type
type InstitutionTypeInfo struct {
	Slug            string // Database value (e.g., "training_institute")
	Name            string // Canonical name (e.g., "Training Institute")
	DisplayName     string // UI display (e.g., "🏫 Training Institute")
	Description     string
	Icon            string
	Color           string
	SortOrder       int
	MetaTitle       string
	MetaDescription string
}

// InstitutionTypes - the source of truth for institution types
var InstitutionTypes = []InstitutionTypeInfo{
	{
		Slug:            "training_institute",
		Name:            "Training Institute",
		DisplayName:     "🏫 Training Institute",
		Description:     "Private training providers, bootcamps, and skill development centers",
		Icon:            "graduation-cap",
		Color:           "#4F46E5", // Indigo
		SortOrder:       1,
		MetaTitle:       "Training Institutes in Kenya | Nuruvent",
		MetaDescription: "Find and book training institutes in Kenya. Professional development and skill-building courses.",
	},
	{
		Slug:            "college",
		Name:            "College / University",
		DisplayName:     "🎓 College / University",
		Description:     "Higher education institutions offering professional courses and degrees",
		Icon:            "university",
		Color:           "#7C3AED", // Purple
		SortOrder:       2,
		MetaTitle:       "Colleges & Universities in Kenya | Nuruvent",
		MetaDescription: "Discover colleges and universities in Kenya offering professional courses and training programs.",
	},
	{
		Slug:            "professional_body",
		Name:            "Professional Body",
		DisplayName:     "💼 Professional Body",
		Description:     "Professional associations offering CPD and certification events",
		Icon:            "briefcase",
		Color:           "#059669", // Emerald
		SortOrder:       3,
		MetaTitle:       "Professional Bodies in Kenya | Nuruvent",
		MetaDescription: "Connect with professional bodies in Kenya. CPD events, certifications, and networking.",
	},
	{
		Slug:            "ngo",
		Name:            "NGO / Non-Profit",
		DisplayName:     "❤️ NGO / Non-Profit",
		Description:     "Non-governmental organizations conducting training and community events",
		Icon:            "heart",
		Color:           "#DC2626", // Red
		SortOrder:       4,
		MetaTitle:       "NGOs & Non-Profits in Kenya | Nuruvent",
		MetaDescription: "Discover NGOs and non-profit organizations in Kenya offering training and community programs.",
	},
	{
		Slug:            "corporate",
		Name:            "Corporate",
		DisplayName:     "🏢 Corporate",
		Description:     "Corporate training departments and HR teams",
		Icon:            "building",
		Color:           "#2563EB", // Blue
		SortOrder:       5,
		MetaTitle:       "Corporate Training in Kenya | Nuruvent",
		MetaDescription: "Find corporate training providers in Kenya. Professional development and team building.",
	},
	{
		Slug:            "government",
		Name:            "Government Agency",
		DisplayName:     "🏛️ Government Agency",
		Description:     "Government departments and agencies offering training",
		Icon:            "landmark",
		Color:           "#D97706", // Amber
		SortOrder:       6,
		MetaTitle:       "Government Agencies in Kenya | Nuruvent",
		MetaDescription: "Discover government agencies in Kenya offering training and professional development.",
	},
}

// InstitutionTypeMap for quick lookups
var InstitutionTypeMap = map[string]InstitutionTypeInfo{
	"training_institute": InstitutionTypes[0],
	"college":            InstitutionTypes[1],
	"professional_body":  InstitutionTypes[2],
	"ngo":                InstitutionTypes[3],
	"corporate":          InstitutionTypes[4],
	"government":         InstitutionTypes[5],
}

// AllInstitutionTypeSlugs returns all valid institution type slugs
func AllInstitutionTypeSlugs() []string {
	values := make([]string, len(InstitutionTypes))
	for i, it := range InstitutionTypes {
		values[i] = it.Slug
	}
	return values
}

// GetInstitutionTypeInfo returns InstitutionTypeInfo by slug
func GetInstitutionTypeInfo(slug string) (InstitutionTypeInfo, bool) {
	info, ok := InstitutionTypeMap[slug]
	return info, ok
}

// IsValidInstitutionType checks if an institution type slug is valid
func IsValidInstitutionType(slug string) bool {
	_, ok := InstitutionTypeMap[slug]
	return ok
}