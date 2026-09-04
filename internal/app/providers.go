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

	profileDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"

	profileHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/delivery/handler"
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

// provideAppDependencies assembles the root application dependencies
func provideAppDependencies(
	cfg *config.Config,
	db *gorm.DB,
	app *fiber.App,
	storageClient *storage.Client,
	redisClient *redis.Client,
	enforcer *authorization.Enforcer,
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
	policyManager authDomain.PolicyManager,
	authService authService.Service,
	authTokenService authDomain.TokenService,
	eventsService eventsService.Service,
	profileSvc profileDomain.Service,
	mediaService mediaService.Service,
	authHandler *authHandler.AuthHandler,
	eventsHandler *eventsHandler.EventHandler,
	profileHandler *profileHandler.ProfileHandler,
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
		AuthService:       authService,
		AuthTokenService:  authTokenService,
		EventsService:     eventsService,
		ProfileService:    profileSvc,
		MediaService:      mediaService,
		AuthHandler:       authHandler,
		EventsHandler:     eventsHandler,
		ProfileHandler:    profileHandler,
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
// EVENTS PERMISSION ADAPTER - Implements eventsDomain.PermissionChecker
// ============================================================

type EventsPermissionAdapter struct {
	permSvc authDomain.PermissionChecker
}

func NewEventsPermissionAdapter(permSvc authDomain.PermissionChecker) eventsDomain.PermissionChecker {
	return &EventsPermissionAdapter{permSvc: permSvc}
}

// ============================================================
// CORE PERMISSION METHODS
// ============================================================

// HasPermission checks if a user has a specific permission in a scope
func (a *EventsPermissionAdapter) HasPermission(ctx context.Context, userID string, scope eventsDomain.Scope, resource, action string) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, resource, action)
}

// HasAnyPermission checks if a user has any of the given permissions in a scope
func (a *EventsPermissionAdapter) HasAnyPermission(ctx context.Context, userID string, scope eventsDomain.Scope, resource string, actions ...string) (bool, error) {
	authScope := a.convertScope(scope)
	for _, action := range actions {
		allowed, err := a.permSvc.HasPermission(ctx, userID, authScope, resource, action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the given permissions in a scope
func (a *EventsPermissionAdapter) HasAllPermissions(ctx context.Context, userID string, scope eventsDomain.Scope, resource string, actions ...string) (bool, error) {
	authScope := a.convertScope(scope)
	for _, action := range actions {
		allowed, err := a.permSvc.HasPermission(ctx, userID, authScope, resource, action)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// ============================================================
// CONVENIENCE METHODS - CREATE
// ============================================================

// CanCreateEvent checks if user can create events in a scope
func (a *EventsPermissionAdapter) CanCreateEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "create")
}

// ============================================================
// CONVENIENCE METHODS - READ
// ============================================================

// CanReadAllEvents checks if user can read ALL events in a scope
func (a *EventsPermissionAdapter) CanReadAllEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "read_all")
}

// CanReadOwnEvents checks if user can read OWN events in a scope
func (a *EventsPermissionAdapter) CanReadOwnEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "read_own")
}

// CanReadEvent checks if user can read events in a scope (ALL or OWN)
func (a *EventsPermissionAdapter) CanReadEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	allowed, err := a.CanReadAllEvents(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanReadOwnEvents(ctx, userID, scope)
}

// CanViewCreator checks if user can view creator details
func (a *EventsPermissionAdapter) CanViewCreator(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, "event", "view_creator")
}

// ============================================================
// CONVENIENCE METHODS - UPDATE
// ============================================================

// CanUpdateAllEvents checks if user can update ALL events in a scope
func (a *EventsPermissionAdapter) CanUpdateAllEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "update_all")
}

// CanUpdateOwnEvents checks if user can update OWN events in a scope
func (a *EventsPermissionAdapter) CanUpdateOwnEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "update_own")
}

// CanUpdateEvent checks if user can update events in a scope (ALL or OWN)
func (a *EventsPermissionAdapter) CanUpdateEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	allowed, err := a.CanUpdateAllEvents(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanUpdateOwnEvents(ctx, userID, scope)
}

// ============================================================
// CONVENIENCE METHODS - DELETE
// ============================================================

// CanDeleteAllEvents checks if user can delete ALL events in a scope
func (a *EventsPermissionAdapter) CanDeleteAllEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "delete_all")
}

// CanDeleteOwnEvents checks if user can delete OWN events in a scope
func (a *EventsPermissionAdapter) CanDeleteOwnEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "delete_own")
}

// CanDeleteEvent checks if user can delete events in a scope (ALL or OWN)
func (a *EventsPermissionAdapter) CanDeleteEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	allowed, err := a.CanDeleteAllEvents(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanDeleteOwnEvents(ctx, userID, scope)
}

// ============================================================
// CONVENIENCE METHODS - PUBLISH
// ============================================================

// CanPublishAllEvents checks if user can publish ALL events in a scope
func (a *EventsPermissionAdapter) CanPublishAllEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "publish_all")
}

// CanPublishOwnEvents checks if user can publish OWN events in a scope
func (a *EventsPermissionAdapter) CanPublishOwnEvents(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasPermission(ctx, userID, scope, "event", "publish_own")
}

// CanPublishEvent checks if user can publish events in a scope (ALL or OWN)
func (a *EventsPermissionAdapter) CanPublishEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	allowed, err := a.CanPublishAllEvents(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanPublishOwnEvents(ctx, userID, scope)
}

// ============================================================
// CONVENIENCE METHODS - MANAGEMENT
// ============================================================

// CanManageEvent checks if user can manage events in a scope (Admin/Manager only)
func (a *EventsPermissionAdapter) CanManageEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	return a.HasAnyPermission(ctx, userID, scope, "event",
		"update_all", "delete_all", "manage")
}

// CanViewEvent checks if user can view events in a scope (ALL or OWN)
func (a *EventsPermissionAdapter) CanViewEvent(ctx context.Context, userID string, scope eventsDomain.Scope) (bool, error) {
	allowed, err := a.CanReadAllEvents(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanReadOwnEvents(ctx, userID, scope)
}

// ============================================================
// HELPER METHODS
// ============================================================

// convertScope converts eventsDomain.Scope to authDomain.Scope
func (a *EventsPermissionAdapter) convertScope(scope eventsDomain.Scope) authDomain.Scope {
	if scope.IsPersonal() {
		return authDomain.NewPersonalTeamScope(scope.ID)
	}
	if scope.IsInstitution() {
		return authDomain.NewInstitutionTeamScope(scope.ID)
	}
	return authDomain.NewPlatformScope()
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

// ============================================================
// EVENTS PROFILE ADAPTER
// ============================================================

// EventsProfileAdapter adapts profile service to events domain UserInfoProvider
type EventsProfileAdapter struct {
	profileSvc profileDomain.Service
}

func NewEventsProfileAdapter(profileSvc profileDomain.Service) eventsDomain.UserInfoProvider {
	return &EventsProfileAdapter{profileSvc: profileSvc}
}

func (a *EventsProfileAdapter) GetUserByID(ctx context.Context, userID string) (*eventsDomain.UserInfo, error) {
	if userID == "" {
		return nil, nil
	}

	user, err := a.profileSvc.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &eventsDomain.UserInfo{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
	}, nil
}

func (a *EventsProfileAdapter) GetUserByIDWithDetails(ctx context.Context, userID string) (*eventsDomain.UserInfo, error) {
	if userID == "" {
		return nil, nil
	}

	user, err := a.profileSvc.GetUserProfileWithDetails(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &eventsDomain.UserInfo{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Phone:       user.Phone,
		AccountType: user.AccountType,
		AvatarURL:   user.AvatarURL,
	}, nil
}

func (a *EventsProfileAdapter) GetInstitutionByID(ctx context.Context, institutionID string) (*eventsDomain.InstitutionInfo, error) {
	if institutionID == "" {
		return nil, nil
	}

	institution, err := a.profileSvc.GetInstitutionProfile(ctx, institutionID)
	if err != nil {
		return nil, err
	}
	if institution == nil {
		return nil, nil
	}

	return &eventsDomain.InstitutionInfo{
		ID:          institution.ID,
		Name:        institution.Name,
		DisplayName: institution.DisplayName,
		Slug:        institution.Slug,
		LogoURL:     institution.LogoURL,
	}, nil
}

// ============================================================
// PROFILE PERMISSION ADAPTER
// ============================================================

// ProfilePermissionAdapter adapts auth domain permission checker to profile domain
type ProfilePermissionAdapter struct {
	permSvc authDomain.PermissionChecker
}

func NewProfilePermissionAdapter(permSvc authDomain.PermissionChecker) profileDomain.PermissionChecker {
	return &ProfilePermissionAdapter{permSvc: permSvc}
}

// ============================================================
// CORE PERMISSION METHODS
// ============================================================

// HasPermission checks if a user has a specific permission in a scope
func (a *ProfilePermissionAdapter) HasPermission(ctx context.Context, userID string, scope profileDomain.Scope, resource, action string) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, resource, action)
}

// HasAnyPermission checks if a user has any of the given permissions in a scope
func (a *ProfilePermissionAdapter) HasAnyPermission(ctx context.Context, userID string, scope profileDomain.Scope, resource string, actions ...string) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasAnyPermission(ctx, userID, authScope, resource, actions...)
}

// HasAllPermissions checks if a user has all of the given permissions in a scope
func (a *ProfilePermissionAdapter) HasAllPermissions(ctx context.Context, userID string, scope profileDomain.Scope, resource string, actions ...string) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasAllPermissions(ctx, userID, authScope, resource, actions...)
}

// ============================================================
// PROFILE PERMISSIONS - READ
// ============================================================

// CanReadAllProfiles checks if user can read ALL profiles in a scope
func (a *ProfilePermissionAdapter) CanReadAllProfiles(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, "profile", "read_all")
}

// CanReadOwnProfile checks if user can read OWN profile in a scope
func (a *ProfilePermissionAdapter) CanReadOwnProfile(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, "profile", "read_own")
}

// CanReadProfile checks if user can read profiles in a scope (ALL or OWN)
func (a *ProfilePermissionAdapter) CanReadProfile(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	allowed, err := a.CanReadAllProfiles(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanReadOwnProfile(ctx, userID, scope)
}

// ============================================================
// PROFILE PERMISSIONS - UPDATE
// ============================================================

// CanUpdateAllProfiles checks if user can update ALL profiles in a scope
func (a *ProfilePermissionAdapter) CanUpdateAllProfiles(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, "profile", "update_all")
}

// CanUpdateOwnProfile checks if user can update OWN profile in a scope
func (a *ProfilePermissionAdapter) CanUpdateOwnProfile(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.HasPermission(ctx, userID, authScope, "profile", "update_own")
}

// CanUpdateProfile checks if user can update profiles in a scope (ALL or OWN)
func (a *ProfilePermissionAdapter) CanUpdateProfile(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
    // First check if user has exact update permission
    authScope := a.convertScope(scope)
    allowed, err := a.permSvc.HasPermission(ctx, userID, authScope, "profile", "update")
    if err == nil && allowed {
        return true, nil
    }
    
    // Then check update_all
    allowed, err = a.permSvc.HasPermission(ctx, userID, authScope, "profile", "update_all")
    if err == nil && allowed {
        return true, nil
    }
    
    // Finally check update_own
    return a.permSvc.HasPermission(ctx, userID, authScope, "profile", "update_own")
}



// ============================================================
// PROFILE PERMISSIONS - MANAGEMENT
// ============================================================

// CanManageProfile checks if user can manage profiles in a scope
func (a *ProfilePermissionAdapter) CanManageProfile(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	return a.HasAnyPermission(ctx, userID, scope, "profile",
		"update_all", "delete_all", "manage")
}

// CanViewProfile checks if user can view profiles in a scope (ALL or OWN)
func (a *ProfilePermissionAdapter) CanViewProfile(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	allowed, err := a.CanReadAllProfiles(ctx, userID, scope)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}
	return a.CanReadOwnProfile(ctx, userID, scope)
}

// ============================================================
// ✅ TEAM ROLE CHECKS
// ============================================================

// IsTeamAdmin checks if user is an admin in the scope
func (a *ProfilePermissionAdapter) IsTeamAdmin(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.IsTeamAdmin(ctx, userID, authScope)
}

// IsEventManager checks if user is an event manager in the scope
func (a *ProfilePermissionAdapter) IsEventManager(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.IsEventManager(ctx, userID, authScope)
}

// IsTeamMember checks if user is a team member in the scope
func (a *ProfilePermissionAdapter) IsTeamMember(ctx context.Context, userID string, scope profileDomain.Scope) (bool, error) {
	authScope := a.convertScope(scope)
	return a.permSvc.IsTeamMember(ctx, userID, authScope)
}

// ============================================================
// ✅ USER INFORMATION METHODS
// ============================================================

// GetUserInstitutionTeamIDs returns all institution team IDs where a user has roles
func (a *ProfilePermissionAdapter) GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error) {
	return a.permSvc.GetUserInstitutionTeamIDs(ctx, userID)
}

// GetUserPersonalTeamID returns the personal team ID for a user
func (a *ProfilePermissionAdapter) GetUserPersonalTeamID(ctx context.Context, userID string) (string, error) {
	return userID, nil
}

// ============================================================
// HELPER METHODS
// ============================================================

// convertScope converts profileDomain.Scope to authDomain.Scope
func (a *ProfilePermissionAdapter) convertScope(scope profileDomain.Scope) authDomain.Scope {
	if scope.IsPersonal() {
		return authDomain.NewPersonalTeamScope(scope.ID)
	}
	if scope.IsInstitution() {
		return authDomain.NewInstitutionTeamScope(scope.ID)
	}
	return authDomain.NewPlatformScope()
}

// ============================================================
// PROFILE MEDIA ADAPTER
// ============================================================

// ProfileMediaAdapter adapts media service to profile domain MediaService
type ProfileMediaAdapter struct {
	mediaSvc mediaService.Service
}

func NewProfileMediaAdapter(mediaSvc mediaService.Service) profileDomain.MediaService {
	return &ProfileMediaAdapter{mediaSvc: mediaSvc}
}

func (a *ProfileMediaAdapter) UploadFile(ctx context.Context, cmd profileDomain.UploadMediaCommand) (*profileDomain.MediaInfo, error) {
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

	return &profileDomain.MediaInfo{
		ID:         media.ID,
		URL:        media.URL,
		MediaType:  cmd.MediaTypeName,
		EntityID:   media.EntityID,
		UploadedBy: media.UploadedBy,
		CreatedAt:  media.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (a *ProfileMediaAdapter) GetMediaByID(ctx context.Context, id string) (*profileDomain.MediaInfo, error) {
	media, err := a.mediaSvc.GetMediaByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, nil
	}

	return &profileDomain.MediaInfo{
		ID:         media.ID,
		URL:        media.URL,
		MediaType:  media.MediaTypeID,
		EntityID:   media.EntityID,
		UploadedBy: media.UploadedBy,
		CreatedAt:  media.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (a *ProfileMediaAdapter) GetMediaByEntity(ctx context.Context, entityID string) ([]*profileDomain.MediaInfo, error) {
	mediaList, _, err := a.mediaSvc.GetMediaByEntity(ctx, entityID, 1, 1000)
	if err != nil {
		return nil, err
	}

	result := make([]*profileDomain.MediaInfo, len(mediaList))
	for i, media := range mediaList {
		result[i] = &profileDomain.MediaInfo{
			ID:         media.ID,
			URL:        media.URL,
			MediaType:  media.MediaTypeID,
			EntityID:   media.EntityID,
			UploadedBy: media.UploadedBy,
			CreatedAt:  media.CreatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

func (a *ProfileMediaAdapter) GetMediaTypeByName(ctx context.Context, name string) (*profileDomain.MediaTypeInfo, error) {
	mediaType, err := a.mediaSvc.GetMediaTypeByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, nil
	}

	return &profileDomain.MediaTypeInfo{
		ID:   mediaType.ID,
		Name: mediaType.Name,
		Slug: mediaType.Slug,
	}, nil
}

func (a *ProfileMediaAdapter) DeleteFile(ctx context.Context, id string) error {
	return a.mediaSvc.DeleteFile(ctx, id)
}

func (a *ProfileMediaAdapter) DeleteFilesByEntity(ctx context.Context, entityID string) error {
	return a.mediaSvc.DeleteFilesByEntity(ctx, entityID)
}

func (a *ProfileMediaAdapter) DeleteFilesByEntityAndMediaType(ctx context.Context, entityID, mediaTypeID string) error {
	return a.mediaSvc.DeleteFilesByEntityAndMediaType(ctx, entityID, mediaTypeID)
}