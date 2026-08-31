// internal/shared/types/category.go

package types

// ============================================================
// CATEGORY CONSTANTS
// ============================================================

// CategorySlug constants (with hyphens) - Used for URLs and API routes
const (
	CategoryTechnologySlug = "technology"
	CategoryBusinessSlug   = "business"
	CategoryHealthSlug     = "health"
	CategoryEducationSlug  = "education"
	CategoryFinanceSlug    = "finance"
	CategoryScienceSlug    = "science"
	CategoryArtsSlug       = "arts"
	CategorySportsSlug     = "sports"
)

// CategoryName constants (with underscores) - Used for internal database lookups
const (
	CategoryTechnologyName = "category_technology"
	CategoryBusinessName   = "category_business"
	CategoryHealthName     = "category_health"
	CategoryEducationName  = "category_education"
	CategoryFinanceName    = "category_finance"
	CategoryScienceName    = "category_science"
	CategoryArtsName       = "category_arts"
	CategorySportsName     = "category_sports"
)

// CategoryDisplayName constants - Used for UI display
const (
	CategoryTechnologyDisplayName = "Technology"
	CategoryBusinessDisplayName   = "Business"
	CategoryHealthDisplayName     = "Health & Wellness"
	CategoryEducationDisplayName  = "Education"
	CategoryFinanceDisplayName    = "Finance"
	CategoryScienceDisplayName    = "Science"
	CategoryArtsDisplayName       = "Arts & Culture"
	CategorySportsDisplayName     = "Sports & Fitness"
)

// CategoryDescription constants
const (
	CategoryTechnologyDescription = "Technology and software related events"
	CategoryBusinessDescription   = "Business, management and leadership events"
	CategoryHealthDescription     = "Health, wellness and fitness events"
	CategoryEducationDescription  = "Education, teaching and learning events"
	CategoryFinanceDescription    = "Finance, accounting and investment events"
	CategoryScienceDescription    = "Science and research events"
	CategoryArtsDescription       = "Arts, culture and creative events"
	CategorySportsDescription     = "Sports and fitness events"
)

// CategoryIcon constants
const (
	CategoryTechnologyIcon = "cpu"
	CategoryBusinessIcon   = "briefcase"
	CategoryHealthIcon     = "heart-pulse"
	CategoryEducationIcon  = "graduation-cap"
	CategoryFinanceIcon    = "landmark"
	CategoryScienceIcon    = "flask"
	CategoryArtsIcon       = "palette"
	CategorySportsIcon     = "dumbbell"
)

// CategoryColor constants
const (
	CategoryTechnologyColor = "#3B82F6" // Blue
	CategoryBusinessColor   = "#8B5CF6" // Purple
	CategoryHealthColor     = "#EF4444" // Red
	CategoryEducationColor  = "#F59E0B" // Amber
	CategoryFinanceColor    = "#10B981" // Green
	CategoryScienceColor    = "#06B6D4" // Cyan
	CategoryArtsColor       = "#EC4899" // Pink
	CategorySportsColor     = "#F97316" // Orange
)

// CategorySortOrder constants
const (
	CategoryTechnologySortOrder = 1
	CategoryBusinessSortOrder   = 2
	CategoryHealthSortOrder     = 3
	CategoryEducationSortOrder  = 4
	CategoryFinanceSortOrder    = 5
	CategoryScienceSortOrder    = 6
	CategoryArtsSortOrder       = 7
	CategorySportsSortOrder     = 8
)

// ============================================================
// CATEGORY DEFINITIONS
// ============================================================

// Category represents a category of events
type Category string

// Category constants - These are the NAMES (with underscores)
const (
	CategoryTechnology Category = CategoryTechnologyName // "category_technology"
	CategoryBusiness   Category = CategoryBusinessName   // "category_business"
	CategoryHealth     Category = CategoryHealthName     // "category_health"
	CategoryEducation  Category = CategoryEducationName  // "category_education"
	CategoryFinance    Category = CategoryFinanceName    // "category_finance"
	CategoryScience    Category = CategoryScienceName    // "category_science"
	CategoryArts       Category = CategoryArtsName       // "category_arts"
	CategorySports     Category = CategorySportsName     // "category_sports"
)

// AllCategories lists all valid categories for validation
var AllCategories = []Category{
	CategoryTechnology,
	CategoryBusiness,
	CategoryHealth,
	CategoryEducation,
	CategoryFinance,
	CategoryScience,
	CategoryArts,
	CategorySports,
}

// ============================================================
// BASIC METHODS - On Category
// ============================================================

// String returns the string representation (the name with underscores)
func (c Category) String() string {
	return string(c)
}

// IsValid checks if the category is valid
func (c Category) IsValid() bool {
	for _, cat := range AllCategories {
		if cat == c {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On Category
// ============================================================

// GetName returns the internal name (with underscores)
func (c Category) GetName() string {
	return string(c)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (c Category) GetSlug() string {
	switch c {
	case CategoryTechnology:
		return CategoryTechnologySlug
	case CategoryBusiness:
		return CategoryBusinessSlug
	case CategoryHealth:
		return CategoryHealthSlug
	case CategoryEducation:
		return CategoryEducationSlug
	case CategoryFinance:
		return CategoryFinanceSlug
	case CategoryScience:
		return CategoryScienceSlug
	case CategoryArts:
		return CategoryArtsSlug
	case CategorySports:
		return CategorySportsSlug
	default:
		return string(c)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (c Category) GetDisplayName() string {
	switch c {
	case CategoryTechnology:
		return CategoryTechnologyDisplayName
	case CategoryBusiness:
		return CategoryBusinessDisplayName
	case CategoryHealth:
		return CategoryHealthDisplayName
	case CategoryEducation:
		return CategoryEducationDisplayName
	case CategoryFinance:
		return CategoryFinanceDisplayName
	case CategoryScience:
		return CategoryScienceDisplayName
	case CategoryArts:
		return CategoryArtsDisplayName
	case CategorySports:
		return CategorySportsDisplayName
	default:
		return string(c)
	}
}

// GetDescription returns the description
func (c Category) GetDescription() string {
	switch c {
	case CategoryTechnology:
		return CategoryTechnologyDescription
	case CategoryBusiness:
		return CategoryBusinessDescription
	case CategoryHealth:
		return CategoryHealthDescription
	case CategoryEducation:
		return CategoryEducationDescription
	case CategoryFinance:
		return CategoryFinanceDescription
	case CategoryScience:
		return CategoryScienceDescription
	case CategoryArts:
		return CategoryArtsDescription
	case CategorySports:
		return CategorySportsDescription
	default:
		return ""
	}
}

// GetIcon returns the icon name for this category
func (c Category) GetIcon() string {
	switch c {
	case CategoryTechnology:
		return CategoryTechnologyIcon
	case CategoryBusiness:
		return CategoryBusinessIcon
	case CategoryHealth:
		return CategoryHealthIcon
	case CategoryEducation:
		return CategoryEducationIcon
	case CategoryFinance:
		return CategoryFinanceIcon
	case CategoryScience:
		return CategoryScienceIcon
	case CategoryArts:
		return CategoryArtsIcon
	case CategorySports:
		return CategorySportsIcon
	default:
		return "folder"
	}
}

// GetColor returns the color for this category
func (c Category) GetColor() string {
	switch c {
	case CategoryTechnology:
		return CategoryTechnologyColor
	case CategoryBusiness:
		return CategoryBusinessColor
	case CategoryHealth:
		return CategoryHealthColor
	case CategoryEducation:
		return CategoryEducationColor
	case CategoryFinance:
		return CategoryFinanceColor
	case CategoryScience:
		return CategoryScienceColor
	case CategoryArts:
		return CategoryArtsColor
	case CategorySports:
		return CategorySportsColor
	default:
		return "#6B7280"
	}
}

// GetSortOrder returns the sort order for this category
func (c Category) GetSortOrder() int {
	switch c {
	case CategoryTechnology:
		return CategoryTechnologySortOrder
	case CategoryBusiness:
		return CategoryBusinessSortOrder
	case CategoryHealth:
		return CategoryHealthSortOrder
	case CategoryEducation:
		return CategoryEducationSortOrder
	case CategoryFinance:
		return CategoryFinanceSortOrder
	case CategoryScience:
		return CategoryScienceSortOrder
	case CategoryArts:
		return CategoryArtsSortOrder
	case CategorySports:
		return CategorySportsSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For Category
// ============================================================

// ParseCategory parses a string into a Category
// Expects the name (with underscores), not the slug
func ParseCategory(name string) (Category, bool) {
	c := Category(name)
	if c.IsValid() {
		return c, true
	}
	return "", false
}

// ParseCategoryWithDefault parses a string or returns a default
func ParseCategoryWithDefault(name string, defaultCategory Category) Category {
	if c, ok := ParseCategory(name); ok {
		return c
	}
	return defaultCategory
}

// ParseCategoryBySlug parses a slug string into a Category
func ParseCategoryBySlug(slug string) (Category, bool) {
	for _, c := range AllCategories {
		if c.GetSlug() == slug {
			return c, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllCategoryNames returns all internal category names (with underscores)
func AllCategoryNames() []string {
	names := make([]string, 0, len(AllCategories))
	for _, c := range AllCategories {
		names = append(names, c.GetName())
	}
	return names
}

// AllCategorySlugs returns all category slugs (with hyphens)
func AllCategorySlugs() []string {
	slugs := make([]string, 0, len(AllCategories))
	for _, c := range AllCategories {
		slugs = append(slugs, c.GetSlug())
	}
	return slugs
}

// AllCategoryDisplayNames returns all category display names
func AllCategoryDisplayNames() []string {
	names := make([]string, 0, len(AllCategories))
	for _, c := range AllCategories {
		names = append(names, c.GetDisplayName())
	}
	return names
}