// internal/modules/video/infrastructure/providers/zoom/meeting.go

package zoom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// CREATE MEETING
// ============================================================

// CreateMeeting creates a meeting on the connected host's Zoom
// account.
//
// The host is identified by conn.AccessToken; Zoom uses the token to
// determine whose account the meeting belongs to. The client always
// posts to /users/me/meetings — the "me" being the token owner.
//
// If spec.Timezone is empty, the host's Zoom account timezone is
// used. The domain requires timezone to be set, so this is defensive.
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

	body := createMeetingRequest{
		Topic:     spec.Topic,
		Type:      2, // scheduled meeting
		StartTime: formatZoomTime(spec.StartTime),
		Duration:  int(spec.Duration.Minutes()),
		Timezone:  spec.Timezone,
		Agenda:    spec.Agenda,
		Settings: meetingSettings{
			JoinBeforeHost:   true,
			WaitingRoom:      false,
			ApprovalType:     2, // no Zoom-side registration
			Audio:            "both",
			AutoRecording:    "none",
			MuteUponEntry:    false,
			HostVideo:        false,
			ParticipantVideo: false,
			UsePMI:           false,
		},
	}

	resp, err := c.doAuthedJSON(ctx, http.MethodPost, c.baseURL+"/users/me/meetings",
		conn.AccessToken, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, mapHTTPError(resp)
	}

	var mr meetingResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, fmt.Errorf("decode meeting response: %w", err)
	}

	return toMeetingDomain(conn, &mr, spec)
}

// ============================================================
// DELETE MEETING
// ============================================================

// DeleteMeeting removes a meeting from Zoom.
//
// Zoom returns 204 No Content on success. Returns ErrMeetingNotFound
// if Zoom reports the meeting doesn't exist — the caller may treat
// that as already-deleted.
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

	endpoint := c.baseURL + "/meetings/" + url.PathEscape(externalID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return mapNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	return mapHTTPError(resp)
}

// ============================================================
// HELPERS
// ============================================================

// doAuthedJSON marshals body as JSON and POSTs/PUTs it with a bearer
// token. Callers are responsible for closing resp.Body.
func (c *Client) doAuthedJSON(
	ctx context.Context,
	method, endpoint, accessToken string,
	body any,
) (*http.Response, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.http.Do(req)
}

// toMeetingDomain maps Zoom's response to the domain Meeting type.
func toMeetingDomain(
	conn *videodomain.Connection,
	mr *meetingResponse,
	spec videodomain.MeetingSpec,
) (*videodomain.Meeting, error) {
	externalID := strconv.FormatInt(mr.ID, 10)
	if externalID == "" || externalID == "0" {
		return nil, fmt.Errorf("%w: empty meeting id", videodomain.ErrPlatformRejected)
	}
	if mr.JoinURL == "" {
		return nil, fmt.Errorf("%w: empty join url", videodomain.ErrPlatformRejected)
	}

	now := time.Now().UTC()

	return &videodomain.Meeting{
		UserID:     conn.UserID,
		Platform:   videodomain.PlatformZoom,
		ExternalID: externalID,
		JoinURL:    mr.JoinURL,
		StartURL:   mr.StartURL,
		Password:   mr.Password,
		Topic:      mr.Topic,
		StartTime:  spec.StartTime,
		Duration:   spec.Duration,
		Timezone:   mr.Timezone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// formatZoomTime formats a time.Time for Zoom's API. Zoom expects
// RFC 3339 in UTC, without fractional seconds.
//
// Example: "2026-10-01T06:00:00Z"
func formatZoomTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}