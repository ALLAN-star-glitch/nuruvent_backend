// internal/modules/video/delivery/http/connection_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	videodomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/videodomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/video/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ConnectionHandler exposes host-facing connection management.
type ConnectionHandler struct {
	svc service.Service
}

func NewConnectionHandler(svc service.Service) *ConnectionHandler {
	return &ConnectionHandler{svc: svc}
}

// List handles GET /connections.
func (h *ConnectionHandler) List(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	conns, err := h.svc.ListConnections(c.Context(), userID)
	if err != nil {
		return mapDomainError(c, err)
	}

	out := make([]*ConnectionResponse, 0, len(conns))
	for _, conn := range conns {
		out = append(out, toConnectionResponse(conn))
	}

	return response.Success(c, "Connections retrieved successfully", &ListConnectionsResponse{
		Count:       len(out),
		Connections: out,
	})
}

// Disconnect handles POST /connections/:id/disconnect.
// Disconnect handles POST /connections/:platform/disconnect.
func (h *ConnectionHandler) Disconnect(c fiber.Ctx) error {
	userID := handlerhelper.GetUserIDOptional(c)
	if userID == "" {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	platform := videodomain.Platform(c.Params("platform"))
	if !platform.IsValid() {
		return response.BadRequest(c, "Invalid platform", nil)
	}

	if err := h.svc.Disconnect(c.Context(), service.DisconnectCommand{
		UserID:   userID,
		Platform: platform,
	}); err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Connection disconnected", nil)
}

// Compile-time guard against unused import drift.
var _ = videodomain.Platform("")