// internal/modules/video/infrastructure/providers/googlemeet/attendance.go

package googlemeet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ConferenceRecord is a single recorded conference for a meeting
// code. A single meeting code can produce multiple records if the
// meeting was started more than once.
type ConferenceRecord struct {
	Name      string     // "conferenceRecords/abc123"
	Space     string     // "spaces/efoBh0pzi5IB"
	StartTime time.Time
	EndTime   *time.Time
}

// Participant is a single participant in a conference record.
type Participant struct {
	Name        string     // "conferenceRecords/abc/participants/p1"
	UserID      string     // "users/123456"; empty for anonymous
	DisplayName string     // always populated
	IsAnonymous bool
	JoinedAt    time.Time
	LeftAt      *time.Time
}

// ============================================================
// LIST CONFERENCE RECORDS
// ============================================================

// ListConferenceRecords returns all conference records for a
// meeting code.
//
// meetingCode is the value stored in video_meetings.external_id.
// Google's convention is "spaces/{id}", so if the value is a bare
// ID like "efoBh0pzi5IB", we normalize to "spaces/efoBh0pzi5IB"
// before building the filter.
//
// GET /v2/conferenceRecords?filter=space.name="spaces/{id}"
func (c *Client) ListConferenceRecords(
	ctx context.Context,
	conn *videodomain.Connection,
	meetingCode string,
) ([]ConferenceRecord, error) {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return nil, videodomain.ErrNotConnected
	}
	if strings.TrimSpace(meetingCode) == "" {
		return nil, fmt.Errorf("%w: meeting code is required", videodomain.ErrInvalidMeeting)
	}

	space := meetingCode
	if !strings.HasPrefix(space, "spaces/") {
		space = "spaces/" + space
	}
	filter := fmt.Sprintf(`space.name="%s"`, space)

	var all []ConferenceRecord
	pageToken := ""

	for {
		endpoint := c.baseURL + "/conferenceRecords" +
			"?pageSize=100&filter=" + url.QueryEscape(filter)
		if pageToken != "" {
			endpoint += "&pageToken=" + url.QueryEscape(pageToken)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("build list conference records request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, mapNetworkError(err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, mapHTTPError(resp)
		}

		var list conferenceRecordsListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode conference records: %w", err)
		}
		resp.Body.Close()

		for _, r := range list.ConferenceRecords {
			rec := ConferenceRecord{
				Name:  r.Name,
				Space: r.Space,
			}
			if t, err := time.Parse(time.RFC3339, r.StartTime); err == nil {
				rec.StartTime = t.UTC()
			}
			if r.EndTime != "" {
				if t, err := time.Parse(time.RFC3339, r.EndTime); err == nil {
					u := t.UTC()
					rec.EndTime = &u
				}
			}
			all = append(all, rec)
		}

		if list.NextPageToken == "" {
			break
		}
		pageToken = list.NextPageToken
	}

	return all, nil
}

// ============================================================
// LIST PARTICIPANTS
// ============================================================

// ListParticipants returns all participants for a conference
// record.
//
// recordName is the "name" field from a ConferenceRecord, e.g.
// "conferenceRecords/abc123".
//
// GET /v2/{recordName}/participants
func (c *Client) ListParticipants(
	ctx context.Context,
	conn *videodomain.Connection,
	recordName string,
) ([]Participant, error) {
	if conn == nil || strings.TrimSpace(conn.AccessToken) == "" {
		return nil, videodomain.ErrNotConnected
	}
	if strings.TrimSpace(recordName) == "" {
		return nil, fmt.Errorf("%w: record name is required", videodomain.ErrInvalidMeeting)
	}

	var all []Participant
	pageToken := ""

	for {
		endpoint := c.baseURL + "/" + recordName + "/participants?pageSize=100"
		if pageToken != "" {
			endpoint += "&pageToken=" + url.QueryEscape(pageToken)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("build list participants request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, mapNetworkError(err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, mapHTTPError(resp)
		}

		var list participantsListResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode participants: %w", err)
		}
		resp.Body.Close()

		for _, pt := range list.Participants {
			part := Participant{Name: pt.Name}

			switch {
			case pt.SignedinUser != nil:
				part.UserID = pt.SignedinUser.User
				part.DisplayName = strings.TrimSpace(pt.SignedinUser.DisplayName)
			case pt.AnonymousUser != nil:
				part.DisplayName = strings.TrimSpace(pt.AnonymousUser.DisplayName)
				part.IsAnonymous = true
			case pt.PhoneUser != nil:
				part.DisplayName = strings.TrimSpace(pt.PhoneUser.DisplayName)
				part.IsAnonymous = true
			}

			if part.DisplayName == "" {
				part.DisplayName = "Guest"
			}

			if pt.EarliestStartTime != "" {
				if t, err := time.Parse(time.RFC3339, pt.EarliestStartTime); err == nil {
					part.JoinedAt = t.UTC()
				}
			}
			if pt.LatestEndTime != "" {
				if t, err := time.Parse(time.RFC3339, pt.LatestEndTime); err == nil {
					u := t.UTC()
					part.LeftAt = &u
				}
			}

			all = append(all, part)
		}

		if list.NextPageToken == "" {
			break
		}
		pageToken = list.NextPageToken
	}

	return all, nil
}