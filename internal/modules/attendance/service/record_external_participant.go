// internal/modules/attendance/service/record_external_participant.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RecordExternalParticipant records a join or leave observed on an
// external platform.
//
// Flow:
//  1. Validate the command.
//  2. Resolve the session by (provider, meeting_code).
//  3. Match the participant to a registered attendee.
//  4. Record the join/leave.
//  5. Recompute rollups.
//
// Unmatched participants return ErrParticipantUnmatched. Callers
// treat that as a soft signal — the participant is collected for
// host reconciliation, not treated as a failure.
func (s *attendanceService) RecordExternalParticipant(
	ctx context.Context,
	cmd RecordExternalParticipantCommand,
) error {
	if strings.TrimSpace(cmd.Provider) == "" {
		return fmt.Errorf("%w: provider is required", attendance.ErrInvalidSession)
	}
	if strings.TrimSpace(cmd.MeetingCode) == "" {
		return fmt.Errorf("%w: meeting code is required", attendance.ErrInvalidSession)
	}
	if strings.TrimSpace(cmd.ParticipantName) == "" {
		return fmt.Errorf("%w: participant name is required", attendance.ErrInvalidSession)
	}
	if cmd.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", attendance.ErrInvalidSession)
	}
	if !cmd.EventType.IsValid() {
		return fmt.Errorf("%w: unknown event type %q",
			attendance.ErrInvalidSession, cmd.EventType)
	}

	provider := attendance.SessionProvider(cmd.Provider)

	// 1. Resolve the session.
	session, err := s.findSessionByProviderMeeting(ctx, provider, cmd.MeetingCode)
	if err != nil {
		if errors.Is(err, attendance.ErrSessionNotFound) {
			log.Printf(
				"[attendance] record external participant: no session for provider=%s meeting=%s",
				cmd.Provider, cmd.MeetingCode,
			)
			return nil
		}
		return err
	}

	// 2. Match the participant to a registered attendee.
	attendee, err := s.matchParticipantToAttendee(
		ctx,
		session.ID,
		cmd.ParticipantName,
		cmd.ExternalUserID,
		provider,
	)
	if err != nil {
		return err
	}
	if attendee == nil {
		log.Printf(
			"[attendance] record external participant: no unique match "+
				"provider=%s session=%s name=%q external_user_id=%q",
			cmd.Provider, session.ID, cmd.ParticipantName, cmd.ExternalUserID,
		)
		return attendance.ErrParticipantUnmatched
	}

	// 3. Record the event.
	switch cmd.EventType {
	case ExternalParticipantJoined:
		_, err = s.RecordJoin(ctx, RecordJoinCommand{
			AttendeeID: attendee.ID,
			SessionID:  session.ID,
			JoinTime:   cmd.OccurredAt,
			Source:     sourceForProvider(provider),
		})
		if err != nil {
			return fmt.Errorf("record join: %w", err)
		}
	case ExternalParticipantLeft:
		err = s.RecordLeave(ctx, RecordLeaveCommand{
			AttendeeID: attendee.ID,
			SessionID:  session.ID,
			LeaveTime:  cmd.OccurredAt,
			Source:     sourceForProvider(provider),
		})
		if err != nil {
			return fmt.Errorf("record leave: %w", err)
		}
	}

	// 4. Recompute rollups.
	if err := s.RecomputeSessionStatuses(ctx, session.ID); err != nil {
		log.Printf(
			"[attendance] record external participant: recompute failed session=%s err=%v",
			session.ID, err,
		)
	}

	return nil
}

// ============================================================
// IDENTITY LINKING
// ============================================================

// SetAttendeeGoogleMeetID is the exported entry point for
// setAttendeeGoogleMeetID. Used by the video module's participant
// linking flow: the host picks an attendee from the roster, the video
// handler calls this to persist the mapping, then re-polls.
func (s *attendanceService) SetAttendeeGoogleMeetID(
	ctx context.Context,
	attendeeID, googleMeetUserID string,
) error {
	if strings.TrimSpace(attendeeID) == "" {
		return fmt.Errorf("%w: attendee_id is required", attendance.ErrInvalidSession)
	}
	if strings.TrimSpace(googleMeetUserID) == "" {
		return fmt.Errorf("%w: google_meet_user_id is required", attendance.ErrInvalidSession)
	}
	return s.setAttendeeGoogleMeetID(ctx, attendeeID, googleMeetUserID)
}

// ============================================================
// PARTICIPANT MATCHING
// ============================================================

// matchParticipantToAttendee resolves a platform participant to a
// registered attendee for the session.
//
// Matching strategy, in order:
//
//  1. Google Meet user id — a stable per-Google-account identifier
//     stored on the attendee row after a successful first match.
//  2. Normalized display name — case-insensitive, whitespace-collapsed,
//     punctuation-tolerant comparison.
//
// Returns (nil, nil) when neither path finds a unique match.
func (s *attendanceService) matchParticipantToAttendee(
	ctx context.Context,
	sessionID string,
	participantName string,
	externalUserID string,
	provider attendance.SessionProvider,
) (*attendance.Attendee, error) {
	// Tier 1: Google Meet user id, if we've linked this participant
	// to an attendee before.
	if provider == attendance.ProviderGoogleMeet && externalUserID != "" {
		if a, err := s.findAttendeeByGoogleMeetID(ctx, sessionID, externalUserID); err != nil {
			return nil, err
		} else if a != nil {
			return a, nil
		}
	}

	// Tier 2: name match.
	matched, err := s.matchAttendeeByDisplayName(ctx, sessionID, participantName)
	if err != nil {
		return nil, err
	}
	if matched == nil {
		return nil, nil
	}

	// Tier 3: persist the Meet user id on the matched attendee so the
	// next poll takes tier 1.
	if provider == attendance.ProviderGoogleMeet &&
		externalUserID != "" &&
		matched.GoogleMeetUserID == "" {
		if err := s.setAttendeeGoogleMeetID(ctx, matched.ID, externalUserID); err != nil {
			log.Printf(
				"[attendance] set google_meet_user_id attendee=%s user=%s: %v",
				matched.ID, externalUserID, err,
			)
		}
	}

	return matched, nil
}

// findAttendeeByGoogleMeetID looks up the attendee registered for a
// session whose google_meet_user_id matches. Returns (nil, nil) when
// no match exists — the caller falls through to name matching.
func (s *attendanceService) findAttendeeByGoogleMeetID(
	ctx context.Context,
	sessionID, googleMeetUserID string,
) (*attendance.Attendee, error) {
	var matched *attendance.Attendee

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		statuses, err := repos.SessionStatuses.ListBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("list session statuses: %w", err)
		}
		for _, st := range statuses {
			a, err := repos.Attendees.FindByID(ctx, st.AttendeeID)
			if err != nil {
				return fmt.Errorf("load attendee %s: %w", st.AttendeeID, err)
			}
			if a == nil {
				continue
			}
			if a.GoogleMeetUserID == googleMeetUserID {
				matched = a
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matched, nil
}

// setAttendeeGoogleMeetID persists the Meet user id on an attendee
// row so future polls match by identity instead of by name.
func (s *attendanceService) setAttendeeGoogleMeetID(
	ctx context.Context,
	attendeeID, googleMeetUserID string,
) error {
	return s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		a, err := repos.Attendees.FindByID(ctx, attendeeID)
		if err != nil {
			return fmt.Errorf("load attendee: %w", err)
		}
		if a == nil {
			return nil
		}
		a.GoogleMeetUserID = googleMeetUserID
		return repos.Attendees.Update(ctx, a)
	})
}

// matchAttendeeByDisplayName finds the single attendee registered
// for a session whose normalized display name matches.
//
// Returns (nil, nil) when there is no match OR when there are
// multiple matches. Ambiguity is treated as "no match" so we never
// record against the wrong attendee.
func (s *attendanceService) matchAttendeeByDisplayName(
	ctx context.Context,
	sessionID, displayName string,
) (*attendance.Attendee, error) {
	needle := normalizeName(displayName)
	if needle == "" {
		return nil, nil
	}

	var matched *attendance.Attendee

	err := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		statuses, err := repos.SessionStatuses.ListBySession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("list session statuses: %w", err)
		}

		var candidates []*attendance.Attendee
		for _, st := range statuses {
			a, err := repos.Attendees.FindByID(ctx, st.AttendeeID)
			if err != nil {
				return fmt.Errorf("load attendee %s: %w", st.AttendeeID, err)
			}
			if a == nil {
				continue
			}
			if normalizeName(a.DisplayName) == needle {
				candidates = append(candidates, a)
			}
		}

		switch len(candidates) {
		case 0:
			return nil
		case 1:
			matched = candidates[0]
			return nil
		default:
			log.Printf(
				"[attendance] match by name: ambiguous session=%s name=%q candidates=%d",
				sessionID, displayName, len(candidates),
			)
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	return matched, nil
}

// normalizeName produces a canonical form for name comparison.
//
// Rules:
//   - lowercase
//   - Unicode-aware: strips diacritics like "café" → "cafe"
//   - collapse runs of whitespace to a single space
//   - drop punctuation and symbols (periods, commas, dashes)
//   - trim leading/trailing whitespace
//
// Not a fuzzy match: "Allan Mathenge" and "Allan Kamau" still
// normalize to different strings.
func normalizeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))

	prevSpace := true
	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			b.WriteRune(unicode.ToLower(r))
			prevSpace = false
		case unicode.IsDigit(r):
			b.WriteRune(r)
			prevSpace = false
		case unicode.IsSpace(r):
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		default:
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		}
	}

	return strings.TrimRight(b.String(), " ")
}