// internal/app/cross_module_adapters.go

package app

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/accounts"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/auth"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/events"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/profile"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/team"

	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	profileDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
	profileService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/service"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
)

// ---- AUTH ADAPTERS ----

// NewAuthNotificationAdapter creates a new auth notification adapter
func NewAuthNotificationAdapter(notifSvc notificationDomain.NotificationService) authDomain.NotificationService {
	return auth.NewNotificationAdapter(notifSvc)
}

func provideOrganizerProvider(accounts accountService.Service) eventsDomain.OrganizerProvider {
	return events.NewOrganizerAdapter(accounts)
}

// NewQueueAdapter creates a new queue adapter for auth
func NewQueueAdapter(q notificationDomain.TaskQueue) authDomain.QueueService {
	return auth.NewQueueAdapter(q)
}

// NewAuthTeamAdapter creates a new team adapter for auth
func NewAuthTeamAdapter(teamSvc teamService.Service) authService.TeamService {
	return auth.NewTeamAdapter(teamSvc)
}

// ---- ACCOUNT ADAPTERS ----

// NewAccountAuthAdapter creates a new account auth adapter
func NewAccountAuthAdapter(authSvc authService.Service) accountService.AuthService {
	return accounts.NewAuthAdapter(authSvc)
}

// NewAccountNotificationAdapter creates a new account notification adapter
func NewAccountNotificationAdapter(notifSvc notificationDomain.NotificationService) accountService.NotificationService {
	return accounts.NewNotificationAdapter(notifSvc)
}

// ---- EVENTS ADAPTERS ----

// NewEventsPermissionAdapter creates a new events permission adapter
func NewEventsPermissionAdapter(permChecker authDomain.PermissionChecker) eventsDomain.PermissionChecker {
	return events.NewPermissionAdapter(permChecker)
}

// NewEventsUserInfoAdapter creates a new events user info adapter
func NewEventsUserInfoAdapter(profileSvc profileService.Service) eventsDomain.UserInfoProvider {
	return events.NewUserInfoAdapter(profileSvc)
}

// NewEventsMediaAdapter creates a new events media adapter
func NewEventsMediaAdapter(mediaSvc mediaService.Service) eventsDomain.MediaService {
	return events.NewMediaAdapter(mediaSvc)
}

// ---- PROFILE ADAPTERS ----

// NewProfilePermissionAdapter creates a new profile permission adapter
func NewProfilePermissionAdapter(permChecker authDomain.PermissionChecker) profileDomain.PermissionChecker {
	return profile.NewPermissionAdapter(permChecker)
}

// NewProfileRoleManagerAdapter satisfies profileDomain.RoleManager using auth's RoleManager
func NewProfileRoleManagerAdapter(roleMgr authDomain.RoleManager) profileDomain.RoleManager {
	return profile.NewRoleManagerAdapter(roleMgr)
}

// NewProfileMediaAdapter creates a new profile media adapter
func NewProfileMediaAdapter(mediaSvc mediaService.Service) profileDomain.MediaService {
	return profile.NewMediaAdapter(mediaSvc)
}

// ---- TEAM ADAPTERS ----

// NewTeamAuthAdapter satisfies teamService.AuthService using authDomain.Repository
func NewTeamAuthAdapter(repo authDomain.Repository) teamService.AuthService {
	return team.NewAuthAdapter(repo)
}

// NewTeamCasbinAdapter creates a new team casbin adapter
func NewTeamCasbinAdapter(
	permChecker authDomain.PermissionChecker,
	roleManager authDomain.RoleManager,
	policyManager authDomain.PolicyManager,
) teamService.CasbinService {
	return team.NewCasbinAdapter(permChecker, roleManager, policyManager)
}

// NewTeamNotificationAdapter creates a new team notification adapter
func NewTeamNotificationAdapter(notifSvc notificationDomain.NotificationService) teamService.NotificationService {
	return team.NewTeamNotificationAdapter(notifSvc)
}




