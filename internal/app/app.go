// internal/app/app.go

package app

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/server"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/redis"
)

type App struct {
	*AppDependencies
}

func NewApp() (*App, error) {
	// Initialize Redis first
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
	if err := app.AuthorizationModule.Init(ctx); err != nil {
		return err
	}
	if err := app.AuthModule.Init(ctx); err != nil {
		return err
	}
	if err := app.MediaModule.Init(ctx); err != nil {
		return err
	}
	if err := app.EventsModule.Init(ctx); err != nil {
		return err
	}
	log.Println("All modules initialized successfully")
	return nil
}

func (app *App) SetupRoutes() {
	server.SetupRoutes(
		app.App,
		app.Config,
		app.AuthHandler,
		app.Enforcer,
		app.EventsModule,
	)
	log.Println("Routes registered successfully")
}

func (app *App) Run() error {
	log.Printf("Server starting on port %s", app.Config.Server.Port)
	return app.App.Listen(":" + app.Config.Server.Port)
}

func (app *App) Close() {
	if app.AuthorizationModule != nil {
		app.AuthorizationModule.Close()
	}
	if app.AuthModule != nil {
		app.AuthModule.Close()
	}
	if app.MediaModule != nil {
		app.MediaModule.Close()
	}
	if app.EventsModule != nil {
		app.EventsModule.Close()
	}
	if err := database.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
	redis.Close()
	log.Println("Application closed successfully")
}