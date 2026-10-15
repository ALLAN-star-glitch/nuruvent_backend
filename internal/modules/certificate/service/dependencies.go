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

	// Phase 2: PDF rendering + storage
	PDF     domain.PDFGenerator
	Storage domain.ObjectStorage

	// Phase 5: paid certificates
	Payment       domain.PaymentCharger
	AccountPrices domain.AccountPriceResolver

	// VerifyBaseURL is the public base URL used to build the QR-code
	// verification link. Comes from cfg.App.PublicURL.
	VerifyBaseURL string
}