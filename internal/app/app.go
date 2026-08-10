package app

import (
	"context"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authmiddleware"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/server"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
)

// App wraps the application dependencies
type App struct {
	*AppDependencies
}

// NewApp creates and initializes the application
func NewApp() (*App, error) {
	cfg := config.Load()

	if err := redis.Init(cfg.Redis.URL); err != nil {
		return nil, err
	}
	log.Println("Redis initialized successfully")

	deps, err := InitializeApp()
	if err != nil {
		return nil, err
	}
	return &App{AppDependencies: deps}, nil
}

func (app *App) Init(ctx context.Context) error {
	log.Println("Initializing modules...")
	if app.AuthService != nil {
		log.Println("Auth module ready")
	}
	if app.AccountService != nil {
		log.Println("Account module ready")
	}
	if app.EventsService != nil {
		log.Println("Events module ready")
	}
	if app.MediaService != nil {
		log.Println("Media module ready")
	}
	log.Println("All modules initialized successfully")
	return nil
}

// SetupRoutes registers all API routes
func (app *App) SetupRoutes() {
	authMiddleware := authmiddleware.AuthMiddleware(app.Config.JWT.Secret)
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

func (app *App) Run() error {
	log.Printf("Server starting on port %s", app.Config.Server.Port)
	return app.App.Listen(":" + app.Config.Server.Port)
}

func (app *App) Close() {
	log.Println("Shutting down application...")
	if app.Enforcer != nil {
		app.Enforcer.Close()
	}
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
	redis.Close()
	log.Println("Application closed successfully")
}