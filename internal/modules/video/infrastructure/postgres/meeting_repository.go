// internal/modules/video/infrastructure/postgres/meeting_repository.go

package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// MeetingRepository implements videodomain.MeetingRepository against
// Postgres.
type MeetingRepository struct {
	db *gorm.DB
}

func NewMeetingRepository(db *gorm.DB) *MeetingRepository {
	return &MeetingRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

func (r *MeetingRepository) Create(ctx context.Context, m *videodomain.Meeting) error {
	model := toMeetingModel(m)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create meeting", videodomain.ErrMeetingNotFound)
	}
	return nil
}

func (r *MeetingRepository) Update(ctx context.Context, m *videodomain.Meeting) error {
	res := r.db.WithContext(ctx).
		Model(&VideoMeetingModel{}).
		Where("id = ?", m.ID).
		Updates(map[string]any{
			"user_id":      m.UserID,
			"platform":     string(m.Platform),
			"external_id":  m.ExternalID,
			"join_url":     m.JoinURL,
			"start_url":    m.StartURL,
			"password":     m.Password,
			"topic":        m.Topic,
			"start_time":   m.StartTime,
			"duration_sec": int(m.Duration.Seconds()),
			"timezone":     m.Timezone,
			"updated_at":   m.UpdatedAt,
		})
	if res.Error != nil {
		return translateError(res.Error, "update meeting", videodomain.ErrMeetingNotFound)
	}
	if res.RowsAffected == 0 {
		return videodomain.ErrMeetingNotFound
	}
	return nil
}

func (r *MeetingRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&VideoMeetingModel{})
	if res.Error != nil {
		return translateError(res.Error, "delete meeting", videodomain.ErrMeetingNotFound)
	}
	if res.RowsAffected == 0 {
		return videodomain.ErrMeetingNotFound
	}
	return nil
}

// ============================================================
// READ
// ============================================================

func (r *MeetingRepository) FindByID(ctx context.Context, id string) (*videodomain.Meeting, error) {
	var m VideoMeetingModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, videodomain.ErrMeetingNotFound
	}
	if err != nil {
		return nil, translateError(err, "find meeting", videodomain.ErrMeetingNotFound)
	}
	return toMeetingDomain(&m), nil
}

func (r *MeetingRepository) FindByExternalID(
	ctx context.Context,
	platform videodomain.Platform,
	externalID string,
) (*videodomain.Meeting, error) {
	var m VideoMeetingModel
	err := r.db.WithContext(ctx).
		Where("platform = ? AND external_id = ?", string(platform), externalID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, videodomain.ErrMeetingNotFound
	}
	if err != nil {
		return nil, translateError(err, "find meeting by external id", videodomain.ErrMeetingNotFound)
	}
	return toMeetingDomain(&m), nil
}

// compile-time assertion
var _ videodomain.MeetingRepository = (*MeetingRepository)(nil)