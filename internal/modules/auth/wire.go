// internal/modules/auth/wire.go

package auth

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/domain"
	authRepo "github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/repository"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/handler"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/service"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/queue"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	ProvideRepository,
	ProvideTokenService,
	ProvideAuthService,
	ProvideAuthHandler,
	ProvideAuthModule,
)

func ProvideRepository(db *gorm.DB) domain.Repository {
	return authRepo.NewGormRepo(db)
}

func ProvideTokenService(
	repo domain.Repository,
	cfg *config.Config,
) service.TokenService {
	return service.NewTokenService(repo, cfg)
}

func ProvideAuthService(
	repo domain.Repository,
	tokenService service.TokenService,
	permService *authorization.Service,
	cfg *config.Config,
	emailService *email.EmailService,
	queueClient *queue.Client,
) service.Service {
	return service.NewService(
		repo,
		cfg,
		queueClient,
		permService,
		tokenService,
		emailService,
	)
}

func ProvideAuthHandler(
	svc service.Service,
	cfg *config.Config,
) *handler.Handler {
	return handler.NewHandler(svc, cfg)
}

func ProvideAuthModule(
	h *handler.Handler,
	svc service.Service,
) *Module {
	return &Module{
		handler: h,
		service: svc,
	}
}