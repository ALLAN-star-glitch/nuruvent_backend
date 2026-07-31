// cmd/api/main.go

package main

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/app"
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

func main() {
	// Create application
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer application.Close()

	// Initialize modules
	ctx := context.Background()
	if err := application.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Setup routes
	application.SetupRoutes()

	// Start server
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}