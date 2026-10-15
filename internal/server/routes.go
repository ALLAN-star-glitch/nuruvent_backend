// internal/server/routes.go

package server

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/handler"
	attendancehttp "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/delivery/http"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	certificatehttp "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/delivery/http"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	paymenthttp "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/delivery/http"
	registrationhttp "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/delivery/http"
	teamHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/delivery/handler"
	videohttp "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/delivery/http"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// SetupRoutes registers all API routes
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
	authHandler *authhandler.AuthHandler,
	eventsHandler *eventhandler.EventHandler,
	teamHandler *teamHandler.TeamHandler,
	accountHandler *handler.AccountHandler,
	regHandler *registrationhttp.Handler,
	paymentHandler *paymenthttp.Handler,
	orderHandler *paymenthttp.OrderHandler,
	attendanceHandler *attendancehttp.Handlers,
	videoHandler *videohttp.Handlers,
	certificateHandler *certificatehttp.Handlers,
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

	// Events routes — optionalAuth populates user context for public
	// GETs so private-event visibility rules can apply.
	eventsHandler.RegisterRoutes(api, authMiddleware, authzMiddleware, optionalAuth)

	// Team routes — optionalAuth lets invitation accept/decline work
	// for both anonymous (new user) and authenticated (existing user)
	// callers.
	teamHandler.RegisterRoutes(api, authMiddleware, authzMiddleware, optionalAuth)

	// Account routes
	accountHandler.RegisterRoutes(api, authMiddleware, authzMiddleware)

	// Registration routes — authzMiddleware now attached so
	// account-scoped reads (registration lists, attendee exports) are
	// authorized, not just authenticated.
	registrationhttp.RegisterRoutes(api, regHandler, authMiddleware, authzMiddleware, optionalAuth)

	// Payment routes (orders + payments + webhooks) — authzMiddleware
	// attached so the ledger, stats, single-payment lookup, and refund
	// routes are authorized against the caller's role.
	paymenthttp.RegisterRoutes(api, orderHandler, paymentHandler, authMiddleware, authzMiddleware, optionalAuth)

	sessionCtx := attendancehttp.SessionContextMiddleware(attendanceHandler.Service())
	attendancehttp.RegisterRoutes(
		api,
		attendanceHandler,
		authMiddleware,
		authzMiddleware,
		optionalAuth,
		sessionCtx,
	)

	// Video routes (OAuth connect/callback, connections)
	videohttp.RegisterRoutes(api.Group("/video"), videoHandler, authMiddleware)

	// Certificate routes
	//
	// The certificate module uses both authMiddleware and
	// authzMiddleware at the route level, plus service-level authz
	// for own-vs-all scope decisions. See the doc comment in the
	// module's routes.go for the rationale.
	certificatehttp.RegisterRoutes(api, certificateHandler, authMiddleware, authzMiddleware, optionalAuth)
	certificatehttp.RegisterUserRoutes(api, certificateHandler, authMiddleware, authzMiddleware)
	certificatehttp.RegisterAccountRoutes(api, certificateHandler, authMiddleware, authzMiddleware)

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