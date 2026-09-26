package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type JoinTokenRepository struct {
	db *gorm.DB
}

func NewJoinTokenRepository(db *gorm.DB) *JoinTokenRepository {
	return &JoinTokenRepository{db: db}
}

func (r *JoinTokenRepository) Create(ctx context.Context, t *attendance.JoinToken) error {
	if err := r.db.WithContext(ctx).Create(toJoinTokenModel(t)).Error; err != nil {
		return translateError(err, "create join token", attendance.ErrTokenNotFound)
	}
	return nil
}

func (r *JoinTokenRepository) FindByHash(ctx context.Context, hash string) (*attendance.JoinToken, error) {
	var m JoinTokenModel
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrTokenNotFound
	}
	if err != nil {
		return nil, translateError(err, "find join token", attendance.ErrTokenNotFound)
	}
	return toJoinTokenDomain(&m), nil
}

func (r *JoinTokenRepository) RevokeByAttendeeSession(
	ctx context.Context,
	attendeeID, sessionID string,
	now time.Time,
) error {
	res := r.db.WithContext(ctx).
		Model(&JoinTokenModel{}).
		Where("attendee_id = ? AND session_id = ? AND revoked_at IS NULL",
			attendeeID, sessionID).
		Update("revoked_at", now)
	if res.Error != nil {
		return translateError(res.Error, "revoke join tokens", attendance.ErrTokenNotFound)
	}
	return nil
}

func (r *JoinTokenRepository) RevokeByID(ctx context.Context, id string, now time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&JoinTokenModel{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now)
	if res.Error != nil {
		return translateError(res.Error, "revoke join token", attendance.ErrTokenNotFound)
	}
	if res.RowsAffected == 0 {
		return attendance.ErrTokenNotFound
	}
	return nil
}

func (r *JoinTokenRepository) ListByAttendee(
	ctx context.Context,
	attendeeID string,
) ([]*attendance.JoinToken, error) {
	var models []JoinTokenModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ?", attendeeID).
		Order("issued_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list join tokens", attendance.ErrTokenNotFound)
	}
	out := make([]*attendance.JoinToken, 0, len(models))
	for i := range models {
		out = append(out, toJoinTokenDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion
var _ attendance.JoinTokenRepository = (*JoinTokenRepository)(nil)