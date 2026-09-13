// cmd/api/main.go

package main

import (
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/worker"
)

// @title           Nuruvent API
// @version         1.0.0
// @description     Light Your Training Events. Illuminate Your Growth.
// @description     The all-in-one platform for training events in Kenya.
// @termsOfService  https://nuruvent.com/terms

// @contact.name   Nuruvent Support
// @contact.email  hello@nuruvent.com
// @contact.url    https://nuruvent.com

// @license.name   Proprietary
// @license.url    https://nuruvent.com/license

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

// cmd/api/main.go

func main() {
    // 1. Load config
    cfg := config.Load()
    log.Printf("Starting Nuruvent API in %s mode", cfg.Environment)

    // 2. Create application (API)
    application, err := app.NewApp()
    if err != nil {
        log.Fatalf("Failed to create application: %v", err)
    }
    defer application.Close()

    // 3. Setup routes
    application.SetupRoutes()

    // 4. Start embedded Asynq worker
    stopWorker := worker.StartEmbeddedWorker(cfg)
    defer stopWorker()

    // 5. Start server (blocking)
    log.Println("🚀 Starting HTTP API server...")
    if err := application.Run(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}