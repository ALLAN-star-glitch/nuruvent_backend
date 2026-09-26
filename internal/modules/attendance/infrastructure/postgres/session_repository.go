package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *attendance.Session) error {
	if err := r.db.WithContext(ctx).Create(toSessionModel(s)).Error; err != nil {
		return translateDuplicate(err, "create session", attendance.ErrDuplicateSession)
	}
	return nil
}

func (r *SessionRepository) Update(ctx context.Context, s *attendance.Session) error {
	res := r.db.WithContext(ctx).
		Model(&SessionModel{}).
		Where("id = ?", s.ID).
		Updates(map[string]any{
			"external_type":       s.External.Type,
			"external_id":         s.External.ID,
			"provider_session_id": s.ProviderSessionID,
			"title":               s.Title,
			"scheduled_start":     s.ScheduledStart,
			"scheduled_end":       s.ScheduledEnd,
			"duration_minutes":    s.DurationMinutes,
			"provider":            string(s.Provider),
			"provider_meeting_id": s.ProviderMeetingID,
			"provider_url":        s.ProviderURL,
			"status":              string(s.Status),
			"updated_at":          s.UpdatedAt,
		})
	if res.Error != nil {
		return translateError(res.Error, "update session", attendance.ErrSessionNotFound)
	}
	if res.RowsAffected == 0 {
		return attendance.ErrSessionNotFound
	}
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*attendance.Session, error) {
	var m SessionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrSessionNotFound
	}
	if err != nil {
		return nil, translateError(err, "find session", attendance.ErrSessionNotFound)
	}
	return toSessionDomain(&m), nil
}

// FindByExternalAndProviderSession returns the session belonging to
// the given parent (external ref) with the given provider session ID.
//
// This is the primary identity lookup. A parent entity may have many
// sessions; the provider_session_id selects one.
func (r *SessionRepository) FindByExternalAndProviderSession(
	ctx context.Context,
	ref attendance.ExternalRef,
	providerSessionID string,
) (*attendance.Session, error) {
	var m SessionModel
	err := r.db.WithContext(ctx).
		Where("external_type = ? AND external_id = ? AND provider_session_id = ?",
			ref.Type, ref.ID, providerSessionID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrSessionNotFound
	}
	if err != nil {
		return nil, translateError(err, "find session by external+provider", attendance.ErrSessionNotFound)
	}
	return toSessionDomain(&m), nil
}

// ListByExternalRef returns every session under a parent entity.
func (r *SessionRepository) ListByExternalRef(
	ctx context.Context,
	ref attendance.ExternalRef,
) ([]*attendance.Session, error) {
	var models []SessionModel
	err := r.db.WithContext(ctx).
		Where("external_type = ? AND external_id = ?", ref.Type, ref.ID).
		Order("scheduled_start ASC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list sessions by external", attendance.ErrSessionNotFound)
	}
	out := make([]*attendance.Session, 0, len(models))
	for i := range models {
		out = append(out, toSessionDomain(&models[i]))
	}
	return out, nil
}

func (r *SessionRepository) FindByProviderMeetingID(
	ctx context.Context,
	provider attendance.SessionProvider,
	meetingID string,
) (*attendance.Session, error) {
	var m SessionModel
	err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_meeting_id = ?", string(provider), meetingID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrSessionNotFound
	}
	if err != nil {
		return nil, translateError(err, "find session by meeting", attendance.ErrSessionNotFound)
	}
	return toSessionDomain(&m), nil
}

func (r *SessionRepository) FindActiveBySchedule(
	ctx context.Context,
	before time.Time,
) ([]*attendance.Session, error) {
	var models []SessionModel
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{
			string(attendance.SessionStatusScheduled),
			string(attendance.SessionStatusLive),
		}).
		Where("scheduled_start <= ?", before).
		Order("scheduled_start ASC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "find active sessions", attendance.ErrSessionNotFound)
	}
	out := make([]*attendance.Session, 0, len(models))
	for i := range models {
		out = append(out, toSessionDomain(&models[i]))
	}
	return out, nil
}

func (r *SessionRepository) FindEndedBefore(
	ctx context.Context,
	before time.Time,
) ([]*attendance.Session, error) {
	var models []SessionModel
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{
			string(attendance.SessionStatusScheduled),
			string(attendance.SessionStatusLive),
		}).
		Where("scheduled_end < ?", before).
		Order("scheduled_end ASC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "find ended sessions", attendance.ErrSessionNotFound)
	}
	out := make([]*attendance.Session, 0, len(models))
	for i := range models {
		out = append(out, toSessionDomain(&models[i]))
	}
	return out, nil
}

// compile-time assertion — unused for now, will be replaced with the
// service interface once we bind it.
var _ = fmt.Sprintf