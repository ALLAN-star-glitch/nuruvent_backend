// internal/modules/certificate/domain/certificate_type.go

package domain

// CertificateType identifies the kind of certificate. It is a domain
// value object that replaces the earlier shared/types version so the
// certificate module owns its own vocabulary.
//
// The values mirror the certificate_types lookup table.
type CertificateType string

const (
	CertificateTypeEvent  CertificateType = "certificate_type_event"
	CertificateTypeCourse CertificateType = "certificate_type_course"
	CertificateTypeCPD    CertificateType = "certificate_type_cpd"
)

// AllCertificateTypes lists every valid certificate type.
var AllCertificateTypes = []CertificateType{
	CertificateTypeEvent,
	CertificateTypeCourse,
	CertificateTypeCPD,
}

// String returns the internal name.
func (c CertificateType) String() string { return string(c) }

// IsValid reports whether c is one of the known certificate types.
func (c CertificateType) IsValid() bool {
	for _, t := range AllCertificateTypes {
		if t == c {
			return true
		}
	}
	return false
}

// Slug returns the URL/API slug for the type.
func (c CertificateType) Slug() string {
	switch c {
	case CertificateTypeEvent:
		return "event-certificate"
	case CertificateTypeCourse:
		return "course-certificate"
	case CertificateTypeCPD:
		return "cpd-certificate"
	default:
		return string(c)
	}
}

// DisplayName returns a human-readable name.
func (c CertificateType) DisplayName() string {
	switch c {
	case CertificateTypeEvent:
		return "Event Certificate"
	case CertificateTypeCourse:
		return "Course Certificate"
	case CertificateTypeCPD:
		return "CPD Certificate"
	default:
		return string(c)
	}
}

// ParseCertificateType parses an internal name into a CertificateType.
func ParseCertificateType(name string) (CertificateType, bool) {
	c := CertificateType(name)
	if c.IsValid() {
		return c, true
	}
	return "", false
}

// ParseCertificateTypeBySlug parses a slug into a CertificateType.
func ParseCertificateTypeBySlug(slug string) (CertificateType, bool) {
	for _, c := range AllCertificateTypes {
		if c.Slug() == slug {
			return c, true
		}
	}
	return "", false
}