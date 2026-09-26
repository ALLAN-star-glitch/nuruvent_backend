
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
}

// IssueJoinTokenRequest is the body for POST /sessions/:id/join-tokens.
type IssueJoinTokenRequest struct {
	AttendeeID string `json:"attendee_id"`
	// Grace is expressed as a Go duration string ("24h", "30m").
	// Optional; zero means use the default.
	Grace string `json:"grace,omitempty"`
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
// The raw token is returned once and never again.
type IssueJoinTokenResponse struct {
	RawToken  string    `json:"raw_token"`
	ExpiresAt time.Time `json:"expires_at"`
	JoinURL   string    `json:"join_url"`
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