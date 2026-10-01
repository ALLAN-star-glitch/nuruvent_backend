// internal/modules/attendance/service/video_meeting_id_resolver.go

package service

import (
	"context"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// VideoMeetingIDResolver resolves a video meeting's internal ID from
// a provider meeting code.
//
// The attendance module never reads the video module's tables
// directly. It declares this port here, and the adapter in
// internal/app/adapters implements it by calling the video module.
//
// Implementations must return ("", nil) — not an error — when no
// meeting matches, because many sessions legitimately have no
// linked video meeting (in-person, manual-only, etc.).
type VideoMeetingIDResolver interface {
	ResolveMeetingID(
		ctx context.Context,
		provider attendance.SessionProvider,
		providerMeetingID string,
	) (string, error)
}