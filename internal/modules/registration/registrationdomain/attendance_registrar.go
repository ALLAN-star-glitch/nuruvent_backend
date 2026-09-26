// internal/modules/registration/registrationdomain/attendance_registrar.go

package registrationdomain

import "context"

// AttendanceRegistrar is the registration module's outbound port for
// syncing attendees into the attendance module.
type AttendanceRegistrar interface {
	// RegisterAttendee registers the attendee with the attendance
	// module. Idempotent on ExternalRef. Returns the attendance
	// module's own attendee ID.
	RegisterAttendee(ctx context.Context, cmd RegisterAttendeeForAttendanceCommand) (string, error)

	// RegisterAttendeeForEvent registers the attendee for every
	// session currently under the event.
	RegisterAttendeeForEvent(ctx context.Context, cmd RegisterAttendeeForEventCommand) error

	// IssueJoinTokens issues one join token per session under the
	// event and returns the (session title, join URL) pairs.
	IssueJoinTokens(ctx context.Context, cmd IssueJoinTokensCommand) ([]JoinLink, error)
}

type RegisterAttendeeForAttendanceCommand struct {
	ExternalType string
	ExternalID   string
	DisplayName  string
	Email        string
}

type RegisterAttendeeForEventCommand struct {
	AttendeeID string
	EventID    string
}

type IssueJoinTokensCommand struct {
	AttendeeID string
	EventID    string
}

// JoinLink is a (session title, join URL) pair.
type JoinLink struct {
	SessionTitle string
	URL          string
}