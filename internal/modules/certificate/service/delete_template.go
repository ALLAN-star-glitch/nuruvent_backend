// internal/modules/certificate/service/delete_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// DeleteTemplate implements Service.DeleteTemplate.
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
		// System templates can only be deleted by platform admins.
		ok, err := s.deps.Authz.CanManageSystemTemplates(ctx, userID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
	} else {
		// Account templates require account-level permission.
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
	}

	return s.deps.Templates.SoftDelete(ctx, templateID)
}