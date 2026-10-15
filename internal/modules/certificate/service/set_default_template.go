// internal/modules/certificate/service/set_default_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// SetDefaultTemplate implements Service.SetDefaultTemplate.
//
// Clears any existing default for the account, then marks the target
// template as default. Both operations happen in sequence; the DB
// enforces "one default per account" via a partial unique index, so
// a racing caller would fail rather than produce two defaults.
func (s *serviceImpl) SetDefaultTemplate(
	ctx context.Context,
	userID, templateID string,
) error {
	tmpl, err := s.deps.Templates.FindByID(ctx, templateID)
	if err != nil {
		return err
	}

	if tmpl.IsSystem {
		return domain.ErrCannotEditSystemTemplate
	}

	ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, tmpl.AccountID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}

	if err := s.deps.Templates.ClearDefault(ctx, tmpl.AccountID); err != nil {
		return err
	}

	tmpl.IsDefault = true
	tmpl.UpdatedAt = s.deps.Clock.Now().UTC()
	return s.deps.Templates.Update(ctx, tmpl)
}