// internal/modules/payment/infrastructure/postgres/refund_repository_test.go

package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// createSucceededPayment creates an order + payment in the DB, then
// marks the payment as succeeded.
func createSucceededPayment(t *testing.T, tx *gorm.DB) *paymentdomain.Payment {
	t.Helper()
	ctx := testCtx(t)
	order := createPersistedOrder(t, tx)

	paymentRepo := NewPaymentRepository(tx)
	p := newPaymentFixture(t, order.ID)
	require.NoError(t, p.MarkSucceeded("ref-"+p.ID, time.Now().UTC()))
	require.NoError(t, paymentRepo.Create(ctx, p))
	return p
}

func TestRefundRepository_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewRefundRepository(tx)
		payment := createSucceededPayment(t, tx)

		r := newRefundFixture(t, payment.ID)
		require.NoError(t, repo.Create(ctx, r))

		got, err := repo.FindByID(ctx, r.ID)
		require.NoError(t, err)
		if got.Amount != r.Amount {
			t.Errorf("amount: got %d, want %d", got.Amount, r.Amount)
		}
		if got.Status != paymentdomain.PaymentStatusPending {
			t.Errorf("status: got %s, want pending", got.Status)
		}
	})
}

func TestRefundRepository_Update_ChangesStatus(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewRefundRepository(tx)
		payment := createSucceededPayment(t, tx)

		r := newRefundFixture(t, payment.ID)
		require.NoError(t, repo.Create(ctx, r))

		require.NoError(t, r.MarkSucceeded("provider-refund-1", time.Now().UTC()))
		require.NoError(t, repo.Update(ctx, r))

		got, err := repo.FindByID(ctx, r.ID)
		require.NoError(t, err)
		if got.Status != paymentdomain.PaymentStatusSucceeded {
			t.Errorf("status: got %s, want succeeded", got.Status)
		}
		if got.ProviderReference != "provider-refund-1" {
			t.Errorf("provider_reference: got %q", got.ProviderReference)
		}
	})
}

func TestRefundRepository_ListByPayment_NewestFirst(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewRefundRepository(tx)
		payment := createSucceededPayment(t, tx)

		base := time.Now().UTC().Add(-3 * time.Minute)
		var ids []string
		for i := 0; i < 3; i++ {
			r := newRefundFixture(t, payment.ID, func(r *paymentdomain.Refund) {
				r.CreatedAt = base.Add(time.Duration(i) * time.Minute)
				r.UpdatedAt = base.Add(time.Duration(i) * time.Minute)
			})
			require.NoError(t, repo.Create(ctx, r))
			ids = append(ids, r.ID)
		}

		got, err := repo.ListByPayment(ctx, payment.ID)
		require.NoError(t, err)
		if len(got) != 3 {
			t.Fatalf("expected 3 refunds, got %d", len(got))
		}
		if got[0].ID != ids[2] {
			t.Errorf("expected newest first: got %s, want %s", got[0].ID, ids[2])
		}
	})
}

func TestRefundRepository_FindPendingBefore(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewRefundRepository(tx)
		payment := createSucceededPayment(t, tx)

		oldTime := time.Now().UTC().Add(-time.Hour)

		old := newRefundFixture(t, payment.ID, func(r *paymentdomain.Refund) {
			r.CreatedAt = oldTime
			r.UpdatedAt = oldTime
		})
		require.NoError(t, repo.Create(ctx, old))

		fresh := newRefundFixture(t, payment.ID)
		require.NoError(t, repo.Create(ctx, fresh))

		cutoff := time.Now().UTC().Add(-30 * time.Minute)
		got, err := repo.FindPendingBefore(ctx, cutoff, 10)
		require.NoError(t, err)
		if len(got) != 1 {
			t.Fatalf("expected 1 old refund, got %d", len(got))
		}
		if got[0].ID != old.ID {
			t.Errorf("wrong refund: got %s, want %s", got[0].ID, old.ID)
		}
	})
}

// Silence unused-import warning for "errors" if it ends up unused.
var _ = errors.New