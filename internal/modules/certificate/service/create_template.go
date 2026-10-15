// internal/modules/certificate/service/create_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// CreateTemplate implements Service.CreateTemplate.
func (s *serviceImpl) CreateTemplate(
	ctx context.Context,
	userID string,
	in CreateTemplateInput,
) (*domain.Template, error) {
	ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, in.AccountID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	now := s.deps.Clock.Now().UTC()
	tmpl := &domain.Template{
		ID:          s.deps.IDs.NewID(),
		AccountID:   in.AccountID,
		Name:        in.Name,
		Description: in.Description,
		Design:      in.Design,
		IsDefault:   false, // explicit — never auto-default
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := tmpl.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Templates.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}