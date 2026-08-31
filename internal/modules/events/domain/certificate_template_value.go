// internal/modules/events/domain/certificate_template_value.go

package domain

// CertificateTemplateID is a type alias for certificate template IDs
type CertificateTemplateID string

// CertificateTemplateSlug constants
const (
	CertificateTemplateDefaultSlug      = "default"
	CertificateTemplateProfessionalSlug = "professional"
	CertificateTemplatePremiumSlug      = "premium"
)

// CertificateTemplateDisplayName constants
const (
	CertificateTemplateDefaultDisplayName      = "Standard Completion Certificate"
	CertificateTemplateProfessionalDisplayName = "Professional Certificate"
	CertificateTemplatePremiumDisplayName      = "Premium Certificate"
)

// CertificateTemplateInfo holds metadata for each certificate template
type CertificateTemplateInfo struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	PreviewURL  string
	IsActive    bool
}

// certificateTemplateRegistry is the domain's source of truth for template metadata
var certificateTemplateRegistry = map[CertificateTemplateID]CertificateTemplateInfo{
	"tmpl_cert_default_01": {
		ID:          "tmpl_cert_default_01",
		Slug:        CertificateTemplateDefaultSlug,
		Name:        "cert_template_default",
		DisplayName: CertificateTemplateDefaultDisplayName,
		Description: "Standard certificate template for events",
		PreviewURL:  "https://cdn.example.com/templates/default-preview.png",
		IsActive:    true,
	},
	// Add more templates as needed
}

// GetCertificateTemplateInfo returns template info by ID
func GetCertificateTemplateInfo(id string) (CertificateTemplateInfo, bool) {
	info, ok := certificateTemplateRegistry[CertificateTemplateID(id)]
	return info, ok
}

// GetCertificateTemplateBySlug returns template info by slug
func GetCertificateTemplateBySlug(slug string) (CertificateTemplateInfo, bool) {
	for _, info := range certificateTemplateRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return CertificateTemplateInfo{}, false
}

// AllCertificateTemplateInfos returns all template infos
func AllCertificateTemplateInfos() []CertificateTemplateInfo {
	infos := make([]CertificateTemplateInfo, 0, len(certificateTemplateRegistry))
	for _, info := range certificateTemplateRegistry {
		infos = append(infos, info)
	}
	return infos
}

// ActiveCertificateTemplateInfos returns only active template infos
func ActiveCertificateTemplateInfos() []CertificateTemplateInfo {
	infos := make([]CertificateTemplateInfo, 0)
	for _, info := range certificateTemplateRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}