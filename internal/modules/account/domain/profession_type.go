package domain


// ============================================================
// PROFESSIONAL TYPE - Value Object
// ============================================================

type ProfessionalTypeValue string

const (
	ProfessionalTypeTrainer   ProfessionalTypeValue = "trainer"
	ProfessionalTypeCoach     ProfessionalTypeValue = "coach"
	ProfessionalTypeConsultant ProfessionalTypeValue = "consultant"
	ProfessionalTypeFreelancer ProfessionalTypeValue = "freelancer"
)

// ProfessionalTypeInfo holds metadata for each professional type
type ProfessionalTypeInfo struct {
	Slug        ProfessionalTypeValue
	Name        string
	DisplayName string
	Description string
	Icon        string
	Color       string
	SortOrder   int
	CanHost     bool
	IsActive    bool
}

// Private registry (prevents external mutation)
var professionalTypeRegistry = map[ProfessionalTypeValue]ProfessionalTypeInfo{
	ProfessionalTypeTrainer: {
		Slug:        ProfessionalTypeTrainer,
		Name:        "Trainer",
		DisplayName: "🎓 Trainer",
		Description: "Professional trainer who conducts training sessions",
		Icon:        "graduation-cap",
		Color:       "#4F46E5",
		SortOrder:   1,
		CanHost:     true,
		IsActive:    true,
	},
	ProfessionalTypeCoach: {
		Slug:        ProfessionalTypeCoach,
		Name:        "Coach",
		DisplayName: "👨‍🏫 Coach",
		Description: "Professional coach providing guidance and mentorship",
		Icon:        "user-tie",
		Color:       "#7C3AED",
		SortOrder:   2,
		CanHost:     true,
		IsActive:    true,
	},
	ProfessionalTypeConsultant: {
		Slug:        ProfessionalTypeConsultant,
		Name:        "Consultant",
		DisplayName: "💼 Consultant",
		Description: "Professional consultant providing expert advice",
		Icon:        "briefcase",
		Color:       "#0EA5E9",
		SortOrder:   3,
		CanHost:     true,
		IsActive:    true,
	},
	ProfessionalTypeFreelancer: {
		Slug:        ProfessionalTypeFreelancer,
		Name:        "Freelancer",
		DisplayName: "🖥️ Freelancer",
		Description: "Independent professional offering services",
		Icon:        "laptop",
		Color:       "#F59E0B",
		SortOrder:   4,
		CanHost:     false,
		IsActive:    true,
	},
}

// ============================================================
// DOMAIN METHODS (on ProfessionalTypeValue)
// ============================================================

func (p ProfessionalTypeValue) String() string {
	return string(p)
}

func (p ProfessionalTypeValue) IsValid() bool {
	_, ok := professionalTypeRegistry[p]
	return ok
}

func (p ProfessionalTypeValue) Info() (ProfessionalTypeInfo, bool) {
	info, ok := professionalTypeRegistry[p]
	return info, ok
}

func (p ProfessionalTypeValue) CanHost() bool {
	info, ok := professionalTypeRegistry[p]
	if !ok {
		return false
	}
	return info.CanHost
}

func (p ProfessionalTypeValue) IsActive() bool {
	info, ok := professionalTypeRegistry[p]
	if !ok {
		return false
	}
	return info.IsActive
}

// ============================================================
// READ-ONLY GETTERS
// ============================================================

func ParseProfessionalType(slug string) (ProfessionalTypeValue, bool) {
	t := ProfessionalTypeValue(slug)
	if t.IsValid() {
		return t, true
	}
	return "", false
}

func AllProfessionalTypeInfos() []ProfessionalTypeInfo {
	infos := make([]ProfessionalTypeInfo, 0, len(professionalTypeRegistry))
	for _, info := range professionalTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

func AllProfessionalTypeSlugs() []string {
	slugs := make([]string, 0, len(professionalTypeRegistry))
	for slug := range professionalTypeRegistry {
		slugs = append(slugs, string(slug))
	}
	return slugs
}

func ActiveProfessionalTypeInfos() []ProfessionalTypeInfo {
	infos := make([]ProfessionalTypeInfo, 0)
	for _, info := range professionalTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}