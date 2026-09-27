// internal/modules/video/service/delete_meeting.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// DeleteMeeting removes a meeting from the platform and from local
// storage.
//
// Steps:
//  1. Validate inputs.
//  2. Load the meeting by external ID and verify ownership.
//  3. Resolve the host's current connection for the meeting's
//     platform. If the connection is gone or revoked, skip the
//     platform call and delete locally — nothing to authenticate with.
//  4. Ask the platform to delete the meeting. Best-effort: a failure
//     here is swallowed — the local record must go regardless.
//  5. Delete the local record.
//
// Idempotent: calling with an external ID that no longer exists
// returns nil.
func (s *videoService) DeleteMeeting(
	ctx context.Context,
	cmd DeleteMeetingCommand,
) error {
	if err := validateDeleteMeetingCommand(cmd); err != nil {
		return err
	}

	// 1. Load the meeting by platform + external ID.
	meeting, err := s.deps.Meetings.FindByExternalID(ctx, cmd.Platform, cmd.ExternalID)
	if err != nil {
		if errorsIsMeetingNotFound(err) {
			return nil // already gone locally — idempotent
		}
		return fmt.Errorf("delete meeting: load: %w", err)
	}
	if meeting == nil {
		return nil
	}

	// 2. Ownership check.
	if meeting.UserID != cmd.UserID {
		return fmt.Errorf("%w: meeting does not belong to user",
			videodomain.ErrForbidden)
	}

	// 3. Resolve the host's current connection for this platform.
	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(
		ctx, cmd.UserID, meeting.Platform,
	)
	if err != nil {
		if errorsIsNotFound(err) {
			return s.deleteLocalMeeting(ctx, meeting.ID)
		}
		return fmt.Errorf("delete meeting: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return s.deleteLocalMeeting(ctx, meeting.ID)
	}

	// 4. Best-effort platform delete.
	provisioner, err := s.deps.Clients.ProvisionerFor(meeting.Platform)
	if err != nil {
		return s.deleteLocalMeeting(ctx, meeting.ID)
	}

	fresh, ferr := s.ensureFreshToken(ctx, conn)
	if ferr == nil {
		conn = fresh
		if derr := provisioner.DeleteMeeting(ctx, conn, meeting.ExternalID); derr != nil {
			// A 404 means the meeting is already gone on Zoom —
			// that's fine. Any other error is logged by the caller
			// but doesn't block local deletion.
			if !errors.Is(derr, videodomain.ErrMeetingNotFound) {
				// no-op: delivery layer logs
			}
		}
	}

	// 5. Remove locally.
	return s.deleteLocalMeeting(ctx, meeting.ID)
}

// deleteLocalMeeting removes the meeting row.
func (s *videoService) deleteLocalMeeting(ctx context.Context, meetingID string) error {
	return s.deps.UnitOfWork.Do(ctx, func(repos videodomain.Repositories) error {
		if err := repos.Meetings.Delete(ctx, meetingID); err != nil {
			if errorsIsMeetingNotFound(err) {
				return nil
			}
			return fmt.Errorf("delete meeting: local delete: %w", err)
		}
		return nil
	})
}

func validateDeleteMeetingCommand(cmd DeleteMeetingCommand) error {
	if strings.TrimSpace(cmd.UserID) == "" {
		return fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return fmt.Errorf("%w: invalid platform %q",
			videodomain.ErrUnsupportedPlatform, cmd.Platform)
	}
	if strings.TrimSpace(cmd.ExternalID) == "" {
		return fmt.Errorf("%w: external id is required", videodomain.ErrInvalidMeeting)
	}
	return nil
}