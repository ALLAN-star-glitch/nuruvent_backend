// internal/app/providers.go

package app

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/gorm"

	// Auth Module
	authHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"

	// Account Module
	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/handler"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"

	// Events Module
	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"

	// Media Module
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	// Notification Module
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"

	// Team Module
	teamHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/delivery/handler"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"

	// Shared
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/ai"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
)

// ============================================================
// APP-SPECIFIC PROVIDERS
// ============================================================

// provideFiberAppWithMiddleware creates the Fiber app with middleware
func provideFiberAppWithMiddleware() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Nuruvent API",
		ServerHeader: "Nuruvent",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:8080",
			"https://nuruvent.com",
			"https://www.nuruvent.com",
			"https://nuruvent.vercel.app",
			"https://*.vercel.app",
			"https://staging.nuruvent.com",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))
	return app
}



// ============================================================
// APP DEPENDENCIES
// ============================================================

// provideAppDependencies assembles the root application dependencies
func provideAppDependencies(
	cfg *config.Config,
	db *gorm.DB,
	app *fiber.App,
	storageClient *storage.Client,
	redisClient *redis.Client,
	aiClient *ai.Client,
	enforcer *authorization.Enforcer,
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
	policyManager authDomain.PolicyManager,
	authSvc authService.Service,
	authTokenService authDomain.TokenService,
	accountSvc accountService.Service,
	teamSvc teamService.Service,
	eventsSvc eventsService.Service,
	mediaSvc mediaService.Service,
	notificationSvc notificationDomain.NotificationService,
	authHndlr *authHandler.AuthHandler,
	accountHndlr *accountHandler.AccountHandler,
	teamHndlr *teamHandler.TeamHandler,
	eventsHndlr *eventsHandler.EventHandler,
	// ✅ Added AI Service
	aiSvc teamService.AIService,
	eventsAIService  eventsService.AIService,
	organizerProvider eventsDomain.OrganizerProvider,
	accountsPermissionChecker accountDomain.PermissionChecker,

) *AppDependencies {
	return &AppDependencies{
		Config:            cfg,
		DB:                db,
		App:               app,
		StorageClient:     storageClient,
		RedisClient:       redisClient,
		Enforcer:          enforcer,
		PermissionChecker: permChecker,
		RoleManager:       roleManager,
		PolicyManager:     policyManager,
		AuthService:       authSvc,
		AuthTokenService:  authTokenService,
		AccountService:    accountSvc,
		TeamService:       teamSvc,
		EventsService:     eventsSvc,
		MediaService:      mediaSvc,
		NotificationSvc:   notificationSvc,
		AuthHandler:       authHndlr,
		AccountHandler:    accountHndlr,
		TeamHandler:       teamHndlr,
		EventsHandler:     eventsHndlr,
		AIService:         aiSvc, // ✅ Added AI Service
		OrganizerProvider:	organizerProvider,
		AccountsPermissionChecker: accountsPermissionChecker,
		AIClient:	aiClient,
		EventsAIService: eventsAIService,
	
	}
}