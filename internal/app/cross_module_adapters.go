// internal/app/cross_module_adapters.go

package app

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/accounts"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/auth"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/events"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/team"

	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
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

// NewAccountPermissionAdapter bridges auth's PermissionChecker to
// account's PermissionChecker interface.
//
// The concrete implementation (authorization.PermissionChecker) already
// satisfies both interfaces structurally. This adapter exists so that:
//   - Cross-module wiring stays visible in one place (this file).
//   - The account module depends on its own accountDomain.PermissionChecker
//     interface, not on auth's.
//   - Future changes to either interface are isolated to the adapter.
func NewAccountPermissionAdapter(permChecker authDomain.PermissionChecker) accountDomain.PermissionChecker {
	return accounts.NewPermissionAdapter(permChecker)
}


// NewAccountMediaAdapter bridges the media module to the account domain's
// MediaService port. Used for user avatars (media_type_profile) and
// account logos (media_type_business).
func NewAccountMediaAdapter(mediaSvc mediaService.Service) accountDomain.MediaService {
	return accounts.NewMediaAdapter(mediaSvc)
}

// ---- EVENTS ADAPTERS ----

// NewEventsPermissionAdapter creates a new events permission adapter
func NewEventsPermissionAdapter(permChecker authDomain.PermissionChecker) eventsDomain.PermissionChecker {
	return events.NewPermissionAdapter(permChecker)
}

// NewEventsUserInfoAdapter creates a new events user info adapter
func NewEventsUserInfoAdapter(profileSvc accountService.Service) eventsDomain.UserInfoProvider {
	return events.NewUserInfoAdapter(profileSvc)
}

// NewEventsMediaAdapter creates a new events media adapter
func NewEventsMediaAdapter(mediaSvc mediaService.Service) eventsDomain.MediaService {
	return events.NewMediaAdapter(mediaSvc)
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
	return team.NewCasbinAdapter(permChecker, roleManager)
}

// NewTeamNotificationAdapter creates a new team notification adapter
func NewTeamNotificationAdapter(notifSvc notificationDomain.NotificationService) teamService.NotificationService {
	return team.NewTeamNotificationAdapter(notifSvc)
}