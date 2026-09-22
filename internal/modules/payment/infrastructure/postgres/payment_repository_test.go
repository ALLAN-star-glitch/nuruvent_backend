// internal/modules/payment/infrastructure/postgres/payment_repository_test.go

package postgres

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// createPersistedOrder creates an order in the DB and returns it.
func createPersistedOrder(t *testing.T, tx *gorm.DB) *paymentdomain.Order {
	t.Helper()
	ctx := testCtx(t)
	repo := NewOrderRepository(tx)
	order := newOrderFixture(t)
	require.NoError(t, repo.Create(ctx, order))
	return order
}

// ============================================================
// CREATE + FIND ROUND TRIP
// ============================================================

func TestPaymentRepository_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)
		order := createPersistedOrder(t, tx)

		p := newPaymentFixture(t, order.ID)
		require.NoError(t, repo.Create(ctx, p))

		got, err := repo.FindByID(ctx, p.ID)
		require.NoError(t, err)
		if got.Status != paymentdomain.PaymentStatusPending {
			t.Errorf("status: got %s, want pending", got.Status)
		}
		if got.Amount != p.Amount {
			t.Errorf("amount: got %d, want %d", got.Amount, p.Amount)
		}
	})
}

// ============================================================
// IDEMPOTENCY
// ============================================================

func TestPaymentRepository_Create_RejectsDuplicateIdempotencyKey(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)
		order := createPersistedOrder(t, tx)

		key := uuid.NewString()

		first := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
			p.IdempotencyKey = key
		})
		require.NoError(t, repo.Create(ctx, first))

		second := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
			p.IdempotencyKey = key
		})
		err := repo.Create(ctx, second)
		if !errors.Is(err, paymentdomain.ErrDuplicatePayment) {
			t.Fatalf("expected ErrDuplicatePayment, got %v", err)
		}
	})
}

// Fires two goroutines attempting the same insert concurrently. Exactly
// one must succeed; the other must receive ErrDuplicatePayment.
//
// IMPORTANT: this test does NOT use withTx. GORM's transaction is
// bound to a single database connection, which cannot be shared by
// concurrent goroutines (Postgres returns "conn busy"). Real
// production traffic uses separate connections from the pool.
//
// Instead, we use the raw connection pool and clean up manually.
func TestPaymentRepository_Create_ConcurrentSameKey(t *testing.T) {
	db := newTestDB(t)
	ctx := testCtx(t)

	// Create the parent order first, in its own connection.
	orderRepo := NewOrderRepository(db)
	order := newOrderFixture(t)
	require.NoError(t, orderRepo.Create(ctx, order))

	// Ensure cleanup even if the test fails.
	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM payments WHERE order_id = ?", order.ID).Error
		_ = db.Exec("DELETE FROM order_items WHERE order_id = ?", order.ID).Error
		_ = db.Exec("DELETE FROM orders WHERE id = ?", order.ID).Error
	})

	key := uuid.NewString()

	var wg sync.WaitGroup
	errs := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each goroutine gets its own repository bound to the
			// connection pool — GORM will pull a separate connection
			// for each.
			repo := NewPaymentRepository(db)
			p := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
				p.IdempotencyKey = key
			})
			errs <- repo.Create(ctx, p)
		}()
	}
	wg.Wait()
	close(errs)

	successes := 0
	duplicates := 0
	for err := range errs {
		if err == nil {
			successes++
		} else if errors.Is(err, paymentdomain.ErrDuplicatePayment) {
			duplicates++
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if successes != 1 || duplicates != 1 {
		t.Errorf("expected 1 success and 1 duplicate, got successes=%d duplicates=%d",
			successes, duplicates)
	}
}

// ============================================================
// FIND BY PROVIDER REFERENCE
// ============================================================

func TestPaymentRepository_FindByProviderReference(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)
		order := createPersistedOrder(t, tx)

		ref := "MPESA-ABC123"
		p := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
			require.NoError(t, p.MarkSucceeded(ref, time.Now().UTC()))
		})
		require.NoError(t, repo.Create(ctx, p))

		got, err := repo.FindByProviderReference(ctx, "stub", ref)
		require.NoError(t, err)
		if got.ID != p.ID {
			t.Errorf("wrong payment: got %s, want %s", got.ID, p.ID)
		}
	})
}

func TestPaymentRepository_FindByProviderReference_NotFound(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)

		_, err := repo.FindByProviderReference(ctx, "stub", "does-not-exist")
		if !errors.Is(err, paymentdomain.ErrPaymentNotFound) {
			t.Fatalf("expected ErrPaymentNotFound, got %v", err)
		}
	})
}

// ============================================================
// FIND BY IDEMPOTENCY KEY
// ============================================================

func TestPaymentRepository_FindByIdempotencyKey(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)
		order := createPersistedOrder(t, tx)

		key := uuid.NewString()
		p := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
			p.IdempotencyKey = key
		})
		require.NoError(t, repo.Create(ctx, p))

		got, err := repo.FindByIdempotencyKey(ctx, order.ID, key)
		require.NoError(t, err)
		if got.ID != p.ID {
			t.Errorf("wrong payment")
		}
	})
}

func TestPaymentRepository_FindByIdempotencyKey_NotFound(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)

		_, err := repo.FindByIdempotencyKey(ctx, uuid.NewString(), uuid.NewString())
		if !errors.Is(err, paymentdomain.ErrPaymentNotFound) {
			t.Fatalf("expected ErrPaymentNotFound, got %v", err)
		}
	})
}

// ============================================================
// FIND PENDING EXPIRED
// ============================================================

func TestPaymentRepository_FindPendingExpired(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)
		order := createPersistedOrder(t, tx)

		pastDeadline := time.Now().UTC().Add(-time.Hour)
		expired := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
			p.ExpiresAt = pastDeadline
		})
		require.NoError(t, repo.Create(ctx, expired))

		succeeded := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
			p.ExpiresAt = pastDeadline
			require.NoError(t, p.MarkSucceeded("ref-succeeded", time.Now().UTC()))
		})
		require.NoError(t, repo.Create(ctx, succeeded))

		got, err := repo.FindPendingExpired(ctx, time.Now().UTC(), 10)
		require.NoError(t, err)
		if len(got) != 1 {
			t.Fatalf("expected 1 pending expired, got %d", len(got))
		}
		if got[0].ID != expired.ID {
			t.Errorf("wrong payment: got %s, want %s", got[0].ID, expired.ID)
		}
	})
}

// ============================================================
// FIND BY ORDER ID
// ============================================================

func TestPaymentRepository_FindByOrderID_NewestFirst(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewPaymentRepository(tx)
		order := createPersistedOrder(t, tx)

		base := time.Now().UTC().Add(-3 * time.Minute)
		var ids []string
		for i := 0; i < 3; i++ {
			p := newPaymentFixture(t, order.ID, func(p *paymentdomain.Payment) {
				p.InitiatedAt = base.Add(time.Duration(i) * time.Minute)
				p.CreatedAt = base.Add(time.Duration(i) * time.Minute)
				p.UpdatedAt = base.Add(time.Duration(i) * time.Minute)
				p.ExpiresAt = base.Add(time.Duration(i) * time.Minute).Add(30 * time.Minute)
			})
			require.NoError(t, repo.Create(ctx, p))
			ids = append(ids, p.ID)
		}

		got, err := repo.FindByOrderID(ctx, order.ID)
		require.NoError(t, err)
		if len(got) != 3 {
			t.Fatalf("expected 3 payments, got %d", len(got))
		}
		if got[0].ID != ids[2] {
			t.Errorf("expected newest first: got %s, want %s", got[0].ID, ids[2])
		}
	})
}