package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type AttendeeRepository struct {
	db *gorm.DB
}

func NewAttendeeRepository(db *gorm.DB) *AttendeeRepository {
	return &AttendeeRepository{db: db}
}

func (r *AttendeeRepository) Create(ctx context.Context, a *attendance.Attendee) error {
	if err := r.db.WithContext(ctx).Create(toAttendeeModel(a)).Error; err != nil {
		return translateDuplicate(err, "create attendee", attendance.ErrDuplicateAttendee)
	}
	return nil
}

func (r *AttendeeRepository) Update(ctx context.Context, a *attendance.Attendee) error {
    res := r.db.WithContext(ctx).
        Model(&AttendeeModel{}).
        Where("id = ?", a.ID).
        Updates(map[string]any{
            "display_name":        a.DisplayName,
            "email":               a.Email,
            "google_meet_user_id": a.GoogleMeetUserID,
            "updated_at":          time.Now().UTC(),
        })
    if res.Error != nil {
        return translateError(res.Error, "update attendee", attendance.ErrInvalidSession)
    }
    if res.RowsAffected == 0 {
        return attendance.ErrInvalidSession
    }
    return nil
}

func (r *AttendeeRepository) FindByID(ctx context.Context, id string) (*attendance.Attendee, error) {
	var m AttendeeModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrAttendeeNotFound
	}
	if err != nil {
		return nil, translateError(err, "find attendee", attendance.ErrAttendeeNotFound)
	}
	return toAttendeeDomain(&m), nil
}

func (r *AttendeeRepository) FindByExternalRef(ctx context.Context, ref attendance.ExternalRef) (*attendance.Attendee, error) {
	var m AttendeeModel
	err := r.db.WithContext(ctx).
		Where("external_type = ? AND external_id = ?", ref.Type, ref.ID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrAttendeeNotFound
	}
	if err != nil {
		return nil, translateError(err, "find attendee by external", attendance.ErrAttendeeNotFound)
	}
	return toAttendeeDomain(&m), nil
}

func (r *AttendeeRepository) FindByEmail(ctx context.Context, email string) ([]*attendance.Attendee, error) {
	var models []AttendeeModel
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "find attendees by email", attendance.ErrAttendeeNotFound)
	}
	out := make([]*attendance.Attendee, 0, len(models))
	for i := range models {
		out = append(out, toAttendeeDomain(&models[i]))
	}
	return out, nil
}



func (r *AttendeeRepository) FindByUsername(
	ctx context.Context,
	username string,
) ([]*attendance.Attendee, error) {
	var models []AttendeeModel
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "find attendees by username", attendance.ErrAttendeeNotFound)
	}
	out := make([]*attendance.Attendee, 0, len(models))
	for i := range models {
		out = append(out, toAttendeeDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion
var _ attendance.AttendeeRepository = (*AttendeeRepository)(nil)

// silence unused import when compile-time assertion passes
var _ = fmt.Sprintf