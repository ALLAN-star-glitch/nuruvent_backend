// internal/modules/events/domain/certificate_type_value.go

package domain

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// CertificateTypeValue is an alias for the shared CertificateType
type CertificateTypeValue = types.CertificateType

// Type constants - Re-exported from shared types for convenience
const (
	CertificateTypeEvent  = types.CertificateTypeEvent
	CertificateTypeCourse = types.CertificateTypeCourse
	CertificateTypeCPD    = types.CertificateTypeCPD
)

// AllCertificateTypes re-exported from shared types
var AllCertificateTypes = types.AllCertificateTypes

// CertificateTypeInfo holds metadata for each certificate type
type CertificateTypeInfo struct {
	Slug        string
	Name        string
	DisplayName string
	IsActive    bool
}

// certificateTypeRegistry - Domain specific wrapper
var certificateTypeRegistry = map[types.CertificateType]CertificateTypeInfo{
	types.CertificateTypeEvent: {
		Slug:        types.CertificateTypeEventSlug,
		Name:        types.CertificateTypeEventName,
		DisplayName: types.CertificateTypeEventDisplayName,
		IsActive:    true,
	},
	types.CertificateTypeCourse: {
		Slug:        types.CertificateTypeCourseSlug,
		Name:        types.CertificateTypeCourseName,
		DisplayName: types.CertificateTypeCourseDisplayName,
		IsActive:    true,
	},
	types.CertificateTypeCPD: {
		Slug:        types.CertificateTypeCPDSlug,
		Name:        types.CertificateTypeCPDName,
		DisplayName: types.CertificateTypeCPDDisplayName,
		IsActive:    true,
	},
}

func GetCertificateTypeInfo(certType CertificateTypeValue) (CertificateTypeInfo, bool) {
	info, ok := certificateTypeRegistry[certType]
	return info, ok
}

func GetCertificateTypeSlug(certType CertificateTypeValue) string {
	return certType.GetSlug()
}

func GetCertificateTypeName(certType CertificateTypeValue) string {
	return certType.GetName()
}

func GetCertificateTypeDisplayName(certType CertificateTypeValue) string {
	return certType.GetDisplayName()
}

func IsCertificateTypeValid(certType CertificateTypeValue) bool {
	return certType.IsValid()
}

func AllCertificateTypeInfos() []CertificateTypeInfo {
	infos := make([]CertificateTypeInfo, 0, len(certificateTypeRegistry))
	for _, info := range certificateTypeRegistry {
		infos = append(infos, info)
	}
	return infos
}

func ActiveCertificateTypeInfos() []CertificateTypeInfo {
	infos := make([]CertificateTypeInfo, 0)
	for _, info := range certificateTypeRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func GetCertificateTypeBySlug(slug string) (CertificateTypeInfo, bool) {
	for certType, info := range certificateTypeRegistry {
		if certType.GetSlug() == slug {
			return info, true
		}
	}
	return CertificateTypeInfo{}, false
}

func GetCertificateTypeByName(name string) (CertificateTypeInfo, bool) {
	certType, ok := types.ParseCertificateType(name)
	if !ok {
		return CertificateTypeInfo{}, false
	}
	return GetCertificateTypeInfo(certType)
}

func GetCertificateTypeByDisplayName(displayName string) (CertificateTypeInfo, bool) {
	for _, info := range certificateTypeRegistry {
		if info.DisplayName == displayName {
			return info, true
		}
	}
	return CertificateTypeInfo{}, false
}