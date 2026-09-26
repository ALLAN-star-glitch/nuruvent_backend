package attendancedomain

import (
	"fmt"
	"strings"
	"time"
)

// Attendee is a person whose attendance is tracked, tied to exactly
// one source entity in an external module.
type Attendee struct {
	ID          string
	External    ExternalRef
	DisplayName string
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewAttendee constructs a new attendee with validation.
func NewAttendee(
	id string,
	external ExternalRef,
	displayName, email string,
	now time.Time,
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
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: valid email is required", ErrInvalidAttendee)
	}
	return &Attendee{
		ID:          id,
		External:    external,
		DisplayName: displayName,
		Email:       email,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// HydrateAttendee reconstructs an attendee from persistence without
// re-validating.
func HydrateAttendee(
	id string,
	external ExternalRef,
	displayName, email string,
	createdAt, updatedAt time.Time,
) *Attendee {
	return &Attendee{
		ID:          id,
		External:    external,
		DisplayName: displayName,
		Email:       email,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// UpdateProfile updates mutable fields and stamps UpdatedAt.
func (a *Attendee) UpdateProfile(displayName, email string, now time.Time) error {
	displayName = strings.TrimSpace(displayName)
	email = strings.TrimSpace(strings.ToLower(email))
	if displayName == "" {
		return fmt.Errorf("%w: display name is required", ErrInvalidAttendee)
	}
	if email == "" || !strings.Contains(email, "@") {
		return fmt.Errorf("%w: valid email is required", ErrInvalidAttendee)
	}
	a.DisplayName = displayName
	a.Email = email
	a.UpdatedAt = now
	return nil
}

// MatchesExternal reports whether the attendee is linked to the given
// external reference.
func (a *Attendee) MatchesExternal(ref ExternalRef) bool {
	return a.External.Equals(ref)
}