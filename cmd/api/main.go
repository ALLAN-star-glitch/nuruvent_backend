// cmd/api/main.go

package main

import (
	"context"
	"log"
	_ "time/tzdata" // embed the IANA timezone database in the binary

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

	// 3. Reconcile Casbin grouping rules with account_members.
	//
	// SQL (account_members.role) is the source of truth; Casbin
	// grouping rules are a derived cache. This one-shot pass clears
	// any drift accumulated since the last boot. The mutation hook
	// (RoleAssignment.ReconcileAll) keeps things aligned during the
	// run, after every add / remove / update-role / leave.
	//
	// Non-fatal: if it fails, the server still boots and the next
	// role mutation reconciles.
		if application.AccountService != nil {
		if err := application.AccountService.ReconcileRoles(context.Background()); err != nil {
			log.Printf("⚠️  Startup reconcile failed: %v", err)
		}
	} else {
		log.Println("ℹ️  Startup reconcile skipped: AccountService not wired")
	}

	// 4. Setup routes
	application.SetupRoutes()

	// 5. Start embedded Asynq worker
	stopWorker := worker.StartEmbeddedWorker(cfg, application.CertificateService)
	defer stopWorker()

	// 6. Start the Google Meet attendance poller.
	//    Off by default so dev environments don't hit Google's API
	//    unattended. Enable with AUTO_FETCH_MEET_ATTENDANCE=true.
	if cfg.AutoFetchMeetAttendance {
		pollerCtx, cancelPoller := context.WithCancel(context.Background())
		defer cancelPoller()

		go func() {
			if err := application.VideoService.AutoFetchGoogleMeetAttendance(pollerCtx); err != nil {
				log.Printf("auto-fetch meet poller exited: %v", err)
			}
		}()
		log.Println("✅ Meet attendance poller started")
	} else {
		log.Println("ℹ️  Meet attendance poller disabled (set AUTO_FETCH_MEET_ATTENDANCE=true to enable)")
	}

	// 7. Start server (blocking)
	log.Println("🚀 Starting HTTP API server...")
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}