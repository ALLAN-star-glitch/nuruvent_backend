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
	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/handler"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"

	// Attendance Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance"
	attendanceHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/delivery/http"
	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"

	// Auth Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth"
	authHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"

	// Events Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events"
	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"

	// Media Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	// Notification Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"

	// Payment Module
	payment "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment"
	paymentHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/delivery/http"
	paymentService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"

	// Team Module
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team"
	teamHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/delivery/handler"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"

	// Registration Module
	registration "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration"
	regHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/delivery/http"
	regService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"

	// Shared
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/ai"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/database"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
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
	PermissionChecker authDomain.PermissionChecker
	RoleManager       authDomain.RoleManager
	PolicyManager     authDomain.PolicyManager
	AuthTokenService  authDomain.TokenService
	AuthService       authService.Service
	AccountService    accountService.Service
	TeamService       teamService.Service
	EventsService     eventsService.Service
	MediaService      mediaService.Service
	NotificationSvc   notificationDomain.NotificationService
	AuthHandler       *authHandler.AuthHandler
	AccountHandler    *accountHandler.AccountHandler
	TeamHandler       *teamHandler.TeamHandler
	EventsHandler     *eventsHandler.EventHandler
	AIService         teamService.AIService
	EventsAIService   eventsService.AIService
	OrganizerProvider eventsDomain.OrganizerProvider
	AccountsPermissionChecker accountDomain.PermissionChecker
	AIClient          *ai.Client

	// Registration Module
	RegHandler *regHandler.Handler
	RegService regService.Service

	// Payment Module
	PaymentHandler *paymentHandler.Handler
	OrderHandler   *paymentHandler.OrderHandler
	PaymentService paymentService.Service

	// Attendance Module
	AttendanceHandler *attendanceHandler.Handlers   // 👈 added
	AttendanceService attendanceService.Service     // 👈 added
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
		id.ProviderSet,

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
		registration.ProviderSet,
		payment.ProviderSet,
		attendance.ProviderSet,   // 👈 added

		// ============================================================
		// CROSS-MODULE ADAPTERS - AUTH
		// ============================================================
		NewAuthNotificationAdapter,
		NewQueueAdapter,
		NewAuthTeamAdapter,

		// ============================================================
		// CROSS-MODULE ADAPTERS - PAYMENT
		// ============================================================
		NewPaymentRegistrationConfirmer,
		NewPaymentProviderRegistry,
		NewPaymentUserEmailResolver,
		NewPaymentNotifier,
		NewPaymentPricingResolver,

		// ============================================================
		// CROSS-MODULE ADAPTERS - EVENTS
		// ============================================================
		NewEventsPermissionAdapter,
		NewEventsUserInfoAdapter,
		NewEventsMediaAdapter,
		NewEventsAttendanceRegistrar, 
		NewRegistrationAttendanceRegistrar,
		NewRegistrationUserInfoAdapter,

		// ============================================================
		// CROSS-MODULE ADAPTERS - TEAM
		// ============================================================
		NewTeamAuthAdapter,
		NewTeamCasbinAdapter,
		NewTeamNotificationAdapter,
		NewAccountNotificationAdapter,

		// ============================================================
		// CROSS-MODULE ADAPTERS - REGISTRATION
		// ============================================================
		NewRegistrableResolver,

		// ============================================================
		// CROSS-MODULE ADAPTERS - ACCOUNT
		// ============================================================
		NewAccountAuthAdapter,
		NewAccountPermissionAdapter,
		NewAccountMediaAdapter,

		provideOrganizerProvider,

		// ============================================================
		// FINAL APP DEPENDENCIES
		// ============================================================
		provideAppDependencies,
	)
	return nil, nil
}