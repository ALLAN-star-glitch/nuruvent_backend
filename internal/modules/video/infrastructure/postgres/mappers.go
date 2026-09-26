// internal/modules/video/infrastructure/postgres/mappers.go

package postgres

import (
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// ============================================================
// CONNECTION
// ============================================================

func toConnectionModel(c *videodomain.Connection) *VideoConnectionModel {
	return &VideoConnectionModel{
		ID:                    c.ID,
		UserID:                c.UserID,
		Platform:              string(c.Platform),
		ExternalUserID:        c.ExternalUserID,
		ExternalEmail:         c.ExternalEmail,
		ExternalOrgID:         c.ExternalOrgID,
		AccessTokenEncrypted:  c.AccessToken,  // encrypted by repository layer before this
		RefreshTokenEncrypted: c.RefreshToken, // encrypted by repository layer before this
		TokenExpiresAt:        c.TokenExpiresAt,
		Scopes:                c.Scopes,
		ConnectedAt:           c.ConnectedAt,
		RevokedAt:             c.RevokedAt,
	}
}

func toConnectionDomain(m *VideoConnectionModel) *videodomain.Connection {
	return videodomain.HydrateConnection(
		m.ID,
		m.UserID,
		videodomain.Platform(m.Platform),
		m.ExternalUserID,
		m.ExternalEmail,
		m.ExternalOrgID,
		m.AccessTokenEncrypted,  // decrypted by repository layer before this
		m.RefreshTokenEncrypted, // decrypted by repository layer before this
		m.TokenExpiresAt,
		m.Scopes,
		m.ConnectedAt,
		m.RevokedAt,
	)
}

// ============================================================
// OAUTH STATE
// ============================================================

func toOAuthStateModel(s *videodomain.OAuthState) *VideoOAuthStateModel {
	return &VideoOAuthStateModel{
		State:      s.State,
		UserID:     s.UserID,
		Platform:   string(s.Platform),
		ReturnURL:  s.ReturnURL,
		CreatedAt:  s.CreatedAt,
		ExpiresAt:  s.ExpiresAt,
		ConsumedAt: s.ConsumedAt,
	}
}

func toOAuthStateDomain(m *VideoOAuthStateModel) *videodomain.OAuthState {
	return videodomain.HydrateOAuthState(
		m.State,
		m.UserID,
		videodomain.Platform(m.Platform),
		m.ReturnURL,
		m.CreatedAt,
		m.ExpiresAt,
		m.ConsumedAt,
	)
}

// ============================================================
// MEETING
// ============================================================

func toMeetingModel(m *videodomain.Meeting) *VideoMeetingModel {
	return &VideoMeetingModel{
		ID:          m.ID,
		UserID:      m.UserID,
		Platform:    string(m.Platform),
		ExternalID:  m.ExternalID,
		JoinURL:     m.JoinURL,
		StartURL:    m.StartURL,
		Password:    m.Password,
		Topic:       m.Topic,
		StartTime:   m.StartTime,
		DurationSec: int(m.Duration.Seconds()),
		Timezone:    m.Timezone,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toMeetingDomain(m *VideoMeetingModel) *videodomain.Meeting {
	return videodomain.HydrateMeeting(
		m.ID,
		m.UserID,
		videodomain.Platform(m.Platform),
		m.ExternalID,
		m.JoinURL,
		m.StartURL,
		m.Password,
		m.Topic,
		m.StartTime,
		durationFromSeconds(m.DurationSec),
		m.Timezone,
		m.CreatedAt,
		m.UpdatedAt,
	)
}