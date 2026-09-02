// internal/modules/events/service/service_impl.go

package service

import (


	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
)

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

// ============================================================
// SHARED HELPER FUNCTIONS
// ============================================================

// getScopeFromEvent creates a Scope from an event
func (s *eventService) getScopeFromEvent(event *domain.Event) domain.Scope {
	if event.InstitutionID != nil && *event.InstitutionID != "" {
		return domain.NewInstitutionTeamScope(*event.InstitutionID)
	}
	return domain.NewPersonalTeamScope(event.CreatedBy)
}

// isEmptyTeam checks if no team filter is applied
func (s *eventService) isEmptyTeam(team domain.TeamFilter) bool {
	return team.ID == "" || team.Type == ""
}
