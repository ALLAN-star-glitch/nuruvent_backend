// internal/modules/video/infrastructure/postgres/oauth_state_repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// OAuthStateRepository implements videodomain.OAuthStateRepository
// against Postgres.
type OAuthStateRepository struct {
	db *gorm.DB
}

func NewOAuthStateRepository(db *gorm.DB) *OAuthStateRepository {
	return &OAuthStateRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

func (r *OAuthStateRepository) Create(ctx context.Context, s *videodomain.OAuthState) error {
	model := toOAuthStateModel(s)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create oauth state", videodomain.ErrOAuthStateNotFound)
	}
	return nil
}

// Consume atomically fetches and marks a state as consumed.
//
// Implemented with a single UPDATE ... RETURNING. Two concurrent
// callbacks with the same state cannot both succeed — the second sees
// RowsAffected == 0 and gets ErrOAuthStateConsumed.
func (r *OAuthStateRepository) Consume(
	ctx context.Context,
	state string,
	now time.Time,
) (*videodomain.OAuthState, error) {
	var m VideoOAuthStateModel

	// Atomic claim: only updates if the row exists, hasn't been
	// consumed, and hasn't expired.
	err := r.db.WithContext(ctx).
		Raw(`
			UPDATE video_oauth_states
			SET consumed_at = ?
			WHERE state = ?
			  AND consumed_at IS NULL
			  AND expires_at > ?
			RETURNING state, user_id, platform, return_url,
			          created_at, expires_at, consumed_at
		`, now, state, now).
		Scan(&m).Error

	if err == nil && m.State != "" {
		return toOAuthStateDomain(&m), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("consume oauth state: %w", err)
	}

	// No row was claimed. Distinguish "not found" from "consumed" or
	// "expired" by looking up the existing record.
	var existing VideoOAuthStateModel
	lookupErr := r.db.WithContext(ctx).
		Where("state = ?", state).
		First(&existing).Error
	if errors.Is(lookupErr, gorm.ErrRecordNotFound) {
		return nil, videodomain.ErrOAuthStateNotFound
	}
	if lookupErr != nil {
		return nil, fmt.Errorf("lookup oauth state: %w", lookupErr)
	}
	if existing.ConsumedAt != nil {
		return nil, videodomain.ErrOAuthStateConsumed
	}
	if !now.Before(existing.ExpiresAt) {
		return nil, videodomain.ErrOAuthStateExpired
	}
	// Shouldn't reach here — the UPDATE would have succeeded. Treat
	// as transient failure.
	return nil, fmt.Errorf("consume oauth state: unexpected race")
}

// DeleteExpired removes consumed or expired states older than the
// given cutoff. Called by a background job.
func (r *OAuthStateRepository) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	res := r.db.WithContext(ctx).
		Where("expires_at < ? OR (consumed_at IS NOT NULL AND consumed_at < ?)",
			before, before).
		Delete(&VideoOAuthStateModel{})
	if res.Error != nil {
		return 0, fmt.Errorf("delete expired oauth states: %w", res.Error)
	}
	return int(res.RowsAffected), nil
}

// compile-time assertion
var _ videodomain.OAuthStateRepository = (*OAuthStateRepository)(nil)