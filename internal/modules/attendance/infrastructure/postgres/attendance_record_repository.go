package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type AttendanceRecordRepository struct {
	db *gorm.DB
}

func NewAttendanceRecordRepository(db *gorm.DB) *AttendanceRecordRepository {
	return &AttendanceRecordRepository{db: db}
}

func (r *AttendanceRecordRepository) Create(ctx context.Context, rec *attendance.AttendanceRecord) error {
	if err := r.db.WithContext(ctx).Create(toAttendanceRecordModel(rec)).Error; err != nil {
		return translateError(err, "create attendance record", attendance.ErrInvalidAttendance)
	}
	return nil
}

func (r *AttendanceRecordRepository) Update(ctx context.Context, rec *attendance.AttendanceRecord) error {
	res := r.db.WithContext(ctx).
		Model(&AttendanceRecordModel{}).
		Where("id = ?", rec.ID).
		Updates(map[string]any{
			"leave_time":       rec.LeaveTime,
			"duration_seconds": rec.DurationSeconds,
			"updated_at":       rec.UpdatedAt,
		})
	if res.Error != nil {
		return translateError(res.Error, "update attendance record", attendance.ErrInvalidAttendance)
	}
	if res.RowsAffected == 0 {
		return attendance.ErrInvalidAttendance
	}
	return nil
}

func (r *AttendanceRecordRepository) FindByID(ctx context.Context, id string) (*attendance.AttendanceRecord, error) {
	var m AttendanceRecordModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrInvalidAttendance
	}
	if err != nil {
		return nil, translateError(err, "find attendance record", attendance.ErrInvalidAttendance)
	}
	return toAttendanceRecordDomain(&m), nil
}

func (r *AttendanceRecordRepository) FindByAttendeeSession(
	ctx context.Context,
	attendeeID, sessionID string,
) ([]*attendance.AttendanceRecord, error) {
	var models []AttendanceRecordModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ? AND session_id = ?", attendeeID, sessionID).
		Order("join_time ASC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "find attendance records", attendance.ErrInvalidAttendance)
	}
	out := make([]*attendance.AttendanceRecord, 0, len(models))
	for i := range models {
		out = append(out, toAttendanceRecordDomain(&models[i]))
	}
	return out, nil
}

func (r *AttendanceRecordRepository) FindOpenByAttendeeSession(
	ctx context.Context,
	attendeeID, sessionID string,
) (*attendance.AttendanceRecord, error) {
	var m AttendanceRecordModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ? AND session_id = ? AND leave_time IS NULL", attendeeID, sessionID).
		Order("join_time DESC").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrInvalidAttendance
	}
	if err != nil {
		return nil, translateError(err, "find open record", attendance.ErrInvalidAttendance)
	}
	return toAttendanceRecordDomain(&m), nil
}

func (r *AttendanceRecordRepository) ListOpenBySession(
	ctx context.Context,
	sessionID string,
) ([]*attendance.AttendanceRecord, error) {
	var models []AttendanceRecordModel
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND leave_time IS NULL", sessionID).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list open records", attendance.ErrInvalidAttendance)
	}
	out := make([]*attendance.AttendanceRecord, 0, len(models))
	for i := range models {
		out = append(out, toAttendanceRecordDomain(&models[i]))
	}
	return out, nil
}

func (r *AttendanceRecordRepository) FindBySession(
	ctx context.Context,
	sessionID string,
) ([]*attendance.AttendanceRecord, error) {
	var models []AttendanceRecordModel
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("join_time ASC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "find records by session", attendance.ErrInvalidAttendance)
	}
	out := make([]*attendance.AttendanceRecord, 0, len(models))
	for i := range models {
		out = append(out, toAttendanceRecordDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion
var _ attendance.AttendanceRecordRepository = (*AttendanceRecordRepository)(nil)