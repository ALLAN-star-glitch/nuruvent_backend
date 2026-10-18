// internal/modules/certificate/service/create_template.go

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// CreateTemplate implements Service.CreateTemplate.
//
// Two modes:
//
//   - System templates (in.IsSystem == true): only platform admins can
//     create them. They have no AccountID and are globally visible to
//     every account. The tier is taken from in.SystemSlug, which MUST
//     be one of the three known slugs. The price is derived from the
//     slug via domain.PriceForSlug — callers cannot set an arbitrary
//     price, which guarantees system templates always map back to a
//     known tier.
//
//   - Account templates (in.IsSystem == false): created by an account
//     admin. AccountID is required; the caller must hold
//     certificate:create/update/delete in that account. The tier comes
//     from TierTemplateID (a system template); in.SystemSlug MUST be
//     empty.
//
// In both modes the chosen tier determines which design fields are
// permitted. Fields outside the tier's allowlist are stripped before
// persistence — the UI hides them, but the enforcement happens here so
// no API call can bypass it.
func (s *serviceImpl) CreateTemplate(
	ctx context.Context,
	userID string,
	in CreateTemplateInput,
) (*domain.Template, error) {
	// ---- 1. Authorization + tier resolution ----
	var (
		accountID  *string
		tier       domain.SystemTemplateSlug
		priceMinor *int64
	)

		fmt.Printf("[DEBUG] CreateTemplate Input: IsSystem=%v, SystemSlug=%q, AccountID=%q\n", in.IsSystem, in.SystemSlug, in.AccountID)

	if in.IsSystem {
		// System templates: platform-admin check, no account required.
		ok, err := s.deps.Authz.CanManageSystemTemplates(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}

		// Slug is required and must be a known tier.
		if in.SystemSlug == "" {
			return nil, domain.ErrSystemTemplateSlugRequired
		}
		if !in.SystemSlug.IsValid() {
			return nil, domain.ErrInvalidSystemTemplateSlug
		}

		tier = in.SystemSlug

		// Price is derived from the slug, NOT from the request. This is
		// the invariant that makes tiers canonical.
		p := domain.PriceForSlug(in.SystemSlug)
		priceMinor = &p
	} else {
		// Account templates: standard account-level check.
		if in.AccountID == "" {
			return nil, errors.New("account_id is required for account templates")
		}
		if in.SystemSlug != "" {
			return nil, domain.ErrInvalidSystemTemplateSlug
		}

		accountID = &in.AccountID

		ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, in.AccountID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}

		// Resolve the parent tier's slug for design restrictions.
		// err is already declared above by the authz call.
		tier, err = s.resolveTierSlug(ctx, in.TierTemplateID)
		if err != nil {
			return nil, err
		}

		priceMinor = in.PriceMinor
	}

	// ---- 2. Enforce the tier's design restrictions ----
	design := domain.SanitizeDesign(tier, in.Design)

	// ---- 2b. Normalize background settings (opacity / size / position) ----
	// Ensures opacity is clamped to [1..100], size/position are valid
	// enums, and non-http background URLs are dropped — regardless of
	// what the caller sent.
	design.NormalizeBackground()

	// ---- 3. Build the template ----
// SystemSlug is only set for system templates. Account templates use
// `tier` for design restriction purposes but must NOT carry the slug
// on the Template row — Validate() rejects that.
var systemSlugForTemplate domain.SystemTemplateSlug
if in.IsSystem {
    systemSlugForTemplate = tier
}

now := s.deps.Clock.Now().UTC()
tmpl := &domain.Template{
    ID:          s.deps.IDs.NewID(),
    AccountID:   accountID,
    Name:        in.Name,
    Description: in.Description,
    Design:      design,
    PriceMinor:  priceMinor,
    SystemSlug:  systemSlugForTemplate,   // ← FIX
    IsDefault:   false,
    IsSystem:    in.IsSystem,
    CreatedAt:   now,
    UpdatedAt:   now,
}

	// ---- 4. Validate ----
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}

	// ---- 5. Persist ----
	if err := s.deps.Templates.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// resolveTierSlug looks up the system template by ID and returns its
// SystemSlug. Empty ID → empty slug → no restrictions.
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
		return "", nil
	}
	return t.SystemSlug, nil
}