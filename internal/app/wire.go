//go:build wireinject
// +build wireinject

package app

import (
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/events"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/media"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/server"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/redis"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/storage"
	"github.com/google/wire"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/gorm"
)

// AppDependencies holds all dependencies for the application
type AppDependencies struct {
	Config              *config.Config
	DB                  *gorm.DB
	App                 *fiber.App
	AuthorizationModule *authorization.Module
	AuthModule          *auth.Module
	MediaModule         *media.Module
	EventsModule        *events.Module
	AuthorizationService *authorization.Service
	Enforcer            *authorization.Enforcer
	AuthHandler         *auth.Handler
	StorageClient       *storage.Client
}

// InitializeApp initializes the entire application using Wire
func InitializeApp() (*AppDependencies, error) {
	wire.Build(
		// Configuration
		config.Load,

		// Database
		database.Connect,

		// Redis
		provideRedis,

		// Storage Client
		provideStorageClient,

		// Fiber App
		provideFiberApp,

		// Middleware
		provideMiddleware,

		// Modules
		authorization.ProviderSet,
		media.ProviderSet,
		auth.ProviderSet,
		events.ProviderSet,

		// Server
		server.ProviderSet,

		// App
		provideAppDependencies,
	)
	return nil, nil
}

// ================================================
// PROVIDERS
// ================================================

func provideRedis(cfg *config.Config) error {
	return redis.Init(cfg.Redis.URL)
}

func provideStorageClient(cfg *config.Config) *storage.Client {
	return storage.NewClient(
		cfg.Supabase.URL,
		cfg.Supabase.SecretKey,
		storage.BucketConfig{
			Events:       cfg.Supabase.BucketEvent,
			Businesses:   cfg.Supabase.BucketBusiness,
			Profiles:     cfg.Supabase.BucketProfile,
			Certificates: cfg.Supabase.BucketCertificate,
			Recordings:   cfg.Supabase.BucketRecording,
		},
	)
}

func provideFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      "Nuruvent API",
		ServerHeader: "Nuruvent",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
}

func provideMiddleware(app *fiber.App) *fiber.App {
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
	return app
}

func provideAppDependencies(
	cfg *config.Config,
	db *gorm.DB,
	app *fiber.App,
	authModule *auth.Module,
	authHandler *auth.Handler,
	authorizationModule *authorization.Module,
	authorizationService *authorization.Service,
	enforcer *authorization.Enforcer,
	mediaModule *media.Module,
	eventsModule *events.Module,
	storageClient *storage.Client,
) *AppDependencies {
	return &AppDependencies{
		Config:               cfg,
		DB:                   db,
		App:                  app,
		AuthModule:           authModule,
		AuthHandler:          authHandler,
		AuthorizationModule:  authorizationModule,
		AuthorizationService: authorizationService,
		Enforcer:             enforcer,
		MediaModule:          mediaModule,
		EventsModule:         eventsModule,
		StorageClient:        storageClient,
	}
}