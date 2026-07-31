package auth

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/queue"
	"github.com/gofiber/fiber/v3"
)

type Module struct {
	handler      *Handler
	service      *Service
	tokenService *TokenService
	repo         *Repository
}

func NewModule(
	cfg *config.Config,
	permService *authorization.Service,
	businessService *bizservice.BusinessService,
) *Module {
	queueClient := queue.NewClient(cfg.Redis.URL)
	repo := NewRepository(database.GetDB())
	tokenService := NewTokenService(repo, cfg)

	// Initialize email service for sync fallback
	emailService := email.NewEmailService(cfg.Email.APIKey, cfg.Email.From)

	service := NewService(
		repo,
		cfg,
		queueClient,
		permService,
		tokenService,
		businessService,
		emailService,
	)
	handler := NewHandler(service, repo, cfg, tokenService)

	return &Module{
		handler:      handler,
		service:      service,
		tokenService: tokenService,
		repo:         repo,
	}
}

// SetupRoutes registers all auth routes
func (m *Module) SetupRoutes(router fiber.Router) {
	RegisterAuthRoutes(router, m.handler)
}

// SetupProtectedRoutes registers protected auth routes
func (m *Module) SetupProtectedRoutes(router fiber.Router) {
	RegisterProtectedRoutes(router, m.handler)
}

// GetHandler returns the auth handler
func (m *Module) GetHandler() *Handler {
	return m.handler
}

// GetService returns the auth service
func (m *Module) GetService() *Service {
	return m.service
}

func (m *Module) Init(ctx context.Context) error {
	log.Println("Auth module initialized")
	return nil
}

func (m *Module) Close() {
	log.Println("Auth module closed")
}