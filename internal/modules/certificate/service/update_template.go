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

	// ---- 1. Authorization ----
	if tmpl.IsSystem {
		// System templates can only be edited by platform admins.
		ok, err := s.deps.Authz.CanManageSystemTemplates(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	} else {
		// Account templates require account-level permission.
		if tmpl.AccountID == nil {
			return nil, domain.ErrForbidden
		}
		ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, *tmpl.AccountID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	}

	// ---- 2. Resolve the tier ----
	// System templates don't have a parent tier (they are the tiers).
	// Account templates resolve their chosen tier to enforce restrictions.
	var tier domain.SystemTemplateSlug
	if !tmpl.IsSystem {
		tier, err = s.resolveTierSlug(ctx, in.TierTemplateID)
		if err != nil {
			return nil, err
		}
	}

	// ---- 3. Apply updates ----
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

	// ---- 4. Validate and persist ----
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Templates.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}