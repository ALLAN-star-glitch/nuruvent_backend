// internal/modules/video/delivery/http/dto.go

package http

import "time"

// ============================================================
// REQUESTS
// ============================================================

// BeginConnectRequest is the query for GET /oauth/:platform/connect.
type BeginConnectRequest struct {
	// ReturnURL is where to send the user after a successful
	// connection. Optional.
	ReturnURL string `query:"return_url"`
}

// DisconnectRequest is the body for POST /connections/:id/disconnect.
// Currently empty — the platform is inferred from the connection ID.
// Kept as a struct so future fields (reason, revoke-at-platform flag)
// don't require a signature change.
type DisconnectRequest struct{}

// ============================================================
// RESPONSES
// ============================================================

// ConnectResponse is the response for GET /oauth/:platform/connect.
//
// The handler redirects (302) to AuthorizeURL by default. If the
// client sends Accept: application/json, this body is returned
// instead, so SPAs can open the URL in a popup.
type ConnectResponse struct {
	AuthorizeURL string `json:"authorize_url"`
}

// ConnectionResponse is the wire format for a connection.
//
// Deliberately excludes tokens — FR-V-023.
type ConnectionResponse struct {
	ID              string     `json:"id"`
	Platform        string     `json:"platform"`
	ExternalUserID  string     `json:"external_user_id"`
	ExternalEmail   string     `json:"external_email"`
	ExternalOrgID   string     `json:"external_org_id,omitempty"`
	Scopes          string     `json:"scopes,omitempty"`
	ConnectedAt     time.Time  `json:"connected_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	TokenExpiresAt  time.Time  `json:"token_expires_at"`
}

// ListConnectionsResponse is the response for GET /connections.
type ListConnectionsResponse struct {
	Count       int                   `json:"count"`
	Connections []*ConnectionResponse `json:"connections"`
}