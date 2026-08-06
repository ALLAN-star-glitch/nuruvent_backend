// internal/server/routes.go

package server

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
)

// SetupRoutes registers all API routes
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authHandler *auth.Handler,
	enforcer *authorization.Enforcer,
	eventsModule *events.Module,
) {
	// ================================================
	// 1. SWAGGER - Must be FIRST to avoid 404
	// ================================================
	SetupSwagger(app)

	// ================================================
	// 2. PUBLIC ROUTES (No authentication required)
	// ================================================
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", healthCheck)

	// Root endpoint
	app.Get("/", welcome)

	// Auth routes (public)
	auth.RegisterAuthRoutes(api, authHandler)
	// Events module routes (includes public and protected)
	eventsModule.SetupRoutes(api, enforcer)

	// Media module - NO routes needed (just internal service)
	// Media service is injected into Events module for file operations

	// ================================================
	// 3. PROTECTED ROUTES (Authentication required)
	// ================================================
	protected := api.Group("/")
	protected.Use(auth.AuthMiddleware(cfg.JWT.Secret))

	// Auth protected routes
	auth.RegisterProtectedRoutes(protected, authHandler)

	// ================================================
	// 4. 404 Handler - Must be LAST
	// ================================================
	app.Use(func(c fiber.Ctx) error {
		return response.NotFound(c, "Route not found", nil)
	})
}

func healthCheck(c fiber.Ctx) error {
	return response.Success(c, "Server is healthy", fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func welcome(c fiber.Ctx) error {
	return response.Success(c, "Welcome to Nuruvent API", fiber.Map{
		"service": "Nuruvent API",
		"version": "1.0.0",
		"status":  "running",
	})
}