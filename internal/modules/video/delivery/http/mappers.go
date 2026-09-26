// internal/modules/video/delivery/http/mappers.go

package http

import (
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
)

// toConnectionResponse maps a domain connection to the wire format.
//
// Never emits AccessToken or RefreshToken — FR-V-023.
func toConnectionResponse(c *videodomain.Connection) *ConnectionResponse {
	if c == nil {
		return nil
	}
	return &ConnectionResponse{
		ID:             c.ID,
		Platform:       string(c.Platform),
		ExternalUserID: c.ExternalUserID,
		ExternalEmail:  c.ExternalEmail,
		ExternalOrgID:  c.ExternalOrgID,
		Scopes:         c.Scopes,
		ConnectedAt:    c.ConnectedAt,
		RevokedAt:      c.RevokedAt,
		IsActive:       c.RevokedAt == nil,
		TokenExpiresAt: c.TokenExpiresAt,
	}
}