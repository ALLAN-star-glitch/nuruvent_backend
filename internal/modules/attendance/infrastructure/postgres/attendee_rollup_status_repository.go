package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type AttendeeRollupStatusRepository struct {
	db *gorm.DB
}

func NewAttendeeRollupStatusRepository(db *gorm.DB) *AttendeeRollupStatusRepository {
	return &AttendeeRollupStatusRepository{db: db}
}

func (r *AttendeeRollupStatusRepository) Upsert(
	ctx context.Context,
	s *attendance.AttendeeRollupStatus,
) error {
	model := toRollupStatusModel(s)
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "attendee_id"},
				{Name: "external_type"},
				{Name: "external_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"derived_status",
				"sessions_total",
				"sessions_attended",
				"sessions_confirmed",
				"total_duration_seconds",
				"last_derived_at",
			}),
		}).
		Create(model).Error
	if err != nil {
		return translateError(err, "upsert rollup status", attendance.ErrStatusNotFound)
	}
	return nil
}

func (r *AttendeeRollupStatusRepository) FindByAttendeeExternal(
	ctx context.Context,
	attendeeID string,
	ref attendance.ExternalRef,
) (*attendance.AttendeeRollupStatus, error) {
	var m AttendeeRollupStatusModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ? AND external_type = ? AND external_id = ?",
			attendeeID, ref.Type, ref.ID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrStatusNotFound
	}
	if err != nil {
		return nil, translateError(err, "find rollup status", attendance.ErrStatusNotFound)
	}
	return toRollupStatusDomain(&m), nil
}

func (r *AttendeeRollupStatusRepository) ListByExternal(
	ctx context.Context,
	ref attendance.ExternalRef,
) ([]*attendance.AttendeeRollupStatus, error) {
	var models []AttendeeRollupStatusModel
	err := r.db.WithContext(ctx).
		Where("external_type = ? AND external_id = ?", ref.Type, ref.ID).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list rollup statuses", attendance.ErrStatusNotFound)
	}
	out := make([]*attendance.AttendeeRollupStatus, 0, len(models))
	for i := range models {
		out = append(out, toRollupStatusDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion
var _ attendance.AttendeeRollupStatusRepository = (*AttendeeRollupStatusRepository)(nil)