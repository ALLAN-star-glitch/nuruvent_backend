// internal/modules/events/service/service_impl.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

type eventService struct {
	repo        domain.Repository
	permChecker domain.PermissionChecker
	mediaSvc    domain.MediaService
	userInfo    domain.UserInfoProvider 
	validator   *validation.Validator
	organizer domain.OrganizerProvider
}

func NewService(
	repo domain.Repository,
	permChecker domain.PermissionChecker,
	userInfo domain.UserInfoProvider,
	mediaSvc domain.MediaService,
	organizerProvider domain.OrganizerProvider,
) Service {
	return &eventService{
		repo:        repo,
		permChecker: permChecker,
		mediaSvc:    mediaSvc,
		userInfo: userInfo,
		validator:   validation.New(),
		organizer: organizerProvider,
	}
}

// ============================================================
// SHARED HELPER FUNCTIONS
// ============================================================





func (s *eventService) GetCategories(ctx context.Context) ([]*domain.Category, error) {
	return s.repo.GetAllCategories(ctx)
}