// internal/app/providers.go

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/gorm"

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

			// Production - Your custom domain
			"https://nuruvent.com",
			"https://www.nuruvent.com",

			// Vercel preview deployments
			"https://nuruvent.vercel.app",
			"https://*.vercel.app",

			// Staging
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
	eventsService eventsService.Service,
	mediaService mediaService.Service,
	authHandler *authHandler.AuthHandler,
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
		EventsService:    eventsService,
		MediaService:     mediaService,
		AuthHandler:      authHandler,
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


// ============================================================
// EVENTS PERMISSION ADAPTER - Updated for new team-based auth
// ============================================================

type EventsPermissionAdapter struct {
    permSvc authDomain.PermissionService
}

func NewEventsPermissionAdapter(permSvc authDomain.PermissionService) eventsDomain.PermissionChecker {
    return &EventsPermissionAdapter{permSvc: permSvc}
}

// ✅ CanManagePersonalEvent - checks if user can manage their own personal events
func (a *EventsPermissionAdapter) CanManagePersonalEvent(ctx context.Context, userID string) bool {
    // User can manage their own personal events if they are the account admin
    return a.permSvc.IsPersonalTeamAdmin(ctx, userID, userID)
}

func (a *EventsPermissionAdapter) CanManageEvent(ctx context.Context, userID, teamID string) bool {
    return a.permSvc.CanManageTeamEvent(ctx, teamID, userID)
}

func (a *EventsPermissionAdapter) CanUpdateEvent(ctx context.Context, userID, teamID string) bool {
    // Try personal team first
    if a.permSvc.HasPersonalTeamPermission(ctx, userID, teamID, "event", "update") {
        return true
    }
    return a.permSvc.HasInstitutionTeamPermission(ctx, userID, teamID, "event", "update")
}

func (a *EventsPermissionAdapter) CanDeleteEvent(ctx context.Context, userID, teamID string) bool {
    if a.permSvc.HasPersonalTeamPermission(ctx, userID, teamID, "event", "delete") {
        return true
    }
    return a.permSvc.HasInstitutionTeamPermission(ctx, userID, teamID, "event", "delete")
}

func (a *EventsPermissionAdapter) CanViewEvent(ctx context.Context, userID, teamID string) bool {
    if a.permSvc.HasPersonalTeamPermission(ctx, userID, teamID, "event", "read") {
        return true
    }
    return a.permSvc.HasInstitutionTeamPermission(ctx, userID, teamID, "event", "read")
}

// ============================================================
// EVENTS MEDIA ADAPTER - Using pure domain types
// ============================================================

type EventsMediaAdapter struct {
	mediaSvc mediaService.Service
}

func NewEventsMediaAdapter(mediaSvc mediaService.Service) eventsDomain.MediaService {
	return &EventsMediaAdapter{mediaSvc: mediaSvc}
}

func (a *EventsMediaAdapter) UploadFile(ctx context.Context, cmd eventsDomain.UploadMediaCommand) (*eventsDomain.MediaInfo, error) {
	mediaCmd := mediaService.UploadCommand{
		File:          cmd.File,
		FileName:      cmd.FileName,
		ContentType:   cmd.ContentType,
		MediaTypeName: cmd.MediaTypeName,
		EntityID:      cmd.EntityID,
		UploadedBy:    cmd.UploadedBy,
	}

	media, err := a.mediaSvc.UploadFile(ctx, mediaCmd)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, nil
	}

	return &eventsDomain.MediaInfo{
		ID:         media.ID,
		URL:        media.URL,
		MediaType:  cmd.MediaTypeName,
		EntityID:   media.EntityID,
		UploadedBy: media.UploadedBy,
		CreatedAt:  media.CreatedAt.Format(time.RFC3339),
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

	mediaTypeName := ""
	mediaType, err := a.mediaSvc.GetMediaTypeByID(ctx, media.MediaTypeID)
	if err == nil && mediaType != nil {
		mediaTypeName = mediaType.Name
	}

	return &eventsDomain.MediaInfo{
		ID:         media.ID,
		URL:        media.URL,
		MediaType:  mediaTypeName,
		EntityID:   media.EntityID,
		UploadedBy: media.UploadedBy,
		CreatedAt:  media.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (a *EventsMediaAdapter) GetMediaByEntity(ctx context.Context, entityID string) ([]*eventsDomain.MediaInfo, error) {
	mediaList, _, err := a.mediaSvc.GetMediaByEntity(ctx, entityID, 1, 100)
	if err != nil {
		return nil, err
	}

	if len(mediaList) == 0 {
		return []*eventsDomain.MediaInfo{}, nil
	}

	result := make([]*eventsDomain.MediaInfo, len(mediaList))
	for i, media := range mediaList {
		mediaTypeName := ""
		mediaType, err := a.mediaSvc.GetMediaTypeByID(ctx, media.MediaTypeID)
		if err == nil && mediaType != nil {
			mediaTypeName = mediaType.Name
		}

		result[i] = &eventsDomain.MediaInfo{
			ID:         media.ID,
			URL:        media.URL,
			MediaType:  mediaTypeName,
			EntityID:   media.EntityID,
			UploadedBy: media.UploadedBy,
			CreatedAt:  media.CreatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (a *EventsMediaAdapter) GetMediaTypeByName(ctx context.Context, name string) (*eventsDomain.MediaTypeInfo, error) {
	mediaType, err := a.mediaSvc.GetMediaTypeByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, nil
	}
	return &eventsDomain.MediaTypeInfo{
		ID:   mediaType.ID,
		Name: mediaType.Name,
		Slug: mediaType.Slug,
	}, nil
}

func (a *EventsMediaAdapter) DeleteFile(ctx context.Context, id string) error {
	return a.mediaSvc.DeleteFile(ctx, id)
}

func (a *EventsMediaAdapter) DeleteFilesByEntity(ctx context.Context, entityID string) error {
	return a.mediaSvc.DeleteFilesByEntity(ctx, entityID)
}

func (a *EventsMediaAdapter) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error {
	return a.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, entityID, mediaTypeID)
}

// ============================================================
// BYTES READER WRAPPER
// ============================================================

type bytesReaderWrapper struct {
	*bytes.Reader
}

func (b *bytesReaderWrapper) Close() error {
	return nil
}

func (b *bytesReaderWrapper) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (b *bytesReaderWrapper) Stat() (fs.FileInfo, error) {
	return nil, nil
}

// ============================================================
// AUTH NOTIFICATION ADAPTER - UNIFIED
// ============================================================

type AuthNotificationAdapter struct {
	notifSvc notificationdomain.NotificationService
}

func NewAuthNotificationAdapter(notifSvc notificationdomain.NotificationService) authDomain.NotificationService {
	return &AuthNotificationAdapter{notifSvc: notifSvc}
}

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

// Welcome emails
func (a *AuthNotificationAdapter) SendIndividualWelcome(ctx context.Context, req authDomain.SendWelcomeRequest) error {
	notifReq := notificationdomain.SendWelcomeRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendIndividualWelcome(ctx, notifReq)
}

func (a *AuthNotificationAdapter) SendInstitutionWelcome(ctx context.Context, req authDomain.SendInstitutionWelcomeRequest) error {
	notifReq := notificationdomain.SendInstitutionWelcomeRequest{
		To:               req.To,
		AdminName:        req.AdminName,
		InstitutionName:  req.InstitutionName,
		InstitutionEmail: req.InstitutionEmail,
	}
	return a.notifSvc.SendInstitutionWelcome(ctx, notifReq)
}

func (a *AuthNotificationAdapter) SendInstitutionKYCWelcome(ctx context.Context, req authDomain.SendInstitutionKYCWelcomeRequest) error {
	notifReq := notificationdomain.SendInstitutionKYCWelcomeRequest{
		To:              req.To,
		AdminName:       req.AdminName,
		InstitutionName: req.InstitutionName,
		InstitutionType: req.InstitutionType,
	}
	return a.notifSvc.SendInstitutionKYCWelcome(ctx, notifReq)
}

func (a *AuthNotificationAdapter) SendNewInstitutionAccountNotification(ctx context.Context, req authDomain.SendNewInstitutionAccountRegistrationRequest) error {
	notifReq := notificationdomain.SendNewInstitutionAccountRegistrationRequest{
		TO:                  req.To,
		NewAccountAdminName: req.NewAccountAdminName,
		InstitutionName:     req.InstitutionName,
		InstitutionType:     req.InstitutionType,
	}
	return a.notifSvc.SendNewInstitutionAccountNotification(ctx, notifReq)
}

func (a *AuthNotificationAdapter) SendNewPersonalAccountNotification(ctx context.Context, req authDomain.SendNewPersonalAccountRegistrationRequest) error {
	notifReq := notificationdomain.SendNewPersonalAccountRegistrationRequest{
		To:                  req.To,
		NewAccountAdminName: req.NewAccountAdminName,
	}
	return a.notifSvc.SendNewPersonalAccountNotification(ctx, notifReq)
}

// Password reset confirm
func (a *AuthNotificationAdapter) SendPasswordResetConfirm(ctx context.Context, req authDomain.SendPasswordResetConfirmRequest) error {
	notifReq := notificationdomain.SendPasswordResetConfirmRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendPasswordResetConfirm(ctx, notifReq)
}

// Login notification
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