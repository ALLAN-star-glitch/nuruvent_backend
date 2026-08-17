// internal/app/providers.go

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/gorm"

	accountHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/delivery/acchandler"
	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"

	authHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdelivery/authhandler"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"

	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"

	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"

	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
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
			// Development
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://localhost:8080",

			// ✅ Production - Your custom domain
			"https://nuruvent.com",
			"https://www.nuruvent.com",

			// ✅ Vercel preview deployments
			"https://nuruvent.vercel.app",
			"https://*.vercel.app",

			// ✅ If you have staging
			"https://staging.nuruvent.com",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))
	return app
}

// provideAppDependencies assembles the root application dependencies
func provideAppDependencies(
	cfg *config.Config,
	db *gorm.DB,
	app *fiber.App,
	storageClient *storage.Client,
	redisClient *redis.Client,
	enforcer *authorization.Enforcer,
	authService authService.Service,
	authTokenService authDomain.TokenService,
	accountService accountService.Service,
	eventsService eventsService.Service,
	mediaService mediaService.Service,
	authHandler *authHandler.AuthHandler,
	accountHandler *accountHandler.AccountHandler,
	eventsHandler *eventsHandler.EventHandler,
) *AppDependencies {
	return &AppDependencies{
		Config:           cfg,
		DB:               db,
		App:              app,
		StorageClient:    storageClient,
		RedisClient:      redisClient,
		Enforcer:         enforcer,
		AuthService:      authService,
		AuthTokenService: authTokenService,
		AccountService:   accountService,
		EventsService:    eventsService,
		MediaService:     mediaService,
		AuthHandler:      authHandler,
		AccountHandler:   accountHandler,
		EventsHandler:    eventsHandler,
	}
}

// ============================================================
// CROSS-MODULE ADAPTERS
// ============================================================

// QueueAdapter adapts notificationdomain.TaskQueue to authDomain.QueueService
type QueueAdapter struct {
	queue notificationdomain.TaskQueue
}

func NewQueueAdapter(queue notificationdomain.TaskQueue) authDomain.QueueService {
	return &QueueAdapter{queue: queue}
}

func (a *QueueAdapter) Enqueue(ctx context.Context, task string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return a.queue.Enqueue(ctx, task, data)
}

func (a *QueueAdapter) EnqueueDelayed(ctx context.Context, task string, payload any, delaySeconds int) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	return a.queue.EnqueueDelayed(ctx, task, data, delaySeconds)
}

// AccountPermissionAdapter adapts authDomain.PermissionService to account domain.PermissionService
type AccountPermissionAdapter struct {
	permSvc authDomain.PermissionService
}

func NewAccountPermissionAdapter(permSvc authDomain.PermissionService) accountDomain.PermissionService {
	return &AccountPermissionAdapter{permSvc: permSvc}
}

func (a *AccountPermissionAdapter) AssignAccountAdminRole(ctx context.Context, accountID, userID string) error {
	return a.permSvc.AssignAccountAdminRole(ctx, accountID, userID)
}

func (a *AccountPermissionAdapter) AssignEventManagerRole(ctx context.Context, accountID, userID string) error {
	return a.permSvc.AssignEventManagerRole(ctx, accountID, userID)
}

func (a *AccountPermissionAdapter) AssignTeamMemberRole(ctx context.Context, accountID, userID string) error {
	return a.permSvc.AssignTeamMemberRole(ctx, accountID, userID)
}

func (a *AccountPermissionAdapter) RemoveRole(ctx context.Context, accountID, userID, role string) error {
	return a.permSvc.RemoveRole(ctx, accountID, userID, role)
}

func (a *AccountPermissionAdapter) HasPermission(ctx context.Context, userID, domain, resource, action string) bool {
	return a.permSvc.HasPermission(ctx, userID, domain, resource, action)
}

func (a *AccountPermissionAdapter) IsAccountAdmin(ctx context.Context, accountID, userID string) bool {
	return a.permSvc.IsAccountAdmin(ctx, accountID, userID)
}

func (a *AccountPermissionAdapter) IsEventManager(ctx context.Context, accountID, userID string) bool {
	return a.permSvc.IsEventManager(ctx, accountID, userID)
}

func (a *AccountPermissionAdapter) IsTeamMember(ctx context.Context, accountID, userID string) bool {
	return a.permSvc.IsTeamMember(ctx, accountID, userID)
}

// EventsPermissionAdapter adapts authDomain.PermissionService to events domain.PermissionChecker
type EventsPermissionAdapter struct {
	permSvc authDomain.PermissionService
}

func NewEventsPermissionAdapter(permSvc authDomain.PermissionService) eventsDomain.PermissionChecker {
	return &EventsPermissionAdapter{permSvc: permSvc}
}

func (a *EventsPermissionAdapter) CanManageEvent(ctx context.Context, userID, eventAccountID string) bool {
	return a.permSvc.CanManageEvent(ctx, eventAccountID, userID)
}

func (a *EventsPermissionAdapter) CanReadEvent(ctx context.Context, userID, eventAccountID string) bool {
	domain := authDomain.AccountDomain(eventAccountID)
	return a.permSvc.HasPermission(ctx, userID, domain, authDomain.ResourceEvent.String(), authDomain.ActionRead.String())
}

func (a *EventsPermissionAdapter) CanUpdateEvent(ctx context.Context, userID, eventAccountID string) bool {
	domain := authDomain.AccountDomain(eventAccountID)
	return a.permSvc.HasPermission(ctx, userID, domain, authDomain.ResourceEvent.String(), authDomain.ActionUpdate.String())
}

func (a *EventsPermissionAdapter) CanDeleteEvent(ctx context.Context, userID, eventAccountID string) bool {
	domain := authDomain.AccountDomain(eventAccountID)
	return a.permSvc.HasPermission(ctx, userID, domain, authDomain.ResourceEvent.String(), authDomain.ActionDelete.String())
}

// EventsMediaAdapter adapts media.Service to events domain.MediaService
type EventsMediaAdapter struct {
	mediaSvc mediaService.Service
}

func NewEventsMediaAdapter(mediaSvc mediaService.Service) eventsDomain.MediaService {
	return &EventsMediaAdapter{mediaSvc: mediaSvc}
}

func (a *EventsMediaAdapter) UploadFile(ctx context.Context, cmd eventsDomain.UploadMediaCommand) (*eventsDomain.MediaInfo, error) {
	var file multipart.File
	var fileHeader *multipart.FileHeader

	if cmd.File != nil {
		if f, ok := cmd.File.(multipart.File); ok {
			file = f
		}
	}

	if cmd.FileHeader != nil {
		if fh, ok := cmd.FileHeader.(*multipart.FileHeader); ok {
			fileHeader = fh
		}
	}

	mediaCmd := mediaService.UploadCommand{
		File:          file,
		FileHeader:    fileHeader,
		MediaTypeName: cmd.MediaTypeName,
		EntityID:      cmd.EntityID,
		UploadedBy:    cmd.UploadedBy,
	}

	media, err := a.mediaSvc.UploadFile(ctx, mediaCmd)
	if err != nil {
		return nil, err
	}

	return &eventsDomain.MediaInfo{
		ID:  media.ID,
		URL: media.URL,
	}, nil
}

func (a *EventsMediaAdapter) GetMediaByID(ctx context.Context, id string) (*eventsDomain.MediaInfo, error) {
	media, err := a.mediaSvc.GetMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, nil
	}
	return &eventsDomain.MediaInfo{
		ID:  media.ID,
		URL: media.URL,
	}, nil
}

func (a *EventsMediaAdapter) DeleteMediaByEntity(ctx context.Context, entityID string) error {
	return a.mediaSvc.DeleteFilesByEntity(ctx, entityID)
}

// ============================================================
// AUTH NOTIFICATION ADAPTER - UNIFIED
// ============================================================

// AuthNotificationAdapter adapts notificationdomain.NotificationService 
// to authDomain.NotificationService
type AuthNotificationAdapter struct {
	notifSvc notificationdomain.NotificationService
}

func NewAuthNotificationAdapter(notifSvc notificationdomain.NotificationService) authDomain.NotificationService {
	return &AuthNotificationAdapter{notifSvc: notifSvc}
}

// ============================================================
// UNIFIED OTP METHOD
// ============================================================

// SendOTP - Unified method for all OTP purposes
func (a *AuthNotificationAdapter) SendOTP(ctx context.Context, req authDomain.SendOTPRequest) error {
	notifReq := notificationdomain.SendOTPRequest{
		To:      req.To,
		Name:    req.Name,
		OTP:     req.OTP,
		Expires: req.Expires,
		Purpose: notificationdomain.VerificationPurpose(req.Purpose),
		Meta:    req.Meta,
	}
	return a.notifSvc.SendOTP(ctx, notifReq)
}

// ============================================================
// WELCOME EMAILS
// ============================================================

func (a *AuthNotificationAdapter) SendIndividualWelcome(ctx context.Context, req authDomain.SendWelcomeRequest) error {
	notifReq := notificationdomain.SendWelcomeRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendIndividualWelcome(ctx, notifReq)
}

func (a *AuthNotificationAdapter) SendInstitutionWelcome(ctx context.Context, req authDomain.SendInstitutionWelcomeRequest) error {
	notifReq := notificationdomain.SendInstitutionWelcomeRequest{
		To:              req.To,
		AdminName:       req.AdminName,
		InstitutionName: req.InstitutionName,
	}
	return a.notifSvc.SendInstitutionWelcome(ctx, notifReq)
}



// ============================================================
// PASSWORD RESET CONFIRM
// ============================================================

func (a *AuthNotificationAdapter) SendPasswordResetConfirm(ctx context.Context, req authDomain.SendPasswordResetConfirmRequest) error {
	notifReq := notificationdomain.SendPasswordResetConfirmRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendPasswordResetConfirm(ctx, notifReq)
}

// ============================================================
// LOGIN NOTIFICATION
// ============================================================

func (a *AuthNotificationAdapter) SendLoginNotification(ctx context.Context, req authDomain.SendLoginNotificationRequest) error {
	notifReq := notificationdomain.SendLoginNotificationRequest{
		To:        req.To,
		Name:      req.Name,
		Time:      req.Time,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
	return a.notifSvc.SendLoginNotification(ctx, notifReq)
}