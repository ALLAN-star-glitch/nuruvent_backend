// internal/shared/types/accounts.go

package types

// ============================================================
// ACCOUNT TYPE CONSTANTS
// ============================================================

type AccountType string

const (
	AccountTypePersonalSlug    = "account-type-personal"
	AccountTypeInstitutionSlug = "account-type-institution"
	AccountTypeInvitedSlug     = "account-type-invited"
)

const (
	AccountTypePersonalName    = "account_type_personal"
	AccountTypeInstitutionName = "account_type_institution"
	AccountTypeInvitedName     = "account_type_invited"
)

const (
	AccountTypePersonalDisplayName    = "Personal Account"
	AccountTypeInstitutionDisplayName = "Institution Account"
	AccountTypeInvitedDisplayName     = "Invited Member"
)

const (
	AccountTypePersonalDescription    = "Individual person (trainer, coach, consultant, freelancer, organizer, learner)"
	AccountTypeInstitutionDescription = "Organization, company, institute, or association"
	AccountTypeInvitedDescription     = "User who joined via a team invitation"
)

const (
	AccountTypePersonalIcon    = "user"
	AccountTypeInstitutionIcon = "building"
	AccountTypeInvitedIcon     = "mail"
)

const (
	AccountTypePersonalColor    = "#4F46E5"
	AccountTypeInstitutionColor = "#7C3AED"
	AccountTypeInvitedColor     = "#6366F1"
)

const (
	AccountTypePersonalSortOrder    = 1
	AccountTypeInstitutionSortOrder = 2
	AccountTypeInvitedSortOrder     = 3
)

const (
	AccountTypePersonal    AccountType = AccountTypePersonalName
	AccountTypeInstitution AccountType = AccountTypeInstitutionName
	AccountTypeInvited     AccountType = AccountTypeInvitedName
)

var AllAccountTypes = []AccountType{
	AccountTypePersonal,
	AccountTypeInstitution,
	AccountTypeInvited,
}

var SelfServiceAccountTypes = []AccountType{
	AccountTypePersonal,
	AccountTypeInstitution,
}

// ============================================================
// BASIC METHODS - On AccountType
// ============================================================

func (a AccountType) String() string { return string(a) }

func (a AccountType) IsValid() bool {
	for _, t := range AllAccountTypes {
		if t == a {
			return true
		}
	}
	return false
}

func (a AccountType) IsPersonal() bool    { return a == AccountTypePersonal }
func (a AccountType) IsInstitution() bool { return a == AccountTypeInstitution }
func (a AccountType) IsInvited() bool     { return a == AccountTypeInvited }

func (a AccountType) IsSelfService() bool {
	for _, t := range SelfServiceAccountTypes {
		if t == a {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On AccountType
// ============================================================

func (a AccountType) GetName() string { return string(a) }

func (a AccountType) GetSlug() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalSlug
	case AccountTypeInstitution:
		return AccountTypeInstitutionSlug
	case AccountTypeInvited:
		return AccountTypeInvitedSlug
	default:
		return string(a)
	}
}

func (a AccountType) GetDisplayName() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalDisplayName
	case AccountTypeInstitution:
		return AccountTypeInstitutionDisplayName
	case AccountTypeInvited:
		return AccountTypeInvitedDisplayName
	default:
		return string(a)
	}
}

func (a AccountType) GetDescription() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalDescription
	case AccountTypeInstitution:
		return AccountTypeInstitutionDescription
	case AccountTypeInvited:
		return AccountTypeInvitedDescription
	default:
		return ""
	}
}

func (a AccountType) GetIcon() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalIcon
	case AccountTypeInstitution:
		return AccountTypeInstitutionIcon
	case AccountTypeInvited:
		return AccountTypeInvitedIcon
	default:
		return "user"
	}
}

func (a AccountType) GetColor() string {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalColor
	case AccountTypeInstitution:
		return AccountTypeInstitutionColor
	case AccountTypeInvited:
		return AccountTypeInvitedColor
	default:
		return "#6B7280"
	}
}

func (a AccountType) GetSortOrder() int {
	switch a {
	case AccountTypePersonal:
		return AccountTypePersonalSortOrder
	case AccountTypeInstitution:
		return AccountTypeInstitutionSortOrder
	case AccountTypeInvited:
		return AccountTypeInvitedSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For AccountType
// ============================================================

func ParseAccountType(name string) (AccountType, bool) {
	t := AccountType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func ParseAccountTypeWithDefault(name string, defaultType AccountType) AccountType {
	if t, ok := ParseAccountType(name); ok {
		return t
	}
	return defaultType
}

func ParseAccountTypeBySlug(slug string) (AccountType, bool) {
	for _, t := range AllAccountTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For AccountType
// ============================================================

func AllAccountTypeNames() []string {
	names := make([]string, 0, len(AllAccountTypes))
	for _, t := range AllAccountTypes {
		names = append(names, t.GetName())
	}
	return names
}

func AllAccountTypeSlugs() []string {
	slugs := make([]string, 0, len(AllAccountTypes))
	for _, t := range AllAccountTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

func AllAccountTypeDisplayNames() []string {
	names := make([]string, 0, len(AllAccountTypes))
	for _, t := range AllAccountTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}

func SelfServiceAccountTypeNames() []string {
	names := make([]string, 0, len(SelfServiceAccountTypes))
	for _, t := range SelfServiceAccountTypes {
		names = append(names, t.GetName())
	}
	return names
}

// ============================================================
// INSTITUTION TYPE CONSTANTS
// ============================================================

type InstitutionType string

// Slug constants
const (
	InstitutionTypeCompanySlug     = "institution-type-company"
	InstitutionTypeInstituteSlug   = "institution-type-institute"
	InstitutionTypeAssociationSlug = "institution-type-association"
	InstitutionTypeSchoolSlug      = "institution-type-school"
	InstitutionTypeUniversitySlug  = "institution-type-university"
	InstitutionTypeCollegeSlug     = "institution-type-college"      // NEW
	InstitutionTypeHighSchoolSlug  = "institution-type-high-school"  // NEW
	InstitutionTypeNgoSlug         = "institution-type-ngo"
	InstitutionTypeGovernmentSlug  = "institution-type-government"
	InstitutionTypeReligiousSlug   = "institution-type-religious"
	InstitutionTypeCommunitySlug   = "institution-type-community"
	InstitutionTypeOtherSlug       = "institution-type-other"
)

// Name constants
const (
	InstitutionTypeCompanyName     = "institution_type_company"
	InstitutionTypeInstituteName   = "institution_type_institute"
	InstitutionTypeAssociationName = "institution_type_association"
	InstitutionTypeSchoolName      = "institution_type_school"
	InstitutionTypeUniversityName  = "institution_type_university"
	InstitutionTypeCollegeName     = "institution_type_college"      // NEW
	InstitutionTypeHighSchoolName  = "institution_type_high_school"  // NEW
	InstitutionTypeNgoName         = "institution_type_ngo"
	InstitutionTypeGovernmentName  = "institution_type_government"
	InstitutionTypeReligiousName   = "institution_type_religious"
	InstitutionTypeCommunityName   = "institution_type_community"
	InstitutionTypeOtherName       = "institution_type_other"
)

// Display names
const (
	InstitutionTypeCompanyDisplayName     = "Company"
	InstitutionTypeInstituteDisplayName   = "Institute"
	InstitutionTypeAssociationDisplayName = "Association"
	InstitutionTypeSchoolDisplayName      = "School"
	InstitutionTypeUniversityDisplayName  = "University"
	InstitutionTypeCollegeDisplayName     = "College"        // NEW
	InstitutionTypeHighSchoolDisplayName  = "High School"    // NEW
	InstitutionTypeNgoDisplayName         = "NGO / Non-Profit"
	InstitutionTypeGovernmentDisplayName  = "Government Agency"
	InstitutionTypeReligiousDisplayName   = "Religious Organization"
	InstitutionTypeCommunityDisplayName   = "Community Organization"
	InstitutionTypeOtherDisplayName       = "Other"
)

// Descriptions
const (
	InstitutionTypeCompanyDescription     = "Corporate organization"
	InstitutionTypeInstituteDescription   = "Educational or research institute"
	InstitutionTypeAssociationDescription = "Professional or trade association"
	InstitutionTypeSchoolDescription      = "Primary or secondary educational institution"
	InstitutionTypeUniversityDescription  = "Higher education institution"
	InstitutionTypeCollegeDescription     = "Tertiary college or vocational institution"   // NEW
	InstitutionTypeHighSchoolDescription  = "Secondary school (high school)"                // NEW
	InstitutionTypeNgoDescription         = "Non-governmental or non-profit organization"
	InstitutionTypeGovernmentDescription  = "Government ministry, department, or agency"
	InstitutionTypeReligiousDescription   = "Church, mosque, or faith-based organization"
	InstitutionTypeCommunityDescription   = "Community group, alumni network, or cooperative"
	InstitutionTypeOtherDescription       = "Any other type of organization"
)

// Icons
const (
	InstitutionTypeCompanyIcon     = "building"
	InstitutionTypeInstituteIcon   = "landmark"
	InstitutionTypeAssociationIcon = "users"
	InstitutionTypeSchoolIcon      = "school"
	InstitutionTypeUniversityIcon  = "graduation-cap"
	InstitutionTypeCollegeIcon     = "building-2"      // NEW
	InstitutionTypeHighSchoolIcon  = "school"          // NEW
	InstitutionTypeNgoIcon         = "heart-handshake"
	InstitutionTypeGovernmentIcon  = "shield"
	InstitutionTypeReligiousIcon   = "church"
	InstitutionTypeCommunityIcon   = "users-round"
	InstitutionTypeOtherIcon       = "circle-help"
)

// Colors
const (
	InstitutionTypeCompanyColor     = "#7C3AED" // Purple
	InstitutionTypeInstituteColor   = "#4F46E5" // Indigo
	InstitutionTypeAssociationColor = "#0EA5E9" // Sky Blue
	InstitutionTypeSchoolColor      = "#F59E0B" // Amber
	InstitutionTypeUniversityColor  = "#10B981" // Emerald
	InstitutionTypeCollegeColor     = "#F472B6" // Pink       // NEW
	InstitutionTypeHighSchoolColor  = "#FB923C" // Orange     // NEW
	InstitutionTypeNgoColor         = "#EC4899" // Pink
	InstitutionTypeGovernmentColor  = "#64748B" // Slate
	InstitutionTypeReligiousColor   = "#A855F7" // Violet
	InstitutionTypeCommunityColor   = "#14B8A6" // Teal
	InstitutionTypeOtherColor       = "#6B7280" // Gray
)

// Sort orders
const (
	InstitutionTypeCompanySortOrder     = 1
	InstitutionTypeInstituteSortOrder   = 2
	InstitutionTypeAssociationSortOrder = 3
	InstitutionTypeSchoolSortOrder      = 4
	InstitutionTypeUniversitySortOrder  = 5
	InstitutionTypeCollegeSortOrder     = 6  // NEW
	InstitutionTypeHighSchoolSortOrder  = 7  // NEW
	InstitutionTypeNgoSortOrder         = 8  // shifted
	InstitutionTypeGovernmentSortOrder  = 9  // shifted
	InstitutionTypeReligiousSortOrder   = 10 // shifted
	InstitutionTypeCommunitySortOrder   = 11 // shifted
	InstitutionTypeOtherSortOrder       = 12 // shifted
)

// ============================================================
// INSTITUTION TYPE DEFINITIONS
// ============================================================

const (
	InstitutionTypeCompany     InstitutionType = InstitutionTypeCompanyName
	InstitutionTypeInstitute   InstitutionType = InstitutionTypeInstituteName
	InstitutionTypeAssociation InstitutionType = InstitutionTypeAssociationName
	InstitutionTypeSchool      InstitutionType = InstitutionTypeSchoolName
	InstitutionTypeUniversity  InstitutionType = InstitutionTypeUniversityName
	InstitutionTypeCollege     InstitutionType = InstitutionTypeCollegeName    // NEW
	InstitutionTypeHighSchool  InstitutionType = InstitutionTypeHighSchoolName // NEW
	InstitutionTypeNgo         InstitutionType = InstitutionTypeNgoName
	InstitutionTypeGovernment  InstitutionType = InstitutionTypeGovernmentName
	InstitutionTypeReligious   InstitutionType = InstitutionTypeReligiousName
	InstitutionTypeCommunity   InstitutionType = InstitutionTypeCommunityName
	InstitutionTypeOther       InstitutionType = InstitutionTypeOtherName
)

var AllInstitutionTypes = []InstitutionType{
	InstitutionTypeCompany,
	InstitutionTypeInstitute,
	InstitutionTypeAssociation,
	InstitutionTypeSchool,
	InstitutionTypeUniversity,
	InstitutionTypeCollege,    // NEW
	InstitutionTypeHighSchool, // NEW
	InstitutionTypeNgo,
	InstitutionTypeGovernment,
	InstitutionTypeReligious,
	InstitutionTypeCommunity,
	InstitutionTypeOther,
}

// ============================================================
// BASIC METHODS - On InstitutionType
// ============================================================

func (i InstitutionType) String() string { return string(i) }

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

func (i InstitutionType) GetName() string { return string(i) }

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
	case InstitutionTypeCollege:
		return InstitutionTypeCollegeSlug
	case InstitutionTypeHighSchool:
		return InstitutionTypeHighSchoolSlug
	case InstitutionTypeNgo:
		return InstitutionTypeNgoSlug
	case InstitutionTypeGovernment:
		return InstitutionTypeGovernmentSlug
	case InstitutionTypeReligious:
		return InstitutionTypeReligiousSlug
	case InstitutionTypeCommunity:
		return InstitutionTypeCommunitySlug
	case InstitutionTypeOther:
		return InstitutionTypeOtherSlug
	default:
		return string(i)
	}
}

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
	case InstitutionTypeCollege:
		return InstitutionTypeCollegeDisplayName
	case InstitutionTypeHighSchool:
		return InstitutionTypeHighSchoolDisplayName
	case InstitutionTypeNgo:
		return InstitutionTypeNgoDisplayName
	case InstitutionTypeGovernment:
		return InstitutionTypeGovernmentDisplayName
	case InstitutionTypeReligious:
		return InstitutionTypeReligiousDisplayName
	case InstitutionTypeCommunity:
		return InstitutionTypeCommunityDisplayName
	case InstitutionTypeOther:
		return InstitutionTypeOtherDisplayName
	default:
		return string(i)
	}
}

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
	case InstitutionTypeCollege:
		return InstitutionTypeCollegeDescription
	case InstitutionTypeHighSchool:
		return InstitutionTypeHighSchoolDescription
	case InstitutionTypeNgo:
		return InstitutionTypeNgoDescription
	case InstitutionTypeGovernment:
		return InstitutionTypeGovernmentDescription
	case InstitutionTypeReligious:
		return InstitutionTypeReligiousDescription
	case InstitutionTypeCommunity:
		return InstitutionTypeCommunityDescription
	case InstitutionTypeOther:
		return InstitutionTypeOtherDescription
	default:
		return ""
	}
}

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
	case InstitutionTypeCollege:
		return InstitutionTypeCollegeIcon
	case InstitutionTypeHighSchool:
		return InstitutionTypeHighSchoolIcon
	case InstitutionTypeNgo:
		return InstitutionTypeNgoIcon
	case InstitutionTypeGovernment:
		return InstitutionTypeGovernmentIcon
	case InstitutionTypeReligious:
		return InstitutionTypeReligiousIcon
	case InstitutionTypeCommunity:
		return InstitutionTypeCommunityIcon
	case InstitutionTypeOther:
		return InstitutionTypeOtherIcon
	default:
		return "building"
	}
}

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
	case InstitutionTypeCollege:
		return InstitutionTypeCollegeColor
	case InstitutionTypeHighSchool:
		return InstitutionTypeHighSchoolColor
	case InstitutionTypeNgo:
		return InstitutionTypeNgoColor
	case InstitutionTypeGovernment:
		return InstitutionTypeGovernmentColor
	case InstitutionTypeReligious:
		return InstitutionTypeReligiousColor
	case InstitutionTypeCommunity:
		return InstitutionTypeCommunityColor
	case InstitutionTypeOther:
		return InstitutionTypeOtherColor
	default:
		return "#6B7280"
	}
}

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
	case InstitutionTypeCollege:
		return InstitutionTypeCollegeSortOrder
	case InstitutionTypeHighSchool:
		return InstitutionTypeHighSchoolSortOrder
	case InstitutionTypeNgo:
		return InstitutionTypeNgoSortOrder
	case InstitutionTypeGovernment:
		return InstitutionTypeGovernmentSortOrder
	case InstitutionTypeReligious:
		return InstitutionTypeReligiousSortOrder
	case InstitutionTypeCommunity:
		return InstitutionTypeCommunitySortOrder
	case InstitutionTypeOther:
		return InstitutionTypeOtherSortOrder
	default:
		return 999
	}
}

// ============================================================
// PARSE FUNCTIONS - For InstitutionType
// ============================================================

func ParseInstitutionType(name string) (InstitutionType, bool) {
	t := InstitutionType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func ParseInstitutionTypeWithDefault(name string, defaultType InstitutionType) InstitutionType {
	if t, ok := ParseInstitutionType(name); ok {
		return t
	}
	return defaultType
}

func ParseInstitutionTypeBySlug(slug string) (InstitutionType, bool) {
	for _, t := range AllInstitutionTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For InstitutionType
// ============================================================

func AllInstitutionTypeNames() []string {
	names := make([]string, 0, len(AllInstitutionTypes))
	for _, t := range AllInstitutionTypes {
		names = append(names, t.GetName())
	}
	return names
}

func AllInstitutionTypeSlugs() []string {
	slugs := make([]string, 0, len(AllInstitutionTypes))
	for _, t := range AllInstitutionTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

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

type ProfessionalType string

// Slug constants
const (
	ProfessionalTypeTrainerSlug    = "professional-type-trainer"
	ProfessionalTypeCoachSlug      = "professional-type-coach"
	ProfessionalTypeConsultantSlug = "professional-type-consultant"
	ProfessionalTypeFreelancerSlug = "professional-type-freelancer"
	ProfessionalTypeOrganizerSlug  = "professional-type-organizer"
	ProfessionalTypeLearnerSlug    = "professional-type-learner"
)

// Name constants
const (
	ProfessionalTypeTrainerName    = "professional_type_trainer"
	ProfessionalTypeCoachName      = "professional_type_coach"
	ProfessionalTypeConsultantName = "professional_type_consultant"
	ProfessionalTypeFreelancerName = "professional_type_freelancer"
	ProfessionalTypeOrganizerName  = "professional_type_organizer"
	ProfessionalTypeLearnerName    = "professional_type_learner"
)

// Display names
const (
	ProfessionalTypeTrainerDisplayName    = "Trainer"
	ProfessionalTypeCoachDisplayName      = "Coach"
	ProfessionalTypeConsultantDisplayName = "Consultant"
	ProfessionalTypeFreelancerDisplayName = "Freelancer"
	ProfessionalTypeOrganizerDisplayName  = "Organizer"
	ProfessionalTypeLearnerDisplayName    = "Learner"
)

// Descriptions
const (
	ProfessionalTypeTrainerDescription    = "Professional trainer who conducts training sessions"
	ProfessionalTypeCoachDescription      = "Professional coach providing guidance and mentorship"
	ProfessionalTypeConsultantDescription = "Professional consultant providing expert advice"
	ProfessionalTypeFreelancerDescription = "Independent professional offering services"
	ProfessionalTypeOrganizerDescription  = "Student union, independent tutor, or alumni group hosting events"
	ProfessionalTypeLearnerDescription    = "Student or professional attending events to learn"
)

// Icons
const (
	ProfessionalTypeTrainerIcon    = "graduation-cap"
	ProfessionalTypeCoachIcon      = "user-tie"
	ProfessionalTypeConsultantIcon = "briefcase"
	ProfessionalTypeFreelancerIcon = "laptop"
	ProfessionalTypeOrganizerIcon  = "users"
	ProfessionalTypeLearnerIcon    = "book-open"
)

// Colors
const (
	ProfessionalTypeTrainerColor    = "#4F46E5"
	ProfessionalTypeCoachColor      = "#7C3AED"
	ProfessionalTypeConsultantColor = "#0EA5E9"
	ProfessionalTypeFreelancerColor = "#F59E0B"
	ProfessionalTypeOrganizerColor  = "#14B8A6"
	ProfessionalTypeLearnerColor    = "#10B981"
)

// Sort orders
const (
	ProfessionalTypeTrainerSortOrder    = 1
	ProfessionalTypeCoachSortOrder      = 2
	ProfessionalTypeConsultantSortOrder = 3
	ProfessionalTypeFreelancerSortOrder = 4
	ProfessionalTypeOrganizerSortOrder  = 5
	ProfessionalTypeLearnerSortOrder    = 6
)

// ============================================================
// PROFESSIONAL TYPE DEFINITIONS
// ============================================================

const (
	ProfessionalTypeTrainer    ProfessionalType = ProfessionalTypeTrainerName
	ProfessionalTypeCoach      ProfessionalType = ProfessionalTypeCoachName
	ProfessionalTypeConsultant ProfessionalType = ProfessionalTypeConsultantName
	ProfessionalTypeFreelancer ProfessionalType = ProfessionalTypeFreelancerName
	ProfessionalTypeOrganizer  ProfessionalType = ProfessionalTypeOrganizerName
	ProfessionalTypeLearner    ProfessionalType = ProfessionalTypeLearnerName
)

var AllProfessionalTypes = []ProfessionalType{
	ProfessionalTypeTrainer,
	ProfessionalTypeCoach,
	ProfessionalTypeConsultant,
	ProfessionalTypeFreelancer,
	ProfessionalTypeOrganizer,
	ProfessionalTypeLearner,
}

// ============================================================
// BASIC METHODS - On ProfessionalType
// ============================================================

func (p ProfessionalType) String() string { return string(p) }

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

func (p ProfessionalType) GetName() string { return string(p) }

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
	case ProfessionalTypeOrganizer:
		return ProfessionalTypeOrganizerSlug
	case ProfessionalTypeLearner:
		return ProfessionalTypeLearnerSlug
	default:
		return string(p)
	}
}

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
	case ProfessionalTypeOrganizer:
		return ProfessionalTypeOrganizerDisplayName
	case ProfessionalTypeLearner:
		return ProfessionalTypeLearnerDisplayName
	default:
		return string(p)
	}
}

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
	case ProfessionalTypeOrganizer:
		return ProfessionalTypeOrganizerDescription
	case ProfessionalTypeLearner:
		return ProfessionalTypeLearnerDescription
	default:
		return ""
	}
}

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
	case ProfessionalTypeOrganizer:
		return ProfessionalTypeOrganizerIcon
	case ProfessionalTypeLearner:
		return ProfessionalTypeLearnerIcon
	default:
		return "user"
	}
}

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
	case ProfessionalTypeOrganizer:
		return ProfessionalTypeOrganizerColor
	case ProfessionalTypeLearner:
		return ProfessionalTypeLearnerColor
	default:
		return "#6B7280"
	}
}

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
	case ProfessionalTypeOrganizer:
		return ProfessionalTypeOrganizerSortOrder
	case ProfessionalTypeLearner:
		return ProfessionalTypeLearnerSortOrder
	default:
		return 999
	}
}

func (p ProfessionalType) CanHost() bool {
	switch p {
	case ProfessionalTypeTrainer, ProfessionalTypeCoach, ProfessionalTypeConsultant, ProfessionalTypeOrganizer:
		return true
	case ProfessionalTypeFreelancer, ProfessionalTypeLearner:
		return false
	default:
		return false
	}
}

// ============================================================
// PARSE FUNCTIONS - For ProfessionalType
// ============================================================

func ParseProfessionalType(name string) (ProfessionalType, bool) {
	t := ProfessionalType(name)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func ParseProfessionalTypeWithDefault(name string, defaultType ProfessionalType) ProfessionalType {
	if t, ok := ParseProfessionalType(name); ok {
		return t
	}
	return defaultType
}

func ParseProfessionalTypeBySlug(slug string) (ProfessionalType, bool) {
	for _, t := range AllProfessionalTypes {
		if t.GetSlug() == slug {
			return t, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For ProfessionalType
// ============================================================

func AllProfessionalTypeNames() []string {
	names := make([]string, 0, len(AllProfessionalTypes))
	for _, t := range AllProfessionalTypes {
		names = append(names, t.GetName())
	}
	return names
}

func AllProfessionalTypeSlugs() []string {
	slugs := make([]string, 0, len(AllProfessionalTypes))
	for _, t := range AllProfessionalTypes {
		slugs = append(slugs, t.GetSlug())
	}
	return slugs
}

func AllProfessionalTypeDisplayNames() []string {
	names := make([]string, 0, len(AllProfessionalTypes))
	for _, t := range AllProfessionalTypes {
		names = append(names, t.GetDisplayName())
	}
	return names
}

func HostProfessionalTypes() []ProfessionalType {
	types := make([]ProfessionalType, 0)
	for _, t := range AllProfessionalTypes {
		if t.CanHost() {
			types = append(types, t)
		}
	}
	return types
}