package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

type WaitlistRepository struct {
	db *gorm.DB
}

func NewWaitlistRepository(db *gorm.DB) registrationdomain.WaitlistRepository {
	return &WaitlistRepository{db: db}
}

func (r *WaitlistRepository) Create(ctx context.Context, w *registrationdomain.WaitlistEntry) error {
	model := &WaitlistModel{
		ID:           w.ID,
		EventID:      w.EventID,
		UserID:       nullableString(w.UserID),
		GuestEmail:   nullableString(w.GuestEmail),
		GuestName:    nullableString(w.GuestName),
		TicketTypeID: nullableString(w.TicketTypeID),
		Position:     w.Position,
		Status:       "waiting",
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create waitlist entry: %w", err)
	}
	return nil
}

func (r *WaitlistRepository) NextPosition(ctx context.Context, eventID string) (int, error) {
	var maxPos *int
	err := r.db.WithContext(ctx).
		Model(&WaitlistModel{}).
		Where("event_id = ?", eventID).
		Select("MAX(position)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, fmt.Errorf("next waitlist position: %w", err)
	}
	if maxPos == nil {
		return 1, nil
	}
	return *maxPos + 1, nil
}

func (r *WaitlistRepository) PeekNext(ctx context.Context, eventID string) (*registrationdomain.WaitlistEntry, error) {
	var model WaitlistModel
	err := r.db.WithContext(ctx).
		Where("event_id = ? AND status = ?", eventID, "waiting").
		Order("position ASC").
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // no entry to promote
	}
	if err != nil {
		return nil, fmt.Errorf("peek next waitlist: %w", err)
	}
	return toWaitlistDomain(&model), nil
}

func (r *WaitlistRepository) ListByEvent(ctx context.Context, eventID string) ([]*registrationdomain.WaitlistEntry, error) {
	var models []WaitlistModel
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("position ASC").
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("list waitlist: %w", err)
	}

	out := make([]*registrationdomain.WaitlistEntry, 0, len(models))
	for i := range models {
		out = append(out, toWaitlistDomain(&models[i]))
	}
	return out, nil
}

func (r *WaitlistRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&WaitlistModel{})
	if res.Error != nil {
		return fmt.Errorf("delete waitlist: %w", res.Error)
	}
	return nil
}