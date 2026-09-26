package postgres

import (
	"context"

	"gorm.io/gorm"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type AttendanceOverrideRepository struct {
	db *gorm.DB
}

func NewAttendanceOverrideRepository(db *gorm.DB) *AttendanceOverrideRepository {
	return &AttendanceOverrideRepository{db: db}
}

func (r *AttendanceOverrideRepository) Create(
	ctx context.Context,
	o *attendance.AttendanceOverride,
) error {
	if err := r.db.WithContext(ctx).Create(toOverrideModel(o)).Error; err != nil {
		return translateError(err, "create override", attendance.ErrInvalidAttendance)
	}
	return nil
}

func (r *AttendanceOverrideRepository) ListByAttendeeSession(
	ctx context.Context,
	attendeeID, sessionID string,
) ([]*attendance.AttendanceOverride, error) {
	var models []AttendanceOverrideModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ? AND session_id = ?", attendeeID, sessionID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list overrides", attendance.ErrInvalidAttendance)
	}
	out := make([]*attendance.AttendanceOverride, 0, len(models))
	for i := range models {
		out = append(out, toOverrideDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion
var _ attendance.AttendanceOverrideRepository = (*AttendanceOverrideRepository)(nil)