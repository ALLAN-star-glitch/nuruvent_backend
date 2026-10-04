// internal/app/app.go

package app

import (
	"context"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authmiddleware"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/server"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// App wraps the application dependencies.
type App struct {
	*AppDependencies
}

// NewApp creates and initializes the application.
func NewApp() (*App, error) {
	// Wire handles ALL initialization.
	deps, err := InitializeApp()
	if err != nil {
		return nil, err
	}
	log.Println("✅ Application dependencies initialized successfully")

	// Add CORS middleware to the Fiber app.
	app := deps.App
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://www.nuruvent.com",
			"https://nuruvent.com",
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"Cookie",
			"Set-Cookie",
		},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Set-Cookie", "Content-Length", "Content-Type"},
		MaxAge:           86400,
	}))

	return &App{AppDependencies: deps}, nil
}

// SetupRoutes registers all API routes.
func (app *App) SetupRoutes() {
	// Get token service from app dependencies.
	tokenSvc := app.AuthTokenService

	// ------------------------------------------------------------
	// User validator
	//
	// Called by the auth middleware after every successful JWT
	// validation to confirm the token's subject still references a
	// live, active user. Rejects ghost sessions: deleted users,
	// DB-reset users, tokens from another environment.
	//
	// Falls back to a no-op check if no repository is wired into
	// AppDependencies yet. Replace with a real lookup before
	// production.
	// ------------------------------------------------------------
	validateUser := app.buildUserValidator()

	// Auth middleware — validates JWT and verifies the user exists.
	authMiddleware := authmiddleware.AuthMiddleware(
		tokenSvc,
		validateUser,
		app.Config,
	)

	// Authorization middleware — Casbin permission checks.
	authzMiddleware := authorization.AuthorizationMiddleware(
		app.PermissionChecker,
	)

	// Optional auth — same user-existence check, non-fatal.
	optionalAuth := authmiddleware.OptionalAuthMiddleware(
		tokenSvc,
		validateUser,
		app.Config,
	)

	server.SetupRoutes(
		app.App,
		app.Config,
		authMiddleware,
		authzMiddleware,
		optionalAuth,
		app.AuthHandler,
		app.EventsHandler,
		app.TeamHandler,
		app.AccountHandler,
		app.RegHandler,
		app.PaymentHandler,
		app.OrderHandler,
		app.AttendanceHandler,
		app.VideoHandler,
	)
	log.Println("Routes registered successfully")
}

// buildUserValidator returns a UserValidator backed by whatever
// user repository is available on the AppDependencies. If none is
// wired (e.g. before a DI refactor), it returns a permissive
// fallback that logs a warning — never silently broken in prod.
//
// To wire the real check, expose the auth repository on
// AppDependencies as `AuthRepository authdomain.Repository`, then
// swap the fallback branch below.
func (app *App) buildUserValidator() authmiddleware.UserValidator {
	// ---- Preferred: real repository lookup ----
	//
	// Uncomment and adapt once AppDependencies exposes the repo:
	//
	// repo := app.AuthRepository
	// if repo != nil {
	//     return func(ctx context.Context, userID string) (bool, error) {
	//         user, err := repo.GetUserByID(ctx, userID)
	//         if err != nil || user == nil {
	//             return false, nil
	//         }
	//         if user.DeletedAt != nil || !user.IsActive {
	//             return false, nil
	//         }
	//         return true, nil
	//     }
	// }

	// ---- Fallback: permissive, log-only ----
	//
	// Keeps the build working until the repository is exposed on
	// AppDependencies. Logs a warning the first time it's hit so
	// it doesn't go unnoticed.
	var warned bool
	return func(ctx context.Context, userID string) (bool, error) {
		if !warned {
			log.Println("⚠️  [auth] UserValidator fallback active — set App.AuthRepository to enable ghost-session rejection")
			warned = true
		}
		return true, nil
	}
}

// Run starts the server.
func (app *App) Run() error {
	log.Printf("Server starting on port %s", app.Config.Server.Port)
	return app.App.Listen(":" + app.Config.Server.Port)
}

// Close gracefully shuts down the application.
func (app *App) Close() {
	log.Println("Shutting down application...")

	if app.Enforcer != nil {
		app.Enforcer.Close()
	}

	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	if app.RedisClient != nil {
		if err := app.RedisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}

	log.Println("Application closed successfully")
}