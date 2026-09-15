package service

import (
    "context"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

func (s *service) GetByID(ctx context.Context, id, actorID string) (*registrationdomain.EventRegistration, error) {
    reg, err := s.deps.EventRegistrations.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if reg.Registration.UserID != "" && reg.Registration.UserID != actorID {
        return nil, registrationdomain.ErrNotOwner
    }
    return reg, nil
}

func (s *service) ListByEvent(
    ctx context.Context,
    eventID, actorID string,
    f ListFilterInput,
) ([]*registrationdomain.EventRegistration, int, error) {
    // Authorization: actor must be the event organizer.
    // Delegated to events module via a port in a later step.
    filter := toDomainFilter(f)
    return s.deps.EventRegistrations.ListByEvent(ctx, eventID, filter)
}

func (s *service) ListByUser(
    ctx context.Context,
    userID string,
    f ListFilterInput,
) ([]*registrationdomain.EventRegistration, int, error) {
    if userID == "" {
        return nil, 0, registrationdomain.ErrNotOwner
    }
    filter := toDomainFilter(f)
    return s.deps.EventRegistrations.ListByUser(ctx, userID, filter)
}

func toDomainFilter(f ListFilterInput) registrationdomain.ListFilter {
    statuses := make([]registrationdomain.Status, 0, len(f.Statuses))
    for _, s := range f.Statuses {
        if parsed, err := registrationdomain.ParseStatus(s); err == nil {
            statuses = append(statuses, parsed)
        }
    }
    page := f.Page
    if page < 1 {
        page = 1
    }
    size := f.PageSize
    if size < 1 {
        size = 20
    }
    return registrationdomain.ListFilter{
        Statuses: statuses,
        Page:     page,
        PageSize: size,
    }
}