package attendancedomain

// ============================================================
// SESSION STATUS
// ============================================================

type SessionStatus string

const (
	SessionStatusScheduled SessionStatus = "scheduled"
	SessionStatusLive      SessionStatus = "live"
	SessionStatusEnded     SessionStatus = "ended"
	SessionStatusCancelled SessionStatus = "cancelled"
)

func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusScheduled, SessionStatusLive, SessionStatusEnded, SessionStatusCancelled:
		return true
	}
	return false
}

func (s SessionStatus) IsFinal() bool {
	return s == SessionStatusEnded || s == SessionStatusCancelled
}

// ============================================================
// SESSION PROVIDER
// ============================================================

type SessionProvider string

const (
	ProviderZoom       SessionProvider = "zoom"
	ProviderGoogleMeet SessionProvider = "google_meet"
	ProviderInPerson   SessionProvider = "in_person"
	ProviderNone       SessionProvider = "none"
)

func (p SessionProvider) IsValid() bool {
	switch p {
	case ProviderZoom, ProviderGoogleMeet, ProviderInPerson, ProviderNone:
		return true
	}
	return false
}

// RequiresMeetingID reports whether the provider needs a meeting ID
// for webhook matching.
func (p SessionProvider) RequiresMeetingID() bool {
	return p == ProviderZoom || p == ProviderGoogleMeet
}

// ============================================================
// ATTENDANCE STATUS
// ============================================================

type AttendanceStatus string

const (
	StatusRegistered AttendanceStatus = "registered"
	StatusJoined     AttendanceStatus = "joined"
	StatusPartial    AttendanceStatus = "partial"
	StatusFull       AttendanceStatus = "full"
	StatusNoShow     AttendanceStatus = "no_show"
	StatusConfirmed  AttendanceStatus = "confirmed"
)

func (s AttendanceStatus) IsValid() bool {
	switch s {
	case StatusRegistered, StatusJoined, StatusPartial, StatusFull, StatusNoShow, StatusConfirmed:
		return true
	}
	return false
}

func (s AttendanceStatus) IsFinal() bool {
	switch s {
	case StatusPartial, StatusFull, StatusNoShow, StatusConfirmed:
		return true
	}
	return false
}

// ============================================================
// ATTENDANCE SOURCE
// ============================================================

type AttendanceSource string

const (
	SourceJoinLink    AttendanceSource = "join_link"
	SourceZoomWebhook AttendanceSource = "zoom_webhook"
	SourceGoogleEvent AttendanceSource = "google_event"
	SourceHostManual  AttendanceSource = "host_manual"
)

func (s AttendanceSource) IsValid() bool {
	switch s {
	case SourceJoinLink, SourceZoomWebhook, SourceGoogleEvent, SourceHostManual:
		return true
	}
	return false
}

// IsHighConfidence reports whether the source implies unambiguous
// attendee identity.
func (s AttendanceSource) IsHighConfidence() bool {
	return s == SourceJoinLink || s == SourceHostManual
}