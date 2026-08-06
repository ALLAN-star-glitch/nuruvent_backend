// internal/modules/auth/module.go

package auth

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/handler"
	authRepo "github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/repository"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/service"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/email"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/queue"
	"github.com/gofiber/fiber/v3"
)

type Module struct {
	handler *handler.Handler
	service service.Service
}

func NewModule(
	cfg *config.Config,
	permService *authorization.Service,
) *Module {
	db := database.GetDB()

	// 1. Initialize repository (implements domain.Repository)
	repo := authRepo.NewGormRepo(db)

	// 2. Initialize token service
	tokenSvc := service.NewTokenService(repo, cfg)

	// 3. Initialize email service
	emailService := email.NewEmailService(cfg.Email.APIKey, cfg.Email.From)

	// 4. Initialize queue client
	queueClient := queue.NewClient(cfg.Redis.URL)

	// 5. Initialize main service
	svc := service.NewService(
		repo,
		cfg,
		queueClient,
		permService,
		tokenSvc,
		emailService,
	)

	// 6. Initialize handler
	h := handler.NewHandler(svc, cfg)

	return &Module{
		handler: h,
		service: svc,
	}
}

// SetupRoutes registers all auth routes (public + protected)
func (m *Module) SetupRoutes(router fiber.Router) {
	// Register public routes
	RegisterAuthRoutes(router, m.handler)
}

// SetupProtectedRoutes registers protected auth routes (authentication required)
func (m *Module) SetupProtectedRoutes(router fiber.Router) {
	RegisterProtectedRoutes(router, m.handler)
}

// GetHandler returns the auth handler
func (m *Module) GetHandler() *handler.Handler {
	return m.handler
}

// GetService returns the auth service
func (m *Module) GetService() service.Service {
	return m.service
}

func (m *Module) Init(ctx context.Context) error {
	log.Println("Auth module initialized")
	return nil
}

func (m *Module) Close() {
	log.Println("Auth module closed")
}