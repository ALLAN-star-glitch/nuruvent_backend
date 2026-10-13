// internal/modules/video/service/get_join_info.go

package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// GetMeetingJoinInfo assembles the credentials a browser needs to
// join a meeting via the embedded Meeting SDK (Client View).
//
// The meeting is looked up by its platform-side external ID (the
// numeric Zoom meeting number), not by Nuruvent's UUID. This is
// because the schedule stores the platform ID, not the UUID.
//
// Role: the meeting owner joins as host (role 1), everyone else joins
// as attendee (role 0). Hosts receive a ZAK in addition to the
// signature; attendees receive only the signature.
//
// Password: Client View requires the *encrypted* pwd= value from the
// meeting's join URL, not the raw human-readable password from the
// `password` column. The two are different strings and Zoom rejects
// the raw one with "password are invalid".
//
// DisplayName: the participant's name as it should appear in the Zoom
// meeting. Resolved from the Nuruvent user record via the Users port.
// Guests depend on this because their browser session has no auth
// slice — without it, the SDK defaults to "Attendee".
func (s *videoService) GetMeetingJoinInfo(
	ctx context.Context,
	cmd GetMeetingJoinInfoCommand,
) (*MeetingJoinInfo, error) {
	if strings.TrimSpace(cmd.UserID) == "" {
		return nil, fmt.Errorf("%w: user id is required", videodomain.ErrUnauthorized)
	}
	if !cmd.Platform.IsValid() {
		return nil, fmt.Errorf("%w: invalid platform",
			videodomain.ErrUnsupportedPlatform)
	}
	if strings.TrimSpace(cmd.ExternalID) == "" {
		return nil, fmt.Errorf("%w: external id is required",
			videodomain.ErrInvalidMeeting)
	}

	meeting, err := s.deps.Meetings.FindByExternalID(
		ctx, cmd.Platform, cmd.ExternalID,
	)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, videodomain.ErrMeetingNotFound
		}
		return nil, fmt.Errorf("get join info: load meeting: %w", err)
	}
	if meeting == nil {
		return nil, videodomain.ErrMeetingNotFound
	}

	if strings.TrimSpace(meeting.ExternalID) == "" {
		return nil, fmt.Errorf("%w: meeting has no external id",
			videodomain.ErrInvalidMeeting)
	}

	role := 0
	if meeting.UserID == cmd.UserID {
		role = 1
	}

	client, err := s.deps.Clients.For(meeting.Platform)
	if err != nil {
		return nil, fmt.Errorf("get join info: %w", err)
	}
	signer, ok := client.(videodomain.MeetingSDKSigner)
	if !ok {
		return nil, fmt.Errorf("%w: platform %s does not support the meeting SDK",
			videodomain.ErrCapabilityMissing, meeting.Platform)
	}

	signature, err := signer.SignedMeetingJWT(meeting.ExternalID, role)
	if err != nil {
		return nil, fmt.Errorf("get join info: sign: %w", err)
	}

	// The join URL carries the encrypted password as ?pwd=... and the
	// region's web endpoint host. Client View needs the encrypted
	// password. Fall back to the raw password if the URL is missing —
	// Zoom will reject it with a specific error, which is easier to
	// debug than a silent failure.
	webEndpoint := ""
	encryptedPwd := meeting.Password
	if meeting.JoinURL != "" {
		if u, err := url.Parse(meeting.JoinURL); err == nil {
			webEndpoint = u.Host
			if p := u.Query().Get("pwd"); p != "" {
				encryptedPwd = p
			}
		}
	}

	// Resolve the participant's display name. Best-effort: if the
	// lookup port isn't wired or the user can't be resolved, the SDK
	// receives an empty string and falls back to its own default
	// ("Attendee"). A failed lookup must never fail the join.
	displayName := ""
	if s.deps.Users != nil {
		name, lookupErr := s.deps.Users.GetDisplayName(ctx, cmd.UserID)
		if lookupErr != nil {
			log.Printf(
				"[join-info] user lookup failed user=%s err=%v",
				cmd.UserID, lookupErr,
			)
		} else {
			displayName = name
		}
	}

	info := &MeetingJoinInfo{
		MeetingNumber: meeting.ExternalID,
		Signature:     signature,
		SDKKey:        signer.ClientID(),
		Password:      encryptedPwd,
		WebEndpoint:   webEndpoint,
		Role:          role,
		DisplayName:   displayName,
	}

	if role == 1 {
		conn, err := s.deps.Connections.FindActiveByUserAndPlatform(
			ctx, cmd.UserID, meeting.Platform,
		)
		if err != nil {
			if errorsIsNotFound(err) {
				return nil, videodomain.ErrNotConnected
			}
			return nil, fmt.Errorf("get join info: find connection: %w", err)
		}
		if conn.RevokedAt != nil {
			return nil, videodomain.ErrConnectionRevoked
		}

		conn, err = s.ensureFreshToken(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("get join info: %w", err)
		}

		zak, err := signer.FetchZAK(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("get join info: fetch zak: %w", err)
		}
		info.ZAK = zak
	}

	return info, nil
}