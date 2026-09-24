// internal/modules/payment/infrastructure/postgres/helpers_test.go

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// errRollbackForTest is a sentinel used to trigger rollback of the
// per-test transaction. Every test that uses withTx rolls back so no
// state leaks between tests.
var errRollbackForTest = errors.New("rollback for test")

// withTx runs fn inside a transaction that is always rolled back. This
// gives each test a clean slate without recreating the database.
func withTx(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB)) {
	t.Helper()

	err := db.Transaction(func(tx *gorm.DB) error {
		fn(tx)
		return errRollbackForTest
	})
	if err != nil && !errors.Is(err, errRollbackForTest) {
		t.Fatalf("transaction failed: %v", err)
	}
}

// ============================================================
// FIXTURE BUILDERS
// ============================================================

// newOrderFixture builds a valid Order for tests.
func newOrderFixture(t *testing.T, opts ...func(*paymentdomain.Order)) *paymentdomain.Order {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	items := []paymentdomain.OrderItem{
		{
			TicketTypeID: uuid.NewString(),
			Quantity:     1,
			UnitPrice:    1500_00,
			Discount:     0,
			LineTotal:    1500_00,
		},
	}

	o, err := paymentdomain.NewOrder(
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
		"",
		"KES",
		items,
		30*time.Minute,
		now,
	)
	require.NoError(t, err)

	for _, opt := range opts {
		opt(o)
	}
	return o
}

// newPaymentFixture builds a valid Payment for tests.
func newPaymentFixture(t *testing.T, orderID string, opts ...func(*paymentdomain.Payment)) *paymentdomain.Payment {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	p, err := paymentdomain.NewPayment(
		uuid.NewString(),
		orderID,
		"stub",
		paymentdomain.PaymentMethodMpesa,
		1500_00,
		"KES",
		uuid.NewString(),
		30*time.Minute,
		now,
	)
	require.NoError(t, err)

	for _, opt := range opts {
		opt(p)
	}
	return p
}

// newRefundFixture builds a valid Refund for tests.
func newRefundFixture(t *testing.T, paymentID string, opts ...func(*paymentdomain.Refund)) *paymentdomain.Refund {
	t.Helper() // Teslls Go this is a helper

	now := time.Now().UTC().Truncate(time.Second)
	r, err := paymentdomain.NewRefund(
		uuid.NewString(),
		paymentID,
		500_00,
		"KES",
		"test refund",
		uuid.NewString(),
		now,
	)
	require.NoError(t, err)

	for _, opt := range opts {
		opt(r)
	}
	return r
}

// newWebhookEventFixture builds a valid WebhookEvent for tests.
func newWebhookEventFixture(t *testing.T, opts ...func(*paymentdomain.WebhookEvent)) *paymentdomain.WebhookEvent {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)
	e := paymentdomain.NewWebhookEvent(
		uuid.NewString(),
		"stub",
		uuid.NewString(),
		true,
		[]byte(`{"status":"succeeded"}`),
		now,
	)

	for _, opt := range opts {
		opt(e)
	}
	return e
}

// testCtx returns a context that cancels when the test finishes.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}