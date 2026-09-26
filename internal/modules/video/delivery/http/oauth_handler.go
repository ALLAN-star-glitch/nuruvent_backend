// internal/modules/video/delivery/http/oauth_handler.go

package http

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// OAuthHandler handles the connect and callback endpoints.
type OAuthHandler struct {
	svc service.Service
}

func NewOAuthHandler(svc service.Service) *OAuthHandler {
	return &OAuthHandler{svc: svc}
}

// Connect handles GET /oauth/:platform/connect.
//
// Default behaviour: 302 redirect to the platform's authorize URL.
// If the client sends Accept: application/json, return the URL in a
// JSON body so SPAs can open it in a popup.
func (h *OAuthHandler) Connect(c fiber.Ctx) error {
	platform := videodomain.Platform(c.Params("platform"))
	if !platform.IsValid() {
		return response.BadRequest(c, "Invalid platform", nil)
	}

	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var query BeginConnectRequest
	_ = c.Bind().Query(&query)

	result, err := h.svc.BeginConnect(c.Context(), service.BeginConnectCommand{
		UserID:    userID,
		Platform:  platform,
		ReturnURL: query.ReturnURL,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	if wantsJSON(c) {
		return response.Success(c, "Authorization URL generated", &ConnectResponse{
			AuthorizeURL: result.AuthorizeURL,
		})
	}
	return c.Redirect().To(result.AuthorizeURL)
}

// Callback handles GET /oauth/:platform/callback.
//
// Public endpoint — the platform's server redirects the user's
// browser here. Authentication is enforced by the single-use OAuth
// state, not by session. See FR-V-050.
func (h *OAuthHandler) Callback(c fiber.Ctx) error {
	state := c.Query("state")
	code := c.Query("code")
	platformErr := c.Query("error")

	result, err := h.svc.HandleCallback(c.Context(), service.CallbackCommand{
		State: state,
		Code:  code,
		Error: platformErr,
	})
	if err != nil {
		return mapDomainError(c, err)
	}

	target := result.ReturnURL
	if target == "" {
		target = "/"
	}
	return c.Redirect().To(target)
}

// wantsJSON reports whether the request asked for a JSON response.
func wantsJSON(c fiber.Ctx) bool {
	return strings.Contains(c.Get("Accept"), "application/json")
}