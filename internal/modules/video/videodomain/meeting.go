// internal/modules/video/videodomain/meeting.go

package videodomain

import (
	"fmt"
	"strings"
	"time"
)

// MeetingSpec is the input the caller provides when creating a
// meeting. It describes what the meeting should be, not how the
// platform represents it.
type MeetingSpec struct {
	Topic      string
	StartTime  time.Time
	Duration   time.Duration
	Timezone   string // IANA, e.g. "Africa/Nairobi"
	Agenda     string
	HostUserID string // the Nuruvent user requesting the meeting
}

// Validate checks the spec before it's sent to a platform.
func (s MeetingSpec) Validate() error {
	if strings.TrimSpace(s.Topic) == "" {
		return fmt.Errorf("%w: topic is required", ErrInvalidMeeting)
	}
	if s.StartTime.IsZero() {
		return fmt.Errorf("%w: start time is required", ErrInvalidMeeting)
	}
	if s.Duration <= 0 {
		return fmt.Errorf("%w: duration must be positive", ErrInvalidMeeting)
	}
	if strings.TrimSpace(s.Timezone) == "" {
		return fmt.Errorf("%w: timezone is required", ErrInvalidMeeting)
	}
	if strings.TrimSpace(s.HostUserID) == "" {
		return fmt.Errorf("%w: host user id is required", ErrInvalidMeeting)
	}
	return nil
}

// Meeting is a scheduled virtual session created on a video platform
// on behalf of a host.
//
// For external platforms (Zoom, Google Meet), ExternalID is the
// platform's meeting ID and JoinURL points at the platform. For
// embeddable providers (LiveKit and future SDKs), ExternalID is the
// room identifier and JoinURL points at a Nuruvent-hosted page.
type Meeting struct {
	ID         string // Nuruvent's own ID
	UserID     string // the Nuruvent user who requested it
	Platform   Platform

	// ExternalID is the platform's identifier for this meeting.
	// Zoom: numeric meeting ID. Google Meet: meeting code.
	// LiveKit: room name. Used to match webhooks.
	ExternalID string

	// JoinURL is what attendees open. For external platforms it
	// points at the platform; for embeddable providers it points at
	// a Nuruvent page.
	JoinURL string

	// StartURL is host-only. Used by the host to start the meeting
	// from the platform. Empty for embeddable providers.
	StartURL string

	// Password is the meeting passcode if the platform provides one.
	Password string

	Topic     string
	StartTime time.Time
	Duration  time.Duration
	Timezone  string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewMeeting constructs a meeting with validation.
func NewMeeting(
	id, userID string,
	platform Platform,
	externalID, joinURL, startURL, password string,
	spec MeetingSpec,
	now time.Time,
) (*Meeting, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidMeeting)
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidMeeting)
	}
	if !platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform %q", ErrInvalidMeeting, platform)
	}
	if strings.TrimSpace(externalID) == "" {
		return nil, fmt.Errorf("%w: external id is required", ErrInvalidMeeting)
	}
	if strings.TrimSpace(joinURL) == "" {
		return nil, fmt.Errorf("%w: join url is required", ErrInvalidMeeting)
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	return &Meeting{
		ID:         id,
		UserID:     userID,
		Platform:   platform,
		ExternalID: externalID,
		JoinURL:    joinURL,
		StartURL:   startURL,
		Password:   password,
		Topic:      spec.Topic,
		StartTime:  spec.StartTime,
		Duration:   spec.Duration,
		Timezone:   spec.Timezone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// HydrateMeeting reconstructs a meeting from persistence without
// re-validating.
func HydrateMeeting(
	id, userID string,
	platform Platform,
	externalID, joinURL, startURL, password string,
	topic string,
	startTime time.Time,
	duration time.Duration,
	timezone string,
	createdAt, updatedAt time.Time,
) *Meeting {
	return &Meeting{
		ID:         id,
		UserID:     userID,
		Platform:   platform,
		ExternalID: externalID,
		JoinURL:    joinURL,
		StartURL:   startURL,
		Password:   password,
		Topic:      topic,
		StartTime:  startTime,
		Duration:   duration,
		Timezone:   timezone,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

// ExternalUser is the platform's view of a connected host, returned
// during OAuth completion.
type ExternalUser struct {
	ID       string
	Email    string
	OrgID    string
	Scopes   string
}