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

	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"

	eventsHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/delivery/eventhandler"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"

	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/storage"

	notificationdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"

	profileHandler "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/delivery/handler"
	profileDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"

	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	teamDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
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
	accountService accountService.Service,
	teamService teamService.Service,
	eventsService eventsService.Service,
	profileSvc profileDomain.Service,
	mediaService mediaService.Service,
	authHandler *authHandler.AuthHandler,
	accountHandler *accountHandler.AccountHandler,
	teamHandler *teamHandler.TeamHandler,
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
		AccountService:    accountService,
		TeamService:       teamService,
		EventsService:     eventsService,
		ProfileService:    profileSvc,
		MediaService:      mediaService,
		AuthHandler:       authHandler,
		AccountHandler:    accountHandler,
		TeamHandler:       teamHandler,
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

// HasPermission checks if a user has a specific permission in a domain
func (a *EventsPermissionAdapter) HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error) {
	return a.permSvc.HasPermission(ctx, userID, domain, resource, action)
}

// HasAnyPermission checks if a user has any of the given permissions in a domain
func (a *EventsPermissionAdapter) HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := a.permSvc.HasPermission(ctx, userID, domain, resource, action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the given permissions in a domain
func (a *EventsPermissionAdapter) HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := a.permSvc.HasPermission(ctx, userID, domain, resource, action)
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
// EVENTS MEDIA ADAPTER
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
// AUTH NOTIFICATION ADAPTER
// ============================================================

type AuthNotificationAdapter struct {
	notifSvc notificationdomain.NotificationService
}

func NewAuthNotificationAdapter(notifSvc notificationdomain.NotificationService) authDomain.NotificationService {
	return &AuthNotificationAdapter{notifSvc: notifSvc}
}

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
		To:                  req.To,
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

func (a *AuthNotificationAdapter) SendPasswordResetConfirm(ctx context.Context, req authDomain.SendPasswordResetConfirmRequest) error {
	notifReq := notificationdomain.SendPasswordResetConfirmRequest{
		To:   req.To,
		Name: req.Name,
	}
	return a.notifSvc.SendPasswordResetConfirm(ctx, notifReq)
}

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
// ACCOUNT - AUTH ADAPTER
// ============================================================

type AccountAuthAdapter struct {
	authSvc authService.Service
}

func NewAccountAuthAdapter(authSvc authService.Service) accountService.AuthService {
	return &AccountAuthAdapter{authSvc: authSvc}
}

func (a *AccountAuthAdapter) GetUserByID(ctx context.Context, userID string) (*accountService.UserResult, error) {
	user, err := a.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &accountService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

func (a *AccountAuthAdapter) GetUserByEmail(ctx context.Context, email string) (*accountService.UserResult, error) {
	user, err := a.authSvc.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &accountService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

func (a *AccountAuthAdapter) UserExists(ctx context.Context, email string) (bool, error) {
	return a.authSvc.UserExists(ctx, email)
}

// ============================================================
// ACCOUNT - PERMISSION ADAPTER
// ============================================================

type AccountPermissionAdapter struct {
	permChecker authDomain.PermissionChecker
	roleManager authDomain.RoleManager
}

func NewAccountPermissionAdapter(
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
) accountService.PermissionChecker {
	return &AccountPermissionAdapter{
		permChecker: permChecker,
		roleManager: roleManager,
	}
}

// HasPermission checks if a user has a specific permission in a domain
func (a *AccountPermissionAdapter) HasPermission(ctx context.Context, domain string, userID, resource, action string) (bool, error) {
	return a.permChecker.HasPermission(ctx, userID, domain, resource, action)
}

// HasAnyPermission checks if a user has any of the given permissions in a domain
func (a *AccountPermissionAdapter) HasAnyPermission(ctx context.Context, domain string, userID, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := a.permChecker.HasPermission(ctx, userID, domain, resource, action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

// HasAllPermissions checks if a user has all of the given permissions in a domain
func (a *AccountPermissionAdapter) HasAllPermissions(ctx context.Context, domain string, userID, resource string, actions ...string) (bool, error) {
	for _, action := range actions {
		allowed, err := a.permChecker.HasPermission(ctx, userID, domain, resource, action)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// CanManageAccount checks if user can manage an account (update, delete, manage)
func (a *AccountPermissionAdapter) CanManageAccount(ctx context.Context, domain string, userID string) (bool, error) {
	return a.HasAnyPermission(ctx, domain, userID, "account", "update", "delete", "manage")
}

// CanManageAccountMembers checks if user can manage account members
func (a *AccountPermissionAdapter) CanManageAccountMembers(ctx context.Context, domain string, userID string) (bool, error) {
	return a.HasAnyPermission(ctx, domain, userID, "member", "create", "update", "delete")
}

// CanViewAccount checks if user can view an account
func (a *AccountPermissionAdapter) CanViewAccount(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.HasPermission(ctx, userID, domain, "account", "read")
}

// IsAccountAdmin checks if user is an account admin in the domain
func (a *AccountPermissionAdapter) IsAccountAdmin(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.IsAccountAdmin(ctx, userID, domain)
}

// IsAccountTrainer checks if user is a trainer in the domain
func (a *AccountPermissionAdapter) IsAccountTrainer(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.IsTrainer(ctx, userID, domain)
}

// GetUserRoles returns all roles for a user in a domain
func (a *AccountPermissionAdapter) GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error) {
	return a.roleManager.GetUserRoles(ctx, userID, domain)
}

// ============================================================
// ACCOUNT - ROLE MANAGER ADAPTER
// ============================================================

type AccountRoleManagerAdapter struct {
	roleManager authDomain.RoleManager
}

func NewAccountRoleManagerAdapter(roleManager authDomain.RoleManager) accountService.RoleManager {
	return &AccountRoleManagerAdapter{roleManager: roleManager}
}

func (a *AccountRoleManagerAdapter) AssignRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.AssignRole(ctx, domain, userID, role)
}

func (a *AccountRoleManagerAdapter) RemoveRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.RemoveRole(ctx, domain, userID, role)
}

func (a *AccountRoleManagerAdapter) GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error) {
	return a.roleManager.GetUserRoles(ctx, userID, domain)
}

// HasRole implements service.RoleManager.
func (a *AccountRoleManagerAdapter) HasRole(ctx context.Context, domain string, userID string, role string) (bool, error) {
	roles, err := a.roleManager.GetUserRoles(ctx, userID, domain)
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

// ============================================================
// TEAM - AUTH ADAPTER
// ============================================================

type TeamAuthAdapter struct {
	authSvc authService.Service
}

func NewTeamAuthAdapter(authSvc authService.Service) teamService.AuthService {
	return &TeamAuthAdapter{authSvc: authSvc}
}

func (a *TeamAuthAdapter) GetUserByID(ctx context.Context, userID string) (*teamService.UserResult, error) {
	user, err := a.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &teamService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

func (a *TeamAuthAdapter) GetUserByEmail(ctx context.Context, email string) (*teamService.UserResult, error) {
	user, err := a.authSvc.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &teamService.UserResult{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Phone:       user.Phone,
		DisplayName: user.DisplayName,
		IsActive:    user.IsActive,
	}, nil
}

func (a *TeamAuthAdapter) UserExists(ctx context.Context, email string) (bool, error) {
	return a.authSvc.UserExists(ctx, email)
}

// ============================================================
// TEAM - PERMISSION ADAPTER
// ============================================================

type TeamPermissionAdapter struct {
	permChecker authDomain.PermissionChecker
	roleManager authDomain.RoleManager
}

func NewTeamPermissionAdapter(
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
) teamService.PermissionService {
	return &TeamPermissionAdapter{
		permChecker: permChecker,
		roleManager: roleManager,
	}
}

func (a *TeamPermissionAdapter) IsTeamAdmin(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.IsAccountAdmin(ctx, userID, domain)
}

func (a *TeamPermissionAdapter) IsTeamTrainer(ctx context.Context, domain string, userID string) (bool, error) {
	return a.permChecker.IsTrainer(ctx, userID, domain)
}

func (a *TeamPermissionAdapter) AssignRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.AssignRole(ctx, domain, userID, role)
}

func (a *TeamPermissionAdapter) RemoveRole(ctx context.Context, domain string, userID, role string) error {
	return a.roleManager.RemoveRole(ctx, domain, userID, role)
}

func (a *TeamPermissionAdapter) GetUserRoles(ctx context.Context, domain string, userID string) ([]string, error) {
	return a.roleManager.GetUserRoles(ctx, userID, domain)
}

// ============================================================
// TEAM - PERMISSION CHECKER ADAPTER
// ============================================================

type TeamPermissionCheckerAdapter struct {
	permChecker authDomain.PermissionChecker
}

func NewTeamPermissionCheckerAdapter(permChecker authDomain.PermissionChecker) teamService.PermissionChecker {
	return &TeamPermissionCheckerAdapter{permChecker: permChecker}
}

func (a *TeamPermissionCheckerAdapter) HasPermission(ctx context.Context, userID string, domain, resource, action string) (bool, error) {
	return a.permChecker.HasPermission(ctx, userID, domain, resource, action)
}

func (a *TeamPermissionCheckerAdapter) IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error) {
	return a.permChecker.IsAccountAdmin(ctx, userID, domain)
}

func (a *TeamPermissionCheckerAdapter) IsTrainer(ctx context.Context, userID string, domain string) (bool, error) {
	return a.permChecker.IsTrainer(ctx, userID, domain)
}

func (a *TeamPermissionCheckerAdapter) HasTeamAccess(ctx context.Context, userID string) (bool, error) {
	return a.permChecker.HasTeamAccess(ctx, userID)
}
