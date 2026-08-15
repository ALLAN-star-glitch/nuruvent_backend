// internal/app/wire.go

//go:build wireinject
// +build wireinject

package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account"
	accountHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/acchandler"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth"
	authHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events"
	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification"
	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/queue"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
)

type AppDependencies struct {
	Config         *config.Config
	DB             *gorm.DB
	App            *fiber.App
	StorageClient  *storage.Client
	RedisClient    *redis.Client          // ✅ Added
	Enforcer       *authorization.Enforcer
	Notification   notificationdomain.NotificationService
	AuthService    authService.Service
	AccountService accountService.Service
	EventsService  eventsService.Service
	MediaService   mediaService.Service
	AuthHandler    *authHandler.AuthHandler
	AccountHandler *accountHandler.AccountHandler
	EventsHandler  *eventsHandler.EventHandler
}

func InitializeApp() (*AppDependencies, error) {
	wire.Build(
		// ============================================================
		// SHARED INFRASTRUCTURE - PROVIDED ONCE
		// ============================================================
		config.ProviderSet,
		database.ProviderSet,
		queue.ProviderSet,
		redis.ProviderSet,      // ✅ Provides *redis.Client
		storage.ProviderSet,

		// ============================================================
		// APP-SPECIFIC
		// ============================================================
		provideFiberAppWithMiddleware,
		provideAppDependencies,

		// ============================================================
		// MODULES
		// ============================================================
		notification.ProviderSet,
		auth.ProviderSet,
		account.ProviderSet,
		events.ProviderSet,
		media.ProviderSet,

		// ============================================================
		// CROSS-MODULE ADAPTERS
		// ============================================================
		NewAuthNotificationAdapter,
		NewQueueAdapter,
		NewAccountPermissionAdapter,
		NewEventsPermissionAdapter,
		NewEventsMediaAdapter,
	)
	return nil, nil
}