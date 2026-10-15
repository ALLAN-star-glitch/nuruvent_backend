// internal/modules/certificate/domain/certificate_type_info.go

package domain

// CertificateTypeInfo is the seed-data shape for a certificate type.
//
// These entries populate the certificate_types lookup table. They
// mirror the CertificateType constants defined in certificate_type.go
// but carry the DB-facing fields (Slug, Name, DisplayName) that the
// type methods already derive.
type CertificateTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	IsActive    bool
}

// AllCertificateTypeInfos returns every certificate type as seed data.
//
// Derived from AllCertificateTypes so the constants and the seed
// registry cannot diverge.
func AllCertificateTypeInfos() []CertificateTypeInfo {
	out := make([]CertificateTypeInfo, 0, len(AllCertificateTypes))
	for _, ct := range AllCertificateTypes {
		out = append(out, CertificateTypeInfo{
			Slug:        ct.Slug(),
			Name:        ct.String(),
			DisplayName: ct.DisplayName(),
			IsActive:    true,
		})
	}
	return out
}

// ActiveCertificateTypeInfos returns only active certificate types.
//
// Every type currently defined is active; the function exists for
// symmetry with ActiveCertificateTemplateInfos and to keep the
// seeder interface uniform.
func ActiveCertificateTypeInfos() []CertificateTypeInfo {
	out := make([]CertificateTypeInfo, 0, len(AllCertificateTypes))
	for _, ct := range AllCertificateTypes {
		out = append(out, CertificateTypeInfo{
			Slug:        ct.Slug(),
			Name:        ct.String(),
			DisplayName: ct.DisplayName(),
			IsActive:    true,
		})
	}
	return out
}