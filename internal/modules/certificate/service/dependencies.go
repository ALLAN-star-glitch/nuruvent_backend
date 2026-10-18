// internal/modules/certificate/service/dependencies.go

package service

import (
	"time"

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

	PDF     domain.PDFGenerator
	Storage domain.ObjectStorage

	Payment       domain.PaymentCharger
	AccountPrices domain.AccountPriceResolver

	// Signed download tokens.
	DownloadTokens   domain.DownloadTokenSigner
	DownloadTokenTTL time.Duration

	// VerifyBaseURL is the frontend URL. Used in QR codes embedded
	// in the PDF and in the footer text ("Verify at …").
	VerifyBaseURL string

	// DownloadBaseURL is the backend API URL. Used to build signed
	// certificate download links that recipients click from email.
	// Distinct from VerifyBaseURL because the two usually live on
	// different hosts (www vs api).
	DownloadBaseURL string

	Notifier domain.NotificationSender

	// Accounts resolves account metadata (e.g. issuer name) when
	// issuing certificates on behalf of an organizer.
	Accounts domain.AccountReader

	RenderJobs domain.RenderJobQueue
}