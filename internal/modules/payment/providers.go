// internal/modules/payment/providers.go

package payment

import (
	"github.com/google/wire"

	paymenthttp "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/delivery/http"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/infrastructure/postgres"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/infrastructure/providers/intasend"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/infrastructure/providers/paystack"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ProviderSet contains all payment module dependencies.
//
// The provider registry and the notifier are NOT provided here —
// they come from the app's composition layer:
//
//   - NewPaymentProviderRegistry  → paymentdomain.ProviderRegistry
//   - NewPaymentNotifier          → paymentdomain.Notifier
//   - NewPaymentUserEmailResolver → paymentdomain.UserEmailResolver
//
// The module depends on those interfaces; the app chooses which
// implementations to inject.
var ProviderSet = wire.NewSet(
	// ============================================================
	// Repositories
	// ============================================================
	postgres.NewOrderRepository,
	wire.Bind(
		new(paymentdomain.OrderRepository),
		new(*postgres.OrderRepository),
	),

	postgres.NewPaymentRepository,
	wire.Bind(
		new(paymentdomain.PaymentRepository),
		new(*postgres.PaymentRepository),
	),

	postgres.NewRefundRepository,
	wire.Bind(
		new(paymentdomain.RefundRepository),
		new(*postgres.RefundRepository),
	),

	postgres.NewWebhookEventRepository,
	wire.Bind(
		new(paymentdomain.WebhookEventRepository),
		new(*postgres.WebhookEventRepository),
	),

	// ============================================================
	// Unit of work
	// ============================================================
	postgres.NewUnitOfWork,
	wire.Bind(
		new(paymentdomain.UnitOfWork),
		new(*postgres.UnitOfWork),
	),

	// ============================================================
	// Provider implementations
	// ============================================================
	intasend.NewProvider,   // ← was flutterwave.NewProvider
	paystack.NewProvider,

	// ============================================================
	// Shared utilities
	// ============================================================
	ProvideSystemClock,
	wire.Bind(
		new(service.Clock),
		new(*service.SystemClock),
	),

	// ============================================================
	// Application layer
	// ============================================================
	ProvideServiceDependencies,
	service.New,

	// ============================================================
	// Delivery layer
	// ============================================================
	paymenthttp.NewHandler,
	paymenthttp.NewOrderHandler,
)

// ============================================================
// DEPENDENCY ASSEMBLY
// ============================================================

// provideServiceDependencies assembles the service.Dependencies struct
// from individually-provided ports.
//
// Every parameter is an interface, not a concrete type. Wire resolves
// the interface from whichever provider was registered — the module's
// own provider set for repositories and utilities, or the app's
// cross-module adapters for the notifier and the registry.
func ProvideServiceDependencies(
	orders paymentdomain.OrderRepository,
	payments paymentdomain.PaymentRepository,
	refunds paymentdomain.RefundRepository,
	webhooks paymentdomain.WebhookEventRepository,

	registry paymentdomain.ProviderRegistry,
	notifier paymentdomain.Notifier,
	uow paymentdomain.UnitOfWork,

	registrations paymentdomain.RegistrationConfirmer,
	registrationPricing paymentdomain.RegistrationPricingResolver,

	idGen id.Generator,
	clock service.Clock,
) service.Dependencies {
	return service.Dependencies{
		Orders:   orders,
		Payments: payments,
		Refunds:  refunds,
		Webhooks: webhooks,

		Providers:  registry,
		Notifier:   notifier,
		UnitOfWork: uow,

		Registrations:       registrations,
		RegistrationPricing: registrationPricing,

		IDGenerator: idGen,
		Clock:       clock,
	}
}

// provideSystemClock provides the Clock implementation.
func ProvideSystemClock() *service.SystemClock {
	return &service.SystemClock{}
}