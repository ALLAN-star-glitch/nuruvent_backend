package registrationdomain

import "context"

// ListFilter carries pagination and filtering options for list queries.
type ListFilter struct {
    Statuses []Status
    Page     int
    PageSize int
}

// RegistrationRepository persists the base Registration entity.
type RegistrationRepository interface {
    Create(ctx context.Context, r *Registration) error
    Update(ctx context.Context, r *Registration) error
    FindByID(ctx context.Context, id string) (*Registration, error)
    FindActiveByUserAndEvent(ctx context.Context, userID, eventID string) (*Registration, error)
    WithTx(ctx context.Context, fn func(RegistrationRepository) error) error
}

// EventRegistrationRepository persists event-specific registrations.
type EventRegistrationRepository interface {
    Create(ctx context.Context, e *EventRegistration) error
    Update(ctx context.Context, e *EventRegistration) error
    FindByID(ctx context.Context, id string) (*EventRegistration, error)
    ListByEvent(ctx context.Context, eventID string, f ListFilter) ([]*EventRegistration, int, error)
    ListByUser(ctx context.Context, userID string, f ListFilter) ([]*EventRegistration, int, error)
}

// WaitlistRepository persists waitlist entries.
type WaitlistRepository interface {
    Create(ctx context.Context, w *WaitlistEntry) error
    NextPosition(ctx context.Context, eventID string) (int, error)
    PeekNext(ctx context.Context, eventID string) (*WaitlistEntry, error)
    ListByEvent(ctx context.Context, eventID string) ([]*WaitlistEntry, error)
    Delete(ctx context.Context, id string) error
}