// internal/modules/attendance/infrastructure/providers/zoom/dto.go

package zoom

// Zoom event names we handle.
const (
	EventEndpointURLValidation = "endpoint.url_validation"
	EventParticipantJoined     = "meeting.participant_joined"
	EventParticipantLeft       = "meeting.participant_left"
)

// webhookEnvelope is the top-level shape of every Zoom webhook
// delivery.
type webhookEnvelope struct {
	Event   string        `json:"event"`
	EventTS int64         `json:"event_ts"`
	Payload webhookPayload `json:"payload"`
}

type webhookPayload struct {
	// URL validation payload uses PlainToken.
	PlainToken string `json:"plainToken,omitempty"`

	// Participant events use Object.
	Object webhookObject `json:"object"`
}

type webhookObject struct {
	ID   string `json:"id"`   // numeric meeting ID
	UUID string `json:"uuid"` // meeting UUID (unique per meeting instance)

	Participant webhookParticipant `json:"participant"`
}

type webhookParticipant struct {
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	Email           string `json:"email"`
	ParticipantUUID string `json:"participant_uuid"`
	JoinTime        string `json:"join_time"`
	LeaveTime       string `json:"leave_time"`
}