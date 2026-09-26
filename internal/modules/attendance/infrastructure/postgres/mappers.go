package postgres

import (
	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

// ============================================================
// ATTENDEE
// ============================================================

func toAttendeeModel(a *attendance.Attendee) *AttendeeModel {
	return &AttendeeModel{
		ID:           a.ID,
		ExternalType: a.External.Type,
		ExternalID:   a.External.ID,
		DisplayName:  a.DisplayName,
		Email:        a.Email,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func toAttendeeDomain(m *AttendeeModel) *attendance.Attendee {
	return attendance.HydrateAttendee(
		m.ID,
		attendance.ExternalRef{Type: m.ExternalType, ID: m.ExternalID},
		m.DisplayName,
		m.Email,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

// ============================================================
// SESSION
// ============================================================

func toSessionModel(s *attendance.Session) *SessionModel {
	return &SessionModel{
		ID:                s.ID,
		ExternalType:      s.External.Type,
		ExternalID:        s.External.ID,
		ProviderSessionID: s.ProviderSessionID,
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

func toSessionDomain(m *SessionModel) *attendance.Session {
	return attendance.HydrateSession(
		m.ID,
		attendance.ExternalRef{Type: m.ExternalType, ID: m.ExternalID},
		m.ProviderSessionID,
		m.Title,
		m.ScheduledStart,
		m.ScheduledEnd,
		m.DurationMinutes,
		attendance.SessionProvider(m.Provider),
		m.ProviderMeetingID,
		m.ProviderURL,
		attendance.SessionStatus(m.Status),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

// ============================================================
// JOIN TOKEN
// ============================================================

func toJoinTokenModel(t *attendance.JoinToken) *JoinTokenModel {
	return &JoinTokenModel{
		ID:         t.ID,
		AttendeeID: t.AttendeeID,
		SessionID:  t.SessionID,
		TokenHash:  t.TokenHash,
		IssuedAt:   t.IssuedAt,
		ExpiresAt:  t.ExpiresAt,
		RevokedAt:  t.RevokedAt,
	}
}

func toJoinTokenDomain(m *JoinTokenModel) *attendance.JoinToken {
	return attendance.HydrateJoinToken(
		m.ID,
		m.AttendeeID,
		m.SessionID,
		m.TokenHash,
		m.IssuedAt,
		m.ExpiresAt,
		m.RevokedAt,
	)
}

// ============================================================
// ATTENDANCE RECORD
// ============================================================

func toAttendanceRecordModel(r *attendance.AttendanceRecord) *AttendanceRecordModel {
	return &AttendanceRecordModel{
		ID:              r.ID,
		AttendeeID:      r.AttendeeID,
		SessionID:       r.SessionID,
		JoinTime:        r.JoinTime,
		LeaveTime:       r.LeaveTime,
		DurationSeconds: r.DurationSeconds,
		Source:          string(r.Source),
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func toAttendanceRecordDomain(m *AttendanceRecordModel) *attendance.AttendanceRecord {
	return attendance.HydrateAttendanceRecord(
		m.ID,
		m.AttendeeID,
		m.SessionID,
		m.JoinTime,
		m.LeaveTime,
		m.DurationSeconds,
		attendance.AttendanceSource(m.Source),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

// ============================================================
// ATTENDEE SESSION STATUS
// ============================================================

func toSessionStatusModel(s *attendance.AttendeeSessionStatus) *AttendeeSessionStatusModel {
	return &AttendeeSessionStatusModel{
		AttendeeID:           s.AttendeeID,
		SessionID:            s.SessionID,
		DerivedStatus:        string(s.DerivedStatus),
		TotalDurationSeconds: s.TotalDurationSeconds,
		HostConfirmed:        s.HostConfirmed,
		ConfirmedStatus:      nullableString(string(s.ConfirmedStatus)),
		ConfirmedBy:          nullableString(s.ConfirmedBy),
		ConfirmedAt:          s.ConfirmedAt,
		ConfirmReason:        s.ConfirmReason,
		LastDerivedAt:        s.LastDerivedAt,
	}
}

func toSessionStatusDomain(m *AttendeeSessionStatusModel) *attendance.AttendeeSessionStatus {
	return &attendance.AttendeeSessionStatus{
		AttendeeID:           m.AttendeeID,
		SessionID:            m.SessionID,
		DerivedStatus:        attendance.AttendanceStatus(m.DerivedStatus),
		TotalDurationSeconds: m.TotalDurationSeconds,
		HostConfirmed:        m.HostConfirmed,
		ConfirmedStatus:      attendance.AttendanceStatus(derefString(m.ConfirmedStatus)),
		ConfirmedBy:          derefString(m.ConfirmedBy),
		ConfirmedAt:          m.ConfirmedAt,
		ConfirmReason:        m.ConfirmReason,
		LastDerivedAt:        m.LastDerivedAt,
	}
}

// ============================================================
// ATTENDEE ROLLUP STATUS
// ============================================================

func toRollupStatusModel(s *attendance.AttendeeRollupStatus) *AttendeeRollupStatusModel {
	return &AttendeeRollupStatusModel{
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

func toRollupStatusDomain(m *AttendeeRollupStatusModel) *attendance.AttendeeRollupStatus {
	return &attendance.AttendeeRollupStatus{
		AttendeeID: m.AttendeeID,
		External: attendance.ExternalRef{
			Type: m.ExternalType,
			ID:   m.ExternalID,
		},
		DerivedStatus:        attendance.AttendanceStatus(m.DerivedStatus),
		SessionsTotal:        m.SessionsTotal,
		SessionsAttended:     m.SessionsAttended,
		SessionsConfirmed:    m.SessionsConfirmed,
		TotalDurationSeconds: m.TotalDurationSeconds,
		LastDerivedAt:        m.LastDerivedAt,
	}
}

// ============================================================
// ATTENDANCE OVERRIDE
// ============================================================

func toOverrideModel(o *attendance.AttendanceOverride) *AttendanceOverrideModel {
	return &AttendanceOverrideModel{
		ID:          o.ID,
		AttendeeID:  o.AttendeeID,
		SessionID:   o.SessionID,
		ActorID:     o.ActorID,
		PriorStatus: string(o.PriorStatus),
		NewStatus:   string(o.NewStatus),
		Reason:      o.Reason,
		CreatedAt:   o.CreatedAt,
	}
}

func toOverrideDomain(m *AttendanceOverrideModel) *attendance.AttendanceOverride {
	return attendance.HydrateAttendanceOverride(
		m.ID,
		m.AttendeeID,
		m.SessionID,
		m.ActorID,
		attendance.AttendanceStatus(m.PriorStatus),
		attendance.AttendanceStatus(m.NewStatus),
		m.Reason,
		m.CreatedAt,
	)
}