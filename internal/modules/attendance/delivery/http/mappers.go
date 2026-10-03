// internal/modules/attendance/delivery/http/mappers.go

package http

import (
	"time"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
)

func toAttendeeResponse(a *attendance.Attendee) *AttendeeResponse {
	return &AttendeeResponse{
		ID:           a.ID,
		ExternalType: a.External.Type,
		ExternalID:   a.External.ID,
		DisplayName:  a.DisplayName,
		Email:        a.Email,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func toSessionResponse(s *attendance.Session) *SessionResponse {
	return &SessionResponse{
		ID:                s.ID,
		ExternalType:      s.External.Type,
		ExternalID:        s.External.ID,
		Title:             s.Title,
		ScheduledStart:    s.ScheduledStart,
		ScheduledEnd:      s.ScheduledEnd,
		DurationMinutes:   s.DurationMinutes,
		Provider:          string(s.Provider),
		ProviderMeetingID: s.ProviderMeetingID,
		ProviderURL:       s.ProviderURL,
		Status:            string(s.Status),
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

func toSessionStatusResponse(s *attendance.AttendeeSessionStatus) *SessionStatusResponse {
	return &SessionStatusResponse{
		AttendeeID:           s.AttendeeID,
		SessionID:            s.SessionID,
		DerivedStatus:        string(s.DerivedStatus),
		EffectiveStatus:      string(s.EffectiveStatus()),
		TotalDurationSeconds: s.TotalDurationSeconds,
		HostConfirmed:        s.HostConfirmed,
		ConfirmedStatus:      string(s.ConfirmedStatus),
		ConfirmedBy:          s.ConfirmedBy,
		ConfirmedAt:          s.ConfirmedAt,
		ConfirmReason:        s.ConfirmReason,
		CertEligible:         s.CertEligible(),
		LastDerivedAt:        s.LastDerivedAt,
		DisplayName:          s.DisplayName,
		Email:                s.Email,
		Phone:                s.Phone,          // ← add
		IsHost:               s.IsHost,
	}
}

func toRollupStatusResponse(s *attendance.AttendeeRollupStatus) *RollupStatusResponse {
	return &RollupStatusResponse{
		AttendeeID:           s.AttendeeID,
		ExternalType:         s.External.Type,
		ExternalID:           s.External.ID,
		DerivedStatus:        string(s.DerivedStatus),
		SessionsTotal:        s.SessionsTotal,
		SessionsAttended:     s.SessionsAttended,
		SessionsConfirmed:    s.SessionsConfirmed,
		TotalDurationSeconds: s.TotalDurationSeconds,
		LastDerivedAt:        s.LastDerivedAt,
	}
}

// ============================================================
// JOIN LINKS
// ============================================================

// toAttendeeSessionLinkResponse converts a service-level
// AttendeeSessionLink to its wire format.
func toAttendeeSessionLinkResponse(l service.AttendeeSessionLink) AttendeeSessionLinkResponse {
	return AttendeeSessionLinkResponse{
		SessionID:   l.SessionID,
		MeetingCode: l.MeetingCode,
		Platform:    string(l.Platform),
		JoinURL:     l.JoinURL,
		ExpiresAt:   l.ExpiresAt,
	}
}

// toAttendeeSessionLinksResponse converts a slice of
// AttendeeSessionLink. Returns a non-nil empty slice for nil input so
// the JSON is `[]` rather than `null`.
func toAttendeeSessionLinksResponse(
	links []service.AttendeeSessionLink,
) []AttendeeSessionLinkResponse {
	out := make([]AttendeeSessionLinkResponse, 0, len(links))
	for _, l := range links {
		out = append(out, toAttendeeSessionLinkResponse(l))
	}
	return out
}

// ============================================================
// EVENT ATTENDEE DIRECTORY
// ============================================================

func toEventAttendeeResponse(a *service.EventAttendeeListItem) EventAttendeeResponse {
	return EventAttendeeResponse{
		AttendeeID:           a.AttendeeID,
		DisplayName:          a.DisplayName,
		Email:                a.Email,
		EffectiveStatus:      string(a.EffectiveStatus),
		SessionsTotal:        a.SessionsTotal,
		SessionsAttended:     a.SessionsAttended,
		SessionsConfirmed:    a.SessionsConfirmed,
		TotalDurationSeconds: int64(a.TotalDurationSeconds),
		RegisteredAt:         a.RegisteredAt.Format(time.RFC3339),
		LastActivityAt:       a.LastActivityAt.Format(time.RFC3339),
		IsHost:               a.IsHost,
		Phone:                a.Phone,
	}
}

func toEventAttendeeDetailResponse(d *service.EventAttendeeDetail) EventAttendeeDetailResponse {
	out := EventAttendeeDetailResponse{
		EventAttendeeResponse: EventAttendeeResponse{
			AttendeeID:           d.AttendeeID,
			DisplayName:          d.DisplayName,
			Email:                d.Email,
			EffectiveStatus:      string(d.EffectiveStatus),
			SessionsTotal:        d.SessionsTotal,
			SessionsAttended:     d.SessionsAttended,
			SessionsConfirmed:    d.SessionsConfirmed,
			TotalDurationSeconds: int64(d.TotalDurationSeconds),
			RegisteredAt:         d.RegisteredAt.Format(time.RFC3339),
			LastActivityAt:       d.LastActivityAt.Format(time.RFC3339),
			IsHost:               d.IsHost,
			Phone: 			  d.Phone,
		},
		Sessions: make([]EventAttendeeSessionResponse, 0, len(d.Sessions)),
	}
	for _, s := range d.Sessions {
		out.Sessions = append(out.Sessions, EventAttendeeSessionResponse{
			SessionID:        s.SessionID,
			Title:            s.Title,
			Provider:         string(s.Provider),
			ScheduledStart:   s.ScheduledStart.Format(time.RFC3339),
			ScheduledEnd:     s.ScheduledEnd.Format(time.RFC3339),
			DerivedStatus:    string(s.DerivedStatus),
			HostConfirmed:    s.HostConfirmed,
			TotalDurationSec: int64(s.TotalDurationSec),
			LastDerivedAt:    s.LastDerivedAt.Format(time.RFC3339),
		})
	}
	return out
}

// ============================================================
// CROSS-EVENT ATTENDEE DIRECTORY
// ============================================================

func toCrossEventAttendeeResponse(a *service.CrossEventAttendeeItem) CrossEventAttendeeResponse {
	return CrossEventAttendeeResponse{
		AttendeeID:           a.AttendeeID,
		DisplayName:          a.DisplayName,
		Email:                a.Email,
		EventID:              a.EventID,
		EventName:            a.EventName,
		EventSlug:            a.EventSlug,
		EventStartDate:       a.EventStartDate.Format(time.RFC3339),
		EffectiveStatus:      string(a.EffectiveStatus),
		SessionsTotal:        a.SessionsTotal,
		SessionsAttended:     a.SessionsAttended,
		SessionsConfirmed:    a.SessionsConfirmed,
		TotalDurationSeconds: a.TotalDurationSeconds,
		RegisteredAt:         a.RegisteredAt.Format(time.RFC3339),
		LastActivityAt:       a.LastActivityAt.Format(time.RFC3339),
		IsHost:               a.IsHost,
		Phone:                a.Phone,
	}
}