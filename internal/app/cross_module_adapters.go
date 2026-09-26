// internal/app/cross_module_adapters.go

package app

import (
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/accounts"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/auth"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/events"
	paymentadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/payment"
	registrationadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/registration"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/team"

	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	authService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	eventsService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	mediaService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/media/service"
	notificationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/notification/notification-domain"
	paymentproviders "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/infrastructure/providers"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/infrastructure/providers/paystack"
	paymentdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	registrationService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
	teamService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
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
func NewAccountPermissionAdapter(permChecker authDomain.PermissionChecker) accountDomain.PermissionChecker {
	return accounts.NewPermissionAdapter(permChecker)
}

// NewAccountMediaAdapter bridges the media module to the account domain's
// MediaService port.
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

// NewEventsAttendanceRegistrar wires the events module's
// AttendanceRegistrar port to the attendance service.
//
// The events service uses this to mirror event schedules into
// attendance sessions on publish and update.
func NewEventsAttendanceRegistrar(
	attendanceSvc attendanceService.Service,
) eventsDomain.AttendanceRegistrar {
	return events.NewAttendanceRegistrarAdapter(attendanceSvc)
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

// ---- REGISTRATION ADAPTERS ----

// NewRegistrableResolver wires the registration module's Registrable
// resolver to the events module.
func NewRegistrableResolver(eventsSvc eventsService.Service) registrationdomain.RegistrableResolver {
	return registrationadapters.NewRegistrableResolver(eventsSvc)
}

// NewRegistrationAttendanceRegistrar wires the registration module's
// AttendanceRegistrar port to the attendance service.
//
// The registration service uses this to register attendees and issue
// join tokens when a registration is confirmed. Join URLs are built
// using cfg.App.PublicURL.
func NewRegistrationAttendanceRegistrar(
	attendanceSvc attendanceService.Service,
	cfg *config.Config,
) registrationdomain.AttendanceRegistrar {
	return registrationadapters.NewAttendanceRegistrarAdapter(
		attendanceSvc,
		cfg.App.PublicURL,
	)
}

// NewRegistrationUserInfoAdapter wires the registration module's
// UserInfoProvider port to the auth service.
//
// The registration service uses this to resolve an authenticated
// user's display name and email when building the attendance record.
func NewRegistrationUserInfoAdapter(
	authSvc authService.Service,
) registrationdomain.UserInfoProvider {
	return registrationadapters.NewUserInfoAdapter(authSvc)
}

// ---- PAYMENT ADAPTERS ----

// NewPaymentRegistrationConfirmer wires the payment module's
// RegistrationConfirmer port to the registration module's service.
func NewPaymentRegistrationConfirmer(
	regSvc registrationService.Service,
) paymentdomain.RegistrationConfirmer {
	return paymentadapters.NewRegistrationConfirmer(regSvc)
}

// NewPaymentProviderRegistry builds the provider registry from
// configured credentials.
//
// Paystack is currently the sole payment provider for both M-Pesa and
// cards — it handles both methods through one hosted integration, so
// there's no per-method fallback to worry about.
func NewPaymentProviderRegistry(
	cfg *config.Config,
) (paymentdomain.ProviderRegistry, error) {
	registry := paymentproviders.NewRegistry()

	if cfg.Paystack.IsConfigured() {
		ps := paystack.NewProvider(cfg.Paystack)

		if err := registry.Register(ps); err != nil {
			return nil, fmt.Errorf("register paystack: %w", err)
		}

		for _, method := range ps.Methods() {
			if method == ps.Method() {
				continue
			}
			if err := registry.RegisterForMethod(ps, method); err != nil {
				return nil, fmt.Errorf("register paystack for %s: %w", method, err)
			}
		}

		log.Println("payment: registered paystack provider (mpesa + card)")
	} else {
		log.Println("payment: paystack not configured (skipping)")
	}

	return registry, nil
}

// NewPaymentUserEmailResolver wires the UserEmailResolver port to the
// account module.
func NewPaymentUserEmailResolver(
	accountSvc accountService.Service,
) paymentdomain.UserEmailResolver {
	return paymentadapters.NewUserEmailResolver(accountSvc)
}

// NewPaymentNotifier wires the payment module's Notifier port to the
// notification module's service.
func NewPaymentNotifier(
	notifSvc notificationDomain.NotificationService,
	userEmails paymentdomain.UserEmailResolver,
) paymentdomain.Notifier {
	return paymentadapters.NewNotifier(notifSvc, userEmails)
}

// NewPaymentPricingResolver wires the payment module's
// RegistrationPricingResolver port to the registration repository.
//
// We inject the repository rather than the service so the adapter can
// load a registration without going through the service's ownership
// check. The payment service enforces ownership itself.
func NewPaymentPricingResolver(
	regRepo registrationdomain.EventRegistrationRepository,
) paymentdomain.RegistrationPricingResolver {
	return paymentadapters.NewPricingResolver(regRepo)
}