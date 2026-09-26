// internal/modules/attendance/delivery/http/mappers.go

package http

import (
	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
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