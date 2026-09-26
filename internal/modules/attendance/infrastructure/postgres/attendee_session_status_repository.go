package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type AttendeeSessionStatusRepository struct {
	db *gorm.DB
}

func NewAttendeeSessionStatusRepository(db *gorm.DB) *AttendeeSessionStatusRepository {
	return &AttendeeSessionStatusRepository{db: db}
}

// Upsert inserts or updates the status row for (attendee, session).
func (r *AttendeeSessionStatusRepository) Upsert(
	ctx context.Context,
	s *attendance.AttendeeSessionStatus,
) error {
	model := toSessionStatusModel(s)
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "attendee_id"},
				{Name: "session_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"derived_status",
				"total_duration_seconds",
				"host_confirmed",
				"confirmed_status",
				"confirmed_by",
				"confirmed_at",
				"confirm_reason",
				"last_derived_at",
			}),
		}).
		Create(model).Error
	if err != nil {
		return translateError(err, "upsert session status", attendance.ErrStatusNotFound)
	}
	return nil
}

func (r *AttendeeSessionStatusRepository) FindByAttendeeSession(
	ctx context.Context,
	attendeeID, sessionID string,
) (*attendance.AttendeeSessionStatus, error) {
	var m AttendeeSessionStatusModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ? AND session_id = ?", attendeeID, sessionID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrStatusNotFound
	}
	if err != nil {
		return nil, translateError(err, "find session status", attendance.ErrStatusNotFound)
	}
	return toSessionStatusDomain(&m), nil
}

func (r *AttendeeSessionStatusRepository) ListBySession(
	ctx context.Context,
	sessionID string,
) ([]*attendance.AttendeeSessionStatus, error) {
	var models []AttendeeSessionStatusModel
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list session statuses", attendance.ErrStatusNotFound)
	}
	out := make([]*attendance.AttendeeSessionStatus, 0, len(models))
	for i := range models {
		out = append(out, toSessionStatusDomain(&models[i]))
	}
	return out, nil
}

func (r *AttendeeSessionStatusRepository) ListByAttendee(
	ctx context.Context,
	attendeeID string,
) ([]*attendance.AttendeeSessionStatus, error) {
	var models []AttendeeSessionStatusModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ?", attendeeID).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list attendee session statuses", attendance.ErrStatusNotFound)
	}
	out := make([]*attendance.AttendeeSessionStatus, 0, len(models))
	for i := range models {
		out = append(out, toSessionStatusDomain(&models[i]))
	}
	return out, nil
}

// ListCertEligibleBySession returns all statuses for a session that
// qualify for a certificate: Full (derived) or Confirmed (host).
func (r *AttendeeSessionStatusRepository) ListCertEligibleBySession(
	ctx context.Context,
	sessionID string,
) ([]*attendance.AttendeeSessionStatus, error) {
	var models []AttendeeSessionStatusModel
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Where("(derived_status = ? OR host_confirmed = true)",
			string(attendance.StatusFull)).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list cert-eligible statuses", attendance.ErrStatusNotFound)
	}
	out := make([]*attendance.AttendeeSessionStatus, 0, len(models))
	for i := range models {
		out = append(out, toSessionStatusDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion
var _ attendance.AttendeeSessionStatusRepository = (*AttendeeSessionStatusRepository)(nil)