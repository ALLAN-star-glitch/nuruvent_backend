// internal/shared/types/certificate.go

package types

// ============================================================
// CERTIFICATE TEMPLATE CONSTANTS
// ============================================================

// CertificateTemplateSlug constants (with hyphens) - Used for URLs and API routes
const (
	CertificateTemplateDefaultSlug      = "default"
	CertificateTemplateProfessionalSlug = "professional"
	CertificateTemplatePremiumSlug      = "premium"
)

// CertificateTemplateName constants (with underscores) - Used for internal database lookups
const (
	CertificateTemplateDefaultName      = "cert_template_default"
	CertificateTemplateProfessionalName = "cert_template_professional"
	CertificateTemplatePremiumName      = "cert_template_premium"
)

// CertificateTemplateDisplayName constants - Used for UI display
const (
	CertificateTemplateDefaultDisplayName      = "Standard Completion Certificate"
	CertificateTemplateProfessionalDisplayName = "Professional Certificate"
	CertificateTemplatePremiumDisplayName      = "Premium Certificate"
)

// CertificateTemplateDescription constants
const (
	CertificateTemplateDefaultDescription      = "Standard certificate template for events"
	CertificateTemplateProfessionalDescription = "Professional certificate template for CPD events"
	CertificateTemplatePremiumDescription      = "Premium certificate template with enhanced design"
)

// ============================================================
// CERTIFICATE TYPE CONSTANTS
// ============================================================

// CertificateType represents the type of certificate
type CertificateType string

// Slug constants (with hyphens)
const (
	CertificateTypeEventSlug   = "event-certificate"
	CertificateTypeCourseSlug  = "course-certificate"
	CertificateTypeCPDSlug     = "cpd-certificate"
)

// Name constants (with underscores)
const (
	CertificateTypeEventName   = "certificate_type_event"
	CertificateTypeCourseName  = "certificate_type_course"
	CertificateTypeCPDName     = "certificate_type_cpd"
)

// Display name constants
const (
	CertificateTypeEventDisplayName   = "Event Certificate"
	CertificateTypeCourseDisplayName  = "Course Certificate"
	CertificateTypeCPDDisplayName     = "CPD Certificate"
)

// ============================================================
// CERTIFICATE TYPE DEFINITIONS
// ============================================================

// CertificateType constants - These are the NAMES (with underscores)
const (
	CertificateTypeEvent  CertificateType = CertificateTypeEventName  // "certificate_type_event"
	CertificateTypeCourse CertificateType = CertificateTypeCourseName // "certificate_type_course"
	CertificateTypeCPD    CertificateType = CertificateTypeCPDName    // "certificate_type_cpd"
)

// AllCertificateTypes lists all valid certificate types for validation
var AllCertificateTypes = []CertificateType{
	CertificateTypeEvent,
	CertificateTypeCourse,
	CertificateTypeCPD,
}

// ============================================================
// BASIC METHODS - On CertificateType
// ============================================================

// String returns the string representation (the name with underscores)
func (c CertificateType) String() string {
	return string(c)
}

// IsValid checks if the certificate type is valid
func (c CertificateType) IsValid() bool {
	for _, ct := range AllCertificateTypes {
		if ct == c {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On CertificateType
// ============================================================

// GetName returns the internal name (with underscores)
func (c CertificateType) GetName() string {
	return string(c)
}

// GetSlug returns the slug (with hyphens)
func (c CertificateType) GetSlug() string {
	switch c {
	case CertificateTypeEvent:
		return CertificateTypeEventSlug
	case CertificateTypeCourse:
		return CertificateTypeCourseSlug
	case CertificateTypeCPD:
		return CertificateTypeCPDSlug
	default:
		return string(c)
	}
}

// GetDisplayName returns the user-facing display name
func (c CertificateType) GetDisplayName() string {
	switch c {
	case CertificateTypeEvent:
		return CertificateTypeEventDisplayName
	case CertificateTypeCourse:
		return CertificateTypeCourseDisplayName
	case CertificateTypeCPD:
		return CertificateTypeCPDDisplayName
	default:
		return string(c)
	}
}

// ============================================================
// PARSE FUNCTIONS - For CertificateType
// ============================================================

// ParseCertificateType parses a string into a CertificateType
func ParseCertificateType(name string) (CertificateType, bool) {
	c := CertificateType(name)
	if c.IsValid() {
		return c, true
	}
	return "", false
}

// ParseCertificateTypeBySlug parses a slug into a CertificateType
func ParseCertificateTypeBySlug(slug string) (CertificateType, bool) {
	for _, c := range AllCertificateTypes {
		if c.GetSlug() == slug {
			return c, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// AllCertificateTypeNames returns all internal certificate type names
func AllCertificateTypeNames() []string {
	names := make([]string, 0, len(AllCertificateTypes))
	for _, c := range AllCertificateTypes {
		names = append(names, c.GetName())
	}
	return names
}

// AllCertificateTypeSlugs returns all certificate type slugs
func AllCertificateTypeSlugs() []string {
	slugs := make([]string, 0, len(AllCertificateTypes))
	for _, c := range AllCertificateTypes {
		slugs = append(slugs, c.GetSlug())
	}
	return slugs
}

// AllCertificateTypeDisplayNames returns all certificate type display names
func AllCertificateTypeDisplayNames() []string {
	names := make([]string, 0, len(AllCertificateTypes))
	for _, c := range AllCertificateTypes {
		names = append(names, c.GetDisplayName())
	}
	return names
}