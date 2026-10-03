// internal/modules/attendance/delivery/http/dto.go

package http

import "time"

// ============================================================
// CONSUMER-FACING REQUESTS
// ============================================================

// RegisterAttendeeRequest is the body for POST /attendees.
type RegisterAttendeeRequest struct {
	ExternalType string `json:"external_type"`
	ExternalID   string `json:"external_id"`
	DisplayName  string `json:"display_name"`
	Email        string `json:"email"`
}

// UpsertSessionRequest is the body for POST /sessions.
type UpsertSessionRequest struct {
	ExternalType      string    `json:"external_type"`
	ExternalID        string    `json:"external_id"`
	ProviderSessionID string    `json:"provider_session_id"`
	Title             string    `json:"title"`
	ScheduledStart    time.Time `json:"scheduled_start"`
	ScheduledEnd      time.Time `json:"scheduled_end"`
	Provider          string    `json:"provider"`
	ProviderMeetingID string    `json:"provider_meeting_id,omitempty"`
	ProviderURL       string    `json:"provider_url,omitempty"`
}

// RegisterAttendeeForExternalRequest is the body for
// POST /attendees/:id/register-for-external.
type RegisterAttendeeForExternalRequest struct {
	ExternalType string `json:"external_type"`
	ExternalID   string `json:"external_id"`

	// PublicBaseURL, when set, is prepended to the raw join token to
	// produce a full join URL. E.g. "https://nuruvent.com" →
	// "https://nuruvent.com/join/nrt_abc...".
	//
	// Leave empty when the caller only needs the raw token or will
	// build the URL itself.
	PublicBaseURL string `json:"public_base_url,omitempty"`

	// LinkGrace is how long after a session ends the issued tokens
	// remain valid, as a Go duration string ("24h", "30m"). Optional;
	// zero uses the service default.
	LinkGrace string `json:"link_grace,omitempty"`
}

// IssueJoinTokenRequest is the body for POST /sessions/:id/join-tokens.
type IssueJoinTokenRequest struct {
	AttendeeID string `json:"attendee_id"`
	// Grace is expressed as a Go duration string ("24h", "30m").
	// Optional; zero means use the default.
	Grace string `json:"grace,omitempty"`

	// PublicBaseURL, when set, is used to build a full join URL in
	// the response.
	PublicBaseURL string `json:"public_base_url,omitempty"`
}

// RevokeJoinTokensRequest is the body for
// DELETE /sessions/:id/join-tokens.
type RevokeJoinTokensRequest struct {
	AttendeeID string `json:"attendee_id"`
}

// ============================================================
// HOST OPERATIONS
// ============================================================

// ConfirmAttendanceRequest is the body for
// POST /sessions/:id/attendance/:attendeeID/confirm.
type ConfirmAttendanceRequest struct {
	Reason string `json:"reason,omitempty"`
}

// OverrideAttendanceRequest is the body for
// POST /sessions/:id/attendance/:attendeeID/override.
type OverrideAttendanceRequest struct {
	NewStatus string `json:"new_status"`
	Reason    string `json:"reason,omitempty"`
}

// BulkConfirmRequest is the body for
// POST /sessions/:id/attendance/bulk-confirm.
type BulkConfirmRequest struct {
	AttendeeIDs []string `json:"attendee_ids"`
	Reason      string   `json:"reason,omitempty"`
}

// ============================================================
// RESPONSES
// ============================================================

// AttendeeResponse is the wire format for an attendee.
type AttendeeResponse struct {
	ID           string    `json:"id"`
	ExternalType string    `json:"external_type"`
	ExternalID   string    `json:"external_id"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SessionResponse is the wire format for a session.
type SessionResponse struct {
	ID                string    `json:"id"`
	ExternalType      string    `json:"external_type"`
	ExternalID        string    `json:"external_id"`
	ProviderSessionID string    `json:"provider_session_id"`
	Title             string    `json:"title"`
	ScheduledStart    time.Time `json:"scheduled_start"`
	ScheduledEnd      time.Time `json:"scheduled_end"`
	DurationMinutes   int       `json:"duration_minutes"`
	Provider          string    `json:"provider"`
	ProviderMeetingID string    `json:"provider_meeting_id,omitempty"`
	ProviderURL       string    `json:"provider_url,omitempty"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// IssueJoinTokenResponse is the wire format for a newly-issued token.
//
// The raw token is returned once and never again. JoinURL is set only
// when the caller supplied a public_base_url.
type IssueJoinTokenResponse struct {
	RawToken  string    `json:"raw_token"`
	ExpiresAt time.Time `json:"expires_at"`
	JoinURL   string    `json:"join_url,omitempty"`
}

// AttendeeSessionLinkResponse is one personalized join URL for one
// (attendee, session) pair. Returned from
// POST /attendees/:id/register-for-external.
//
// JoinURL is populated only when the caller supplied a
// public_base_url. MeetingCode + Platform are always populated so
// callers can build their own URL if needed.
type AttendeeSessionLinkResponse struct {
	SessionID   string    `json:"session_id"`
	MeetingCode string    `json:"meeting_code,omitempty"`
	Platform    string    `json:"platform,omitempty"`
	JoinURL     string    `json:"join_url,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// RegisterAttendeeForExternalResponse is the body of
// POST /attendees/:id/register-for-external.
type RegisterAttendeeForExternalResponse struct {
	AttendeeID string                          `json:"attendee_id"`
	Links      []AttendeeSessionLinkResponse   `json:"links"`
}

// RedeemJoinTokenResponse is the response for GET /join/:token.
type RedeemJoinTokenResponse struct {
	AttendeeID string `json:"attendee_id"`
	SessionID  string `json:"session_id"`
	RedirectTo string `json:"redirect_to"`
}

// SessionStatusResponse is the wire format for a per-session status.
type SessionStatusResponse struct {
	AttendeeID           string     `json:"attendee_id"`
	SessionID            string     `json:"session_id"`
	DerivedStatus        string     `json:"derived_status"`
	EffectiveStatus      string     `json:"effective_status"`
	TotalDurationSeconds int64      `json:"total_duration_seconds"`
	HostConfirmed        bool       `json:"host_confirmed"`
	ConfirmedStatus      string     `json:"confirmed_status,omitempty"`
	ConfirmedBy          string     `json:"confirmed_by,omitempty"`
	ConfirmedAt          *time.Time `json:"confirmed_at,omitempty"`
	ConfirmReason        string     `json:"confirm_reason,omitempty"`
	CertEligible         bool       `json:"cert_eligible"`
	LastDerivedAt        time.Time  `json:"last_derived_at"`
	DisplayName          string     `json:"display_name,omitempty"`
	Email                string     `json:"email,omitempty"`
	Phone				string     `json:"phone,omitempty"`
	IsHost               bool       `json:"is_host"`
}



// AttendeeSummaryResponse is the wire format for an attendee's
// dashboard summary.
type AttendeeSummaryResponse struct {
	Attendee *AttendeeResponse        `json:"attendee"`
	Statuses []*SessionStatusResponse `json:"statuses"`
	Rollups  []*RollupStatusResponse  `json:"rollups"`
}

// RollupStatusResponse is the wire format for a roll-up.
type RollupStatusResponse struct {
	AttendeeID           string    `json:"attendee_id"`
	ExternalType         string    `json:"external_type"`
	ExternalID           string    `json:"external_id"`
	DerivedStatus        string    `json:"derived_status"`
	SessionsTotal        int       `json:"sessions_total"`
	SessionsAttended     int       `json:"sessions_attended"`
	SessionsConfirmed    int       `json:"sessions_confirmed"`
	TotalDurationSeconds int64     `json:"total_duration_seconds"`
	LastDerivedAt        time.Time `json:"last_derived_at"`
}

// ============================================================
// EVENT ATTENDANCE SUMMARY
// ============================================================

// SessionAttendanceSummaryResponse is the per-session aggregate in
// the event summary response.
type SessionAttendanceSummaryResponse struct {
	SessionID          string `json:"session_id"`
	Title              string `json:"title"`
	Provider           string `json:"provider"`
	ScheduledStart     string `json:"scheduled_start"`
	ScheduledEnd       string `json:"scheduled_end"`
	RegisteredCount    int    `json:"registered_count"`
	AttendedCount      int    `json:"attended_count"`
	AvgDurationSeconds int64  `json:"avg_duration_seconds"`
	HasAttendanceData  bool   `json:"has_attendance_data"`
	VideoMeetingID     string `json:"video_meeting_id,omitempty"`
}

// EventAttendanceTotalsResponse is the cross-session aggregate.
type EventAttendanceTotalsResponse struct {
	TotalSessions          int `json:"total_sessions"`
	SessionsWithAttendance int `json:"sessions_with_attendance"`
	UniqueAttendees        int `json:"unique_attendees"`
	TotalAttendanceEvents  int `json:"total_attendance_events"`
}

// EventAttendanceSummaryResponse is what GET /events/:eventId/summary
// returns.
type EventAttendanceSummaryResponse struct {
	Sessions []SessionAttendanceSummaryResponse `json:"sessions"`
	Totals   EventAttendanceTotalsResponse      `json:"totals"`
}


// ============================================================
// EVENT ATTENDEE DIRECTORY — response DTOs
// ============================================================

type EventAttendeesListResponse struct {
	Attendees []EventAttendeeResponse `json:"attendees"`
	Total     int                     `json:"total"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
}

type EventAttendeeResponse struct {
	AttendeeID           string `json:"attendee_id"`
	DisplayName          string `json:"display_name"`
	Email                string `json:"email"`
	EffectiveStatus      string `json:"effective_status"`
	SessionsTotal        int    `json:"sessions_total"`
	SessionsAttended     int    `json:"sessions_attended"`
	SessionsConfirmed    int    `json:"sessions_confirmed"`
	TotalDurationSeconds int64  `json:"total_duration_seconds"`
	RegisteredAt         string `json:"registered_at"`
	LastActivityAt       string `json:"last_activity_at"`
	IsHost               bool   `json:"is_host"`   // ← add
	Phone                string `json:"phone"`     // ← add	
}


type EventAttendeeDetailResponse struct {
	EventAttendeeResponse
	Sessions []EventAttendeeSessionResponse `json:"sessions"`
}

type EventAttendeeSessionResponse struct {
	SessionID        string `json:"session_id"`
	Title            string `json:"title"`
	Provider         string `json:"provider"`
	ScheduledStart   string `json:"scheduled_start"`
	ScheduledEnd     string `json:"scheduled_end"`
	DerivedStatus    string `json:"derived_status"`
	HostConfirmed    bool   `json:"host_confirmed"`
	TotalDurationSec int64  `json:"total_duration_seconds"`
	LastDerivedAt    string `json:"last_derived_at"`
}

type CrossEventAttendeeResponse struct {
	AttendeeID           string `json:"attendee_id"`
	DisplayName          string `json:"display_name"`
	Email                string `json:"email"`
	EventID              string `json:"event_id"`
	EventName            string `json:"event_name"`
	EventSlug            string `json:"event_slug"`
	EventStartDate       string `json:"event_start_date"`
	EffectiveStatus      string `json:"effective_status"`
	SessionsTotal        int    `json:"sessions_total"`
	SessionsAttended     int    `json:"sessions_attended"`
	SessionsConfirmed    int    `json:"sessions_confirmed"`
	TotalDurationSeconds int64  `json:"total_duration_seconds"`
	RegisteredAt         string `json:"registered_at"`
	LastActivityAt       string `json:"last_activity_at"`
	IsHost               bool   `json:"is_host"`   // ← add
	Phone				string `json:"phone"`     // ← add
}