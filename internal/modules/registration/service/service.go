package service

import (
	"context"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/id"
)

type Service interface {
	RegisterForEvent(ctx context.Context, cmd RegisterCommand) (*registrationdomain.EventRegistration, error)
	ConfirmRegistration(ctx context.Context, registrationID string) error
	CancelRegistration(ctx context.Context, cmd CancelCommand) error
	ExpirePending(ctx context.Context, registrationID string) error

	JoinWaitlist(ctx context.Context, cmd JoinWaitlistCommand) (*registrationdomain.WaitlistEntry, error)
	PromoteFromWaitlist(ctx context.Context, eventID string) (*registrationdomain.EventRegistration, error)

	GetByID(ctx context.Context, id, actorID, guestEmail string) (*registrationdomain.EventRegistration, error)
	ListByEvent(ctx context.Context, eventID, actorID string, f ListFilterInput) ([]*registrationdomain.EventRegistration, int, error)
	ListByUser(ctx context.Context, userID string, f ListFilterInput) ([]*registrationdomain.EventRegistration, int, error)
}

// Dependencies groups the ports the service needs.
type Dependencies struct {
	Registrations      registrationdomain.RegistrationRepository
	EventRegistrations registrationdomain.EventRegistrationRepository
	Waitlist           registrationdomain.WaitlistRepository
	Registrables       registrationdomain.RegistrableResolver
	Notifier           registrationdomain.Notifier
	IDGenerator        id.Generator  // ← from shared
	Clock              Clock         // ← stays local (service-specific)
	NumberGenerator    registrationdomain.RegistrationNumberGenerator
	Users      registrationdomain.UserInfoProvider   // 👈 new
	Attendance registrationdomain.AttendanceRegistrar // 👈 new
}

// Clock abstracts time.Now so tests can control time.
type Clock interface {
	Now() time.Time
}

type service struct {
	deps Dependencies
}

func New(deps Dependencies) Service {
	return &service{deps: deps}
}