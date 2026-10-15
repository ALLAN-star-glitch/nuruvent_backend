// internal/modules/certificate/service/dependencies.go

package service

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// Dependencies bundles every outbound port the certificate service
// needs.
type Dependencies struct {
	Certificates domain.CertificateRepository
	Templates    domain.TemplateRepository

	IDs   domain.IDGenerator
	Codes domain.VerificationCodeGenerator
	Clock domain.Clock

	Authz domain.AuthorizationChecker

	// Reserved for later phases:
	// Storage      domain.ObjectStorage
	// PDF          domain.PDFGenerator
	// Notification domain.NotificationSender
}