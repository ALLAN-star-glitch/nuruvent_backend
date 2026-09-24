// internal/modules/payment/infrastructure/postgres/unit_of_work_test.go

package postgres

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// errSimulated triggers a rollback in Do callbacks.
var errSimulated = errors.New("simulated failure")

func TestUnitOfWork_CommitsOnSuccess(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(outer *gorm.DB) {
		ctx := testCtx(t)
		uow := NewUnitOfWork(outer)

		order := newOrderFixture(t)
		payment := newPaymentFixture(t, order.ID)

		err := uow.Do(ctx, func(repos paymentdomain.Repositories) error {
			if err := repos.Orders.Create(ctx, order); err != nil {
				return err
			}
			if err := repos.Payments.Create(ctx, payment); err != nil {
				return err
			}
			return nil
		})
		require.NoError(t, err)

		orderRepo := NewOrderRepository(outer)
		paymentRepo := NewPaymentRepository(outer)

		_, err = orderRepo.FindByID(ctx, order.ID)
		require.NoError(t, err)

		_, err = paymentRepo.FindByID(ctx, payment.ID)
		require.NoError(t, err)
	})
}

func TestUnitOfWork_RollsBackOnError(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(outer *gorm.DB) {
		ctx := testCtx(t)
		uow := NewUnitOfWork(outer)

		order := newOrderFixture(t)

		err := uow.Do(ctx, func(repos paymentdomain.Repositories) error {
			if err := repos.Orders.Create(ctx, order); err != nil {
				return err
			}
			return errSimulated
		})
		if !errors.Is(err, errSimulated) {
			t.Fatalf("expected errSimulated, got %v", err)
		}

		orderRepo := NewOrderRepository(outer)
		_, err = orderRepo.FindByID(ctx, order.ID)
		if !errors.Is(err, paymentdomain.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound (rolled back), got %v", err)
		}
	})
}