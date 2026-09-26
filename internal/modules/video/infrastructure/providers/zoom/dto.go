// internal/modules/video/infrastructure/providers/zoom/dto.go

package zoom

// ============================================================
// OAUTH — TOKEN EXCHANGE
// ============================================================

// tokenResponse is the response from /oauth/token (both the
// authorization_code grant and the refresh_token grant).
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	Scope        string `json:"scope"`
}

// ============================================================
// USER INFO
// ============================================================

// userResponse is the response from GET /users/me.
type userResponse struct {
	ID         string `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	AccountID  string `json:"account_id"`
	Type       int    `json:"type"` // 1=basic, 2=licensed, ...
	Verified   int    `json:"verified"`
	Timezone   string `json:"timezone"`
	Language   string `json:"language"`
	PMPI       int    `json:"pmi"`
	CreatedAt  string `json:"created_at"`
}

// ============================================================
// MEETINGS — CREATE
// ============================================================

// createMeetingRequest is the body for POST /users/me/meetings.
type createMeetingRequest struct {
	Topic      string               `json:"topic"`
	Type       int                  `json:"type"`       // 2 = scheduled
	StartTime  string               `json:"start_time"` // ISO 8601 UTC
	Duration   int                  `json:"duration"`   // minutes
	Timezone   string               `json:"timezone"`
	Agenda     string               `json:"agenda,omitempty"`
	Password   string               `json:"password,omitempty"`
	Settings   meetingSettings      `json:"settings"`
}

// meetingSettings controls the meeting's runtime behavior.
//
// Defaults chosen for training sessions:
//   - join_before_host: attendees can enter before the host
//   - waiting_room: off, so attendees don't wait
//   - approval_type: 2 = no registration required (Nuruvent handles it)
//   - audio: both telephone and computer audio available
//   - auto_recording: none; hosts can enable per meeting if they want
type meetingSettings struct {
	JoinBeforeHost  bool   `json:"join_before_host"`
	WaitingRoom     bool   `json:"waiting_room"`
	ApprovalType    int    `json:"approval_type"`
	Audio           string `json:"audio"`
	AutoRecording   string `json:"auto_recording"`
	MuteUponEntry   bool   `json:"mute_upon_entry"`
	HostVideo       bool   `json:"host_video"`
	ParticipantVideo bool  `json:"participant_video"`
	UsePMI          bool   `json:"use_pmi"`
}

// ============================================================
// MEETINGS — RESPONSE
// ============================================================

// meetingResponse is the response from POST /users/me/meetings and
// from GET /meetings/{id}.
type meetingResponse struct {
	ID       int64  `json:"id"`        // numeric meeting ID
	UUID     string `json:"uuid"`      // per-instance UUID (changes on each start)
	Topic    string `json:"topic"`
	Type     int    `json:"type"`
	HostID   string `json:"host_id"`
	HostEmail string `json:"host_email"`

	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	Timezone  string `json:"timezone"`
	Agenda    string `json:"agenda"`

	JoinURL   string `json:"join_url"`
	StartURL  string `json:"start_url"` // host-only; absent on some responses
	Password  string `json:"password"`
	H323Pass  string `json:"h323_password"`
	PSTNPass  string `json:"pstn_password"`

	Settings meetingSettings `json:"settings"`
}

// ============================================================
// ERRORS
// ============================================================

// zoomError is Zoom's standard error envelope.
//
// Returned by every Zoom endpoint when the request is rejected.
type zoomError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}