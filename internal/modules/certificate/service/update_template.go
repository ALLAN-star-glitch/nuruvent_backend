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
//
// Slug rules:
//
//   - System templates: SystemSlug is optional. If provided it must be
//     one of the three known tiers; changing it also changes the price
//     (derived from the slug). If omitted, the current slug is kept.
//     Callers cannot set an arbitrary PriceMinor on a system template.
//
//   - Account templates: SystemSlug MUST be nil. The tier comes from
//     TierTemplateID. PriceMinor is taken from the request if provided.
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
	// System templates: use the (current or new) slug as the tier.
	// Account templates: resolve the parent tier from TierTemplateID.
	var tier domain.SystemTemplateSlug
	if tmpl.IsSystem {
		tier = tmpl.SystemSlug
	} else {
		tier, err = s.resolveTierSlug(ctx, in.TierTemplateID)
		if err != nil {
			return nil, err
		}
	}

	// ---- 3. Apply common updates ----
	if in.Name != nil {
		tmpl.Name = *in.Name
	}
	if in.Description != nil {
		tmpl.Description = *in.Description
	}
	if in.Design != nil {
		tmpl.Design = domain.SanitizeDesign(tier, *in.Design)
	}

	// ---- 4. Apply slug + price (system vs. account divergence) ----
	if tmpl.IsSystem {
		// Account templates cannot sneak a slug through this path.
		if in.SystemSlug != nil {
			if !in.SystemSlug.IsValid() {
				return nil, domain.ErrInvalidSystemTemplateSlug
			}
			tmpl.SystemSlug = *in.SystemSlug
			tier = tmpl.SystemSlug
			// Re-sanitize the design against the new tier — an admin
			// changing premium → default should lose premium fields.
			if in.Design == nil {
				tmpl.Design = domain.SanitizeDesign(tier, tmpl.Design)
			}
		}
		// Price always follows the slug for system templates.
		p := domain.PriceForSlug(tmpl.SystemSlug)
		tmpl.PriceMinor = &p
	} else {
		if in.SystemSlug != nil {
			return nil, domain.ErrInvalidSystemTemplateSlug
		}
		if in.PriceMinor != nil {
			tmpl.PriceMinor = in.PriceMinor
		}
	}

	// ---- 4b. Normalize background settings ----
	// Runs on every update — including updates that only change the name
	// or description — so legacy rows with unset/malformed background
	// fields get corrected as users re-save their templates.
	tmpl.Design.NormalizeBackground()

	tmpl.UpdatedAt = s.deps.Clock.Now().UTC()

	// ---- 5. Validate and persist ----
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}
	if err := s.deps.Templates.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}