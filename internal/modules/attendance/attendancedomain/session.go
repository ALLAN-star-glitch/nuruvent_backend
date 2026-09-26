package attendancedomain

import (
	"fmt"
	"strings"
	"time"
)

// Session is a scheduled occurrence during which attendance is
// tracked.
//
// A session belongs to a parent external entity (an event, a course
// cohort, etc.) — that's `External`. Within the parent, the session
// may have its own identifier chosen by the consumer — that's
// `ProviderSessionID`. For a multi-day event with four
// `event_schedules`, each schedule is one attendance session, and its
// `ProviderSessionID` is the schedule's own UUID.
//
// `ProviderSessionID` is opaque to the attendance module. It's stored
// for consumer lookups, not interpreted.
type Session struct {
	ID                string
	External          ExternalRef
	ProviderSessionID string
	Title             string
	ScheduledStart    time.Time
	ScheduledEnd      time.Time
	DurationMinutes   int
	Provider          SessionProvider
	ProviderMeetingID string
	ProviderURL       string
	Status            SessionStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewSession constructs a new session with validation.
func NewSession(
	id string,
	external ExternalRef,
	providerSessionID string,
	title string,
	start, end time.Time,
	provider SessionProvider,
	meetingID, providerURL string,
	now time.Time,
) (*Session, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidSession)
	}
	if !external.IsValid() {
		return nil, fmt.Errorf("%w: external reference is required", ErrInvalidSession)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidSession)
	}
	if !end.After(start) {
		return nil, fmt.Errorf("%w: end must be after start", ErrInvalidSession)
	}
	if !provider.IsValid() {
		return nil, fmt.Errorf("%w: invalid provider %q", ErrInvalidSession, provider)
	}
	if provider.RequiresMeetingID() && strings.TrimSpace(meetingID) == "" {
		return nil, fmt.Errorf("%w: meeting id required for %s", ErrInvalidSession, provider)
	}
	if provider == ProviderZoom && strings.TrimSpace(providerURL) == "" {
		return nil, fmt.Errorf("%w: provider url required for zoom", ErrInvalidSession)
	}
	return &Session{
		ID:                id,
		External:          external,
		ProviderSessionID: strings.TrimSpace(providerSessionID),
		Title:             title,
		ScheduledStart:    start,
		ScheduledEnd:      end,
		DurationMinutes:   int(end.Sub(start).Minutes()),
		Provider:          provider,
		ProviderMeetingID: strings.TrimSpace(meetingID),
		ProviderURL:       strings.TrimSpace(providerURL),
		Status:            SessionStatusScheduled,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// HydrateSession reconstructs a session from persistence.
func HydrateSession(
	id string,
	external ExternalRef,
	providerSessionID string,
	title string,
	scheduledStart, scheduledEnd time.Time,
	durationMinutes int,
	provider SessionProvider,
	providerMeetingID, providerURL string,
	status SessionStatus,
	createdAt, updatedAt time.Time,
) *Session {
	return &Session{
		ID:                id,
		External:          external,
		ProviderSessionID: providerSessionID,
		Title:             title,
		ScheduledStart:    scheduledStart,
		ScheduledEnd:      scheduledEnd,
		DurationMinutes:   durationMinutes,
		Provider:          provider,
		ProviderMeetingID: providerMeetingID,
		ProviderURL:       providerURL,
		Status:            status,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}

// MarkLive transitions the session to live. Idempotent.
func (s *Session) MarkLive(now time.Time) {
	if s.Status == SessionStatusScheduled {
		s.Status = SessionStatusLive
		s.UpdatedAt = now
	}
}

// MarkEnded transitions the session to ended. Idempotent.
func (s *Session) MarkEnded(now time.Time) {
	if s.Status == SessionStatusLive || s.Status == SessionStatusScheduled {
		s.Status = SessionStatusEnded
		s.UpdatedAt = now
	}
}

// Cancel transitions the session to cancelled. Idempotent.
func (s *Session) Cancel(now time.Time) {
	if !s.Status.IsFinal() {
		s.Status = SessionStatusCancelled
		s.UpdatedAt = now
	}
}

// TransitionTo determines and applies the lifecycle transition the
// session should be in given the current time.
func (s *Session) TransitionTo(now time.Time) {
	if s.Status == SessionStatusCancelled {
		return
	}
	switch {
	case now.Before(s.ScheduledStart):
		// no-op
	case now.Before(s.ScheduledEnd):
		s.MarkLive(now)
	default:
		s.MarkEnded(now)
	}
}

// MatchesProviderMeeting reports whether the session is served by the
// given provider with the given meeting ID.
func (s *Session) MatchesProviderMeeting(provider SessionProvider, meetingID string) bool {
	return s.Provider == provider && s.ProviderMeetingID == meetingID
}