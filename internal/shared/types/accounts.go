// internal/shared/types/accounts.go

package types

// ============================================================
// ACCOUNT TYPE CONSTANTS
// ============================================================

// AccountType represents the type of account
type AccountType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	AccountTypePersonalSlug    = "account-type-personal"
	AccountTypeInstitutionSlug = "account-type-institution"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	AccountTypePersonalName    = "account_type_personal"
	AccountTypeInstitutionName = "account_type_institution"
)

// Display name constants - Used for UI display
const (
	AccountTypePersonalDisplayName    = "Personal Account"
	AccountTypeInstitutionDisplayName = "Institution Account"
)

// Description constants
const (
	AccountTypePersonalDescription    = "Individual person (trainer, coach, consultant, freelancer)"
	AccountTypeInstitutionDescription = "Organization, company, institute, or association"
)

// Icon constants
const (
	AccountTypePersonalIcon    = "user"
	AccountTypeInstitutionIcon = "building"
)

// Color constants
const (
	AccountTypePersonalColor    = "#4F46E5" // Indigo
	AccountTypeInstitutionColor = "#7C3AED" // Purple
)

// Sort order constants
const (
	AccountTypePersonalSortOrder    = 1
	AccountTypeInstitutionSortOrder = 2
)

// ============================================================
// ACCOUNT TYPE DEFINITIONS
// ============================================================

// AccountType constants - These are the NAMES (with underscores)
const (
	AccountTypePersonal    AccountType = AccountTypePersonalName    // "account_type_personal"
	AccountTypeInstitution AccountType = AccountTypeInstitutionName // "account_type_institution"
)

// AllAccountTypes lists all valid account types for validation
var AllAccountTypes = []AccountType{
	AccountTypePersonal,
	AccountTypeInstitution,
}

// ============================================================
// BASIC METHODS - On AccountType
// ============================================================

// String returns the string representation (the name with underscores)
func (a AccountType) String() string {
	return string(a)
}

// IsValid checks if the account type is valid
func (a AccountType) IsValid() bool {
	for _, t := range AllAccountTypes {
		if t == a {
			return true
		}
	}
	return false
}

// IsPersonal checks if the account type is personal
func (a AccountType) IsPersonal() bool {
	return a == AccountTypePersonal
}

// IsInstitution checks if the account type is institution
func (a AccountType) IsInstitution() bool {
	return a == AccountTypeInstitution
}

// ============================================================
// GETTER METHODS - On AccountType
// ============================================================

// GetName returns the internal name (with underscores)
func (a AccountType) GetName() string {
	return string(a)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (a AccountType) GetSlug() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalSlug
	case AccountTypeInstitution:
		return AccountTypeInstitutionSlug
	default:
		return string(a)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (a AccountType) GetDisplayName() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalDisplayName
	case AccountTypeInstitution:
		return AccountTypeInstitutionDisplayName
	default:
		return string(a)
	}
}

// GetDescription returns the description
func (a AccountType) GetDescription() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalDescription
	case AccountTypeInstitution:
		return AccountTypeInstitutionDescription
	default:
		return ""
	}
}

// GetIcon returns the icon
func (a AccountType) GetIcon() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalIcon
	case AccountTypeInstitution:
		return AccountTypeInstitutionIcon
	default:
		return "user"
	}
}

// GetColor returns the color
func (a AccountType) GetColor() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalColor
	case AccountTypeInstitution:
		return AccountTypeInstitutionColor
	default:
		return "#6B7280"
	}
}

// GetSortOrder returns the sort order
func (a AccountType) GetSortOrder() int {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalSortOrder
	case AccountTypeInstitution:
		return AccountTypeInstitutionSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For AccountType
// ============================================================

// ParseAccountType parses a string into an AccountType
// Expects the name (with underscores), not the slug
func ParseAccountType(name string) (AccountType, bool) {
	t := AccountType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// ParseAccountTypeWithDefault parses a string or returns a default
func ParseAccountTypeWithDefault(name string, defaultType AccountType) AccountType {
	if t, ok := ParseAccountType(name); ok {
		return t
	}
	return defaultType
}

// ParseAccountTypeBySlug parses a slug string into an AccountType
func ParseAccountTypeBySlug(slug string) (AccountType, bool) {
	for _, t := range AllAccountTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllAccountTypeNames returns all internal type names (with underscores)
func AllAccountTypeNames() []string {
	names := make([]string, 0, len(AllAccountTypes))
	for _, t := range AllAccountTypes {
		names = append(names, t.GetName())
	}
	return names
}

// AllAccountTypeSlugs returns all type slugs (with hyphens)
func AllAccountTypeSlugs() []string {
	slugs := make([]string, 0, len(AllAccountTypes))
	for _, t := range AllAccountTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

// AllAccountTypeDisplayNames returns all display names
func AllAccountTypeDisplayNames() []string {
	names := make([]string, 0, len(AllAccountTypes))
	for _, t := range AllAccountTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}

// ============================================================
// INSTITUTION TYPE CONSTANTS
// ============================================================

// InstitutionType represents the type of institution
type InstitutionType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	InstitutionTypeCompanySlug     = "institution-type-company"
	InstitutionTypeInstituteSlug   = "institution-type-institute"
	InstitutionTypeAssociationSlug = "institution-type-association"
	InstitutionTypeSchoolSlug      = "institution-type-school"
	InstitutionTypeUniversitySlug  = "institution-type-university"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	InstitutionTypeCompanyName     = "institution_type_company"
	InstitutionTypeInstituteName   = "institution_type_institute"
	InstitutionTypeAssociationName = "institution_type_association"
	InstitutionTypeSchoolName      = "institution_type_school"
	InstitutionTypeUniversityName  = "institution_type_university"
)

// Display name constants - Used for UI display
const (
	InstitutionTypeCompanyDisplayName     = "Company"
	InstitutionTypeInstituteDisplayName   = "Institute"
	InstitutionTypeAssociationDisplayName = "Association"
	InstitutionTypeSchoolDisplayName      = "School"
	InstitutionTypeUniversityDisplayName  = "University"
)

// Description constants
const (
	InstitutionTypeCompanyDescription     = "Corporate organization"
	InstitutionTypeInstituteDescription   = "Educational or research institute"
	InstitutionTypeAssociationDescription = "Professional or trade association"
	InstitutionTypeSchoolDescription      = "Primary or secondary educational institution"
	InstitutionTypeUniversityDescription  = "Higher education institution"
)

// Icon constants
const (
	InstitutionTypeCompanyIcon     = "building"
	InstitutionTypeInstituteIcon   = "landmark"
	InstitutionTypeAssociationIcon = "users"
	InstitutionTypeSchoolIcon      = "school"
	InstitutionTypeUniversityIcon  = "graduation-cap"
)

// Color constants
const (
	InstitutionTypeCompanyColor     = "#7C3AED" // Purple
	InstitutionTypeInstituteColor   = "#4F46E5" // Indigo
	InstitutionTypeAssociationColor = "#0EA5E9" // Sky Blue
	InstitutionTypeSchoolColor      = "#F59E0B" // Amber
	InstitutionTypeUniversityColor  = "#10B981" // Emerald
)

// Sort order constants
const (
	InstitutionTypeCompanySortOrder     = 1
	InstitutionTypeInstituteSortOrder   = 2
	InstitutionTypeAssociationSortOrder = 3
	InstitutionTypeSchoolSortOrder      = 4
	InstitutionTypeUniversitySortOrder  = 5
)

// ============================================================
// INSTITUTION TYPE DEFINITIONS
// ============================================================

// InstitutionType constants - These are the NAMES (with underscores)
const (
	InstitutionTypeCompany     InstitutionType = InstitutionTypeCompanyName     // "institution_type_company"
	InstitutionTypeInstitute   InstitutionType = InstitutionTypeInstituteName   // "institution_type_institute"
	InstitutionTypeAssociation InstitutionType = InstitutionTypeAssociationName // "institution_type_association"
	InstitutionTypeSchool      InstitutionType = InstitutionTypeSchoolName      // "institution_type_school"
	InstitutionTypeUniversity  InstitutionType = InstitutionTypeUniversityName  // "institution_type_university"
)

// AllInstitutionTypes lists all valid institution types for validation
var AllInstitutionTypes = []InstitutionType{
	InstitutionTypeCompany,
	InstitutionTypeInstitute,
	InstitutionTypeAssociation,
	InstitutionTypeSchool,
	InstitutionTypeUniversity,
}

// ============================================================
// BASIC METHODS - On InstitutionType
// ============================================================

// String returns the string representation (the name with underscores)
func (i InstitutionType) String() string {
	return string(i)
}

// IsValid checks if the institution type is valid
func (i InstitutionType) IsValid() bool {
	for _, t := range AllInstitutionTypes {
		if t == i {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On InstitutionType
// ============================================================

// GetName returns the internal name (with underscores)
func (i InstitutionType) GetName() string {
	return string(i)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (i InstitutionType) GetSlug() string {
	switch i {
	case InstitutionTypeCompany:
		return InstitutionTypeCompanySlug
	case InstitutionTypeInstitute:
		return InstitutionTypeInstituteSlug
	case InstitutionTypeAssociation:
		return InstitutionTypeAssociationSlug
	case InstitutionTypeSchool:
		return InstitutionTypeSchoolSlug
	case InstitutionTypeUniversity:
		return InstitutionTypeUniversitySlug
	default:
		return string(i)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (i InstitutionType) GetDisplayName() string {
	switch i {
	case InstitutionTypeCompany:
		return InstitutionTypeCompanyDisplayName
	case InstitutionTypeInstitute:
		return InstitutionTypeInstituteDisplayName
	case InstitutionTypeAssociation:
		return InstitutionTypeAssociationDisplayName
	case InstitutionTypeSchool:
		return InstitutionTypeSchoolDisplayName
	case InstitutionTypeUniversity:
		return InstitutionTypeUniversityDisplayName
	default:
		return string(i)
	}
}

// GetDescription returns the description
func (i InstitutionType) GetDescription() string {
	switch i {
	case InstitutionTypeCompany:
		return InstitutionTypeCompanyDescription
	case InstitutionTypeInstitute:
		return InstitutionTypeInstituteDescription
	case InstitutionTypeAssociation:
		return InstitutionTypeAssociationDescription
	case InstitutionTypeSchool:
		return InstitutionTypeSchoolDescription
	case InstitutionTypeUniversity:
		return InstitutionTypeUniversityDescription
	default:
		return ""
	}
}

// GetIcon returns the icon
func (i InstitutionType) GetIcon() string {
	switch i {
	case InstitutionTypeCompany:
		return InstitutionTypeCompanyIcon
	case InstitutionTypeInstitute:
		return InstitutionTypeInstituteIcon
	case InstitutionTypeAssociation:
		return InstitutionTypeAssociationIcon
	case InstitutionTypeSchool:
		return InstitutionTypeSchoolIcon
	case InstitutionTypeUniversity:
		return InstitutionTypeUniversityIcon
	default:
		return "building"
	}
}

// GetColor returns the color
func (i InstitutionType) GetColor() string {
	switch i {
	case InstitutionTypeCompany:
		return InstitutionTypeCompanyColor
	case InstitutionTypeInstitute:
		return InstitutionTypeInstituteColor
	case InstitutionTypeAssociation:
		return InstitutionTypeAssociationColor
	case InstitutionTypeSchool:
		return InstitutionTypeSchoolColor
	case InstitutionTypeUniversity:
		return InstitutionTypeUniversityColor
	default:
		return "#6B7280"
	}
}

// GetSortOrder returns the sort order
func (i InstitutionType) GetSortOrder() int {
	switch i {
	case InstitutionTypeCompany:
		return InstitutionTypeCompanySortOrder
	case InstitutionTypeInstitute:
		return InstitutionTypeInstituteSortOrder
	case InstitutionTypeAssociation:
		return InstitutionTypeAssociationSortOrder
	case InstitutionTypeSchool:
		return InstitutionTypeSchoolSortOrder
	case InstitutionTypeUniversity:
		return InstitutionTypeUniversitySortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For InstitutionType
// ============================================================

// ParseInstitutionType parses a string into an InstitutionType
// Expects the name (with underscores), not the slug
func ParseInstitutionType(name string) (InstitutionType, bool) {
	t := InstitutionType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// ParseInstitutionTypeWithDefault parses a string or returns a default
func ParseInstitutionTypeWithDefault(name string, defaultType InstitutionType) InstitutionType {
	if t, ok := ParseInstitutionType(name); ok {
		return t
	}
	return defaultType
}

// ParseInstitutionTypeBySlug parses a slug string into an InstitutionType
func ParseInstitutionTypeBySlug(slug string) (InstitutionType, bool) {
	for _, t := range AllInstitutionTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllInstitutionTypeNames returns all internal type names (with underscores)
func AllInstitutionTypeNames() []string {
	names := make([]string, 0, len(AllInstitutionTypes))
	for _, t := range AllInstitutionTypes {
		names = append(names, t.GetName())
	}
	return names
}

// AllInstitutionTypeSlugs returns all type slugs (with hyphens)
func AllInstitutionTypeSlugs() []string {
	slugs := make([]string, 0, len(AllInstitutionTypes))
	for _, t := range AllInstitutionTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

// AllInstitutionTypeDisplayNames returns all display names
func AllInstitutionTypeDisplayNames() []string {
	names := make([]string, 0, len(AllInstitutionTypes))
	for _, t := range AllInstitutionTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}

// ============================================================
// PROFESSIONAL TYPE CONSTANTS
// ============================================================

// ProfessionalType represents the type of professional
type ProfessionalType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	ProfessionalTypeTrainerSlug    = "professional-type-trainer"
	ProfessionalTypeCoachSlug      = "professional-type-coach"
	ProfessionalTypeConsultantSlug = "professional-type-consultant"
	ProfessionalTypeFreelancerSlug = "professional-type-freelancer"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	ProfessionalTypeTrainerName    = "professional_type_trainer"
	ProfessionalTypeCoachName      = "professional_type_coach"
	ProfessionalTypeConsultantName = "professional_type_consultant"
	ProfessionalTypeFreelancerName = "professional_type_freelancer"
)

// Display name constants - Used for UI display
const (
	ProfessionalTypeTrainerDisplayName    = "Trainer"
	ProfessionalTypeCoachDisplayName      = "Coach"
	ProfessionalTypeConsultantDisplayName = "Consultant"
	ProfessionalTypeFreelancerDisplayName = "Freelancer"
)

// Description constants
const (
	ProfessionalTypeTrainerDescription    = "Professional trainer who conducts training sessions"
	ProfessionalTypeCoachDescription      = "Professional coach providing guidance and mentorship"
	ProfessionalTypeConsultantDescription = "Professional consultant providing expert advice"
	ProfessionalTypeFreelancerDescription = "Independent professional offering services"
)

// Icon constants
const (
	ProfessionalTypeTrainerIcon    = "graduation-cap"
	ProfessionalTypeCoachIcon      = "user-tie"
	ProfessionalTypeConsultantIcon = "briefcase"
	ProfessionalTypeFreelancerIcon = "laptop"
)

// Color constants
const (
	ProfessionalTypeTrainerColor    = "#4F46E5" // Indigo
	ProfessionalTypeCoachColor      = "#7C3AED" // Purple
	ProfessionalTypeConsultantColor = "#0EA5E9" // Sky Blue
	ProfessionalTypeFreelancerColor = "#F59E0B" // Amber
)

// Sort order constants
const (
	ProfessionalTypeTrainerSortOrder    = 1
	ProfessionalTypeCoachSortOrder      = 2
	ProfessionalTypeConsultantSortOrder = 3
	ProfessionalTypeFreelancerSortOrder = 4
)

// ============================================================
// PROFESSIONAL TYPE DEFINITIONS
// ============================================================

// ProfessionalType constants - These are the NAMES (with underscores)
const (
	ProfessionalTypeTrainer    ProfessionalType = ProfessionalTypeTrainerName    // "professional_type_trainer"
	ProfessionalTypeCoach      ProfessionalType = ProfessionalTypeCoachName      // "professional_type_coach"
	ProfessionalTypeConsultant ProfessionalType = ProfessionalTypeConsultantName // "professional_type_consultant"
	ProfessionalTypeFreelancer ProfessionalType = ProfessionalTypeFreelancerName // "professional_type_freelancer"
)

// AllProfessionalTypes lists all valid professional types for validation
var AllProfessionalTypes = []ProfessionalType{
	ProfessionalTypeTrainer,
	ProfessionalTypeCoach,
	ProfessionalTypeConsultant,
	ProfessionalTypeFreelancer,
}

// ============================================================
// BASIC METHODS - On ProfessionalType
// ============================================================

// String returns the string representation (the name with underscores)
func (p ProfessionalType) String() string {
	return string(p)
}

// IsValid checks if the professional type is valid
func (p ProfessionalType) IsValid() bool {
	for _, t := range AllProfessionalTypes {
		if t == p {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On ProfessionalType
// ============================================================

// GetName returns the internal name (with underscores)
func (p ProfessionalType) GetName() string {
	return string(p)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (p ProfessionalType) GetSlug() string {
	switch p {
	case ProfessionalTypeTrainer:
		return ProfessionalTypeTrainerSlug
	case ProfessionalTypeCoach:
		return ProfessionalTypeCoachSlug
	case ProfessionalTypeConsultant:
		return ProfessionalTypeConsultantSlug
	case ProfessionalTypeFreelancer:
		return ProfessionalTypeFreelancerSlug
	default:
		return string(p)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (p ProfessionalType) GetDisplayName() string {
	switch p {
	case ProfessionalTypeTrainer:
		return ProfessionalTypeTrainerDisplayName
	case ProfessionalTypeCoach:
		return ProfessionalTypeCoachDisplayName
	case ProfessionalTypeConsultant:
		return ProfessionalTypeConsultantDisplayName
	case ProfessionalTypeFreelancer:
		return ProfessionalTypeFreelancerDisplayName
	default:
		return string(p)
	}
}

// GetDescription returns the description
func (p ProfessionalType) GetDescription() string {
	switch p {
	case ProfessionalTypeTrainer:
		return ProfessionalTypeTrainerDescription
	case ProfessionalTypeCoach:
		return ProfessionalTypeCoachDescription
	case ProfessionalTypeConsultant:
		return ProfessionalTypeConsultantDescription
	case ProfessionalTypeFreelancer:
		return ProfessionalTypeFreelancerDescription
	default:
		return ""
	}
}

// GetIcon returns the icon
func (p ProfessionalType) GetIcon() string {
	switch p {
	case ProfessionalTypeTrainer:
		return ProfessionalTypeTrainerIcon
	case ProfessionalTypeCoach:
		return ProfessionalTypeCoachIcon
	case ProfessionalTypeConsultant:
		return ProfessionalTypeConsultantIcon
	case ProfessionalTypeFreelancer:
		return ProfessionalTypeFreelancerIcon
	default:
		return "user"
	}
}

// GetColor returns the color
func (p ProfessionalType) GetColor() string {
	switch p {
	case ProfessionalTypeTrainer:
		return ProfessionalTypeTrainerColor
	case ProfessionalTypeCoach:
		return ProfessionalTypeCoachColor
	case ProfessionalTypeConsultant:
		return ProfessionalTypeConsultantColor
	case ProfessionalTypeFreelancer:
		return ProfessionalTypeFreelancerColor
	default:
		return "#6B7280"
	}
}

// GetSortOrder returns the sort order
func (p ProfessionalType) GetSortOrder() int {
	switch p {
	case ProfessionalTypeTrainer:
		return ProfessionalTypeTrainerSortOrder
	case ProfessionalTypeCoach:
		return ProfessionalTypeCoachSortOrder
	case ProfessionalTypeConsultant:
		return ProfessionalTypeConsultantSortOrder
	case ProfessionalTypeFreelancer:
		return ProfessionalTypeFreelancerSortOrder
	default:
		return 999
	}
}

// CanHost returns whether this professional type can host events
func (p ProfessionalType) CanHost() bool {
	switch p {
	case ProfessionalTypeTrainer, ProfessionalTypeCoach, ProfessionalTypeConsultant:
		return true
	case ProfessionalTypeFreelancer:
		return false
	default:
		return false
	}
}

// ============================================================
// PARSE FUNCTIONS - For ProfessionalType
// ============================================================

// ParseProfessionalType parses a string into a ProfessionalType
// Expects the name (with underscores), not the slug
func ParseProfessionalType(name string) (ProfessionalType, bool) {
	t := ProfessionalType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

// ParseProfessionalTypeWithDefault parses a string or returns a default
func ParseProfessionalTypeWithDefault(name string, defaultType ProfessionalType) ProfessionalType {
	if t, ok := ParseProfessionalType(name); ok {
		return t
	}
	return defaultType
}

// ParseProfessionalTypeBySlug parses a slug string into a ProfessionalType
func ParseProfessionalTypeBySlug(slug string) (ProfessionalType, bool) {
	for _, t := range AllProfessionalTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllProfessionalTypeNames returns all internal type names (with underscores)
func AllProfessionalTypeNames() []string {
	names := make([]string, 0, len(AllProfessionalTypes))
	for _, t := range AllProfessionalTypes {
		names = append(names, t.GetName())
	}
	return names
}

// AllProfessionalTypeSlugs returns all type slugs (with hyphens)
func AllProfessionalTypeSlugs() []string {
	slugs := make([]string, 0, len(AllProfessionalTypes))
	for _, t := range AllProfessionalTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

// AllProfessionalTypeDisplayNames returns all display names
func AllProfessionalTypeDisplayNames() []string {
	names := make([]string, 0, len(AllProfessionalTypes))
	for _, t := range AllProfessionalTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}

// HostProfessionalTypes returns only professional types that can host events
func HostProfessionalTypes() []ProfessionalType {
	types := make([]ProfessionalType, 0)
	for _, t := range AllProfessionalTypes {
		if t.CanHost() {
			types = append(types, t)
		}
	}
	return types
}