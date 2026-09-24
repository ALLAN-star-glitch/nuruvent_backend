package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

type RegistrationRepository struct {
	db *gorm.DB
}


func NewRegistrationRepository(db *gorm.DB) registrationdomain.RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Create(ctx context.Context, reg *registrationdomain.Registration) error {
	statusID, err := r.resolveStatusID(ctx, reg.Status)
	if err != nil {
		return err
	}

	model := toRegistrationModel(reg, statusID)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create registration: %w", err)
	}
	return nil
}

func (r *RegistrationRepository) Update(ctx context.Context, reg *registrationdomain.Registration) error {
	statusID, err := r.resolveStatusID(ctx, reg.Status)
	if err != nil {
		return err
	}

	model := toRegistrationModel(reg, statusID)

	res := r.db.WithContext(ctx).
		Model(&RegistrationModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"user_id":             model.UserID,
			"guest_email":         model.GuestEmail,
			"guest_name":          model.GuestName,
			"guest_phone":         model.GuestPhone,
			"status_id":           model.StatusID,
			"currency":            model.Currency,
			"subtotal":            model.Subtotal,
			"discount_total":      model.DiscountTotal,
			"total_amount":        model.TotalAmount,
			"updated_at":          model.UpdatedAt,
			"confirmed_at":        model.ConfirmedAt,
			"cancelled_at":        model.CancelledAt,
			"cancelled_by":        model.CancelledBy,
			"cancellation_reason": model.CancellationReason,
		})

	if res.Error != nil {
		return fmt.Errorf("update registration: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return registrationdomain.ErrRegistrationNotFound
	}
	return nil
}

func (r *RegistrationRepository) FindByID(ctx context.Context, id string) (*registrationdomain.Registration, error) {
	var model RegistrationModel
	err := r.db.WithContext(ctx).
		Preload("Status").
		Where("id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, registrationdomain.ErrRegistrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find registration: %w", err)
	}
	return toRegistrationDomain(&model)
}

func (r *RegistrationRepository) FindActiveByUserAndEvent(ctx context.Context, userID, eventID string) (*registrationdomain.Registration, error) {
	var model RegistrationModel
	err := r.db.WithContext(ctx).
		Model(&RegistrationModel{}).
		Joins("JOIN event_registrations er ON er.registration_id = registrations.id").
		Joins("JOIN registration_statuses rs ON rs.id = registrations.status_id").
		Where("registrations.user_id = ?", userID).
		Where("er.event_id = ?", eventID).
		Where("rs.slug IN ?", []string{"pending", "confirmed"}).
		Where("registrations.deleted_at IS NULL").
		Preload("Status").
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, registrationdomain.ErrRegistrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find active registration: %w", err)
	}
	return toRegistrationDomain(&model)
}

func (r *RegistrationRepository) WithTx(
	ctx context.Context,
	fn func(
		registrationdomain.RegistrationRepository,
		registrationdomain.EventRegistrationRepository,
		registrationdomain.WaitlistRepository,
	) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(
			&RegistrationRepository{db: tx},
			&EventRegistrationRepository{db: tx},
			&WaitlistRepository{db: tx},
		)
	})
}

func (r *RegistrationRepository) resolveStatusID(
	ctx context.Context,
	status registrationdomain.Status,
) (string, error) {
	var row RegistrationStatusModel
	err := r.db.WithContext(ctx).
		Where("slug = ?", status.GetSlug()).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("registration status %q not found", status.GetSlug())
		}
		return "", fmt.Errorf("resolve status %q: %w", status.GetSlug(), err)
	}
	return row.ID, nil
}


