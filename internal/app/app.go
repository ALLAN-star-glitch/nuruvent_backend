package app

import (
	"context"
	"log"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/server"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/redis"
	
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// App represents the application with all dependencies
type App struct {
	Config   *config.Config
	App      *fiber.App

	// Modules
	AuthorizationModule *authorization.Module
	AuthModule          *auth.Module
	BusinessModule      *business.Module

	// Services (convenience access)
	AuthorizationService *authorization.Service
	Enforcer            *authorization.Enforcer
	AuthHandler         *auth.Handler
}

// NewApp creates a new application instance with all dependencies initialized
func NewApp() (*App, error) {
	// 1. Load configuration
	cfg := config.Load()
	log.Printf("Starting Nuruvent API in %s mode", cfg.Environment)

	// 2. Initialize Redis
	if err := redis.Init(cfg.Redis.URL); err != nil {
		return nil, err
	}
	log.Println("Redis initialized successfully")

	// 3. Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}
	log.Println("Database connected successfully")

	// 4. Initialize authorization Module
	permModule, err := authorization.NewModule(db, cfg)
	if err != nil {
		return nil, err
	}
	enforcer := permModule.GetEnforcer()
	permService := permModule.GetService()

	// 5. Initialize Business Module (needed for auth)
	businessModule := business.NewModule(cfg, permService, enforcer)
	businessService := businessModule.GetBusinessService()

	// 6. Initialize Auth Module with business service
	authModule := auth.NewModule(cfg, permService, businessService)
	authHandler := authModule.GetHandler()

	// 7. Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Nuruvent API",
		ServerHeader: "Nuruvent",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// 8. Setup middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:8080",
			"https://nuruvent.vercel.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	return &App{
		Config:               cfg,
		App:                  app,
		AuthorizationModule:  permModule,
		AuthModule:           authModule,
		BusinessModule:       businessModule,
		AuthorizationService: permService,
		Enforcer:             enforcer,
		AuthHandler:          authHandler,
	}, nil
}

// Init initializes all modules
func (app *App) Init(ctx context.Context) error {
	// Initialize authorization module
	if err := app.AuthorizationModule.Init(ctx); err != nil {
		return err
	}

	// Initialize auth module
	if err := app.AuthModule.Init(ctx); err != nil {
		return err
	}

	// Initialize business module
	if err := app.BusinessModule.Init(ctx); err != nil {
		return err
	}

	log.Println("All modules initialized successfully")
	return nil
}

// SetupRoutes registers all routes
func (app *App) SetupRoutes() {
	// Setup API routes with all handlers (includes Swagger)
	server.SetupRoutes(
		app.App,
		app.Config,
		app.AuthHandler,
		app.Enforcer,
		app.BusinessModule,
	)

	log.Println("Routes registered successfully")
}

// Run starts the server
func (app *App) Run() error {
	log.Printf("Server starting on port %s", app.Config.Server.Port)
	log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", app.Config.Server.Port)
	return app.App.Listen(":" + app.Config.Server.Port)
}

// Close cleans up all resources
func (app *App) Close() {
	log.Println("Closing application...")

	// Close modules
	if app.AuthorizationModule != nil {
		app.AuthorizationModule.Close()
	}
	if app.AuthModule != nil {
		app.AuthModule.Close()
	}
	if app.BusinessModule != nil {
		app.BusinessModule.Close()
	}

	// Close Redis
	redis.Close()

	// Close database connection
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Application closed successfully")
}