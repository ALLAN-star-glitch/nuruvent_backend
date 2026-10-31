// internal/modules/registration/service/list.go

package service

import (
	"context"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// ============================================================
// GET BY ID
// ============================================================

// GetByID fetches a single registration by ID.
//
// Authorization:
//   - If actorID is set (authenticated user), the caller must be the
//     registration owner (reg.UserID == actorID).
//   - If actorID is empty (guest), the caller must supply the guestEmail
//     that matches the email stored on the registration (case-insensitive).
func (s *service) GetByID(
	ctx context.Context,
	id, actorID, guestEmail string,
) (*registrationdomain.EventRegistration, error) {
	er, err := s.deps.EventRegistrations.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if er == nil || er.Registration == nil {
		return nil, registrationdomain.ErrRegistrationNotFound
	}

	reg := er.Registration

	// ---- Authenticated user path ----
	if actorID != "" {
		// Owner check
		if reg.UserID == actorID {
			return er, nil
		}

		// TODO: Admin/team-level access check (via PermissionChecker port)
		// when dashboard needs cross-user reads on this endpoint.

		return nil, registrationdomain.ErrRegistrationNotFound
	}

	// ---- Guest path ----
	// Guests have no actorID; authorization comes from matching the
	// supplied email against the one stored on the registration.
	if guestEmail != "" {
		if reg.GuestEmail != "" && strings.EqualFold(reg.GuestEmail, guestEmail) {
			return er, nil
		}
	}

	return nil, registrationdomain.ErrRegistrationNotFound
}

// ============================================================
// LIST BY EVENT
// ============================================================

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

// ============================================================
// LIST BY USER
// ============================================================

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

// ============================================================
// HELPERS
// ============================================================

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