package service

import (
	"context"
	"log"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// BACKGROUND MEET ATTENDANCE POLLER
// ============================================================
//
// Google Meet does not push participant events. Attendance has to be
// pulled. This poller runs every AutoFetchInterval and fetches
// attendance for every Meet meeting whose scheduled window is
// currently active.
//
// Zoom is not polled — its participants arrive via webhook.
//
// The poller is best-effort. A failure on one meeting logs and
// continues; the next tick will retry.

const (
	// AutoFetchInterval is how often the poller wakes up.
	AutoFetchInterval = 2 * time.Minute

	// AutoFetchGrace is how far outside a meeting's scheduled window
	// the poller still polls. Catches meetings that start early or
	// run over, and Google's post-meeting consistency window.
	AutoFetchGrace = 15 * time.Minute
)

// AutoFetchGoogleMeetAttendance runs the Meet attendance poller.
//
// Blocks until ctx is cancelled.
func (s *videoService) AutoFetchGoogleMeetAttendance(ctx context.Context) error {
	log.Printf(
		"[video] auto-fetch meet: starting, interval=%s grace=%s",
		AutoFetchInterval, AutoFetchGrace,
	)

	// Run once immediately so a live meeting doesn't wait a full
	// interval for its first fetch.
	s.pollActiveMeetMeetings(ctx)

	ticker := time.NewTicker(AutoFetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[video] auto-fetch meet: context cancelled, stopping")
			return nil
		case <-ticker.C:
			s.pollActiveMeetMeetings(ctx)
		}
	}
}

// pollActiveMeetMeetings fetches attendance for every Meet meeting
// whose scheduled window overlaps [now - grace, now + grace].
func (s *videoService) pollActiveMeetMeetings(ctx context.Context) {
	now := time.Now().UTC()
	from := now.Add(-AutoFetchGrace)
	to := now.Add(AutoFetchGrace)

	meetings, err := s.deps.Meetings.ListActiveGoogleMeetMeetings(ctx, from, to)
	if err != nil {
		log.Printf("[video] auto-fetch meet: list meetings: %v", err)
		return
	}
	if len(meetings) == 0 {
		return
	}

	log.Printf(
		"[video] auto-fetch meet: %d active meeting(s) in window",
		len(meetings),
	)

	for _, m := range meetings {
		// Each meeting polls in its own goroutine so a slow Meet
		// response on one doesn't block the loop.
		go s.autoFetchOneMeeting(ctx, m)
	}
}

// autoFetchOneMeeting runs FetchGoogleMeetAttendance for a single
// meeting. Best-effort: failures are logged, not propagated.
func (s *videoService) autoFetchOneMeeting(
	ctx context.Context,
	m *videodomain.Meeting,
) {
	if m == nil || m.ID == "" || m.UserID == "" {
		return
	}

	// Re-check the platform defensively — the repository filter
	// should guarantee Meet-only, but the poller should not fail if
	// it ever receives something else.
	if m.Platform != videodomain.PlatformGoogleMeet {
		return
	}

	result, err := s.FetchGoogleMeetAttendance(ctx, FetchGoogleMeetAttendanceCommand{
		UserID:    m.UserID,
		MeetingID: m.ID,
	})
	if err != nil {
		log.Printf(
			"[video] auto-fetch meet: meeting=%s err=%v",
			m.ID, err,
		)
		return
	}

	log.Printf(
		"[video] auto-fetch meet: meeting=%s records=%d participants=%d dispatched=%d unmatched=%d",
		m.ID,
		result.ConferenceRecords,
		result.Participants,
		result.EventsDispatched,
		len(result.UnmatchedParticipants),
	)
}