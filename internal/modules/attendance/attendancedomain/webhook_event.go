package attendancedomain

import "time"

// WebhookEvent is the normalized form of a provider webhook.
// Providers send wildly different payloads; each provider adapter
// translates its own into this shape.
type WebhookEvent struct {
	// ProviderName identifies the source provider.
	ProviderName SessionProvider

	// ProviderEventID is a stable ID from the provider for
	// deduplication. The service rejects events it has already seen.
	ProviderEventID string

	// ProviderMeetingID identifies the session.
	ProviderMeetingID string

	// EventType is "joined" or "left".
	EventType WebhookEventType

	// ParticipantEmail and ParticipantName are used to match the
	// event to an attendee when the token isn't available.
	ParticipantEmail string
	ParticipantName  string

	// OccurredAt is when the event happened at the provider.
	OccurredAt time.Time

	// Raw is the provider's raw payload, stored for audit.
	Raw map[string]any
}

type WebhookEventType string

const (
	WebhookEventJoined WebhookEventType = "joined"
	WebhookEventLeft   WebhookEventType = "left"
)

func (t WebhookEventType) IsValid() bool {
	return t == WebhookEventJoined || t == WebhookEventLeft
}