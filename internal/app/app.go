// internal/app/app.go

package app

import (
	"context"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authmiddleware"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/server"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/gofiber/fiber/v3/middleware/cors" // ✅ Import CORS
)

// App wraps the application dependencies
type App struct {
	*AppDependencies
}

// NewApp creates and initializes the application
func NewApp() (*App, error) {
	// Wire handles ALL initialization (config, database, redis, etc.)
	deps, err := InitializeApp()
	if err != nil {
		return nil, err
	}
	log.Println("✅ Application dependencies initialized successfully")

	// ✅ Add CORS middleware to the Fiber app
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
		},
		AllowCredentials: true, // ✅ Required for cookies
		ExposeHeaders:    []string{"Set-Cookie"},
		MaxAge:           86400,
	}))

	return &App{AppDependencies: deps}, nil
}

// Init initializes all modules
func (app *App) Init(ctx context.Context) error {
	log.Println("Initializing modules...")
	if app.AuthService != nil {
		log.Println("✅ Auth module ready")
	}
	if app.AccountService != nil {
		log.Println("✅ Account module ready")
	}
	if app.EventsService != nil {
		log.Println("✅ Events module ready")
	}
	if app.MediaService != nil {
		log.Println("✅ Media module ready")
	}
	if app.Notification != nil {
		log.Println("✅ Notification module ready")
	}
	log.Println("All modules initialized successfully")
	return nil
}

// SetupRoutes registers all API routes
func (app *App) SetupRoutes() {
	// Get token service from app dependencies
	tokenSvc := app.AuthTokenService
	
	// Create middleware with token service
	authMiddleware := authmiddleware.AuthMiddleware(tokenSvc)
	authzMiddleware := authorization.AuthorizationMiddleware(app.Enforcer)

	server.SetupRoutes(
		app.App,
		app.Config,
		authMiddleware,
		authzMiddleware,
		app.AuthHandler,
		app.AccountHandler,
		app.EventsHandler,
	)
	log.Println("Routes registered successfully")
}

// Run starts the server
func (app *App) Run() error {
	log.Printf("Server starting on port %s", app.Config.Server.Port)
	return app.App.Listen(":" + app.Config.Server.Port)
}

// Close gracefully shuts down the application
func (app *App) Close() {
	log.Println("Shutting down application...")
	
	if app.Enforcer != nil {
		app.Enforcer.Close()
	}
	
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
	
	// Close Redis using the injected client
	if app.RedisClient != nil {
		if err := app.RedisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}
	
	log.Println("Application closed successfully")
}