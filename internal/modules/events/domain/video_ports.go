// internal/modules/events/domain/video_ports.go

package domain

import (
	"context"
	"time"
)

// VideoMeetingCreator is the port the events service uses to create,
// update, and delete meetings on a host's connected video platform.
//
// Implemented in internal/app/adapters/events/video.go by wrapping the
// video module's Service. The events module never imports the video
// module directly — this interface is the boundary.
type VideoMeetingCreator interface {
	// IsConnected reports whether the user has an active connection
	// for the given platform. Platforms use the same slugs as the
	// video module: "zoom", "google_meet".
	IsConnected(ctx context.Context, userID string, platform string) (bool, error)

	// CreateMeeting creates a meeting on the user's connected account.
	// Returns the platform meeting ID and the join URL.
	CreateMeeting(ctx context.Context, req VideoMeetingRequest) (*VideoMeetingResult, error)

	// UpdateMeeting updates an existing meeting's topic, start time,
	// duration, and timezone. The join URL is unchanged, so attendees
	// keep the same link.
	//
	// The external ID passed here must be the platform-side meeting ID
	// (video_meetings.external_id), not Nuruvent's internal UUID.
	UpdateMeeting(ctx context.Context, req UpdateVideoMeetingRequest) (*VideoMeetingResult, error)

	// DeleteMeeting removes a meeting from the platform. Best-effort
	// from the caller's point of view.
	DeleteMeeting(ctx context.Context, req VideoMeetingDeleteRequest) error
}

// VideoMeetingRequest is the input to CreateMeeting.
type VideoMeetingRequest struct {
	UserID    string // the host whose account creates the meeting
	Platform  string // "zoom", "google_meet"
	Topic     string
	StartTime time.Time
	Duration  time.Duration
	Timezone  string
	Agenda    string
}

// VideoMeetingResult is the output of CreateMeeting and UpdateMeeting.
type VideoMeetingResult struct {
	MeetingID string // Nuruvent-side meeting ID (video_meetings.id)
	ExternalID string // platform-side meeting ID (Zoom numeric ID, etc.)
	JoinURL   string
	StartURL  string
}

// UpdateVideoMeetingRequest is the input to UpdateMeeting.
//
// ExternalID is the platform-side meeting ID — the same value that
// was stored on schedule.video_meeting_id when the meeting was created.
type UpdateVideoMeetingRequest struct {
	UserID     string
	Platform   string
	ExternalID string
	Topic      string
	StartTime  time.Time
	Duration   time.Duration
	Timezone   string
	Agenda     string
}

// VideoMeetingDeleteRequest is the input to DeleteMeeting.
//
// ExternalID is the platform-side meeting ID.
type VideoMeetingDeleteRequest struct {
	UserID     string
	Platform   string
	ExternalID string
}