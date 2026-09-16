// internal/modules/registration/providers.go

package registration

import (
	"github.com/google/wire"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/delivery/http"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/infrastructure/notifier"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

var ProviderSet = wire.NewSet(
	// ============================================================
	// Infrastructure Layer
	// ============================================================

	postgres.NewRegistrationRepository,
	postgres.NewEventRegistrationRepository,
	postgres.NewWaitlistRepository,

	postgres.NewRegistrationNumberGenerator,
	wire.Bind(
		new(registrationdomain.RegistrationNumberGenerator),
		new(*postgres.RegistrationNumberGenerator),
	),

	notifier.NewNoop,
	wire.Bind(
		new(registrationdomain.Notifier),
		new(*notifier.Noop),
	),

	// RegistrableResolver is provided by the app's cross-module
	// adapters (NewRegistrableResolver) — not from within this module.

	// ============================================================
	// Shared utilities
	// ============================================================

	id.NewUUIDGenerator,
	wire.Bind(
		new(id.Generator),
		new(*id.UUIDGenerator),
	),

	ProvideSystemClock,
	ProvideServiceDependencies,

	// ============================================================
	// Application Layer
	// ============================================================

	service.New,

	// ============================================================
	// Delivery Layer
	// ============================================================

	http.NewHandler,
)

func ProvideServiceDependencies(
	registrations registrationdomain.RegistrationRepository,
	eventRegistrations registrationdomain.EventRegistrationRepository,
	waitlist registrationdomain.WaitlistRepository,
	registrables registrationdomain.RegistrableResolver,
	notifier registrationdomain.Notifier,
	idGen id.Generator,
	clock service.Clock,
	numberGen registrationdomain.RegistrationNumberGenerator,
) service.Dependencies {
	return service.Dependencies{
		Registrations:      registrations,
		EventRegistrations: eventRegistrations,
		Waitlist:           waitlist,
		Registrables:       registrables,
		Notifier:           notifier,
		IDGenerator:        idGen,
		Clock:              clock,
		NumberGenerator:    numberGen,
	}
}

func ProvideSystemClock() service.Clock {
	return service.SystemClock{}
}