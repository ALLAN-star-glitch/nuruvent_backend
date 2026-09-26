// internal/modules/attendance/infrastructure/providers/zoom/provider.go

package zoom

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// ProviderName is the identifier used in the attendance module for
// sessions served by Zoom.
const ProviderName = "zoom"

// MaxSignatureAge is how old a signature can be before Zoom's webhook
// delivery is rejected. Zoom recommends a 5-minute window; anything
// older is treated as a replay.
const MaxSignatureAge = 5 * time.Minute

// Provider implements the attendance.ProviderAdapter interface for
// Zoom.
//
// The cfg is validated by the caller — the wiring only constructs
// this adapter when cfg.IsConfigured() is true.
type Provider struct {
	cfg config.ZoomConfig
	// now is injected so tests can control time. Production uses
	// time.Now via NewProvider.
	now func() time.Time
}

// NewProvider constructs a Zoom adapter.
//
// cfg carries the Secret Token (used for webhook signature
// verification) plus Account ID / Client ID / Client Secret (reserved
// for future REST API calls, unused for webhook-only operation).
func NewProvider(cfg config.ZoomConfig) *Provider {
	return &Provider{
		cfg: cfg,
		now: time.Now,
	}
}

// Provider returns the attendance session provider this adapter
// handles.
func (p *Provider) Provider() attendance.SessionProvider {
	return attendance.ProviderZoom
}

// ParseWebhook verifies the Zoom signature and translates the payload
// into a normalized attendance.WebhookEvent.
//
// Errors:
//   - Missing signature headers → ErrInvalidSession wrapped
//   - Signature mismatch → same
//   - Timestamp outside the replay window → same
//   - Unknown event type → same
//
// Zoom's URL validation event (sent once when the subscription is
// created) is returned as a *URLValidationError. The HTTP handler
// detects it, calls (*Provider).HandleURLValidation, and responds
// with the returned body.
func (p *Provider) ParseWebhook(
	ctx context.Context,
	payload []byte,
	headers map[string]string,
) (*attendance.WebhookEvent, error) {
	// 1. Verify the signature.
	if err := p.verifySignature(payload, headers); err != nil {
		return nil, err
	}

	// 2. Parse the envelope.
	var env webhookEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return nil, fmt.Errorf("%w: malformed payload: %v", attendance.ErrInvalidSession, err)
	}

	// 3. Handle URL validation (endpoint verification).
	if env.Event == EventEndpointURLValidation {
		return nil, errURLValidation(env.Payload.PlainToken)
	}

	// 4. Normalize the event type.
	var eventType attendance.WebhookEventType
	switch env.Event {
	case EventParticipantJoined:
		eventType = attendance.WebhookEventJoined
	case EventParticipantLeft:
		eventType = attendance.WebhookEventLeft
	default:
		return nil, fmt.Errorf("%w: unhandled zoom event %q", attendance.ErrInvalidSession, env.Event)
	}

	// 5. Validate required fields.
	if env.Payload.Object.ID == "" {
		return nil, fmt.Errorf("%w: missing meeting id", attendance.ErrInvalidSession)
	}
	if env.Payload.Object.Participant.Email == "" {
		return nil, fmt.Errorf("%w: missing participant email", attendance.ErrInvalidSession)
	}

	// 6. Compute the timestamp.
	occurredAt := p.occurredAtFor(env, eventType)

	// 7. Synthesize a stable event ID for dedup. Zoom does not expose
	//    a per-event ID for participant webhooks; the meeting UUID +
	//    participant UUID + event type is unique per delivery and
	//    stable across redeliveries.
	providerEventID := synthesizeEventID(
		env.Payload.Object.UUID,
		env.Payload.Object.Participant.ParticipantUUID,
		string(eventType),
	)

	return &attendance.WebhookEvent{
		ProviderName:      attendance.ProviderZoom,
		ProviderEventID:   providerEventID,
		ProviderMeetingID: env.Payload.Object.ID,
		EventType:         eventType,
		ParticipantEmail:  strings.ToLower(strings.TrimSpace(env.Payload.Object.Participant.Email)),
		ParticipantName:   strings.TrimSpace(env.Payload.Object.Participant.UserName),
		OccurredAt:        occurredAt,
		Raw:               nil,
	}, nil
}

// occurredAtFor computes the timestamp for a webhook event.
//
// Preference order:
//  1. The participant's join_time (for joined) or leave_time (for left).
//  2. The envelope's event_ts (Unix milliseconds).
//  3. The provider's current clock.
func (p *Provider) occurredAtFor(
	env webhookEnvelope,
	eventType attendance.WebhookEventType,
) time.Time {
	participant := env.Payload.Object.Participant

	if eventType == attendance.WebhookEventLeft && participant.LeaveTime != "" {
		if t, err := parseZoomTime(participant.LeaveTime); err == nil {
			return t
		}
	}
	if eventType == attendance.WebhookEventJoined && participant.JoinTime != "" {
		if t, err := parseZoomTime(participant.JoinTime); err == nil {
			return t
		}
	}

	// Fallback: envelope timestamp (Unix milliseconds).
	if env.EventTS > 0 {
		return time.UnixMilli(env.EventTS).UTC()
	}

	// Last resort: current time. This is defensive — a webhook with
	// no participant timestamps and no event_ts is malformed, but we
	// prefer to record something rather than reject.
	return p.now().UTC()
}

// verifySignature checks the x-zm-signature header against the
// HMAC-SHA256 of the payload using the configured secret.
//
// Constant-time comparison. Timestamp freshness check. Missing headers
// fail closed.
func (p *Provider) verifySignature(payload []byte, headers map[string]string) error {
	signature := headerValue(headers, "x-zm-signature")
	timestamp := headerValue(headers, "x-zm-request-timestamp")

	if signature == "" || timestamp == "" {
		return fmt.Errorf("%w: missing zoom signature headers", attendance.ErrInvalidSession)
	}

	// Replay window check uses a parsed (and possibly normalized)
	// timestamp, but the HMAC message below uses the raw string.
	tsMs, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid zoom timestamp", attendance.ErrInvalidSession)
	}
	// Normalize seconds → milliseconds if Zoom sent a shorter value.
	tsForCheck := tsMs
	if tsForCheck < 1_000_000_000_000 {
		tsForCheck *= 1000
	}
	ts := time.UnixMilli(tsForCheck)
	if diff := p.now().Sub(ts); diff > MaxSignatureAge || diff < -MaxSignatureAge {
		return fmt.Errorf("%w: zoom signature outside replay window (diff=%v)",
			attendance.ErrInvalidSession, diff)
	}

	// Compute expected signature using the EXACT header string.
	// Do not reconstruct the timestamp from parsed fields — Zoom
	// signs with the raw header value including milliseconds.
	message := "v0:" + timestamp + ":" + string(payload)
	mac := hmac.New(sha256.New, []byte(p.cfg.SecretToken))
	mac.Write([]byte(message))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))

	// 🔻 DEBUG — remove after fixing
	log.Printf("[zoom] SIG DEBUG header=%q expected=%q secret_len=%d",
		signature, expected, len(p.cfg.SecretToken))
	log.Printf("[zoom] SIG DEBUG message=%q", message)
	// 🔺 DEBUG

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("%w: zoom signature mismatch", attendance.ErrInvalidSession)
	}

	return nil
}

// headerValue looks up a header by case-insensitive name. HTTP headers
// are case-insensitive but Go maps aren't, so we normalize on both
// sides.
func headerValue(headers map[string]string, name string) string {
	name = strings.ToLower(name)
	for k, v := range headers {
		if strings.ToLower(k) == name {
			return v
		}
	}
	return ""
}

// synthesizeEventID builds a stable dedup key from a Zoom webhook.
//
// Zoom does not provide a per-event ID for participant_joined/left
// events. The meeting UUID + participant UUID + event type is unique
// per delivery and stable across redeliveries.
func synthesizeEventID(meetingUUID, participantUUID, eventType string) string {
	h := sha256.Sum256([]byte(meetingUUID + ":" + participantUUID + ":" + eventType))
	return hex.EncodeToString(h[:])
}

// parseZoomTime handles Zoom's timestamp format, which is RFC 3339
// with a trailing "Z".
func parseZoomTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}