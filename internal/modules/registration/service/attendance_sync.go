package service

import (
	"context"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

// syncRegistrationToAttendance registers the attendee with the
// attendance module, enrolls them for the event's sessions, and
// issues one join token per session.
//
// Best-effort: any failure is logged and swallowed. The registration
// itself is authoritative; a reconciliation job can catch missed
// attendance records later.
func (s *service) syncRegistrationToAttendance(
	ctx context.Context,
	reg *registrationdomain.EventRegistration,
) {
	if s.deps.Attendance == nil {
		return
	}
	if reg == nil || reg.Registration == nil {
		return
	}

	regID := reg.Registration.ID

	displayName, email := s.attendeeIdentity(ctx, reg)
	if email == "" {
		log.Printf("[registration] skipping attendance sync: no email for registration %s", regID)
		return
	}

	attendeeID, err := s.deps.Attendance.RegisterAttendee(ctx, registrationdomain.RegisterAttendeeForAttendanceCommand{
		ExternalType: "event_registration",
		ExternalID:   regID,
		DisplayName:  displayName,
		Email:        email,
	})
	if err != nil {
		log.Printf("[registration] attendance.RegisterAttendee failed for %s: %v", regID, err)
		return
	}

	if err := s.deps.Attendance.RegisterAttendeeForEvent(ctx, registrationdomain.RegisterAttendeeForEventCommand{
		AttendeeID: attendeeID,
		EventID:    reg.EventID,
	}); err != nil {
		log.Printf("[registration] attendance.RegisterAttendeeForEvent failed for %s: %v", regID, err)
		return
	}

	links, err := s.deps.Attendance.IssueJoinTokens(ctx, registrationdomain.IssueJoinTokensCommand{
		AttendeeID: attendeeID,
		EventID:    reg.EventID,
	})
	if err != nil {
		log.Printf("[registration] attendance.IssueJoinTokens failed for %s: %v", regID, err)
		return
	}

	reg.JoinLinks = links
	log.Printf("[registration] attendance sync complete for %s: %d session(s)", regID, len(links))
}

// attendeeIdentity extracts the display name and email to use for the
// attendance record.
func (s *service) attendeeIdentity(
	ctx context.Context,
	reg *registrationdomain.EventRegistration,
) (displayName, email string) {
	if reg == nil || reg.Registration == nil {
		return "", ""
	}
	r := reg.Registration

	// Guest path.
	if r.GuestEmail != "" {
		name := r.GuestName
		if name == "" {
			name = r.GuestEmail
		}
		return name, r.GuestEmail
	}

	// Authenticated path.
	if r.UserID == "" {
		return "", ""
	}
	if s.deps.Users == nil {
		return "", ""
	}
	name, mail, err := s.deps.Users.GetUserInfo(ctx, r.UserID)
	if err != nil {
		log.Printf("[registration] user lookup failed for %s: %v", r.UserID, err)
		return "", ""
	}
	if name == "" {
		name = mail
	}
	return name, mail
}