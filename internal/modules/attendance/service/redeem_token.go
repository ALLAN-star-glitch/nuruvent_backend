// internal/modules/attendance/service/redeem_token.go

package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// RedeemJoinToken validates a raw join token and returns the Nuruvent
// meeting URL the browser should be redirected to.
//
// It does NOT write an attendance record. The click-through only
// proves the user opened the link; actual presence is captured by the
// platform's webhook (Zoom) or polling (Google Meet). Writing an
// attendance row here produced one orphaned row per click — no
// left_at, no duration — that inflated the session summary and never
// closed.
//
// The `join_tokens.redeemed_at` column (set by FindByHash or a
// separate MarkRedeemed call — depending on the repository) records
// that the click happened, for auditing.
//
// Safe to call multiple times: each redemption returns the same
// redirect URL and does not create any side effects beyond the token
// lookup itself.
func (s *attendanceService) RedeemJoinToken(
	ctx context.Context,
	rawToken string,
) (*RedeemResult, error) {
	if rawToken == "" {
		return nil, fmt.Errorf("%w: raw token is required", attendance.ErrInvalidToken)
	}

	now := s.deps.Clock.Now()
	hash := s.deps.TokenGenerator.Hash(rawToken)

	var result *RedeemResult

	txErr := s.deps.UnitOfWork.Do(ctx, func(repos attendance.Repositories) error {
		token, err := repos.JoinTokens.FindByHash(ctx, hash)
		if err != nil {
			return fmt.Errorf("lookup token: %w", err)
		}

		if !token.IsActive(now) {
			if token.RevokedAt != nil {
				return attendance.ErrTokenRevoked
			}
			return attendance.ErrTokenExpired
		}

		session, err := repos.Sessions.FindByID(ctx, token.SessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}

		if session.Status == attendance.SessionStatusCancelled {
			return fmt.Errorf("%w: session is cancelled", attendance.ErrTokenRevoked)
		}

		// No attendance record is written here. The click-through
		// is the audit signal, not the attendance signal.

		redirect, err := s.buildJoinRedirect(session)
		if err != nil {
			return err
		}

		result = &RedeemResult{
			AttendeeID: token.AttendeeID,
			SessionID:  token.SessionID,
			RedirectTo: redirect,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return result, nil
}

// buildJoinRedirect produces the frontend URL the browser should land
// on after a successful token redemption.
//
// Virtual sessions go to /meeting/:code — the Nuruvent page that
// hosts the embedded SDK flow. In-person sessions go to the event
// dashboard, since there's nothing to join remotely.
//
// The URL is absolute (scheme + host from AppConfig.PublicURL)
// because the HTTP handler responds with a 302 and the browser
// resolves relative Location headers against the request host. The
// request host is the backend's — not the frontend's — so a relative
// redirect would land the user on /meeting/... on the backend, where
// no such route exists.
//
// The path segment is the "meeting code" the frontend expects:
//
//   - Zoom: the numeric meeting ID (e.g. "71911238429") — already what
//     provider_meeting_id holds.
//   - Google Meet: the URL code (e.g. "dwt-neok-pgk"), NOT the space
//     resource name ("spaces/23_mkbr6Xy4B") that provider_meeting_id
//     holds. The space name is what Meet's webhook payloads use; the
//     code is what the frontend embeds.
func (s *attendanceService) buildJoinRedirect(session *attendance.Session) (string, error) {
	frontendBase := strings.TrimRight(s.deps.AppConfig.PublicURL, "/")
	if frontendBase == "" {
		frontendBase = "http://localhost:3000"
	}

	eventReturnPath := "/dashboard/events/" + session.External.ID

	if !session.Provider.RequiresMeetingID() {
		return frontendBase + eventReturnPath, nil
	}

	code := session.ProviderMeetingID
	if session.Provider == attendance.ProviderGoogleMeet {
		if c := meetCodeFromURL(session.ProviderURL); c != "" {
			code = c
		}
	}

	if code == "" {
		return "", fmt.Errorf(
			"%w: session has no meeting id for provider %q",
			attendance.ErrInvalidToken, session.Provider,
		)
	}

	params := url.Values{}
	params.Set("name", session.EventDisplayName)
	params.Set("host", session.OrganizerDisplayName)
	params.Set("return", eventReturnPath)
	params.Set("platform", string(session.Provider))

	return frontendBase + fmt.Sprintf("/meeting/%s?%s",
		code, params.Encode()), nil
}

// meetCodeFromURL extracts the meeting code from a Google Meet URL.
//
// "https://meet.google.com/dwt-neok-pgk"        → "dwt-neok-pgk"
// "meet.google.com/abc-defg-hij?authuser=0"     → "abc-defg-hij"
//
// Returns "" if the URL doesn't contain a recognizable code.
var meetCodeRe = regexp.MustCompile(`meet\.google\.com/([a-z]{3}-[a-z]{4}-[a-z]{3})`)

func meetCodeFromURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if m := meetCodeRe.FindStringSubmatch(u); len(m) >= 2 {
		return m[1]
	}
	return ""
}