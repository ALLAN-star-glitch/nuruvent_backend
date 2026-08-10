package server

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/acchandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// SetupRoutes registers all API routes
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	authHandler *authhandler.AuthHandler,
	accountHandler *acchandler.AccountHandler,
	eventsHandler *eventhandler.EventHandler,
) {
	// ================================================
	// 1. SWAGGER - Must be FIRST to avoid 404
	// ================================================
	SetupSwagger(app)

	// ================================================
	// 2. API ROUTES
	// ================================================
	api := app.Group("/api/v1")

	// Health check (public)
	api.Get("/health", healthCheck)

	// Root endpoint (public)
	app.Get("/", welcome)

	// ================================================
	// 3. MODULE ROUTES
	// ================================================

	// Auth routes (public + protected)
	authHandler.RegisterRoutes(api, authMiddleware)

	// Account routes
	accountHandler.RegisterRoutes(api, authMiddleware, authzMiddleware)

	// Events routes
	eventsHandler.RegisterRoutes(api, authMiddleware, authzMiddleware)

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