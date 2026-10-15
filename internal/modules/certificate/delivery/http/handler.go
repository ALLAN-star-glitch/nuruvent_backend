// internal/modules/certificate/delivery/http/handler.go

package http

import (
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
)

// Handler is the base struct shared by all certificate HTTP handlers.
type Handler struct {
	svc certificateService.Service
}

// NewHandler constructs a Handler.
func NewHandler(svc certificateService.Service) *Handler {
	return &Handler{svc: svc}
}