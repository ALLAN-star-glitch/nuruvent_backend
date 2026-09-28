// internal/modules/video/delivery/http/meeting_sdk_handler.go

package http

import (
	"log"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// MeetingSDKHandler exposes the two endpoints the embedded Meeting SDK
// needs: one to issue a signed join token, and one to fetch the host's
// ZAK token.
type MeetingSDKHandler struct {
	svc service.Service
}

func NewMeetingSDKHandler(svc service.Service) *MeetingSDKHandler {
	return &MeetingSDKHandler{svc: svc}
}

// Signature handles POST /meetings/signature.
//
// Returns a signed JWT that authorizes the browser to join a specific
// meeting with a specific role. The frontend passes it to the Meeting
// SDK's join() call.
func (h *MeetingSDKHandler) Signature(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var req GenerateSignatureRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	platform := videodomain.Platform(req.Platform)
	if !platform.IsValid() {
		return response.BadRequest(c, "Invalid platform", nil)
	}

	result, err := h.svc.GenerateMeetingSignature(c.Context(), service.GenerateMeetingSignatureCommand{
		UserID:        userID,
		Platform:      platform,
		MeetingNumber: req.MeetingNumber,
		Role:          req.Role,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Signature generated", &SignatureResponse{
		Signature: result.Signature,
		SDKKey:    result.SDKKey,
	})
}

// ZAK handles GET /meetings/zak.
//
// Returns a Zoom Access Key token for the authenticated user's active
// connection. The frontend passes it to the Meeting SDK's join() call
// when the user is the host.
//
// Requires the user:read:zak scope on the OAuth app. If the scope is
// missing, Zoom returns 400 and the response is a platform rejection.
func (h *MeetingSDKHandler) ZAK(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	platform := videodomain.Platform(c.Query("platform"))
	if !platform.IsValid() {
		return response.BadRequest(c, "Invalid platform", nil)
	}

	result, err := h.svc.FetchMeetingZAK(c.Context(), service.FetchMeetingZAKCommand{
		UserID:   userID,
		Platform: platform,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "ZAK fetched", &ZAKResponse{
		ZAK: result.ZAK,
	})
}


// JoinInfo handles GET /meetings/:id/join-info?platform=zoom.
//
// Returns the credentials a browser needs to join the meeting via
// the embedded Meeting SDK. The :id path segment is the platform's
// external meeting ID (the numeric Zoom meeting number), matching
// what is stored on the schedule as video_meeting_id.
func (h *MeetingSDKHandler) JoinInfo(c fiber.Ctx) error {
	// --- Request arrival diagnostics -------------------------------
	// Logged before any validation so we can see exactly what the
	// proxy forwarded to Fiber, including the raw query string.
	log.Printf(
		"[join-info] hit method=%s path=%q originalURL=%q rawQuery=%q params=%v query=%v",
		c.Method(),
		c.Path(),
		c.OriginalURL(),
		string(c.Request().URI().QueryString()),
		c.Route().Params,
		c.Queries(),
	)
	// ---------------------------------------------------------------

	userID := handlerhelper.GetUserIDOptional(c)
	log.Printf("[join-info] userID=%q", userID)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	externalID := c.Params("id")
	log.Printf("[join-info] externalID=%q", externalID)
	if externalID == "" {
		return response.BadRequest(c, "Meeting ID is required", nil)
	}

	rawPlatform := c.Query("platform")
	platform := videodomain.Platform(rawPlatform)
	log.Printf(
		"[join-info] rawPlatform=%q parsedPlatform=%q isValid=%v",
		rawPlatform,
		string(platform),
		platform.IsValid(),
	)
	if !platform.IsValid() {
		return response.BadRequest(c, "Invalid platform", nil)
	}

	log.Printf(
		"[join-info] calling service userID=%q platform=%q externalID=%q",
		userID, platform, externalID,
	)

	info, err := h.svc.GetMeetingJoinInfo(c.Context(), service.GetMeetingJoinInfoCommand{
		UserID:     userID,
		Platform:   platform,
		ExternalID: externalID,
	})
	if err != nil {
		log.Printf("[join-info] service error: %T %v", err, err)
		return mapDomainError(c, err)
	}

	log.Printf(
    "[join-info] success meetingNumber=%q role=%d hasZAK=%v passwordLen=%d sdkKeyLen=%d sigLen=%d",
    info.MeetingNumber,
    info.Role,
    info.ZAK != "",
    len(info.Password),
    len(info.SDKKey),
    len(info.Signature),
)

	return response.Success(c, "Join info retrieved", &JoinInfoResponse{
		MeetingNumber: info.MeetingNumber,
		Signature:     info.Signature,
		SDKKey:        info.SDKKey,
		Password:      info.Password,
		WebEndpoint:   info.WebEndpoint,
		ZAK:           info.ZAK,
		Role:          info.Role,
	})
}














