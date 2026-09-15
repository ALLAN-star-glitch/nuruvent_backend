// internal/modules/registration/registrationdomain/waitlist.go

package registrationdomain

import (
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// WAITLIST STATUS
// ============================================================

// WaitlistStatus describes the lifecycle of a waitlist entry.
// Aliased to the shared types.WaitlistStatus so the domain and the DB
// reference table use the same value type.
type WaitlistStatus = types.WaitlistStatus

const (
	WaitlistStatusWaiting   = types.WaitlistStatusWaiting
	WaitlistStatusOffered   = types.WaitlistStatusOffered
	WaitlistStatusConverted = types.WaitlistStatusConverted
	WaitlistStatusExpired   = types.WaitlistStatusExpired
	WaitlistStatusCancelled = types.WaitlistStatusCancelled
)

// ============================================================
// ENTITY
// ============================================================

// WaitlistEntry is a queued request to register when capacity opens.
type WaitlistEntry struct {
	ID               string
	EventID          string
	UserID           string // empty when guest
	GuestEmail       string // empty when user
	GuestName        string
	TicketTypeID     string // desired ticket type, may be empty
	Position         int
	Status           WaitlistStatus
	CreatedAt        time.Time
	NotifiedAt       *time.Time
	OfferedAt        *time.Time
	OfferedExpiresAt *time.Time
	PromotedAt       *time.Time // set when the entry becomes a registration
}

// NewWaitlistEntry creates a new entry in the waiting state.
func NewWaitlistEntry(
	id, eventID string,
	userID, guestEmail, guestName string,
	ticketTypeID string,
	position int,
	now time.Time,
) (*WaitlistEntry, error) {
	if userID == "" && guestEmail == "" {
		return nil, ErrIdentityRequired
	}
	if userID != "" && guestEmail != "" {
		return nil, fmt.Errorf("%w: both user and guest identity set", ErrIdentityRequired)
	}
	if guestEmail != "" && guestName == "" {
		return nil, fmt.Errorf("%w: guest name is required with guest email", ErrIdentityRequired)
	}
	if position < 1 {
		return nil, fmt.Errorf("waitlist position must be >= 1")
	}

	return &WaitlistEntry{
		ID:           id,
		EventID:      eventID,
		UserID:       userID,
		GuestEmail:   guestEmail,
		GuestName:    guestName,
		TicketTypeID: ticketTypeID,
		Position:     position,
		Status:       WaitlistStatusWaiting,
		CreatedAt:    now,
	}, nil
}

// HydrateWaitlistEntry reconstructs a WaitlistEntry from persistence.
// No validation — the DB is the source of truth.
func HydrateWaitlistEntry(
	id, eventID string,
	userID, guestEmail, guestName string,
	ticketTypeID string,
	position int,
	status WaitlistStatus,
	createdAt time.Time,
	notifiedAt, offeredAt, offeredExpiresAt, promotedAt *time.Time,
) *WaitlistEntry {
	return &WaitlistEntry{
		ID:               id,
		EventID:          eventID,
		UserID:           userID,
		GuestEmail:       guestEmail,
		GuestName:        guestName,
		TicketTypeID:     ticketTypeID,
		Position:         position,
		Status:           status,
		CreatedAt:        createdAt,
		NotifiedAt:       notifiedAt,
		OfferedAt:        offeredAt,
		OfferedExpiresAt: offeredExpiresAt,
		PromotedAt:       promotedAt,
	}
}

// ============================================================
// LIFECYCLE
// ============================================================

// MarkNotified records that the user was told they are on the waitlist.
func (w *WaitlistEntry) MarkNotified(now time.Time) {
	w.NotifiedAt = &now
}

// Offer records that a spot has opened and the user is being offered it.
func (w *WaitlistEntry) Offer(expiresAt, now time.Time) error {
	if w.Status != WaitlistStatusWaiting {
		return fmt.Errorf("%w: can only offer a waiting entry", ErrInvalidStatusTransition)
	}
	if !w.Status.CanTransitionTo(WaitlistStatusOffered) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, w.Status, WaitlistStatusOffered)
	}
	w.Status = WaitlistStatusOffered
	w.OfferedAt = &now
	w.OfferedExpiresAt = &expiresAt
	return nil
}

// MarkPromoted records that this entry has been converted into a registration.
func (w *WaitlistEntry) MarkPromoted(now time.Time) error {
	if !w.Status.CanTransitionTo(WaitlistStatusConverted) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, w.Status, WaitlistStatusConverted)
	}
	w.Status = WaitlistStatusConverted
	w.PromotedAt = &now
	return nil
}

// Expire marks an offered entry as expired (user didn't accept in time).
func (w *WaitlistEntry) Expire(now time.Time) error {
	if !w.Status.CanTransitionTo(WaitlistStatusExpired) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, w.Status, WaitlistStatusExpired)
	}
	w.Status = WaitlistStatusExpired
	return nil
}

// Cancel transitions to cancelled. Terminal.
func (w *WaitlistEntry) Cancel() error {
	if !w.Status.CanTransitionTo(WaitlistStatusCancelled) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, w.Status, WaitlistStatusCancelled)
	}
	w.Status = WaitlistStatusCancelled
	return nil
}

// ============================================================
// QUERIES
// ============================================================

// IsPromoted reports whether the entry has been converted into a registration.
func (w *WaitlistEntry) IsPromoted() bool {
	return w.Status == WaitlistStatusConverted
}

// IsWaiting reports whether the entry is still waiting for a spot.
func (w *WaitlistEntry) IsWaiting() bool {
	return w.Status == WaitlistStatusWaiting
}

// IsGuest reports whether the entry is for a guest (no user account).
func (w *WaitlistEntry) IsGuest() bool { return w.UserID == "" }