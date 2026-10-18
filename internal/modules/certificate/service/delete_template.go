// internal/modules/certificate/service/delete_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// DeleteTemplate implements Service.DeleteTemplate.
//
// System templates are protected from deletion: they are tier parents,
// and any account template that references one would lose its tier
// resolution (and therefore its design restrictions) if the parent
// disappeared. Admins can still edit system templates (name, design,
// price are derived from slug) but cannot delete them.
func (s *serviceImpl) DeleteTemplate(
	ctx context.Context,
	userID, templateID string,
) error {
	tmpl, err := s.deps.Templates.FindByID(ctx, templateID)
	if err != nil {
		return err
	}

	// ---- 1. Authorization ----
	if tmpl.IsSystem {
		// System templates can only be touched by platform admins —
		// and even then, only edited, never deleted.
		ok, err := s.deps.Authz.CanManageSystemTemplates(ctx, userID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}

		// Reject deletion regardless of admin status. System templates
		// are tier parents; removing one would orphan every account
		// template that references it.
		return domain.ErrCannotDeleteSystemTemplate
	}

	// ---- 2. Account templates: standard account-level check ----
	if tmpl.AccountID == nil {
		return domain.ErrForbidden
	}
	ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, *tmpl.AccountID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}

	return s.deps.Templates.SoftDelete(ctx, templateID)
}