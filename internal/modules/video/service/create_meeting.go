// internal/modules/video/service/create_meeting.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// CreateMeeting creates a meeting on the connected host's platform
// account.
//
// Steps:
//  1. Validate inputs and spec.
//  2. Load the host's active connection for the requested platform.
//  3. Ensure the access token is fresh (refresh if near expiry).
//  4. Resolve the platform's MeetingProvisioner.
//  5. Ask the platform to create the meeting.
//  6. Rebuild the Meeting through the domain constructor so
//     validation and invariants apply, then persist.
//
// The caller (events module) persists the returned Meeting reference
// against the event. This module owns only the meeting record.
//
// On the rare case where the platform call succeeds but the local
// persist fails, the Meeting is returned alongside the error so the
// caller can decide whether to retry the persist or clean up the
// orphaned external meeting.
func (s *videoService) CreateMeeting(
	ctx context.Context,
	cmd CreateMeetingCommand,
) (*videodomain.Meeting, error) {
	if err := validateCreateMeetingCommand(cmd); err != nil {
		return nil, err
	}

	// 1. Resolve the host's active connection.
	conn, err := s.deps.Connections.FindActiveByUserAndPlatform(
		ctx, cmd.UserID, cmd.Platform,
	)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrNotConnected
		}
		return nil, fmt.Errorf("create meeting: find connection: %w", err)
	}
	if conn.RevokedAt != nil {
		return nil, videodomain.ErrConnectionRevoked
	}

	// 2. Ensure the access token is fresh.
	conn, err = s.ensureFreshToken(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("create meeting: %w", err)
	}

	// 3. Resolve the provisioner.
	provisioner, err := s.deps.Clients.ProvisionerFor(conn.Platform)
	if err != nil {
		return nil, fmt.Errorf("create meeting: %w", err)
	}

	// 4. Ask the platform. The client returns a partially-populated
	//    Meeting — its job is to fill the platform-side fields
	//    (ExternalID, JoinURL, StartURL, Password). Nuruvent-side
	//    fields are ours to set.
	platformMeeting, err := provisioner.CreateMeeting(ctx, conn, cmd.Spec)
	if err != nil {
		return nil, fmt.Errorf("create meeting: platform call: %w", err)
	}
	if platformMeeting == nil {
		return nil, errors.New("create meeting: platform returned nil meeting")
	}

	// 5. Rebuild through the domain constructor. This enforces every
	//    invariant (non-empty ExternalID and JoinURL, valid platform,
	//    valid spec) even if the provider client was sloppy. Any
	//    provider bug becomes a loud error here, not a corrupt row.
	now := s.deps.Clock.Now()
	meeting, err := videodomain.NewMeeting(
		s.deps.IDs.NewID(),
		cmd.UserID,
		conn.Platform,
		platformMeeting.ExternalID,
		platformMeeting.JoinURL,
		platformMeeting.StartURL,
		platformMeeting.Password,
		cmd.Spec,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create meeting: %w", err)
	}

	// 6. Persist locally.
	if err := s.deps.Meetings.Create(ctx, meeting); err != nil {
		return meeting, fmt.Errorf("create meeting: persist local record: %w", err)
	}

	return meeting, nil
}

// validateCreateMeetingCommand checks required fields before any I/O.
// Matches the style of BeginConnect / Disconnect: returns a domain
// sentinel wrapped with a human message.
func validateCreateMeetingCommand(cmd CreateMeetingCommand) error {
	if strings.TrimSpace(cmd.UserID) == "" {
		return fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return fmt.Errorf("%w: invalid platform %q",
			videodomain.ErrUnsupportedPlatform, cmd.Platform)
	}
	if cmd.Spec.HostUserID == "" {
		// The spec carries its own host; it must match the command.
		// Callers who forget to set it get a clear error rather than
		// a silently-wrong host attribution.
		return fmt.Errorf("%w: spec.host_user_id is required",
			videodomain.ErrInvalidMeeting)
	}
	if cmd.Spec.HostUserID != cmd.UserID {
		return fmt.Errorf("%w: spec.host_user_id must match user_id",
			videodomain.ErrInvalidMeeting)
	}
	if err := cmd.Spec.Validate(); err != nil {
		return err
	}
	return nil
}