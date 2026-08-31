// internal/modules/events/service/service_impl.go

package service

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

// ============================================================
// SERVICE IMPLEMENTATION
// ============================================================

type eventService struct {
	repo        domain.Repository
	permChecker domain.PermissionChecker
	mediaSvc    domain.MediaService
	validator   *validation.Validator
}

func NewService(
	repo domain.Repository,
	permChecker domain.PermissionChecker,
	mediaSvc domain.MediaService,
) Service {
	return &eventService{
		repo:        repo,
		permChecker: permChecker,
		mediaSvc:    mediaSvc,
		validator:   validation.New(),
	}
}