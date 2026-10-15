// internal/modules/certificate/service/list_templates.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// ListTemplates implements Service.ListTemplates.
//
// Returns the account's own templates plus every system template.
// System templates are appended so the UI can always show the
// Nuruvent defaults as fallback options.
func (s *serviceImpl) ListTemplates(
	ctx context.Context,
	userID, accountID string,
) ([]*domain.Template, error) {
	ok, err := s.deps.Authz.CanReadAll(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrForbidden
	}

	accountTemplates, err := s.deps.Templates.FindByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	systemTemplates := make([]*domain.Template, 0, len(domain.AllSystemTemplates))
	for _, slug := range domain.AllSystemTemplates {
		t, err := s.deps.Templates.FindSystemBySlug(ctx, slug)
		if err != nil {
			// A missing system template is not fatal — skip and
			// keep the list usable. Admin-level seeding should
			// ensure all three exist.
			continue
		}
		systemTemplates = append(systemTemplates, t)
	}

	// Account templates first, then system templates.
	out := make([]*domain.Template, 0, len(accountTemplates)+len(systemTemplates))
	out = append(out, accountTemplates...)
	out = append(out, systemTemplates...)
	return out, nil
}