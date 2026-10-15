// internal/modules/certificate/service/update_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// UpdateTemplate implements Service.UpdateTemplate.
func (s *serviceImpl) UpdateTemplate(
	ctx context.Context,
	userID, templateID string,
	in UpdateTemplateInput,
) (*domain.Template, error) {
	tmpl, err := s.deps.Templates.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// System templates are not editable.
	if tmpl.IsSystem {
		return nil, domain.ErrCannotEditSystemTemplate
	}

	ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, tmpl.AccountID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	// Apply partial updates — nil fields mean "no change".
	if in.Name != nil {
		tmpl.Name = *in.Name
	}
	if in.Description != nil {
		tmpl.Description = *in.Description
	}
	if in.Design != nil {
		tmpl.Design = *in.Design
	}

	tmpl.UpdatedAt = s.deps.Clock.Now().UTC()

	if err := tmpl.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Templates.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}