// internal/modules/certificate/service/dependencies.go

package service

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// Dependencies bundles every outbound port the certificate service
// needs.
type Dependencies struct {
	Certificates  domain.CertificateRepository
	Templates     domain.TemplateRepository

	IDs   domain.IDGenerator
	Codes domain.VerificationCodeGenerator
	Clock domain.Clock

	Authz domain.AuthorizationChecker

	PDF     domain.PDFGenerator
	Storage domain.ObjectStorage

	Payment       domain.PaymentCharger
	AccountPrices domain.AccountPriceResolver

	// Phase 4: signed download tokens
	DownloadTokens   domain.DownloadTokenSigner
	DownloadTokenTTL time.Duration

	VerifyBaseURL string

	Notifier domain.NotificationSender
}