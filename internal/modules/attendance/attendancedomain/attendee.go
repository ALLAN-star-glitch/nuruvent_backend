package attendancedomain

import (
	"fmt"
	"strings"
	"time"
)

// Attendee is a person whose attendance is tracked, tied to exactly
// one source entity in an external module.
//
// Email is required for registration-based attendees (event
// registrations, course enrolments) because those sources always
// collect it. It may be empty for attendees observed from a platform
// that does not expose email — e.g. Google Meet participants, whose
// identity is a Google user resource ID rather than an email
// address.
type Attendee struct {
	ID               string
	External         ExternalRef
	DisplayName      string
	Email            string
	Phone            string        // ← add
	Username         string
	GoogleMeetUserID string
	IsHost           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewAttendee constructs a new attendee with validation.
func NewAttendee(
	id string,
	external ExternalRef,
	displayName, email, phone, username string,
	isHost bool,
	now time.Time,
	googleMeetUserID string,
) (*Attendee, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidAttendee)
	}
	if !external.IsValid() {
		return nil, fmt.Errorf("%w: external reference is required", ErrInvalidAttendee)
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, fmt.Errorf("%w: display name is required", ErrInvalidAttendee)
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if external.EmailRequired() {
		if email == "" || !strings.Contains(email, "@") {
			return nil, fmt.Errorf(
				"%w: valid email is required for %s",
				ErrInvalidAttendee, external.Type,
			)
		}
	}
	return &Attendee{
		ID:               id,
		External:         external,
		DisplayName:      displayName,
		Email:            email,
		Phone:            phone,
		Username:         strings.TrimSpace(username),
		GoogleMeetUserID: googleMeetUserID,
		IsHost:           isHost,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// HydrateAttendee reconstructs an attendee from persistence without
// re-validating.
func HydrateAttendee(
	id string,
	external ExternalRef,
	displayName, email, phone, username string,
	isHost bool,
	googleMeetUserID string,
	createdAt, updatedAt time.Time,
) *Attendee {
	return &Attendee{
		ID:               id,
		External:         external,
		DisplayName:      displayName,
		Email:            email,
		Phone:            phone,
		Username:         username,
		GoogleMeetUserID: googleMeetUserID,
		IsHost:           isHost,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

// UpdateProfile updates mutable fields and stamps UpdatedAt.
//
// The email rule follows the same logic as NewAttendee: required for
// registration-based refs, optional for platform observations.
func (a *Attendee) UpdateProfile(displayName, email string, phone string, now time.Time) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return fmt.Errorf("%w: display name is required", ErrInvalidAttendee)
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if a.External.EmailRequired() {
		if email == "" || !strings.Contains(email, "@") {
			return fmt.Errorf(
				"%w: valid email is required for %s",
				ErrInvalidAttendee, a.External.Type,
			)
		}
	}
	a.DisplayName = displayName
	a.Email = email
    a.Phone = strings.TrimSpace(phone)
	a.UpdatedAt = now
	return nil
}

// MatchesExternal reports whether the attendee is linked to the given
// external reference.
func (a *Attendee) MatchesExternal(ref ExternalRef) bool {
	return a.External.Equals(ref)
}