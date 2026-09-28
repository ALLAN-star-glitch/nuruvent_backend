// internal/modules/video/infrastructure/providers/googlemeet/meeting.go

package googlemeet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// CREATE MEETING
// ============================================================

// CreateMeeting provisions a new Google Meet space.
//
// POST /v2/spaces with an empty body creates a new space. The
// response contains a resource name ("spaces/abc123") and a
// meetingUri ("https://meet.google.com/abc-mnop-xyz").
//
// Google Meet spaces are persistent: they are not scheduled
// meetings. The following Meeting fields are therefore empty in
// the returned object:
//
//   - StartURL
//   - Password
//
// Topic, StartTime, Duration, and Timezone are populated from the
// spec for local record-keeping, but are never sent to Google.
// The events module owns scheduling; Google Meet provides only the
// room.
//
// Rate limit: spaces.create is limited to 10 requests per minute
// per user per project. When Google returns 429, mapHTTPError
// produces ErrPlatformRateLimited.
func (c *Client) CreateMeeting(
	ctx context.Context,
	conn *videodomain.Connection,
	spec videodomain.MeetingSpec,
) (*videodomain.Meeting, error) {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return nil, videodomain.ErrNotConnected
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/spaces",
		strings.NewReader("{}"),
	)
	if err != nil {
		return nil, fmt.Errorf("build create space request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}

	var sr spaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decode space response: %w", err)
	}

	return toMeetingDomain(conn, &sr, spec)
}

// ============================================================
// UPDATE MEETING
// ============================================================

// UpdateMeeting is a no-op for Google Meet.
//
// Google Meet's PATCH endpoint updates space configuration (access
// type, entry point, moderation) — not scheduling metadata. Topic,
// start time, duration, and timezone cannot be changed because
// they do not exist on a Meet space.
//
// The method returns a Meeting constructed from the spec plus the
// external ID, with JoinURL left empty so the service preserves
// the existing value from its own record. This mirrors the Zoom
// provider, which also cannot retrieve the join URL from a PATCH
// response.
//
// If a future requirement needs to update space configuration
// (e.g. accessType), this is the place to add the network call.
func (c *Client) UpdateMeeting(
	ctx context.Context,
	conn *videodomain.Connection,
	externalID string,
	spec videodomain.MeetingSpec,
) (*videodomain.Meeting, error) {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return nil, videodomain.ErrNotConnected
	}
	if strings.TrimSpace(externalID) == "" {
		return nil, fmt.Errorf("%w: external id is required", videodomain.ErrInvalidMeeting)
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	// Google Meet spaces have no editable scheduling fields.
	// Return a Meeting constructed from the spec plus the external
	// ID; the service layer preserves the existing JoinURL.
	now := time.Now().UTC()
	return &videodomain.Meeting{
		UserID:     conn.UserID,
		Platform:   videodomain.PlatformGoogleMeet,
		ExternalID: externalID,
		JoinURL:    "", // preserved by the caller
		StartURL:   "",
		Password:   "",
		Topic:      spec.Topic,
		StartTime:  spec.StartTime,
		Duration:   spec.Duration,
		Timezone:   spec.Timezone,
		UpdatedAt:  now,
	}, nil
}

// ============================================================
// DELETE MEETING
// ============================================================

// DeleteMeeting ends an active conference within a Google Meet
// space, if one exists.
//
// Google Meet spaces are persistent and cannot be deleted via the
// public REST API. The closest operation is endActiveConference,
// which terminates any ongoing meeting in the space. If no
// conference is active, Google returns 200 with an empty body.
//
// This differs from Zoom, where delete removes the meeting
// entirely. For Google Meet, "delete" means "end any active
// session"; the space itself remains.
//
// A 404 is treated as success: the space no longer exists, so
// there is nothing to end. This mirrors the Zoom provider's
// handling of ErrMeetingNotFound on delete.
func (c *Client) DeleteMeeting(
	ctx context.Context,
	conn *videodomain.Connection,
	externalID string,
) error {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return videodomain.ErrNotConnected
	}
	if strings.TrimSpace(externalID) == "" {
		return fmt.Errorf("%w: external id is required", videodomain.ErrInvalidMeeting)
	}

	endpoint := c.baseURL + "/" + externalID + ":endActiveConference"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader("{}"),
	)
	if err != nil {
		return fmt.Errorf("build end conference request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return mapNetworkError(err)
	}
	defer resp.Body.Close()

	// 404 means the space no longer exists — treat as success.
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mapHTTPError(resp)
	}
	return nil
}

// ============================================================
// HELPERS
// ============================================================

// toMeetingDomain maps a Google Meet space response into the
// domain Meeting type.
//
// The scheduling fields (Topic, StartTime, Duration, Timezone) are
// taken from the spec for local record-keeping. They are never
// sent to Google and never returned by Google; the events module
// owns them.
//
// An empty space name or meeting URI is rejected as
// ErrPlatformRejected, because a space without either is not
// usable.
func toMeetingDomain(
	conn *videodomain.Connection,
	sr *spaceResponse,
	spec videodomain.MeetingSpec,
) (*videodomain.Meeting, error) {
	if strings.TrimSpace(sr.Name) == "" {
		return nil, fmt.Errorf("%w: empty space name", videodomain.ErrPlatformRejected)
	}
	if strings.TrimSpace(sr.MeetingURI) == "" {
		return nil, fmt.Errorf("%w: empty meeting uri", videodomain.ErrPlatformRejected)
	}

	now := time.Now().UTC()

	return &videodomain.Meeting{
		UserID:     conn.UserID,
		Platform:   videodomain.PlatformGoogleMeet,
		ExternalID: sr.Name,       // "spaces/abc123"
		JoinURL:    sr.MeetingURI, // "https://meet.google.com/abc-mnop-xyz"
		StartURL:   "",            // Google has no host-only start URL
		Password:   "",            // Google spaces have no password
		Topic:      spec.Topic,
		StartTime:  spec.StartTime,
		Duration:   spec.Duration,
		Timezone:   spec.Timezone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}