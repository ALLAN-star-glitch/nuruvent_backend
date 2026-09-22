// internal/modules/payment/service/fakes_test.go

package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ============================================================
// ORDER REPOSITORY FAKE
// ============================================================

type fakeOrderRepository struct {
	createFunc                   func(ctx context.Context, o *paymentdomain.Order) error
	updateFunc                   func(ctx context.Context, o *paymentdomain.Order) error
	findByIDFunc                 func(ctx context.Context, id string) (*paymentdomain.Order, error)
	findActiveByRegistrationFunc func(ctx context.Context, regID string) (*paymentdomain.Order, error)
	findExpiredFunc              func(ctx context.Context, before time.Time, limit int) ([]*paymentdomain.Order, error)

	mu      sync.Mutex
	created []*paymentdomain.Order
	updated []*paymentdomain.Order
}

func (f *fakeOrderRepository) Create(ctx context.Context, o *paymentdomain.Order) error {
	f.mu.Lock()
	f.created = append(f.created, o)
	f.mu.Unlock()
	if f.createFunc != nil {
		return f.createFunc(ctx, o)
	}
	return nil
}

func (f *fakeOrderRepository) Update(ctx context.Context, o *paymentdomain.Order) error {
	f.mu.Lock()
	f.updated = append(f.updated, o)
	f.mu.Unlock()
	if f.updateFunc != nil {
		return f.updateFunc(ctx, o)
	}
	return nil
}

func (f *fakeOrderRepository) FindByID(ctx context.Context, id string) (*paymentdomain.Order, error) {
	if f.findByIDFunc != nil {
		return f.findByIDFunc(ctx, id)
	}
	return nil, paymentdomain.ErrOrderNotFound
}

func (f *fakeOrderRepository) FindActiveByRegistration(ctx context.Context, regID string) (*paymentdomain.Order, error) {
	if f.findActiveByRegistrationFunc != nil {
		return f.findActiveByRegistrationFunc(ctx, regID)
	}
	return nil, paymentdomain.ErrOrderNotFound
}

func (f *fakeOrderRepository) FindExpired(ctx context.Context, before time.Time, limit int) ([]*paymentdomain.Order, error) {
	if f.findExpiredFunc != nil {
		return f.findExpiredFunc(ctx, before, limit)
	}
	return nil, nil
}

// ============================================================
// PAYMENT REPOSITORY FAKE
// ============================================================

type fakePaymentRepository struct {
	createFunc                  func(ctx context.Context, p *paymentdomain.Payment) error
	updateFunc                  func(ctx context.Context, p *paymentdomain.Payment) error
	findByIDFunc                func(ctx context.Context, id string) (*paymentdomain.Payment, error)
	findByProviderReferenceFunc func(ctx context.Context, provider, ref string) (*paymentdomain.Payment, error)
	findByIdempotencyKeyFunc    func(ctx context.Context, orderID, key string) (*paymentdomain.Payment, error)

	mu      sync.Mutex
	created []*paymentdomain.Payment
	updated []*paymentdomain.Payment
}

func (f *fakePaymentRepository) Create(ctx context.Context, p *paymentdomain.Payment) error {
	f.mu.Lock()
	f.created = append(f.created, p)
	f.mu.Unlock()
	if f.createFunc != nil {
		return f.createFunc(ctx, p)
	}
	return nil
}

func (f *fakePaymentRepository) Update(ctx context.Context, p *paymentdomain.Payment) error {
	f.mu.Lock()
	f.updated = append(f.updated, p)
	f.mu.Unlock()
	if f.updateFunc != nil {
		return f.updateFunc(ctx, p)
	}
	return nil
}

func (f *fakePaymentRepository) FindByID(ctx context.Context, id string) (*paymentdomain.Payment, error) {
	if f.findByIDFunc != nil {
		return f.findByIDFunc(ctx, id)
	}
	return nil, paymentdomain.ErrPaymentNotFound
}

func (f *fakePaymentRepository) FindByProviderReference(ctx context.Context, provider, ref string) (*paymentdomain.Payment, error) {
	if f.findByProviderReferenceFunc != nil {
		return f.findByProviderReferenceFunc(ctx, provider, ref)
	}
	return nil, paymentdomain.ErrPaymentNotFound
}

func (f *fakePaymentRepository) FindByIdempotencyKey(ctx context.Context, orderID, key string) (*paymentdomain.Payment, error) {
	if f.findByIdempotencyKeyFunc != nil {
		return f.findByIdempotencyKeyFunc(ctx, orderID, key)
	}
	return nil, paymentdomain.ErrPaymentNotFound
}

func (f *fakePaymentRepository) FindPendingExpired(ctx context.Context, before time.Time, limit int) ([]*paymentdomain.Payment, error) {
	return nil, nil
}

func (f *fakePaymentRepository) FindByOrderID(ctx context.Context, orderID string) ([]*paymentdomain.Payment, error) {
	return nil, nil
}

// ============================================================
// REFUND REPOSITORY FAKE
// ============================================================

type fakeRefundRepository struct {
	createFunc        func(ctx context.Context, r *paymentdomain.Refund) error
	updateFunc        func(ctx context.Context, r *paymentdomain.Refund) error
	findByIDFunc      func(ctx context.Context, id string) (*paymentdomain.Refund, error)
	listByPaymentFunc func(ctx context.Context, paymentID string) ([]*paymentdomain.Refund, error)

	mu      sync.Mutex
	created []*paymentdomain.Refund
	updated []*paymentdomain.Refund
}

func (f *fakeRefundRepository) Create(ctx context.Context, r *paymentdomain.Refund) error {
	f.mu.Lock()
	f.created = append(f.created, r)
	f.mu.Unlock()
	if f.createFunc != nil {
		return f.createFunc(ctx, r)
	}
	return nil
}

func (f *fakeRefundRepository) Update(ctx context.Context, r *paymentdomain.Refund) error {
	f.mu.Lock()
	f.updated = append(f.updated, r)
	f.mu.Unlock()
	if f.updateFunc != nil {
		return f.updateFunc(ctx, r)
	}
	return nil
}

func (f *fakeRefundRepository) FindByID(ctx context.Context, id string) (*paymentdomain.Refund, error) {
	if f.findByIDFunc != nil {
		return f.findByIDFunc(ctx, id)
	}
	return nil, paymentdomain.ErrRefundNotFound
}

func (f *fakeRefundRepository) ListByPayment(ctx context.Context, paymentID string) ([]*paymentdomain.Refund, error) {
	if f.listByPaymentFunc != nil {
		return f.listByPaymentFunc(ctx, paymentID)
	}
	return nil, nil
}

func (f *fakeRefundRepository) FindPendingBefore(ctx context.Context, before time.Time, limit int) ([]*paymentdomain.Refund, error) {
	return nil, nil
}

// ============================================================
// WEBHOOK EVENT REPOSITORY FAKE
// ============================================================

type fakeWebhookEventRepository struct {
	recordIfNewFunc func(ctx context.Context, e *paymentdomain.WebhookEvent) error
	updateFunc      func(ctx context.Context, e *paymentdomain.WebhookEvent) error
	findByIDFunc    func(ctx context.Context, id string) (*paymentdomain.WebhookEvent, error)

	mu      sync.Mutex
	created []*paymentdomain.WebhookEvent
	updated []*paymentdomain.WebhookEvent
}

func (f *fakeWebhookEventRepository) RecordIfNew(ctx context.Context, e *paymentdomain.WebhookEvent) error {
	f.mu.Lock()
	f.created = append(f.created, e)
	f.mu.Unlock()
	if f.recordIfNewFunc != nil {
		return f.recordIfNewFunc(ctx, e)
	}
	return nil
}

func (f *fakeWebhookEventRepository) Update(ctx context.Context, e *paymentdomain.WebhookEvent) error {
	f.mu.Lock()
	f.updated = append(f.updated, e)
	f.mu.Unlock()
	if f.updateFunc != nil {
		return f.updateFunc(ctx, e)
	}
	return nil
}

func (f *fakeWebhookEventRepository) FindByID(ctx context.Context, id string) (*paymentdomain.WebhookEvent, error) {
	if f.findByIDFunc != nil {
		return f.findByIDFunc(ctx, id)
	}
	return nil, paymentdomain.ErrWebhookEventNotFound
}

func (f *fakeWebhookEventRepository) FindUnprocessed(ctx context.Context, limit int) ([]*paymentdomain.WebhookEvent, error) {
	return nil, nil
}

// ============================================================
// PROVIDER REGISTRY FAKE
// ============================================================

type fakeProviderRegistry struct {
	byNameFunc   func(name string) (paymentdomain.PaymentProvider, error)
	byMethodFunc func(method paymentdomain.PaymentMethod) (paymentdomain.PaymentProvider, error)
}

func (f *fakeProviderRegistry) ByName(name string) (paymentdomain.PaymentProvider, error) {
	if f.byNameFunc != nil {
		return f.byNameFunc(name)
	}
	return nil, paymentdomain.ErrProviderNotFound
}

func (f *fakeProviderRegistry) ByMethod(method paymentdomain.PaymentMethod) (paymentdomain.PaymentProvider, error) {
	if f.byMethodFunc != nil {
		return f.byMethodFunc(method)
	}
	return nil, paymentdomain.ErrProviderNotFound
}

// ============================================================
// PROVIDER FAKE
// ============================================================

type fakeProvider struct {
	name   string
	method paymentdomain.PaymentMethod

	initiateFunc     func(ctx context.Context, req paymentdomain.InitiateRequest) (*paymentdomain.InitiateResult, error)
	verifyFunc       func(ctx context.Context, ref string) (*paymentdomain.ProviderStatus, error)
	refundFunc       func(ctx context.Context, req paymentdomain.RefundRequest) (*paymentdomain.RefundResult, error)
	parseWebhookFunc func(ctx context.Context, payload []byte, headers map[string]string) (*paymentdomain.WebhookEventData, error)
}

func (f *fakeProvider) Name() string                        { return f.name }
func (f *fakeProvider) Method() paymentdomain.PaymentMethod { return f.method }

func (f *fakeProvider) Initiate(ctx context.Context, req paymentdomain.InitiateRequest) (*paymentdomain.InitiateResult, error) {
	if f.initiateFunc != nil {
		return f.initiateFunc(ctx, req)
	}
	return &paymentdomain.InitiateResult{
		ProviderReference: "fake-ref",
		Status:            paymentdomain.PaymentStatusPending,
	}, nil
}

func (f *fakeProvider) Verify(ctx context.Context, ref string) (*paymentdomain.ProviderStatus, error) {
	if f.verifyFunc != nil {
		return f.verifyFunc(ctx, ref)
	}
	return &paymentdomain.ProviderStatus{
		Reference: ref,
		Status:    paymentdomain.PaymentStatusPending,
	}, nil
}

func (f *fakeProvider) Refund(ctx context.Context, req paymentdomain.RefundRequest) (*paymentdomain.RefundResult, error) {
	if f.refundFunc != nil {
		return f.refundFunc(ctx, req)
	}
	return &paymentdomain.RefundResult{
		ProviderReference: "fake-refund-ref",
		Status:            paymentdomain.PaymentStatusSucceeded,
	}, nil
}

func (f *fakeProvider) ParseWebhook(ctx context.Context, payload []byte, headers map[string]string) (*paymentdomain.WebhookEventData, error) {
	if f.parseWebhookFunc != nil {
		return f.parseWebhookFunc(ctx, payload, headers)
	}
	return &paymentdomain.WebhookEventData{
		ProviderEventID:   "fake-evt",
		ProviderReference: "fake-ref",
		OurPaymentID:      "fake-payment-id",
		Status:            paymentdomain.PaymentStatusSucceeded,
	}, nil
}

// ============================================================
// NOTIFIER FAKE
// ============================================================

type fakeNotifier struct {
	mu sync.Mutex

	initiated []*paymentdomain.Payment
	succeeded []*paymentdomain.Payment
	failed    []*paymentdomain.Payment
	expired   []*paymentdomain.Payment
	refunded  []*paymentdomain.Refund
}

func (f *fakeNotifier) PaymentInitiated(ctx context.Context, p *paymentdomain.Payment, o *paymentdomain.Order) error {
	f.mu.Lock()
	f.initiated = append(f.initiated, p)
	f.mu.Unlock()
	return nil
}

func (f *fakeNotifier) PaymentSucceeded(ctx context.Context, p *paymentdomain.Payment, o *paymentdomain.Order) error {
	f.mu.Lock()
	f.succeeded = append(f.succeeded, p)
	f.mu.Unlock()
	return nil
}

func (f *fakeNotifier) PaymentFailed(ctx context.Context, p *paymentdomain.Payment, o *paymentdomain.Order) error {
	f.mu.Lock()
	f.failed = append(f.failed, p)
	f.mu.Unlock()
	return nil
}

func (f *fakeNotifier) PaymentExpired(ctx context.Context, p *paymentdomain.Payment, o *paymentdomain.Order) error {
	f.mu.Lock()
	f.expired = append(f.expired, p)
	f.mu.Unlock()
	return nil
}

func (f *fakeNotifier) RefundIssued(ctx context.Context, r *paymentdomain.Refund, p *paymentdomain.Payment) error {
	f.mu.Lock()
	f.refunded = append(f.refunded, r)
	f.mu.Unlock()
	return nil
}

// ============================================================
// REGISTRATION CONFIRMER FAKE
// ============================================================

type fakeRegistrationConfirmer struct {
	confirmFunc func(ctx context.Context, regID string) error
	expireFunc  func(ctx context.Context, regID string) error

	mu        sync.Mutex
	confirmed []string
	expired   []string
}

func (f *fakeRegistrationConfirmer) ConfirmRegistration(ctx context.Context, regID string) error {
	f.mu.Lock()
	f.confirmed = append(f.confirmed, regID)
	f.mu.Unlock()
	if f.confirmFunc != nil {
		return f.confirmFunc(ctx, regID)
	}
	return nil
}

func (f *fakeRegistrationConfirmer) ExpirePending(ctx context.Context, regID string) error {
	f.mu.Lock()
	f.expired = append(f.expired, regID)
	f.mu.Unlock()
	if f.expireFunc != nil {
		return f.expireFunc(ctx, regID)
	}
	return nil
}

// ============================================================
// UNIT OF WORK FAKE
// ============================================================

type fakeUnitOfWork struct {
	deps   *Dependencies
	doFunc func(ctx context.Context, fn func(paymentdomain.Repositories) error) error
}

func (f *fakeUnitOfWork) Do(ctx context.Context, fn func(paymentdomain.Repositories) error) error {
	if f.doFunc != nil {
		return f.doFunc(ctx, fn)
	}
	// Default: execute the callback with the same repos (no real tx).
	return fn(paymentdomain.Repositories{
		Orders:   f.deps.Orders,
		Payments: f.deps.Payments,
		Refunds:  f.deps.Refunds,
		Webhooks: f.deps.Webhooks,
	})
}

// ============================================================
// DETERMINISTIC GENERATORS
// ============================================================

// fixedIDGenerator returns IDs from a fixed list in order.
type fixedIDGenerator struct {
	mu    sync.Mutex
	ids   []string
	index int
}

func newFixedIDGenerator(ids ...string) *fixedIDGenerator {
	return &fixedIDGenerator{ids: ids}
}

func (g *fixedIDGenerator) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.index >= len(g.ids) {
		return "fallback-id"
	}
	id := g.ids[g.index]
	g.index++
	return id
}

// fixedClock returns a fixed time for every call.
type fixedClock struct {
	t time.Time
}

func newFixedClock() *fixedClock {
	return &fixedClock{t: time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)}
}

func (c *fixedClock) Now() time.Time { return c.t }

// ============================================================
// TEST HARNESS
// ============================================================

// newTestService wires a service with fresh fakes and returns both
// the service and the deps so tests can inspect captured calls.
func newTestService(t *testing.T, customize func(*Dependencies)) (Service, *Dependencies) {
	t.Helper()

	deps := &Dependencies{
		Orders:        &fakeOrderRepository{},
		Payments:      &fakePaymentRepository{},
		Refunds:       &fakeRefundRepository{},
		Webhooks:      &fakeWebhookEventRepository{},
		Providers:     &fakeProviderRegistry{},
		Notifier:      &fakeNotifier{},
		Registrations: &fakeRegistrationConfirmer{},
		IDGenerator:   newFixedIDGenerator("id-1", "id-2", "id-3", "id-4", "id-5"),
		Clock:         newFixedClock(),
	}
	deps.UnitOfWork = &fakeUnitOfWork{deps: deps}

	if customize != nil {
		customize(deps)
	}

	return New(*deps), deps
}

// ============================================================
// FIXTURES
// ============================================================

// newPendingOrder builds a valid pending order for tests.
func newPendingOrder(t *testing.T) *paymentdomain.Order {
	t.Helper()

	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)
	items := []paymentdomain.OrderItem{
		{
			TicketTypeID: "tkt-1",
			Quantity:     1,
			UnitPrice:    1500_00,
			LineTotal:    1500_00,
		},
	}

	o, err := paymentdomain.NewOrder(
		"order-1",
		"reg-1",
		"user-1",
		"",
		"KES",
		items,
		30*time.Minute,
		now,
	)
	if err != nil {
		t.Fatalf("newPendingOrder: %v", err)
	}
	return o
}

// newPendingPayment builds a valid pending payment for tests.
func newPendingPayment(t *testing.T, orderID string) *paymentdomain.Payment {
	t.Helper()

	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)
	p, err := paymentdomain.NewPayment(
		"pay-1",
		orderID,
		"flutterwave",
		paymentdomain.PaymentMethodMpesa,
		1500_00,
		"KES",
		"idem-1",
		30*time.Minute,
		now,
	)
	if err != nil {
		t.Fatalf("newPendingPayment: %v", err)
	}
	return p
}