// internal/modules/video/service/fetch_google_meet_attendance.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	googlemeet "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/infrastructure/providers/googlemeet"
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// FetchGoogleMeetAttendanceCommand is the input to
// FetchGoogleMeetAttendance.
type FetchGoogleMeetAttendanceCommand struct {
	UserID    string
	MeetingID string
}

// UnmatchedParticipant is a Meet participant that couldn't be
// resolved to a registered attendee on the last poll.
type UnmatchedParticipant struct {
	GoogleMeetUserID string
	DisplayName      string
	JoinedAt         time.Time
	LeftAt           *time.Time
}

// FetchGoogleMeetAttendanceResult summarizes what was fetched and
// ingested.
type FetchGoogleMeetAttendanceResult struct {
	MeetingNumber         string
	ConferenceRecords     int
	Participants          int
	EventsDispatched      int
	UnmatchedParticipants []UnmatchedParticipant
}

// FetchGoogleMeetAttendance polls Google Meet for conference records
// and participants, and dispatches normalized events to the
// attendance module.
func (s *videoService) FetchGoogleMeetAttendance(
	ctx context.Context,
	cmd FetchGoogleMeetAttendanceCommand,
) (*FetchGoogleMeetAttendanceResult, error) {
	if strings.TrimSpace(cmd.UserID) == "" {
		return nil, videodomain.ErrUnauthorized
	}
	if strings.TrimSpace(cmd.MeetingID) == "" {
		return nil, fmt.Errorf("%w: meeting id is required", videodomain.ErrInvalidMeeting)
	}

	meeting, err := s.deps.Meetings.FindByID(ctx, cmd.MeetingID)
	if err != nil {
		return nil, fmt.Errorf("fetch attendance: load meeting: %w", err)
	}
	if meeting == nil {
		return nil, videodomain.ErrMeetingNotFound
	}
	if meeting.UserID != cmd.UserID {
		return nil, videodomain.ErrUnauthorized
	}
	if meeting.Platform != videodomain.PlatformGoogleMeet {
		return nil, fmt.Errorf(
			"%w: attendance polling is only supported for google_meet",
			videodomain.ErrUnsupportedPlatform,
		)
	}

	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(
		ctx, cmd.UserID, videodomain.PlatformGoogleMeet,
	)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrNotConnected
		}
		return nil, fmt.Errorf("fetch attendance: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return nil, videodomain.ErrConnectionRevoked
	}
	conn, err = s.ensureFreshToken(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("fetch attendance: %w", err)
	}

	client, err := s.deps.Clients.For(videodomain.PlatformGoogleMeet)
	if err != nil {
		return nil, fmt.Errorf("fetch attendance: %w", err)
	}
	gm, ok := client.(*googlemeet.Client)
	if !ok {
		return nil, fmt.Errorf(
			"%w: google_meet provider does not expose attendance APIs",
			videodomain.ErrCapabilityMissing,
		)
	}

	meetingCode := meeting.ExternalID
	if strings.TrimSpace(meetingCode) == "" {
		return nil, fmt.Errorf("%w: meeting has no external id", videodomain.ErrInvalidMeeting)
	}

	records, err := gm.ListConferenceRecords(ctx, conn, meetingCode)
	if err != nil {
		return nil, fmt.Errorf("fetch attendance: list records: %w", err)
	}

	result := &FetchGoogleMeetAttendanceResult{
		MeetingNumber:     meetingCode,
		ConferenceRecords: len(records),
	}
	if len(records) == 0 {
		return result, nil
	}

	for _, rec := range records {
		participants, err := gm.ListParticipants(ctx, conn, rec.Name)
		if err != nil {
			log.Printf(
				"[video] fetch attendance: list participants record=%s err=%v",
				rec.Name, err,
			)
			continue
		}
		result.Participants += len(participants)

		log.Printf("[video] fetch attendance: record=%s participants=%d recordEndTime=%v",
			rec.Name, len(participants), rec.EndTime)

		for _, p := range participants {
			// Debug log so we can see what Meet actually returned
			// for this participant. Remove once the pipeline is
			// confirmed healthy.
			log.Printf(
				"[video] fetch attendance: participant name=%q user=%q anon=%v joined=%v left=%v",
				p.DisplayName, p.UserID, p.IsAnonymous, p.JoinedAt, p.LeftAt,
			)

			if p.JoinedAt.IsZero() {
				continue
			}

			// ---- JOIN ----
			err := s.deps.Attendance.RecordExternalParticipant(
				ctx,
				RecordParticipantCommand{
					Provider:        string(videodomain.PlatformGoogleMeet),
					MeetingCode:     meetingCode,
					ParticipantName: p.DisplayName,
					ExternalUserID:  p.UserID,
					OccurredAt:      p.JoinedAt,
					EventType:       EventTypeJoined,
				},
			)
			if err != nil {
				// Unmatched participant — collect it for the roster
				// UI so a host can link the Meet identity to a
				// registered attendee. The adapter translates the
				// attendance module's internal sentinel into the
				// video module's ErrParticipantUnmatched before this
				// call returns, so the check is against our own
				// sentinel, not the attendance domain's.
				if errors.Is(err, ErrParticipantUnmatched) {
					result.UnmatchedParticipants = append(
						result.UnmatchedParticipants,
						UnmatchedParticipant{
							GoogleMeetUserID: p.UserID,
							DisplayName:      p.DisplayName,
							JoinedAt:         p.JoinedAt,
							LeftAt:           p.LeftAt,
						},
					)
					// No leave to record either — nothing was stored.
					continue
				}
				log.Printf(
					"[video] fetch attendance: record join name=%q err=%v",
					p.DisplayName, err,
				)
				continue
			}
			result.EventsDispatched++

			// ---- LEAVE ----
			//
			// Prefer the participant's own latest end time. Fall back
			// to the conference record's EndTime when Meet hasn't
			// populated the per-participant field yet — this happens
			// for participants still connected when the host ends the
			// meeting, and immediately after the meeting ends while
			// Google is still finalizing per-participant records.
			leaveAt := p.LeftAt
			if leaveAt == nil && rec.EndTime != nil {
				leaveAt = rec.EndTime
			}

			if leaveAt == nil {
				// Meeting is still live and this participant hasn't
				// left. Nothing to close yet — a later poll will
				// catch the leave.
				continue
			}

			// Don't try to close with a time before the join. Guard
			// against clock skew or mis-ordered data.
			if !leaveAt.After(p.JoinedAt) {
				log.Printf(
					"[video] fetch attendance: skip invalid leave name=%q joined=%v leave=%v",
					p.DisplayName, p.JoinedAt, *leaveAt,
				)
				continue
			}

			if err := s.deps.Attendance.RecordExternalParticipant(
				ctx,
				RecordParticipantCommand{
					Provider:        string(videodomain.PlatformGoogleMeet),
					MeetingCode:     meetingCode,
					ParticipantName: p.DisplayName,
					ExternalUserID:  p.UserID,
					OccurredAt:      *leaveAt,
					EventType:       EventTypeLeft,
				},
			); err != nil {
				log.Printf(
					"[video] fetch attendance: record leave name=%q err=%v",
					p.DisplayName, err,
				)
				continue
			}
			result.EventsDispatched++
		}
	}

	return result, nil
}

// ============================================================
// LINK PARTICIPANT
// ============================================================

// LinkParticipantCommand is the input to LinkParticipant.
type LinkParticipantCommand struct {
	UserID           string
	MeetingID        string
	AttendeeID       string
	GoogleMeetUserID string
}

// LinkParticipant binds a Meet participant's Google user id to a
// registered attendee, then re-polls the meeting so the newly-linked
// identity is picked up and the join is recorded.
//
// Used by the roster's "unmatched participants" UI: the host picks
// the attendee, we persist the identity, we re-poll, done.
func (s *videoService) LinkParticipant(
	ctx context.Context,
	cmd LinkParticipantCommand,
) (*FetchGoogleMeetAttendanceResult, error) {
	if strings.TrimSpace(cmd.UserID) == "" {
		return nil, videodomain.ErrUnauthorized
	}
	if strings.TrimSpace(cmd.MeetingID) == "" {
		return nil, fmt.Errorf("%w: meeting id is required", videodomain.ErrInvalidMeeting)
	}
	if strings.TrimSpace(cmd.AttendeeID) == "" {
		return nil, fmt.Errorf("%w: attendee_id is required", videodomain.ErrInvalidMeeting)
	}
	if strings.TrimSpace(cmd.GoogleMeetUserID) == "" {
		return nil, fmt.Errorf("%w: google_meet_user_id is required", videodomain.ErrInvalidMeeting)
	}

	meeting, err := s.deps.Meetings.FindByID(ctx, cmd.MeetingID)
	if err != nil {
		return nil, fmt.Errorf("link participant: load meeting: %w", err)
	}
	if meeting == nil {
		return nil, videodomain.ErrMeetingNotFound
	}
	if meeting.UserID != cmd.UserID {
		return nil, videodomain.ErrUnauthorized
	}
	if meeting.Platform != videodomain.PlatformGoogleMeet {
		return nil, fmt.Errorf(
			"%w: identity linking is only supported for google_meet",
			videodomain.ErrUnsupportedPlatform,
		)
	}

	// Persist the identity on the attendee row.
	if err := s.deps.Attendance.SetAttendeeGoogleMeetID(
		ctx, cmd.AttendeeID, cmd.GoogleMeetUserID,
	); err != nil {
		return nil, fmt.Errorf("link participant: set identity: %w", err)
	}

	// Re-poll. Now that the identity is stored, the matcher will
	// resolve this participant to the attendee and record the join.
	return s.FetchGoogleMeetAttendance(ctx, FetchGoogleMeetAttendanceCommand{
		UserID:    cmd.UserID,
		MeetingID: cmd.MeetingID,
	})
}

// ============================================================
// GET UNMATCHED PARTICIPANTS
// ============================================================

// GetUnmatchedParticipantsCommand is the input to
// GetUnmatchedParticipants.
type GetUnmatchedParticipantsCommand struct {
	UserID    string
	MeetingID string
}

// GetUnmatchedParticipants polls the meeting and returns any
// participants that couldn't be resolved to a registered attendee.
//
// Used by the roster dialog to surface a "link to attendee" UI.
// Because the poll also records matched joins/leaves as a side
// effect, the same call doubles as a refresh of the roster state.
// GetUnmatchedParticipants polls the meeting and returns any
// participants that couldn't be resolved to a registered attendee.
//
// Multiple participation events for the same Google user are
// collapsed into a single entry — one row per physical Google
// account, keeping the most recent join/leave window.
//
// Used by the roster dialog to surface a "link to attendee" UI.
// Because the poll also records matched joins/leaves as a side
// effect, the same call doubles as a refresh of the roster state.
func (s *videoService) GetUnmatchedParticipants(
	ctx context.Context,
	cmd GetUnmatchedParticipantsCommand,
) ([]UnmatchedParticipant, error) {
	result, err := s.FetchGoogleMeetAttendance(ctx, FetchGoogleMeetAttendanceCommand{
		UserID:    cmd.UserID,
		MeetingID: cmd.MeetingID,
	})
	if err != nil {
		return nil, err
	}

	// Collapse multiple participation events per Google user into
	// one entry. The user resource name is stable per Google account
	// across conference records, so this maps to one physical
	// person even if they joined the meeting space more than once.
	byUserID := make(map[string]UnmatchedParticipant)

	for _, p := range result.UnmatchedParticipants {
		if p.GoogleMeetUserID == "" {
			// No stable id — keep as a distinct entry using
			// name+timestamp as a composite key.
			key := fmt.Sprintf("anon:%s:%d", p.DisplayName, p.JoinedAt.Unix())
			byUserID[key] = p
			continue
		}

		existing, ok := byUserID[p.GoogleMeetUserID]
		if !ok {
			byUserID[p.GoogleMeetUserID] = p
			continue
		}

		// Take the most recent join and display name.
		if p.JoinedAt.After(existing.JoinedAt) {
			existing.JoinedAt = p.JoinedAt
			existing.DisplayName = p.DisplayName
		}

		// Prefer the most recent non-nil leave time.
		if p.LeftAt != nil {
			if existing.LeftAt == nil || p.LeftAt.After(*existing.LeftAt) {
				existing.LeftAt = p.LeftAt
			}
		}

		byUserID[p.GoogleMeetUserID] = existing
	}

	out := make([]UnmatchedParticipant, 0, len(byUserID))
	for _, p := range byUserID {
		out = append(out, p)
	}

	// Most recent first.
	sort.Slice(out, func(i, j int) bool {
		return out[i].JoinedAt.After(out[j].JoinedAt)
	})

	return out, nil
}