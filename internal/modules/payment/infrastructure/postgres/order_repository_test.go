// internal/modules/payment/infrastructure/postgres/order_repository_test.go

package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ============================================================
// CREATE + FIND ROUND TRIP
// ============================================================

func TestOrderRepository_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		order := newOrderFixture(t)
		require.NoError(t, repo.Create(ctx, order))

		got, err := repo.FindByID(ctx, order.ID)
		require.NoError(t, err)

		if got.Status != order.Status {
			t.Errorf("status: got %s, want %s", got.Status, order.Status)
		}
		if got.TotalAmount != order.TotalAmount {
			t.Errorf("total: got %d, want %d", got.TotalAmount, order.TotalAmount)
		}
		if len(got.Items) != len(order.Items) {
			t.Errorf("items count: got %d, want %d", len(got.Items), len(order.Items))
		}
	})
}

func TestOrderRepository_Create_PersistsItems(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		order := newOrderFixture(t, func(o *paymentdomain.Order) {
			o.Items = []paymentdomain.OrderItem{
				{TicketTypeID: uuid.NewString(), Quantity: 2, UnitPrice: 1500_00, LineTotal: 3000_00},
				{TicketTypeID: uuid.NewString(), Quantity: 1, UnitPrice: 5000_00, Discount: 500_00, LineTotal: 4500_00},
			}
			o.Subtotal = 8000_00
			o.DiscountTotal = 500_00
			o.TotalAmount = 7500_00
		})
		require.NoError(t, repo.Create(ctx, order))

		got, err := repo.FindByID(ctx, order.ID)
		require.NoError(t, err)
		if len(got.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(got.Items))
		}
	})
}

// ============================================================
// FIND ACTIVE BY REGISTRATION
// ============================================================

func TestOrderRepository_FindActiveByRegistration_ReturnsPending(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		regID := uuid.NewString()
		order := newOrderFixture(t, func(o *paymentdomain.Order) {
			o.RegistrationID = regID
		})
		require.NoError(t, repo.Create(ctx, order))

		got, err := repo.FindActiveByRegistration(ctx, regID)
		require.NoError(t, err)
		if got.ID != order.ID {
			t.Errorf("wrong order returned: got %s, want %s", got.ID, order.ID)
		}
	})
}

func TestOrderRepository_FindActiveByRegistration_IgnoresPaidOrders(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		regID := uuid.NewString()
		order := newOrderFixture(t, func(o *paymentdomain.Order) {
			o.RegistrationID = regID
		})
		require.NoError(t, repo.Create(ctx, order))

		require.NoError(t, order.MarkPaid(time.Now().UTC()))
		require.NoError(t, repo.Update(ctx, order))

		_, err := repo.FindActiveByRegistration(ctx, regID)
		if !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})
}

func TestOrderRepository_FindActiveByRegistration_NoOrder(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		_, err := repo.FindActiveByRegistration(ctx, uuid.NewString())
		if !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})
}

// ============================================================
// UNIQUE CONSTRAINT: ONE ACTIVE ORDER PER REGISTRATION
// ============================================================

func TestOrderRepository_Create_EnforcesActiveUniqueness(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		regID := uuid.NewString()

		first := newOrderFixture(t, func(o *paymentdomain.Order) {
			o.RegistrationID = regID
		})
		require.NoError(t, repo.Create(ctx, first))

		second := newOrderFixture(t, func(o *paymentdomain.Order) {
			o.RegistrationID = regID
		})
		err := repo.Create(ctx, second)
		if err == nil {
			t.Fatal("expected error on second pending order, got nil")
		}
	})
}

// ============================================================
// FIND EXPIRED
// ============================================================

func TestOrderRepository_FindExpired(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		pastDeadline := time.Now().UTC().Add(-time.Hour)
		expired := newOrderFixture(t, func(o *paymentdomain.Order) {
			o.ExpiresAt = pastDeadline
		})
		require.NoError(t, repo.Create(ctx, expired))

		fresh := newOrderFixture(t)
		require.NoError(t, repo.Create(ctx, fresh))

		got, err := repo.FindExpired(ctx, time.Now().UTC(), 10)
		require.NoError(t, err)

		if len(got) != 1 {
			t.Fatalf("expected 1 expired order, got %d", len(got))
		}
		if got[0].ID != expired.ID {
			t.Errorf("wrong order: got %s, want %s", got[0].ID, expired.ID)
		}
	})
}

func TestOrderRepository_FindExpired_RespectsLimit(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		pastDeadline := time.Now().UTC().Add(-time.Hour)
		for i := 0; i < 5; i++ {
			o := newOrderFixture(t, func(o *paymentdomain.Order) {
				o.ExpiresAt = pastDeadline
			})
			require.NoError(t, repo.Create(ctx, o))
		}

		got, err := repo.FindExpired(ctx, time.Now().UTC(), 3)
		require.NoError(t, err)
		if len(got) != 3 {
			t.Errorf("expected 3 orders (limit), got %d", len(got))
		}
	})
}

// ============================================================
// UPDATE
// ============================================================

func TestOrderRepository_Update_ChangesStatus(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		order := newOrderFixture(t)
		require.NoError(t, repo.Create(ctx, order))

		require.NoError(t, order.MarkPaid(time.Now().UTC()))
		require.NoError(t, repo.Update(ctx, order))

		got, err := repo.FindByID(ctx, order.ID)
		require.NoError(t, err)
		if got.Status != paymentdomain.OrderStatusPaid {
			t.Errorf("status: got %s, want %s", got.Status, paymentdomain.OrderStatusPaid)
		}
		if got.PaidAt == nil {
			t.Error("paid_at should be set")
		}
	})
}

func TestOrderRepository_Update_NotFound(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewOrderRepository(tx)

		order := newOrderFixture(t) // not persisted
		err := repo.Update(ctx, order)
		if !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})
}