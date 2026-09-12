// internal/app/wire.go

//go:build wireinject
// +build wireinject

package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"gorm.io/gorm"

	// Account Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account"
	accountHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/handler"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	accountDomain  "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	

	// Auth Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth"
	authHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"


	// Events Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events"
	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	eventsDomain  "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"


	// Media Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	// Notification Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"


	// Team Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team"
	teamHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/delivery/handler"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"

	// Shared
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/queue"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/ai"

)


// ============================================================
// APP DEPENDENCIES
// ============================================================

type AppDependencies struct {
	Config             *config.Config
	DB                 *gorm.DB
	App                *fiber.App
	StorageClient      *storage.Client
	RedisClient        *redis.Client
	Enforcer           *authorization.Enforcer
	PermissionChecker  authDomain.PermissionChecker
	RoleManager        authDomain.RoleManager
	PolicyManager      authDomain.PolicyManager
	AuthTokenService   authDomain.TokenService
	AuthService        authService.Service
	AccountService     accountService.Service
	TeamService        teamService.Service
	EventsService      eventsService.Service
	MediaService       mediaService.Service
	NotificationSvc    notificationDomain.NotificationService
	AuthHandler        *authHandler.AuthHandler
	AccountHandler     *accountHandler.AccountHandler
	TeamHandler        *teamHandler.TeamHandler
	EventsHandler      *eventsHandler.EventHandler
	AIService          teamService.AIService
	EventsAIService		eventsService.AIService
	OrganizerProvider	eventsDomain.OrganizerProvider
	AccountsPermissionChecker		accountDomain.PermissionChecker
	AIClient					*ai.Client
}

// ============================================================
// INITIALIZE APP
// ============================================================

func InitializeApp() (*AppDependencies, error) {
	wire.Build(
		// ============================================================
		// SHARED INFRASTRUCTURE
		// ============================================================
		config.ProviderSet,
		database.ProviderSet,
		queue.ProviderSet,
		redis.ProviderSet,
		storage.ProviderSet,
		ai.ProviderSet,  

		// ============================================================
		// APP-SPECIFIC
		// ============================================================
		provideFiberAppWithMiddleware,

		// ============================================================
		// MODULES
		// ============================================================
		auth.ProviderSet,
		account.ProviderSet,
		events.ProviderSet,
		media.ProviderSet,
		notification.ProviderSet,
		team.ProviderSet,

		// ============================================================
		// CROSS-MODULE ADAPTERS - AUTH
		// ============================================================
		NewAuthNotificationAdapter,
		NewQueueAdapter,
		NewAuthTeamAdapter,

		// ============================================================
		// CROSS-MODULE ADAPTERS - ACCOUNT
		// ============================================================
		NewAccountAuthAdapter,
		NewAccountPermissionAdapter,
		NewAccountMediaAdapter,

		// ============================================================
		// CROSS-MODULE ADAPTERS - EVENTS
		// ============================================================
		NewEventsPermissionAdapter,
		NewEventsUserInfoAdapter,
		NewEventsMediaAdapter,

		// ============================================================
		// CROSS-MODULE ADAPTERS - TEAM
		// ============================================================
		NewTeamAuthAdapter,
		NewTeamCasbinAdapter,
		NewTeamNotificationAdapter,
		NewAccountNotificationAdapter,


		provideOrganizerProvider,


		// ============================================================
		// FINAL APP DEPENDENCIES
		// ============================================================
		provideAppDependencies,
	)
	return nil, nil
}