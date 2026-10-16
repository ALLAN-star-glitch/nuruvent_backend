// internal/modules/certificate/service/update_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// UpdateTemplate implements Service.UpdateTemplate.
//
// Tier restrictions are re-applied on every update, so switching to a
// lower tier strips fields that tier doesn't allow.
func (s *serviceImpl) UpdateTemplate(
	ctx context.Context,
	userID string,
	templateID string,
	in UpdateTemplateInput,
) (*domain.Template, error) {
	tmpl, err := s.deps.Templates.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, tmpl.AccountID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	if tmpl.IsSystem {
		return nil, domain.ErrForbidden
	}

	// Resolve the tier (may have changed).
	tier, err := s.resolveTierSlug(ctx, in.TierTemplateID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		tmpl.Name = *in.Name
	}
	if in.Description != nil {
		tmpl.Description = *in.Description
	}
	if in.Design != nil {
		tmpl.Design = domain.SanitizeDesign(tier, *in.Design)
	}
	if in.PriceMinor != nil {
		tmpl.PriceMinor = in.PriceMinor
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