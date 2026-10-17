package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// ListSystemTemplates returns all system (platform-owned) templates.
//
// Authorization is enforced by the HTTP layer (platform-domain
// authzMiddleware). Defense-in-depth: the service also checks.
func (s *serviceImpl) ListSystemTemplates(ctx context.Context) ([]*domain.Template, error) {
	return s.deps.Templates.FindAllSystem(ctx)
}