package attendancedomain

import "time"

// WebhookEvent is the normalized form of a provider webhook or poll
// result. Providers send wildly different payloads; each provider
// adapter translates its own into this shape.
type WebhookEvent struct {
	// ProviderName identifies the source provider.
	ProviderName SessionProvider

	// ProviderEventID is a stable ID from the provider for
	// deduplication. The service rejects events it has already seen.
	ProviderEventID string

	// ProviderMeetingID identifies the session.
	ProviderMeetingID string

	// EventType is "joined", "left", or "ended".
	EventType WebhookEventType

	// ParticipantEmail and ParticipantName are used to match the
	// event to an attendee when the token isn't available.
	//
	// ParticipantEmail may be empty when the source platform does
	// not expose an email (e.g. Google Meet).
	ParticipantEmail string
	ParticipantName  string




		// ParticipantCustomerKey is the value passed to the Meeting SDK
	// as `customerKey` on join. Zoom echoes it back in webhooks as
	// `participant.customer_key`. Nuruvent sets this to the user's
	// username so the webhook can match the participant to a
	// registered attendee without exposing the username in the
	// visible display name.
	//
	// Empty when the participant joined via a plain link or when the
	// provider doesn't support this field.
	ParticipantCustomerKey string

	// ExternalUserID is the provider's user identifier when the
	// platform exposes one but not an email. Google Meet uses the
	// Google user resource name ("users/123456"). Empty for
	// platforms that supply email (Zoom) or for anonymous
	// participants.
	ExternalUserID string

	// OccurredAt is when the event happened at the provider.
	OccurredAt time.Time

	// Raw is the provider's raw payload, stored for audit.
	Raw map[string]any
}

type WebhookEventType string

const (
	WebhookEventJoined  WebhookEventType = "joined"
	WebhookEventLeft    WebhookEventType = "left"
	WebhookEventStarted WebhookEventType = "started"
	WebhookEventEnded   WebhookEventType = "ended"
)

func (t WebhookEventType) IsValid() bool {
	return t == WebhookEventJoined ||
		t == WebhookEventLeft ||
		t == WebhookEventStarted ||
		t == WebhookEventEnded
}