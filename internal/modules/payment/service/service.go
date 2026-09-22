// internal/modules/payment/service/service.go

package service

import (
	"context"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

// ============================================================
// SERVICE INTERFACE
// ============================================================

// Service is the application-layer interface for the payment module.
type Service interface {
	InitiatePayment(ctx context.Context, cmd InitiateCommand) (*paymentdomain.Payment, error)
	ConfirmPayment(ctx context.Context, paymentID string) error
	FailPayment(ctx context.Context, paymentID, reason string) error
	Refund(ctx context.Context, cmd RefundCommand) error
	HandleWebhook(
		ctx context.Context,
		provider string,
		payload []byte,
		headers map[string]string,
	) error
}

// ============================================================
// DEPENDENCIES
// ============================================================

// Dependencies groups the ports the service needs.
type Dependencies struct {
	// Repositories
	Orders   paymentdomain.OrderRepository
	Payments paymentdomain.PaymentRepository
	Refunds  paymentdomain.RefundRepository
	Webhooks paymentdomain.WebhookEventRepository

	// Cross-cutting
	Providers  paymentdomain.ProviderRegistry
	Notifier   paymentdomain.Notifier
	UnitOfWork paymentdomain.UnitOfWork

	// Cross-module
	Registrations paymentdomain.RegistrationConfirmer

	// Utilities
	IDGenerator id.Generator  // ← from shared
	Clock       Clock         // ← stays local
}

// ============================================================
// LOCAL UTILITY INTERFACES
// ============================================================

// Clock abstracts time.Now so tests can control time.
type Clock interface {
	Now() time.Time
}

// ============================================================
// IMPLEMENTATION
// ============================================================

type service struct {
	deps Dependencies
}

func New(deps Dependencies) Service {
	return &service{deps: deps}
}

// ============================================================
// DEFAULT IMPLEMENTATIONS
// ============================================================

// SystemClock returns the current UTC time. Production Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }