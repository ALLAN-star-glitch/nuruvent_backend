// internal/modules/video/infrastructure/providers/googlemeet/dto.go

package googlemeet

// ============================================================
// TOKEN ENDPOINT
// ============================================================

// tokenResponse is Google's token endpoint response.
//
// Returned by POST https://oauth2.googleapis.com/token for both
// the authorization_code and refresh_token grants.
//
// RefreshToken is present on the initial code exchange when
// access_type=offline was requested. On subsequent refreshes,
// Google may or may not rotate it — when it does, the new value
// appears here; when it does not, the field is absent.
//
// IDToken is a signed JWT containing user identity. It is only
// present when the "openid" scope was requested.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token,omitempty"`
}

// ============================================================
// USERINFO ENDPOINT
// ============================================================

// userInfoResponse is Google's OpenID Connect userinfo response.
//
// Returned by GET https://www.googleapis.com/oauth2/v3/userinfo.
// Used only as a fallback when the token response did not include
// an ID token.
//
// Sub is the stable, unique Google user identifier. Email is the
// user's primary email address. Both are requested via the
// "openid email" scopes.
type userInfoResponse struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

// ============================================================
// ID TOKEN CLAIMS
// ============================================================

// idTokenClaims is the subset of Google ID token claims Nuruvent
// uses.
//
// Google's ID token is a JWT: header.payload.signature. The
// payload is base64url-encoded JSON. The full claim set is large;
// only sub and email are needed.
type idTokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

// ============================================================
// MEET SPACES
// ============================================================

// spaceResponse is Google Meet's Space resource.
//
// Returned by POST /v2/spaces and GET /v2/spaces/{name}.
//
// Only the fields Nuruvent uses are declared. The full resource
// also includes config, activeConference, and telephony objects,
// which are not needed for meeting provisioning.
type spaceResponse struct {
	// Name is the resource name, e.g. "spaces/jQCFfuBOdN5z".
	// This is the stable identifier used in subsequent API calls.
	Name string `json:"name"`

	// MeetingURI is the join URL, e.g.
	// "https://meet.google.com/abc-mnop-xyz". Both hosts and
	// attendees use this same URL.
	MeetingURI string `json:"meetingUri"`

	// MeetingCode is the typeable alias, e.g. "abc-mnop-xyz".
	// A user can enter this code at meet.google.com to join.
	MeetingCode string `json:"meetingCode"`
}

// ============================================================
// ERROR RESPONSES
// ============================================================

// googleErrorObject is Google's structured error for API
// endpoints (Meet, userinfo).
//
// Shape:
//
//	{
//	  "error": {
//	    "code": 429,
//	    "message": "Quota exceeded",
//	    "status": "RESOURCE_EXHAUSTED"
//	  }
//	}
type googleErrorObject struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// googleErrorString is Google's flat error for the OAuth token
// endpoint.
//
// Shape:
//
//	{
//	  "error": "invalid_grant",
//	  "error_description": "Token has been expired or revoked."
//	}
type googleErrorString struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}