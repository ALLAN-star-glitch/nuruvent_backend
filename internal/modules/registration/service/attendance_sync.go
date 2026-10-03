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

	displayName, email, username := s.attendeeIdentity(ctx, reg)
	if email == "" {
		log.Printf("[registration] skipping attendance sync: no email for registration %s", regID)
		return
	}

	// Resolve the attendee's phone. Guests have no user record, so
	// the value stays empty. Authenticated registrations pull the
	// number from the live user record — this keeps it in sync with
	// any update the user makes to their profile after registering.
	phone := s.attendeePhone(ctx, reg)


	log.Printf("[reg-sync] %s user=%q phone=%q email=%q",
    regID, reg.Registration.UserID, phone, email)

	attendeeID, err := s.deps.Attendance.RegisterAttendee(ctx, registrationdomain.RegisterAttendeeForAttendanceCommand{
		ExternalType: "event_registration",
		ExternalID:   regID,
		DisplayName:  displayName,
		Email:        email,
		Phone:        phone,
		Username:     username,
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
) (displayName, email, username string) {
	if reg == nil || reg.Registration == nil {
		return "", "", ""
	}
	r := reg.Registration

	// Guest path.
	if r.GuestEmail != "" {
		name := r.GuestName
		if name == "" {
			name = r.GuestEmail
		}
		return name, r.GuestEmail, ""
	}

	// Authenticated path.
	if r.UserID == "" {
		return "", "", ""
	}
	if s.deps.Users == nil {
		return "", "", ""
	}
	name, mail, err := s.deps.Users.GetUserInfo(ctx, r.UserID)
	if err != nil {
		log.Printf("[registration] user lookup failed for %s: %v", r.UserID, err)
		return "", "", ""
	}
	if name == "" {
		name = mail
	}
	// Username lookup — see step 4 below for the port change.
	uname := ""
	if u, err := s.deps.Users.GetUsername(ctx, r.UserID); err == nil {
		uname = u
	} else {
		log.Printf("[registration] username lookup failed for %s: %v", r.UserID, err)
	}
	return name, mail, uname
}

// attendeePhone resolves the phone number to attach to the attendance
// record. Guests have none. Authenticated registrations read it from
// the live user record so post-registration profile updates propagate
// on the next sync.
//
// Missing or failed lookups return "" rather than erroring — the
// caller's email-based guard is the only hard requirement, and a
// missing phone shouldn't block attendance tracking.
func (s *service) attendeePhone(
	ctx context.Context,
	reg *registrationdomain.EventRegistration,
) string {
	if reg == nil || reg.Registration == nil {
		return ""
	}
	r := reg.Registration

	// Guest path — no user record to look up.
	if r.GuestEmail != "" {
		return ""
	}

	// Authenticated path.
	if r.UserID == "" {
		return ""
	}
	if s.deps.Users == nil {
		return ""
	}

	phone, err := s.deps.Users.GetPhone(ctx, r.UserID)
	if err != nil {
		log.Printf("[registration] phone lookup failed for %s: %v", r.UserID, err)
		return ""
	}
	return phone
}