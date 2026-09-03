// internal/app/wire.go

//go:build wireinject
// +build wireinject

package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth"
	authHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	profileHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/delivery/handler"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification"
	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"

	// Profile module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile"
	profileDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/queue"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"
)

// ============================================================
// APP DEPENDENCIES
// ============================================================

type AppDependencies struct {
	Config            *config.Config
	DB                *gorm.DB
	App               *fiber.App
	StorageClient     *storage.Client
	RedisClient       *redis.Client
	Enforcer          *authorization.Enforcer
	PermissionChecker authdomain.PermissionChecker
	RoleManager       authdomain.RoleManager
	PolicyManager     authdomain.PolicyManager
	AuthTokenService  authdomain.TokenService
	Notification      notificationdomain.NotificationService
	ProfileService    profileDomain.Service
	AuthService       authService.Service
	EventsService     eventsService.Service
	MediaService      mediaService.Service
	AuthHandler       *authHandler.AuthHandler
	EventsHandler     *eventsHandler.EventHandler
	ProfileHandler    *profileHandler.ProfileHandler
}

// ============================================================
// PROVIDER FUNCTIONS
// ============================================================

// provideEventsPermissionAdapter creates the events permission adapter
func provideEventsPermissionAdapter(permChecker authdomain.PermissionChecker) eventsDomain.PermissionChecker {
	return NewEventsPermissionAdapter(permChecker)
}

// provideEventsProfileAdapter creates the events profile adapter
func provideEventsProfileAdapter(profileSvc profileDomain.Service) eventsDomain.UserInfoProvider {
	return NewEventsProfileAdapter(profileSvc)
}

// provideEventsMediaAdapter creates the events media adapter
func provideEventsMediaAdapter(mediaSvc mediaService.Service) eventsDomain.MediaService {
	return NewEventsMediaAdapter(mediaSvc)
}

// ✅ provideProfilePermissionAdapter creates the profile permission adapter
func provideProfilePermissionAdapter(permChecker authdomain.PermissionChecker) profileDomain.PermissionChecker {
	return NewProfilePermissionAdapter(permChecker)
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

		// ============================================================
		// APP-SPECIFIC
		// ============================================================
		provideFiberAppWithMiddleware,

		// ============================================================
		// MODULES
		// ============================================================
		auth.ProviderSet,
		events.ProviderSet,
		media.ProviderSet,
		notification.ProviderSet,
		profile.ProviderSet,

		// ============================================================
		// CROSS-MODULE ADAPTERS
		// ============================================================
		NewAuthNotificationAdapter,
		NewQueueAdapter,
		provideEventsPermissionAdapter,
		provideEventsProfileAdapter,
		provideEventsMediaAdapter,
		provideProfilePermissionAdapter, // ✅ Add this

		// ============================================================
		// FINAL APP DEPENDENCIES
		// ============================================================
		provideAppDependencies,
	)
	return nil, nil
}