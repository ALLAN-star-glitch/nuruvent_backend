// internal/modules/payment/infrastructure/postgres/webhook_event_repository_test.go

package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

func TestWebhookEventRepository_RecordIfNew(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewWebhookEventRepository(tx)

		e := newWebhookEventFixture(t)
		require.NoError(t, repo.RecordIfNew(ctx, e))

		got, err := repo.FindByID(ctx, e.ID)
		require.NoError(t, err)
		if got.ProviderEventID != e.ProviderEventID {
			t.Errorf("provider_event_id: got %q", got.ProviderEventID)
		}
		if len(got.Payload) != len(e.Payload) {
			t.Errorf("payload length: got %d, want %d", len(got.Payload), len(e.Payload))
		}
	})
}

func TestWebhookEventRepository_RecordIfNew_RejectsDuplicate(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewWebhookEventRepository(tx)

		providerEventID := "evt-" + time.Now().Format("150405.000")

		first := newWebhookEventFixture(t, func(e *paymentdomain.WebhookEvent) {
			e.ProviderEventID = providerEventID
		})
		require.NoError(t, repo.RecordIfNew(ctx, first))

		second := newWebhookEventFixture(t, func(e *paymentdomain.WebhookEvent) {
			e.ProviderEventID = providerEventID
		})
		err := repo.RecordIfNew(ctx, second)
		if !errors.Is(err, paymentdomain.ErrDuplicateWebhook) {
			t.Fatalf("expected ErrDuplicateWebhook, got %v", err)
		}
	})
}

func TestWebhookEventRepository_Update_MarksProcessed(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewWebhookEventRepository(tx)

		e := newWebhookEventFixture(t)
		require.NoError(t, repo.RecordIfNew(ctx, e))

		e.MarkProcessed(time.Now().UTC())
		require.NoError(t, repo.Update(ctx, e))

		got, err := repo.FindByID(ctx, e.ID)
		require.NoError(t, err)
		if !got.IsProcessed() {
			t.Errorf("expected IsProcessed true, got false")
		}
	})
}

func TestWebhookEventRepository_Update_MarksFailed(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewWebhookEventRepository(tx)

		e := newWebhookEventFixture(t)
		require.NoError(t, repo.RecordIfNew(ctx, e))

		e.MarkFailed("simulated processing error")
		require.NoError(t, repo.Update(ctx, e))

		got, err := repo.FindByID(ctx, e.ID)
		require.NoError(t, err)
		if !got.HasFailed() {
			t.Errorf("expected HasFailed true, got false")
		}
		if got.ProcessingError != "simulated processing error" {
			t.Errorf("processing_error: got %q", got.ProcessingError)
		}
	})
}

func TestWebhookEventRepository_FindUnprocessed(t *testing.T) {
	db := newTestDB(t)
	withTx(t, db, func(tx *gorm.DB) {
		ctx := testCtx(t)
		repo := NewWebhookEventRepository(tx)

		pending := newWebhookEventFixture(t)
		require.NoError(t, repo.RecordIfNew(ctx, pending))

		processed := newWebhookEventFixture(t)
		require.NoError(t, repo.RecordIfNew(ctx, processed))
		processed.MarkProcessed(time.Now().UTC())
		require.NoError(t, repo.Update(ctx, processed))

		failed := newWebhookEventFixture(t)
		require.NoError(t, repo.RecordIfNew(ctx, failed))
		failed.MarkFailed("boom")
		require.NoError(t, repo.Update(ctx, failed))

		got, err := repo.FindUnprocessed(ctx, 10)
		require.NoError(t, err)

		if len(got) != 2 {
			t.Fatalf("expected 2 unprocessed events, got %d", len(got))
		}

		ids := map[string]bool{got[0].ID: true, got[1].ID: true}
		if !ids[pending.ID] {
			t.Errorf("pending event missing from results")
		}
		if !ids[failed.ID] {
			t.Errorf("failed event missing from results")
		}
		if ids[processed.ID] {
			t.Errorf("processed event should not appear")
		}
	})
}