// internal/modules/video/service/get_join_info.go

package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// GetMeetingJoinInfo assembles the credentials a browser needs to
// join a meeting via the embedded Meeting SDK.
//
// The meeting is looked up by its platform-side external ID (the
// numeric Zoom meeting number), not by Nuruvent's UUID. This is
// because the schedule stores the platform ID, not the UUID.
//
// The caller's role is derived from the meeting's UserID: the
// meeting owner joins as host (role 1), everyone else joins as
// attendee (role 0). Hosts receive a ZAK in addition to the
// signature; attendees receive only the signature.
//
// Password handling: Zoom's Meeting SDK 6.x expects the *raw*
// meeting password (the value shown in the Zoom web client as the
// user-typed password), not the encrypted pwd= value from the join
// URL. The DB column `password` stores the raw value.
//
// (The encrypted pwd= value is only used by Zoom's older SDKs and
// by the SDK Embed product. If Zoom ever changes this for 6.x, the
// extraction from join_url is here for reference:
//
//	if p := u.Query().Get("pwd"); p != "" { password = p }
// )
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

	// Derive the region's web endpoint from the join URL host so the
	// SDK connects to the correct Zoom cluster (e.g. us04web.zoom.us).
	webEndpoint := ""
	if meeting.JoinURL != "" {
		if u, err := url.Parse(meeting.JoinURL); err == nil {
			webEndpoint = u.Host
		}
	}

	info := &MeetingJoinInfo{
		MeetingNumber: meeting.ExternalID,
		Signature:     signature,
		SDKKey:        signer.ClientID(),
		Password:      meeting.Password, // raw password — what SDK 6.x expects
		WebEndpoint:   webEndpoint,
		Role:          role,
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