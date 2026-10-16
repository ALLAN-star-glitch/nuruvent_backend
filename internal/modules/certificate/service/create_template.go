// internal/modules/certificate/service/create_template.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// CreateTemplate implements Service.CreateTemplate.
//
// The chosen tier determines which design fields are permitted. Fields
// outside the tier's allowlist are stripped before persistence — the
// UI hides them, but the enforcement happens here so no API call can
// bypass it.
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

	// Resolve the tier from the chosen tier template (if any).
	tier, err := s.resolveTierSlug(ctx, in.TierTemplateID)
	if err != nil {
		return nil, err
	}

	// Enforce the tier's design restrictions.
	design := domain.SanitizeDesign(tier, in.Design)

	now := s.deps.Clock.Now().UTC()
	tmpl := &domain.Template{
		ID:          s.deps.IDs.NewID(),
		AccountID:   in.AccountID,
		Name:        in.Name,
		Description: in.Description,
		Design:      design,
		PriceMinor:  in.PriceMinor,
		IsDefault:   false,
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

// resolveTierSlug looks up the system template by ID and returns its
// SystemSlug. Empty ID → empty slug → no restrictions (account-owned
// template without a tier).
func (s *serviceImpl) resolveTierSlug(
	ctx context.Context,
	tierTemplateID string,
) (domain.SystemTemplateSlug, error) {
	if tierTemplateID == "" {
		return "", nil
	}
	t, err := s.deps.Templates.FindByID(ctx, tierTemplateID)
	if err != nil {
		return "", err
	}
	if !t.IsSystem {
		// Not a system template — no restrictions.
		return "", nil
	}
	return t.SystemSlug, nil
}