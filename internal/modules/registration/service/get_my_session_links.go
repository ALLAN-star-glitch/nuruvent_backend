// internal/modules/registration/service/get_my_session_links.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// SessionLinkGroup ties a set of session links to the event the user
// is registered for.
type SessionLinkGroup struct {
	RegistrationID string
	EventID        string
	EventName      string
	EventSlug      string
	EventDate      time.Time
	Links          []registrationdomain.JoinLink
}

// SessionLinksResult is what GetMySessionLinks returns.
type SessionLinksResult struct {
	Groups []SessionLinkGroup
}

// GetMySessionLinks returns one group per confirmed registration,
// each containing a fresh personalized join link per session under
// the event.
//
// The links are newly issued on every call. Tokens issued by prior
// calls remain valid until their own expiry.
//
// Only confirmed registrations are considered. Pending and cancelled
// registrations are excluded — there's nothing to join yet, or
// nothing to join any more.
func (s *service) GetMySessionLinks(
	ctx context.Context,
	userID string,
) (*SessionLinksResult, error) {
	if userID == "" {
		return nil, errors.New("user id is required")
	}

	regs, _, err := s.deps.EventRegistrations.ListByUser(
		ctx,
		userID,
		registrationdomain.ListFilter{
			Statuses: []registrationdomain.Status{registrationdomain.StatusConfirmed},
			Page:     1,
			PageSize: 100,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list user registrations: %w", err)
	}

	groups := make([]SessionLinkGroup, 0, len(regs))

	for _, reg := range regs {
		if reg == nil || reg.Registration == nil {
			continue
		}

		// Resolve the event for the display name and slug.
		registrable, err := s.deps.Registrables.Resolve(ctx, "event", reg.EventID)
		if err != nil {
			log.Printf("[session-links] resolve event %s: %v", reg.EventID, err)
			continue
		}

		// Find or create the attendee for this registration.
		// RegisterAttendee is idempotent — a repeat call for the
		// same (external_type, external_id) returns the existing row.
		displayName, email, username := s.attendeeIdentity(ctx, reg)
		if email == "" {
			log.Printf("[session-links] no email for registration %s", reg.Registration.ID)
			continue
		}

		attendeeID, err := s.deps.Attendance.RegisterAttendee(
			ctx,
			registrationdomain.RegisterAttendeeForAttendanceCommand{
				ExternalType: "event_registration",
				ExternalID:   reg.Registration.ID,
				DisplayName:  displayName,
				Email:        email,
				Username:     username,
			},
		)
		if err != nil {
			log.Printf("[session-links] register attendee for %s: %v", reg.Registration.ID, err)
			continue
		}

		// Issue fresh join tokens for every session under the event.
		joinLinks, err := s.deps.Attendance.IssueJoinTokens(
			ctx,
			registrationdomain.IssueJoinTokensCommand{
				AttendeeID: attendeeID,
				EventID:    reg.EventID,
			},
		)
		if err != nil {
			log.Printf("[session-links] issue tokens for %s: %v", reg.Registration.ID, err)
			continue
		}

		groups = append(groups, SessionLinkGroup{
			RegistrationID: reg.Registration.ID,
			EventID:        reg.EventID,
			EventName:      registrable.DisplayName(),
			EventSlug:      registrable.Slug(),
			EventDate:      registrable.StartDate(),
			Links:          joinLinks,
		})
	}

	return &SessionLinksResult{Groups: groups}, nil
}