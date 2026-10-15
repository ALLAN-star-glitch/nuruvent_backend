// internal/modules/certificate/providers.go

package certificate

import (
	"github.com/google/wire"

	certificateDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	certificateClock "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/infrastructure/clock"
	certificateID "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/infrastructure/id"
	certificatePostgres "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/infrastructure/postgres"
	certificateVC "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/infrastructure/verificationcode"
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	certificateHTTP "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/delivery/http"
)

// ProviderSet is the certificate module's wire provider set.
var ProviderSet = wire.NewSet(
	// Persistence
	certificatePostgres.NewCertificateRepository,
	wire.Bind(new(certificateDomain.CertificateRepository), new(*certificatePostgres.CertificateRepository)),

	certificatePostgres.NewTemplateRepository,
	wire.Bind(new(certificateDomain.TemplateRepository), new(*certificatePostgres.TemplateRepository)),

	// Infrastructure adapters
	certificateClock.NewSystemClock,
	wire.Bind(new(certificateDomain.Clock), new(*certificateClock.SystemClock)),

	certificateVC.NewCryptoGenerator,
	wire.Bind(new(certificateDomain.VerificationCodeGenerator), new(*certificateVC.CryptoGenerator)),

	certificateID.NewUUIDGenerator,
	wire.Bind(new(certificateDomain.IDGenerator), new(*certificateID.UUIDGenerator)),

	// Service dependencies bundle
	ProvideDependencies,

	// Service — returns the Service interface directly.
	certificateService.NewService,

		// HTTP handlers
	certificateHTTP.NewHandlers,
)

// ProvideDependencies assembles the service.Dependencies bundle.
func ProvideDependencies(
	certs certificateDomain.CertificateRepository,
	templates certificateDomain.TemplateRepository,
	ids certificateDomain.IDGenerator,
	codes certificateDomain.VerificationCodeGenerator,
	clock certificateDomain.Clock,
	authz certificateDomain.AuthorizationChecker,
) certificateService.Dependencies {
	return certificateService.Dependencies{
		Certificates: certs,
		Templates:    templates,
		IDs:          ids,
		Codes:        codes,
		Clock:        clock,
		Authz:        authz,
	}
}