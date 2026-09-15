package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

type EventRegistrationRepository struct {
	db *gorm.DB
}

func NewEventRegistrationRepository(db *gorm.DB) registrationdomain.EventRegistrationRepository {
	return &EventRegistrationRepository{db: db}
}

func (r *EventRegistrationRepository) Create(ctx context.Context, er *registrationdomain.EventRegistration) error {
	model := &EventRegistrationModel{
		RegistrationID: er.Registration.ID,
		EventID:        er.EventID,
		UserID:         nullableString(er.Registration.UserID),
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create event registration: %w", err)
	}

	// Insert ticket selections
	if err := r.replaceTickets(ctx, er.Registration.ID, er.Selections); err != nil {
		return err
	}
	return nil
}

func (r *EventRegistrationRepository) Update(ctx context.Context, er *registrationdomain.EventRegistration) error {
	// Update base row
	res := r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Where("registration_id = ?", er.Registration.ID).
		Updates(map[string]any{
			"event_id": er.EventID,
			"user_id":  nullableString(er.Registration.UserID),
		})
	if res.Error != nil {
		return fmt.Errorf("update event registration: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return registrationdomain.ErrRegistrationNotFound
	}

	// Replace ticket selections (delete + reinsert)
	if err := r.replaceTickets(ctx, er.Registration.ID, er.Selections); err != nil {
		return err
	}
	return nil
}

func (r *EventRegistrationRepository) FindByID(ctx context.Context, id string) (*registrationdomain.EventRegistration, error) {
	var model EventRegistrationModel
	err := r.db.WithContext(ctx).
		Preload("Registration").
		Preload("Registration.Status").
		Where("registration_id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, registrationdomain.ErrRegistrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find event registration: %w", err)
	}

	tickets, err := r.loadTickets(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEventRegistrationDomain(&model, tickets)
}

func (r *EventRegistrationRepository) ListByEvent(ctx context.Context, eventID string, f registrationdomain.ListFilter) ([]*registrationdomain.EventRegistration, int, error) {
	query := r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Where("event_id = ?", eventID)

	// Optional status filter via join
	if len(f.Statuses) > 0 {
		slugs := statusesToSlugs(f.Statuses)
		query = query.
			Joins("JOIN registrations r ON r.id = event_registrations.registration_id").
			Joins("JOIN registration_statuses rs ON rs.id = r.status_id").
			Where("rs.slug IN ?", slugs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count event registrations: %w", err)
	}

	var models []EventRegistrationModel
	err := query.
		Preload("Registration").
		Preload("Registration.Status").
		Order("event_registrations.created_at DESC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Find(&models).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list event registrations: %w", err)
	}

	out := make([]*registrationdomain.EventRegistration, 0, len(models))
	for i := range models {
		tickets, err := r.loadTickets(ctx, models[i].RegistrationID)
		if err != nil {
			return nil, 0, err
		}
		er, err := toEventRegistrationDomain(&models[i], tickets)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, er)
	}
	return out, int(total), nil
}

func (r *EventRegistrationRepository) ListByUser(ctx context.Context, userID string, f registrationdomain.ListFilter) ([]*registrationdomain.EventRegistration, int, error) {
	query := r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Joins("JOIN registrations r ON r.id = event_registrations.registration_id").
		Where("r.user_id = ?", userID)

	if len(f.Statuses) > 0 {
		slugs := statusesToSlugs(f.Statuses)
		query = query.
			Joins("JOIN registration_statuses rs ON rs.id = r.status_id").
			Where("rs.slug IN ?", slugs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count user registrations: %w", err)
	}

	var models []EventRegistrationModel
	err := query.
		Preload("Registration").
		Preload("Registration.Status").
		Order("event_registrations.created_at DESC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Find(&models).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list user registrations: %w", err)
	}

	out := make([]*registrationdomain.EventRegistration, 0, len(models))
	for i := range models {
		tickets, err := r.loadTickets(ctx, models[i].RegistrationID)
		if err != nil {
			return nil, 0, err
		}
		er, err := toEventRegistrationDomain(&models[i], tickets)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, er)
	}
	return out, int(total), nil
}

// ============================================================
// HELPERS
// ============================================================

func (r *EventRegistrationRepository) replaceTickets(ctx context.Context, registrationID string, selections []registrationdomain.TicketSelection) error {
	// Delete existing
	if err := r.db.WithContext(ctx).
		Where("registration_id = ?", registrationID).
		Delete(&EventRegistrationTicketModel{}).Error; err != nil {
		return fmt.Errorf("delete tickets: %w", err)
	}

	// Insert new
	models := make([]EventRegistrationTicketModel, 0, len(selections))
	for _, s := range selections {
		models = append(models, EventRegistrationTicketModel{
			RegistrationID: registrationID,
			TicketTypeID:   s.TicketTypeID,
			Quantity:       s.Quantity,
			UnitPrice:      s.UnitPrice,
			Discount:       s.Discount,
		})
	}
	if len(models) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return fmt.Errorf("insert tickets: %w", err)
	}
	return nil
}

func (r *EventRegistrationRepository) loadTickets(ctx context.Context, registrationID string) ([]registrationdomain.TicketSelection, error) {
	var models []EventRegistrationTicketModel
	err := r.db.WithContext(ctx).
		Where("registration_id = ?", registrationID).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("load tickets: %w", err)
	}

	out := make([]registrationdomain.TicketSelection, 0, len(models))
	for _, m := range models {
		out = append(out, registrationdomain.TicketSelection{
			TicketTypeID: m.TicketTypeID,
			Quantity:     m.Quantity,
			UnitPrice:    m.UnitPrice,
			Discount:     m.Discount,
		})
	}
	return out, nil
}

func statusesToSlugs(statuses []registrationdomain.Status) []string {
	slugs := make([]string, 0, len(statuses))
	for _, s := range statuses {
		slugs = append(slugs, s.GetSlug())
	}
	return slugs
}