package registrationdomain

import (
	"fmt"
	"time"
)

// Registration is the base entity: identity + lifecycle + pricing. Concrete
// registration types (event, course) extend it.
type Registration struct {
	ID                 string
	RegistrationNumber string
	UserID             string // empty when guest
	GuestEmail         string // empty when user
	GuestName          string
	GuestPhone         string
	Status             Status
	Currency           string
	TotalAmount        int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ConfirmedAt        *time.Time
	CancelledAt        *time.Time
	CancelledBy        string
	CancellationReason string
}

// NewRegistration creates a new registration with validation. Status is
// derived from requiresPayment: pending for paid, confirmed for free.
func NewRegistration(
	id string,
	registrationNumber string,
	userID string,
	guestEmail string,
	guestName string,
	guestPhone string,
	pricing PricingSnapshot,
	requiresPayment bool,
	now time.Time,
) (*Registration, error) {
	if userID == "" && guestEmail == "" {
		return nil, ErrIdentityRequired
	}
	if userID != "" && guestEmail != "" {
		return nil, fmt.Errorf("%w: both user and guest identity set", ErrIdentityRequired)
	}
	if guestEmail != "" && guestName == "" {
		return nil, fmt.Errorf("%w: guest name is required with guest email", ErrIdentityRequired)
	}

	status := StatusConfirmed
	if requiresPayment {
		status = StatusPending
	}

	r := &Registration{
		ID:                 id,
		RegistrationNumber: registrationNumber,
		UserID:             userID,
		GuestEmail:         guestEmail,
		GuestName:          guestName,
		GuestPhone:         guestPhone,
		Status:             status,
		Currency:           pricing.Currency,
		TotalAmount:        pricing.Total,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if status == StatusConfirmed {
		r.ConfirmedAt = &now
	}

	return r, nil
}

// HydrateRegistration reconstructs a Registration from persistence. No
// validation — the DB is the source of truth.
func HydrateRegistration(
	id, registrationNumber string,
	userID, guestEmail, guestName, guestPhone string,
	status Status,
	currency string,
	totalAmount int64,
	createdAt, updatedAt time.Time,
	confirmedAt, cancelledAt *time.Time,
	cancelledBy, cancellationReason string,
) *Registration {
	return &Registration{
		ID:                 id,
		RegistrationNumber: registrationNumber,
		UserID:             userID,
		GuestEmail:         guestEmail,
		GuestName:          guestName,
		GuestPhone:         guestPhone,
		Status:             status,
		Currency:           currency,
		TotalAmount:        totalAmount,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		ConfirmedAt:        confirmedAt,
		CancelledAt:        cancelledAt,
		CancelledBy:        cancelledBy,
		CancellationReason: cancellationReason,
	}
}

// Confirm transitions to confirmed. Idempotent when already confirmed.
func (r *Registration) Confirm(now time.Time) error {
	if r.Status == StatusConfirmed {
		return nil
	}
	if !r.Status.CanTransitionTo(StatusConfirmed) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, r.Status, StatusConfirmed)
	}
	r.Status = StatusConfirmed
	r.ConfirmedAt = &now
	r.UpdatedAt = now
	return nil
}

// Cancel transitions to cancelled. Records the actor and reason.
func (r *Registration) Cancel(actorID, reason string, now time.Time) error {
	if r.Status == StatusCancelled {
		return nil
	}
	if !r.Status.CanTransitionTo(StatusCancelled) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, r.Status, StatusCancelled)
	}
	r.Status = StatusCancelled
	r.CancelledAt = &now
	r.CancelledBy = actorID
	r.CancellationReason = reason
	r.UpdatedAt = now
	return nil
}

// Expire transitions a pending registration to expired. Only valid from pending.
func (r *Registration) Expire(now time.Time) error {
	if r.Status != StatusPending {
		return fmt.Errorf("%w: only pending can expire", ErrInvalidStatusTransition)
	}
	r.Status = StatusExpired
	r.UpdatedAt = now
	return nil
}

// Refund transitions a confirmed registration to refunded.
func (r *Registration) Refund(now time.Time) error {
	if !r.Status.CanTransitionTo(StatusRefunded) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, r.Status, StatusRefunded)
	}
	r.Status = StatusRefunded
	r.UpdatedAt = now
	return nil
}

// IsActive reports whether the registration counts toward capacity.
func (r *Registration) IsActive() bool { return r.Status.IsActive() }

// IsGuest reports whether the registration is for a guest (no user account).
func (r *Registration) IsGuest() bool { return r.UserID == "" }