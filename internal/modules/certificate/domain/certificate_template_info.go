// internal/modules/certificate/domain/certificate_template_info.go

package domain

// CertificateTemplateInfo is the seed-data shape for a system
// certificate template.
//
// These entries populate the certificate_templates table with the
// Nuruvent-provided defaults. Account-owned templates are created at
// runtime through the service and are not part of this registry.
//
// The `Slug` is the internal identifier used by code and by the
// certificate resolver. The `Name` is the customer-facing label shown
// in pickers and previews — it must not leak the internal identifier.
type CertificateTemplateInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	PreviewURL  string
	IsActive    bool
}

// certificateTemplateRegistry is the source of truth for system
// template metadata. Every template here is seeded with
// is_system = TRUE.
var certificateTemplateRegistry = []CertificateTemplateInfo{
	{
		Slug:        string(SystemTemplateDefault),
		Name:        "Standard Completion Certificate",
		DisplayName: "Standard Completion Certificate",
		Description: "Standard certificate template for events",
		PreviewURL:  "https://cdn.example.com/templates/default-preview.png",
		IsActive:    true,
	},
	{
		Slug:        string(SystemTemplateProfessional),
		Name:        "Professional Certificate",
		DisplayName: "Professional Certificate",
		Description: "Professional certificate template for CPD events",
		PreviewURL:  "https://cdn.example.com/templates/professional-preview.png",
		IsActive:    true,
	},
	{
		Slug:        string(SystemTemplatePremium),
		Name:        "Premium Certificate",
		DisplayName: "Premium Certificate",
		Description: "Premium certificate template with enhanced design",
		PreviewURL:  "https://cdn.example.com/templates/premium-preview.png",
		IsActive:    true,
	},
}

// AllCertificateTemplateInfos returns every system template entry.
func AllCertificateTemplateInfos() []CertificateTemplateInfo {
	out := make([]CertificateTemplateInfo, len(certificateTemplateRegistry))
	copy(out, certificateTemplateRegistry)
	return out
}

// ActiveCertificateTemplateInfos returns only active system templates.
func ActiveCertificateTemplateInfos() []CertificateTemplateInfo {
	out := make([]CertificateTemplateInfo, 0, len(certificateTemplateRegistry))
	for _, info := range certificateTemplateRegistry {
		if info.IsActive {
			out = append(out, info)
		}
	}
	return out
}