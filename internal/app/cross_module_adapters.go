// internal/app/cross_module_adapters.go

package app

import (
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/accounts"
	attendanceadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/attendance"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/auth"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/events"
	paymentadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/payment"
	registrationadapters "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/registration"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/team"

	accountDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	accountService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	attendanceService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
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
	videoService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"

	registrationnotifier "github.com/ALLAN-star-glitch/nuruvent-backend/internal/app/adapters/registration"
	registrationDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"

	attendanceDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
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

// NewAccountRoleAssignmentAdapter bridges the auth module's Casbin
// enforcer to the account domain's RoleAssignment port.
//
// Used by the account service to keep Casbin in sync whenever account
// membership changes — assign on add, replace on role change, revoke
// on remove or leave, and revoke-all on account deletion.
func NewAccountRoleAssignmentAdapter(
	enforcer *authorization.Enforcer,
) accountDomain.RoleAssignment {
	return accounts.NewRoleAssignmentAdapter(enforcer)
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
func NewRegistrationUserInfoAdapter(
	authSvc authService.Service,
) registrationdomain.UserInfoProvider {
	return registrationadapters.NewUserInfoAdapter(authSvc)
}

// NewRegistrationPermissionAdapter bridges auth's PermissionChecker to
// the registration module's PermissionChecker port.
func NewRegistrationPermissionAdapter(
	permChecker authDomain.PermissionChecker,
) registrationDomain.PermissionChecker {
	return registrationadapters.NewPermissionCheckerAdapter(permChecker)
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
func NewPaymentPricingResolver(
	regRepo registrationdomain.EventRegistrationRepository,
) paymentdomain.RegistrationPricingResolver {
	return paymentadapters.NewPricingResolver(regRepo)
}

// ---- VIDEO ADAPTERS ----

func NewEventsVideoAdapter(
	videoSvc videoService.Service,
) eventsDomain.VideoMeetingCreator {
	return events.NewVideoAdapter(videoSvc)
}

// NewVideoAttendanceAdapter wires the video module's
// ParticipantRecorder port to the attendance service.
func NewVideoAttendanceAdapter(
	attendanceSvc attendanceService.Service,
) videoService.ParticipantRecorder {
	return attendanceadapters.NewAttendanceToVideo(attendanceSvc)
}

// NewAttendanceVideoMeetingIDResolver wires the attendance module's
// VideoMeetingIDResolver port to the video service.
func NewAttendanceVideoMeetingIDResolver(
	meetings videodomain.MeetingRepository,
) attendanceService.VideoMeetingIDResolver {
	return attendanceadapters.NewVideoMeetingIDResolver(meetings)
}

func NewRegistrationNotifier(
	notifSvc notificationDomain.NotificationService,
	users registrationDomain.UserInfoProvider,
) registrationDomain.Notifier {
	return registrationnotifier.NewNotifier(notifSvc, users)
}

// NewAttendancePermissionAdapter bridges auth's PermissionChecker to
// the attendance module's PermissionChecker port.
func NewAttendancePermissionAdapter(
	permChecker authDomain.PermissionChecker,
) attendanceDomain.PermissionChecker {
	return attendanceadapters.NewPermissionCheckerAdapter(permChecker)
}

// NewAttendanceRegistrationLookup wires the attendance module's
// RegistrationLookup port to the registration service.
func NewAttendanceRegistrationLookup(
	eventRegs registrationDomain.EventRegistrationRepository,
) attendanceDomain.RegistrationLookup {
	return attendanceadapters.NewRegistrationLookupAdapter(eventRegs)
}

// NewEventsVideoIdentityAdapter bridges the video service to the
// events domain's VideoIdentityProvider port.
func NewEventsVideoIdentityAdapter(
	videoSvc videoService.Service,
) eventsDomain.VideoIdentityProvider {
	return events.NewVideoIdentityAdapter(videoSvc)
}

// NewPaymentBillingResolver wires the payment module's
// RegistrationBillingResolver port to the registration service.
func NewPaymentBillingResolver(
	regSvc registrationService.Service,
) paymentdomain.RegistrationBillingResolver {
	return paymentadapters.NewBillingResolver(regSvc)
}

func NewAttendanceAuthSessionIssuer(
	authSvc authService.Service,
) attendanceDomain.AuthSessionIssuer {
	return attendanceadapters.NewAuthSessionIssuerAdapter(authSvc)
}

// NewEventsReader wires the attendance module's EventsReader port to
// the events service via a lazy holder — breaks the wire-time cycle
// (video → attendance → events → video).
func NewEventsReader(
	holder *attendanceadapters.EventsServiceHolder,
) attendanceDomain.EventsReader {
	return attendanceadapters.NewEventsReaderAdapter(holder.Resolver())
}

// NewEventsTeamMembershipAdapter bridges the team service to the events
// module's TeamMembershipChecker port.
//
// Used by the events service to gate private events to team members.
func NewEventsTeamMembershipAdapter(
	teamSvc teamService.Service,
) eventsDomain.TeamMembershipChecker {
	return events.NewTeamMembershipAdapter(teamSvc)
}