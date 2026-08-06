// internal/modules/auth/wire.go

package auth

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/account/accountrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/authrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/queue"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet defines all dependencies for the auth module
var ProviderSet = wire.NewSet(
	ProvideAuthRepository,
	ProvideTokenService,
	ProvideAuthService,
	ProvideAuthHandler,
)

// ProvideAuthRepository provides the auth repository
func ProvideAuthRepository(db *gorm.DB) *authrepo.Repository {
	return authrepo.NewRepository(db)
}

// ProvideTokenService provides the token service
func ProvideTokenService(
	authRepo *authrepo.Repository,
	accountRepo *accountrepo.Repository,
	cfg *config.Config,
) *TokenService {
	return NewTokenService(authRepo, accountRepo, cfg)
}

// ProvideAuthService provides the auth service
func ProvideAuthService(
	authRepo *authrepo.Repository,
	accountRepo *accountrepo.Repository,
	tokenService *TokenService,
	permService *authorization.Service,
	cfg *config.Config,
	emailService *email.EmailService,
	queueClient *queue.Client,
) *Service {
	return NewService(
		authRepo,
		accountRepo,
		tokenService,
		permService,
		cfg,
		emailService,
		queueClient,
	)
}

// ProvideAuthHandler provides the auth handler
func ProvideAuthHandler(service *Service) *Handler {
	return NewHandler(service)
}